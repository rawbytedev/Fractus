package dbflat

import "fmt"

// Expose DBFLAT dynamic encoding/decoding functions
type api struct {
	enc *Encoder
	dec *Decoder
	buf *buffer
}

// reuse buffer and avoid allocs
type buffer struct {
	fields   []FieldValue
	schemaID uint64
	hotTags  []uint16
	outBuf   []byte
}

func initAPI() *api {
	return &api{enc: &Encoder{}, dec: &Decoder{}, buf: &buffer{}}
}

func (a *api) CreateInstance() {
	a.enc = &Encoder{}
	a.dec = &Decoder{}
	a.buf = &buffer{}
}

func (a *api) AddField(tag uint16, compFlags uint16, field []byte) {
	a.buf.fields = append(a.buf.fields, FieldValue{Tag: tag, CompFlags: compFlags, Payload: field})
}

func (a *api) AddSchema(id uint64) {
	a.buf.schemaID = id
}

func (a *api) Encode() {
	var err error
	a.buf.outBuf, err = a.enc.EncodeRecord(a.buf.schemaID, a.buf.hotTags, a.buf.fields)
	if err != nil {
		fmt.Printf("Error %s", err)
	}
}

// Decode the content
func (a *api) Decode() {
	var err error
	id := ReadSchema(a.buf.outBuf)
	r, err := a.dec.DecodeRecord(a.buf.outBuf, nil)
	if err != nil {
		fmt.Printf("error: %s", err)
	}
	reflet(id, r)
}

// reflet map into usable struct
func reflet(schemaid uint64, fmap map[uint16][]byte) {

}

// Tale in the schema
func ParseSchema() {

}

func (a *api) AddBytes(encoded []byte) {
	a.buf.outBuf = append(a.buf.outBuf, encoded...)
}

// reset Buffer
func (a *api) Reset() {
	a.buf.fields = a.buf.fields[:0]
	a.buf.hotTags = a.buf.hotTags[:0]
	a.buf.outBuf = a.buf.outBuf[:0]
	a.buf.schemaID = uint64(0)
}
