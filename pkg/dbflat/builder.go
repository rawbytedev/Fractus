package dbflat

type Builder struct {
	enc       *Encoder       // zero-alloc, reusable
	dec       *Decoder       // for validation or round-trip
	buf       []byte         // output buffer
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

func NewBuilder(buff []byte) *Builder {
	if buff != nil {
		return &Builder{buf: buff, enc: NewEncoder(), dec: NewDecoder()}
	}
	return &Builder{buf: make([]byte, 0), enc: NewEncoder(), dec: NewDecoder()}
}

func NewEncoder() *Encoder {
	return &Encoder{}
}
func NewDecoder() *Decoder {
	return &Decoder{}
}

func (b *Builder) AddField(tag uint16, compFlags uint16, field []byte, hot bool) {
	b.fields = append(b.fields, pendingField{Tag: tag, Hot: hot, CompFlags: compFlags, Value: field})
}

func (b *Builder) AddFieldOffset(tag uint16, compFlags uint16, field []byte, hot bool, offset []uint32) {
	if !hot {
		hot = false
	}
	b.fields = append(b.fields, pendingField{Tag: tag, Hot: hot, CompFlags: compFlags, Value: field})
	b.offsetTbl = append(b.offsetTbl, fieldIndex{Tag: tag, offsets: offset})
}

func (b *Builder) Commit(schemaID uint64, flags uint16) error {
	var hotTags []uint16
	b.enc.headerflag = flags
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
	out, err := b.enc.EncodeRecordFull(schemaID, hotTags, field)
	b.out = append(b.out, out...)
	out = out[:0]
	if err != nil {
		return err
	}
	return err
}

func (b *Builder) Output() []byte {
	return b.out
}
func (b *Builder) Reset() {
	b.fields = b.fields[:0]
	b.out = b.out[:0]
	b.buf = b.buf[:0]
}

/*
func (b *Builder) Validate() (bool, error){

}*/
