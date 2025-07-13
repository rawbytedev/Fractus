package dbflat

type Inspector struct {
	buf     []byte            // raw dbflat message(encoded)
	offsets map[uint16]uint32 // tag → offset if offset table present
	fields  map[uint16][]byte // tag → raw field data
	hotset  []uint16
	index   []FieldValue // keep track of fields
	idx     *int
	outbuf  []byte   // decoded dbflat data hold single data at a time
	dec     *Decoder // Reuse buffers
}

// Parse parses raw dbflat buffer and returns an Inspector
func Inspect(buf []byte, dec *Decoder) (*Inspector, error) {
	return &Inspector{buf: buf, dec: dec}, nil
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

// Perform lazy scan and fill some fields
func (i *Inspector) Next() bool {
	
	if i.idx != nil {

	}
	return true
}

// Must be run before Next()
func (i *Inspector) Scan() {

}
