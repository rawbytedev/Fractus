package dbflat

import (
	"bytes"
	"encoding/binary"
	"sort"
)

// buildHotBitmap for tags 1–8
// HotFields
func buildHotBitmap(tags []uint16) byte {
	var bm byte
	for _, t := range tags {
		if t >= 1 && t <= 8 {
			bm |= 1 << (t - 1)
		}
	}
	return bm
}

// EncodeRecord packs fields into DBFlat bytes
func EncodeRecord(schemaID uint64, hotTags []uint16, fields []FieldValue) ([]byte, error) {
	// 1) Sort fields by Tag for canonical bytes
	sort.Slice(fields, func(i, j int) bool { return fields[i].Tag < fields[j].Tag })

	// 2) Build Data Section & record offsets
	dataBuf := new(bytes.Buffer)
	offsets := make([]uint32, len(fields))

	for i, f := range fields {
		// pad to 8-byte alignement if align8 is set
		pad := align(dataBuf.Len(), 8) - dataBuf.Len()
		if pad > 0 {
			dataBuf.Write((make([]byte, pad)))
		}
		offsets[i] = uint32(dataBuf.Len())

		// variable-length
		if f.CompFlags&ArrayMask != 0 {
			if f.CompFlags&^ArrayMask != CompRaw {
				// compress if needed
				comp, err := compressData(f.CompFlags, f.Payload)
				if err != nil {
					return nil, err
				}
				// prefix compressed size // compress then append size taken by compression
				writeVarUint(dataBuf, uint64(len(comp)))
				dataBuf.Write(comp)
			} else {
				writeVarUint(dataBuf, uint64(len(f.Payload)))
				dataBuf.Write(f.Payload)
			}

		} else {
			// fixed-width raw
			dataBuf.Write(f.Payload)
		}
	}

	// 3) Build VTable
	vtSize := len(fields) * 8 // each Vtable entry is 8B
	vtBuf := make([]byte, vtSize)
	idx := 0
	for i, f := range fields {
		binary.LittleEndian.PutUint16(vtBuf[idx:], f.Tag)
		binary.LittleEndian.PutUint16(vtBuf[idx+2:], f.CompFlags)
		binary.LittleEndian.PutUint32(vtBuf[idx+4:], offsets[i])
		idx += 8
	}

	// 4) Build Header
	header := Header{
		Magic:       MagicV1,
		Version:     VersionV1,
		Flags:       0x0001 | 0x0002, // align8 + schemaID
		SchemaID:    schemaID,
		HotBitmap:   buildHotBitmap(hotTags),
		VTableSlots: byte(len(fields)),
		DataOffset:  uint16(HeaderSize + len(vtBuf)),
		VTableOff:   uint32(HeaderSize),
	}

	// 5) Assemble final bytes
	out := new(bytes.Buffer)
	binary.Write(out, binary.LittleEndian, header)
	out.Write(vtBuf)
	out.Write(dataBuf.Bytes())

	return out.Bytes(), nil
}

// Testing
type Encoder struct {
	headerBuf []byte
	vtBuf     []byte
	dataBuf   []byte
	fieldBuf  []byte   // for per-field varints
	offsets   []uint32 //reused
}

func (e *Encoder) EncodeRecord(schemaID uint64, hotTags []uint16, fields []FieldValue) ([]byte, error) {
	sort.Slice(fields, func(i, j int) bool { return fields[i].Tag < fields[j].Tag })

	// Reset buffers
	e.headerBuf = e.headerBuf[:0]
	e.vtBuf = e.vtBuf[:0]
	e.dataBuf = e.dataBuf[:0]
	e.fieldBuf = e.fieldBuf[:0]

	// Ensure offsets slice fits
	if cap(e.offsets) < len(fields) {
		e.offsets = make([]uint32, len(fields))
	} else {
		e.offsets = e.offsets[:len(fields)]
	}

	// --- Encode field payloads ---
	for i, f := range fields {
		// Align to 8 bytes if flag set
		pad := align(len(e.dataBuf), 8) - len(e.dataBuf)
		e.dataBuf = append(e.dataBuf, zeroPadding[:pad]...)

		e.offsets[i] = uint32(len(e.dataBuf))

		// Compress or array logic
		switch {
		case f.CompFlags&CompressionMask != 0:
			compressed, err := compressData(f.CompFlags, f.Payload)
			if err != nil {
				return nil, err
			}
			e.fieldBuf = e.fieldBuf[:0]
			e.dataBuf = append(e.dataBuf, e.writeVarUint(uint64(len(compressed)))...)
			e.dataBuf = append(e.dataBuf, compressed...)

		case f.CompFlags&ArrayMask != 0:
			e.fieldBuf = e.fieldBuf[:0]
			e.dataBuf = append(e.dataBuf, e.writeVarUint(uint64(len(f.Payload)))...)
			e.dataBuf = append(e.dataBuf, f.Payload...)

		default:
			e.dataBuf = append(e.dataBuf, f.Payload...)
		}
	}

	// --- Encode vtable ---
	vtSize := len(fields) * 8
	if cap(e.vtBuf) < vtSize {
		e.vtBuf = make([]byte, vtSize)
	}
	e.vtBuf = e.vtBuf[:vtSize]
	for i, f := range fields {
		idx := i * 8
		binary.LittleEndian.PutUint16(e.vtBuf[idx:], f.Tag)
		binary.LittleEndian.PutUint16(e.vtBuf[idx+2:], f.CompFlags)
		binary.LittleEndian.PutUint32(e.vtBuf[idx+4:], e.offsets[i])
	}

	// --- Encode header ---
	e.headerBuf = encodeHeader(e.headerBuf[:0], Header{
		Magic:       MagicV1,
		Version:     VersionV1,
		Flags:       0x0001 | 0x0002,
		SchemaID:    schemaID,
		HotBitmap:   buildHotBitmap(hotTags),
		VTableSlots: byte(len(fields)),
		DataOffset:  uint16(HeaderSize + len(e.vtBuf)),
		VTableOff:   uint32(HeaderSize),
	})

	// --- Final payload ---
	total := len(e.headerBuf) + len(e.vtBuf) + len(e.dataBuf)
	out := make([]byte, 0, total)
	out = append(out, e.headerBuf...)
	out = append(out, e.vtBuf...)
	out = append(out, e.dataBuf...)

	return out, nil
}

func encodeHeader(buf []byte, h Header) []byte {
	buf = append(buf, make([]byte, HeaderSize)...)
	binary.LittleEndian.PutUint32(buf[0:], h.Magic)
	binary.LittleEndian.PutUint16(buf[4:], h.Version)
	binary.LittleEndian.PutUint16(buf[6:], h.Flags)
	binary.LittleEndian.PutUint64(buf[8:], h.SchemaID)
	buf[16] = h.HotBitmap
	buf[17] = h.VTableSlots
	binary.LittleEndian.PutUint16(buf[18:], h.DataOffset)
	binary.LittleEndian.PutUint32(buf[20:], h.VTableOff)
	return buf
}

var zeroPadding = [8]byte{}

func (e *Encoder) writeVarUint(x uint64) []byte {
	e.fieldBuf = e.fieldBuf[:0]
	for x >= 0x80 {
		e.fieldBuf = append(e.fieldBuf, byte(x)|0x80)
		x >>= 7
	}
	e.fieldBuf = append(e.fieldBuf, byte(x))
	return e.fieldBuf
}