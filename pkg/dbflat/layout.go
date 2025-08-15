package dbflat

import (
	"encoding/binary"
	"slices"
	"sort"
)

type EncodeStrategyType int

const (
	FullVTable EncodeStrategyType = iota
	HotVTable
	TagWalk
)

type LayoutPlan struct {
	Fields      []FieldValue       // input fields
	HotTags     []uint16           // tags to prioritize
	Strategy    EncodeStrategyType // enum: FullVTable, HotVTable, TagWalk
	HeaderFlags uint16             // layout control
	SchemaID    uint64             // optional
}
type LayoutEngine struct {
	GenPayload        func(fields []FieldValue) ([]byte, []OffsetMap) //field, tag + comflag + offset
	GenVTable         func([]OffsetMap) []byte
	GenTagWalkPayload func([]FieldValue) []byte   // tag+comflag+field...
	BuildHeader       func(Header, []byte) []byte // customHeader+vtable
	buf               []byte                      // buffer
}

// Default Engine
var DefautEngine = LayoutEngine{
	GenPayload:        GenPayloads,
	GenVTable:         GeneVtables,
	GenTagWalkPayload: GenTagWalk,
}

type LayoutDefaultEngine struct {
	plan       *LayoutPlan
	engine     LayoutEngine
	vtablebuf  []byte // vtablebuffer
	headerbuf  []byte // headerbuffer
	payloadbuf []byte // payloadbuffer
}

func (e *LayoutDefaultEngine) Init() {
	e.engine = LayoutEngine{
		GenPayload:        GenPayloads,
		GenVTable:         GeneVtables,
		GenTagWalkPayload: GenTagWalk,
	}
}

// Start a layout Plan using default settings
func (e *LayoutDefaultEngine) Launch() []byte {
	switch e.plan.Strategy {
	case FullVTable:
		payload, offsets := DefautEngine.GenPayload(e.plan.Fields)
		vtable := DefautEngine.GenVTable(offsets)
		header := BuildHeader(vtable, e.plan)
		return Join(header, vtable, payload, nil)
	case HotVTable:
		hot, cold := PartitionFields(e.plan.Fields, e.plan.HotTags)
		payloadHot, hotOffsets := DefautEngine.GenPayload(hot)
		vtable := DefautEngine.GenVTable(hotOffsets)
		tagwalk := DefautEngine.GenTagWalkPayload(cold)
		header := BuildHeader(vtable, e.plan)
		return Join(header, vtable, payloadHot, tagwalk)

	case TagWalk:
		tagwalk := DefautEngine.GenTagWalkPayload(e.plan.Fields)
		header := BuildHeader(nil, e.plan)
		return Join(header, nil, nil, tagwalk)
	}
	return nil
}

type OffsetMap struct {
	tag      uint16 // tag
	compflag uint16 // compression flags
	offset   uint32 // start offset
}

// GeneratePayloads
func GenPayloads(fields []FieldValue) ([]byte, []OffsetMap) {
	var tmp []byte
	next := 0
	var fieldbuf []byte
	var offmap []OffsetMap
	for _, field := range fields {
		if field.CompFlags&ArrayMask != 0 {
			fieldbuf = writeVarUint(fieldbuf, uint64(len(field.Payload)))
			tmp = append(tmp, fieldbuf...)
			fieldbuf = fieldbuf[:0]
		}
		tmp = append(tmp, field.Payload...)
		offmap = append(offmap, OffsetMap{tag: field.Tag, compflag: field.CompFlags, offset: uint32(next)})
		next = len(tmp)
	}
	return tmp, offmap
}

// Generate Vtables
func GeneVtables(offsets []OffsetMap) []byte {
	vtSize := len(offsets) * 8
	vtBuf := make([]byte, vtSize)
	for i, offmap := range offsets {
		idx := i * 8
		binary.LittleEndian.PutUint16(vtBuf[idx:], offmap.tag)
		binary.LittleEndian.PutUint16(vtBuf[idx+2:], offmap.compflag)
		binary.LittleEndian.PutUint32(vtBuf[idx+4:], offmap.offset)
	}
	return vtBuf
}

