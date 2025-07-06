package dbflat

import "bytes"

// still building not error free
type Builder struct {
	enc       *Encoder       // zero-alloc, reusable
	dec       *Decoder       // for validation or round-trip
	buf       *bytes.Buffer  // output buffer
	fields    []pendingField // scratch state
	offsetTbl []fieldIndex   // optional field-to-offset
	out       []byte         // encoded data
}

type pendingField struct {
	Tag       uint16
	Hot       bool
	CompFlags uint16
	Value     []byte // encoded value
}
type fieldIndex struct {
	Tag     uint16
	offsets []uint32
}

func NewBuilder(buff *bytes.Buffer) *Builder {
	if buff != nil {
		return &Builder{buf: buff}
	}
	return &Builder{}
}

func (b *Builder) AddField(tag uint16, compFlags uint16, field []byte, hot bool) {
	b.fields = append(b.fields, pendingField{Tag: tag, Hot: hot, CompFlags: compFlags, Value: field})
}

func (b *Builder) AddFieldOffset(tag uint16, compFlags uint16, field []byte, hot bool, offset []uint32) {
	if hot == false {
		hot = false
	}
	b.fields = append(b.fields, pendingField{Tag: tag, Hot: hot, CompFlags: compFlags, Value: field})
	b.offsetTbl = append(b.offsetTbl, fieldIndex{Tag: tag, offsets: offset})
}

func (b *Builder) Commit(schemaID uint64, flags ...uint32) error {
	var hotTags []uint16
	var field []FieldValue
	for _, a := range b.fields {
		field = append(field, FieldValue{Tag: a.Tag, CompFlags: a.CompFlags, Payload: a.Value})
		if a.Hot {
			if !(a.Tag > 8 || a.Tag == 0) {
				hotTags = append(hotTags, a.Tag)
			}
		}
	}
	var err error
	b.out, err = b.enc.EncodeRecord(schemaID, hotTags, field)
	return err
}
