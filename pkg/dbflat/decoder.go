package dbflat

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type Decoder struct {
	blob      []byte
	fieldsOut map[uint16][]byte
	raw       []byte
}

// Decode Hot Field with partial decoding and return a []byte
// This operation gives out O(1) speedfunc ReadHotField(buf []byte, tag uint16, width int) ([]byte, error) {
func (d *Decoder) ReadHotField(buf []byte, tag uint16, width int) ([]byte, error) {
	if tag == 0 || tag > 8 {
		return nil, fmt.Errorf("invalid hot field tag: %d", tag)
	}

	d.blob = d.blob[:0]
	// Parse header
	h, err := ParseHeader(buf)
	if err != nil {
		return nil, err
	}
	if (h.HotBitmap>>(tag-1))&1 == 0 {
		return nil, fmt.Errorf("tag %d is not a hot field", tag)
	}
	// Compute slot and data offsets
	slotIdx := int(tag - 1)
	slotOffset := int(h.VTableOff) + slotIdx*SlotSize
	if slotOffset+SlotSize > len(buf) {
		return nil, errors.New("vtable slot out of bounds")
	}

	// Pull out compFlags and offset
	compFlags := binary.LittleEndian.Uint16(buf[slotOffset+2:])
	offset := binary.LittleEndian.Uint32(buf[slotOffset+4:])
	ptr := int(h.DataOffset) + int(offset)

	// Align if needed
	if h.Flags&0x0001 != 0 {
		ptr = align(ptr, 8)
	}
	if ptr >= len(buf) {
		return nil, errors.New("data pointer out of bounds")
	}

	// Handle compression/array flags
	if compFlags&ArrayMask != 0 || compFlags&^ArrayMask != CompRaw {
		size, n := readVarUint(buf[ptr:])
		ptr += n
		if ptr+int(size) > len(buf) {
			return nil, errors.New("compressed blob out of bounds")
		}
		d.blob = d.blob[:0]
		d.blob = buf[ptr : ptr+int(size)]
		if compFlags&^ArrayMask != CompRaw {
			return decompressData(compFlags, d.blob, int(size))
		}
		return d.blob, nil
	}

	// Fixed-width hot field
	if ptr+width > len(buf) {
		return nil, errors.New("fixed-width field out of bounds")
	}
	return buf[ptr : ptr+width], nil
}

func ReadSchema(buf []byte) uint64{
	return binary.LittleEndian.Uint64(buf[8:])
}
//Parse Header from buf; zero copy
func ParseHeader(buf []byte) (Header, error) {
	if len(buf) < HeaderSize {
		return Header{}, errors.New("buffer too short for header")
	}
	h := Header{}
	h.Magic = binary.LittleEndian.Uint32(buf[0:])
	h.Version = binary.BigEndian.Uint16(buf[4:])
	h.Flags = binary.LittleEndian.Uint16(buf[6:])
	h.SchemaID = binary.LittleEndian.Uint64(buf[8:])
	h.HotBitmap = buf[16]
	h.VTableSlots = buf[17]
	h.DataOffset = binary.LittleEndian.Uint16(buf[18:])
	h.VTableOff = binary.LittleEndian.Uint32(buf[20:])
	return h, nil
}
func (d *Decoder) DecodeRecord(buf []byte, fmaps map[uint16]int) (map[uint16][]byte, error) {
	// 1) Header
	h, err := ParseHeader(buf)
	if err != nil {
		return nil, err
	}
	if h.Magic != MagicV1 {
		return nil, errors.New("invalid magic")
	}
	// 2) Carve out VTable
	vtStart := int(h.VTableOff)
	slotCnt := int(h.VTableSlots)
	// 3) parse each slot
	dataStart := int(h.DataOffset)

	if d.fieldsOut == nil {
		d.fieldsOut = make(map[uint16][]byte)
	} else {
		// clear map
		for k := range d.fieldsOut {
			delete(d.fieldsOut, k)
		}
	}
	for i := range slotCnt {
		base := vtStart + i*SlotSize
		tag := binary.LittleEndian.Uint16(buf[base:])
		cf := binary.LittleEndian.Uint16(buf[base+2:])
		off := binary.LittleEndian.Uint32(buf[base+4:])
		ptr := dataStart + int(off)
		// handle alignement (if set)
		if h.Flags&0x0001 != 0 {
			ptr = align(ptr, 8)
		}
		// 4) Decode payload
		if cf&ArrayMask != 0 || cf&^ArrayMask != CompRaw {
			// read uncompressedSize varint
			size, n := readVarUint(buf[ptr:])
			ptr += n
			// compressed blob ends at next slot/data boundary or len(buf)
			// assume till end for simplicity
			d.raw = buf[ptr : ptr+int(size)]
			if cf&CompressionMask != CompRaw {
				decoded, err := decompressData(cf, d.raw, int(size))
				if err != nil {
					return nil, err
				}
				d.raw = decoded
			}
		} else {
			// fixed-width: get width from compFlags/type map
			/*fixedWidthMap := GetFixedWidthMap(false, fmaps)
			if fixedWidthMap != nil {
				width := fixedWidthMap[tag]
				d.blob = buf[ptr : ptr+width]
			} else {*/
			width := fixedWidth(tag)
			d.raw = buf[ptr : ptr+width]
		}
		d.fieldsOut[tag] = d.raw

	}
	return d.fieldsOut, nil
}
