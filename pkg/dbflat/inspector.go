package dbflat

import (
	"encoding/binary"
	"maps"
)

type Inspector struct {
	buf       []byte            // raw dbflat message(encoded)
	offsets   map[uint16]int    // tag → offset
	fields    map[uint16][]byte // tag → raw field data
	idx       int               // track number of tags
	nexoffset int               // store next offset
	curoffset int               // store current offset
	outbuf    []byte            // decoded dbflat data hold single data at a time
	dec       *Decoder          // Reuse buffers
}

// Parse parses raw dbflat buffer and returns an Inspector
func Inspect(buf []byte, dec *Decoder) (*Inspector, error) {
	return &Inspector{buf: buf, dec: dec}, nil
}
func NewInspect(buf []byte) *Inspector {
	return &Inspector{buf: buf, dec: NewDecoder()}
}

// Insert new encoded bytes then reset all maps
func (i *Inspector) Insert(buf []byte) {
	i.buf = i.buf[:0]
	i.buf = buf
}

// Reset all fields including encoded/decoded bytes
func (i *Inspector) Reset() {
	i.buf = i.buf[:0]
	i.outbuf = i.outbuf[:0]
	i.offsets = nil
}

// GetField returns raw bytes for tag
// Debug Version
func (i *Inspector) GetFieldD(tag uint16) ([]byte, error) {
	var err error
	i.outbuf, err = i.dec.FindWithTag(i.buf, tag)
	if err != nil {
		return nil, err
	}
	return i.outbuf, nil
}

// GetField returns raw bytes for tag
func (i *Inspector) GetField(tag uint16) []byte {
	var err error
	i.outbuf, err = i.dec.FindWithTag(i.buf, tag)
	if err != nil {
		return nil
	}
	return i.outbuf
}

// --- Tagwalk -----
// Perform lazy scan and fill some fields
func (i *Inspector) Next() bool {
	if i.idx != 0 {
		i.curoffset = i.nexoffset
		tmpmap, offset, err := i.dec.DecodeRecordTagWalk(i.buf, i.nexoffset, nil)
		if err != nil {
			// reset cursor
			i.idx = 0
			return false
		}
		maps.Copy(i.fields, tmpmap)
		i.nexoffset = offset
		i.idx = i.idx + 1
		return true
	}
	//i.idx = 0
	i.fields = map[uint16][]byte{}
	i.offsets = map[uint16]int{}
	tmpmap, noffset, err := i.dec.DecodeRecordTagWalk(i.buf, i.curoffset, nil)
	if err != nil {
		return false
	}
	for k, v := range tmpmap {
		i.fields[k] = v
		if i.idx != 0 {
			i.offsets[k] = i.idx
		}
		i.offsets[k] = noffset
	}
	i.nexoffset = noffset
	i.idx = i.idx + 1
	return true
}

// Returns the tag of the field at cursor position
func (i *Inspector) Peek() uint16 {
	return i.PeekWithOffset(i.curoffset) // base
}
func (i *Inspector) FlagsPeek() uint16 {
	return i.PeekWithOffset(i.curoffset + 2) // +2
}
func (i *Inspector) binary() []byte {
	return i.fields[i.Peek()]
}

// reconstruct FieldValue
func (i *Inspector) Field() FieldValue {
	return FieldValue{Tag: i.Peek(), CompFlags: i.FlagsPeek(), Payload: i.binary()}
}

// This is a helper function
// takes the offset and return the tag/compflags/varint of the fields
func (i *Inspector) PeekWithOffset(off int) uint16 {
	return binary.LittleEndian.Uint16(i.buf[off:])
}

// Resets all variable / Used for new buffer
func (i *Inspector) Scan() {
	for k := range i.fields {
		delete(i.fields, k)
	}
	for k := range i.offsets {
		delete(i.offsets, k)
	}
	i.idx = 0
	i.curoffset = 0
	i.nexoffset = 0

}