// Generate TagWalk
func GenTagWalk(fields []FieldValue) []byte {
	var tmp []byte
	var varint []byte
	for _, field := range fields {
		tmp = append(tmp, ToBytes(field.Tag)...)
		tmp = append(tmp, ToBytes(field.CompFlags)...)
		if field.CompFlags&ArrayMask != 0 {
			tmp = append(tmp, writeVarUint(varint, uint64(len(field.Payload)))...)
			varint = varint[0:]
		}
		tmp = append(tmp, field.Payload...)

	}
	return tmp
}

// used for tests
func LaunchPlan(plan *LayoutPlan) []byte {

	switch plan.Strategy {
	case FullVTable:
		payload, offsets := DefautEngine.GenPayload(plan.Fields)
		vtable := DefautEngine.GenVTable(offsets)
		header := BuildHeader(vtable, plan)
		return Join(header, vtable, payload, nil)

	case HotVTable:
		hot, cold := PartitionFields(plan.Fields, plan.HotTags)
		payloadHot, hotOffsets := DefautEngine.GenPayload(hot)
		vtable := DefautEngine.GenVTable(hotOffsets)
		tagwalk := DefautEngine.GenTagWalkPayload(cold)
		header := BuildHeader(vtable, plan)
		return Join(header, vtable, payloadHot, tagwalk)

	case TagWalk:
		tagwalk := DefautEngine.GenTagWalkPayload(plan.Fields)
		header := BuildHeader(nil, plan)
		return Join(header, nil, nil, tagwalk)
	}
	return nil
}

func Join(header *Header, vtable []byte, payload []byte, tagwalk []byte) []byte {
	var buffer []byte
	if header != nil {
		buffer = encodeHeader(buffer, *header)
	}
	if vtable != nil {
		buffer = append(buffer, vtable...)
	}
	if payload != nil {
		buffer = append(buffer, payload...)
	}
	if tagwalk != nil {
		buffer = append(buffer, tagwalk...)
	}
	return buffer
}

// Works only for FullMode and HotVtable
// Tagwalk doesn't need a header
func enHeader(h *Header) []byte {
	buf := make([]byte, HeaderSize) //Max size for header
	if h.Flags&FlagNoSchemaID != 0 {
		buf = append(buf, make([]byte, HeaderSize-8)...)
	} else {
		buf = append(buf, make([]byte, HeaderSize)...)
	}
	binary.LittleEndian.PutUint32(buf[0:], h.Magic)
	binary.LittleEndian.PutUint16(buf[4:], h.Version)
	binary.LittleEndian.PutUint16(buf[6:], h.Flags)
	if h.Flags&FlagNoSchemaID != 0 {
		buf[8] = h.HotBitmap
		buf[9] = h.VTableSlots
		binary.LittleEndian.PutUint16(buf[10:], h.DataOffset)
		binary.LittleEndian.PutUint32(buf[12:], h.VTableOff)
		return buf
	} else {
		binary.LittleEndian.PutUint64(buf[8:], h.SchemaID)
		buf[16] = h.HotBitmap
		buf[17] = h.VTableSlots
		binary.LittleEndian.PutUint16(buf[18:], h.DataOffset)
		binary.LittleEndian.PutUint32(buf[20:], h.VTableOff)
		return buf
	}
}

// Not fully implemented
func PartitionFields(fields []FieldValue, u []uint16) ([]FieldValue, []FieldValue) {
	var cold []FieldValue
	if len(fields) < len(u) {
		return nil, nil
	}
	hot := slices.DeleteFunc(fields, func(s FieldValue) bool {
		return true
	})

	return hot, cold
}
func Sort(fields []FieldValue) []FieldValue {
	if !isSortedByTag(fields) {
		sort.Slice(fields, func(i, j int) bool { return fields[i].Tag < fields[j].Tag })
	}
	return fields
}

func BuildHeader(vtable []byte, plan *LayoutPlan) *Header {
	if vtable != nil {
		h := &Header{
			Magic:       MagicV1,
			Version:     VersionV1,
			Flags:       plan.HeaderFlags,
			SchemaID:    plan.SchemaID,
			HotBitmap:   buildHotBitmap(plan.HotTags),
			VTableSlots: byte(len(vtable) / 8),
			DataOffset:  uint16(HeaderSize + len(vtable)),
			VTableOff:   uint32(HeaderSize),
		}
		return h
	}
	return nil

}

func BuildHeaderCustom(vtable []byte, u uint16) Header {
	panic("unimplemented")
}
