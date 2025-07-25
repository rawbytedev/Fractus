package dbflat

import (
	"bytes"
	"sort"
	"testing"

	"github.com/klauspost/compress/zstd"
)

func Order(field []FieldValue) []FieldValue {
	if !isSortedByTag(field) {
		sort.Slice(field, func(i, j int) bool { return field[i].Tag < field[j].Tag })
	}
	return field
}
func TestHeaderDecode(t *testing.T) {
	field := makeTestFields("heavy")
	schemaID := uint64(112)
	hotTags := []uint16{
		uint16(1),
		uint16(2),
	}
	e := &Encoder{headerflag: 0x0001}

	enc, err := e.EncodeRecordFull(schemaID, hotTags, field)
	if err != nil {
		t.Fatal(err)
	}
	head, err := ParseHeader(enc)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if head.Magic != MagicV1 {
		t.Fatalf("Expected: %d  got %d", MagicV1, head.Magic)
	}
	if !(head.Flags&FlagPadding != 0) {
		t.Fatalf("Expected: %d  got %d", FlagPadding, head.Flags)
	}
	if head.SchemaID != schemaID {
		t.Fatalf("Expected: %d got %d", schemaID, head.SchemaID)
	}
	if (head.Version >> 8) != VersionV1 {
		t.Fatalf("Expected: %d got %d", VersionV1, head.Version)
	}
}

func TestDecodeHotFieldWithPadding(t *testing.T) {
	field := makeTestFields("skinny")
	schemaID := uint64(112)
	hotTags := []uint16{
		uint16(1),
		uint16(2),
	}
	e := &Encoder{headerflag: 0x0001}
	var d Decoder
	enc, err := e.EncodeRecordFull(schemaID, hotTags, field)
	if err != nil {
		t.Fatal(err)
	}
	for i := range hotTags {
		a, err := d.ReadHotField(enc, uint16(i+1), 0)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(a, field[i].Payload) {
			t.Failed()
		}
	}
	t.Logf("size with padding: %d", len(enc))
}
func TestDecodeHotFieldNoPadding(t *testing.T) {
	field := makeTestFields("skinny")
	schemaID := uint64(112)
	hotTags := []uint16{
		uint16(1),
		uint16(2),
	}
	e := &Encoder{}
	var d Decoder
	enc, err := e.EncodeRecordFull(schemaID, hotTags, field)
	if err != nil {
		t.Fatal(err)
	}
	for i := range hotTags {
		a, err := d.ReadHotField(enc, uint16(i+1), 0)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(a, field[i].Payload) {
			t.Failed()
		}
	}
	t.Logf("size without padding: %d", len(enc))
}

func TestComp(t *testing.T) {
	a, err := zstd.NewWriter(nil)
	s := a.EncodeAll([]byte("TestCompression"), nil)
	if err != nil {
		t.Fatal(err)
	}
	dec, err := zstd.NewReader(nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = dec.DecodeAll(s, nil)
	if err != nil {
		t.Fatal(err)
	}
	//t.Log(string(res))
}

func TestWriter(t *testing.T) {
	a, _ := Write(uint32(1000))
	b, _ := ReadAny(a, TypeUint32)
	if uint32(1000) != b {
		t.Fatalf("Writer error")
	}
}

func makeTestFields(shape string) []FieldValue {
	switch shape {
	case "skinny":
		a, _ := Write(uint32(300))
		return []FieldValue{
			{Tag: uint16(1), Payload: []byte("Hello I'm Test 1"), CompFlags: 0x8000},
			{Tag: uint16(2), Payload: []byte("Hello I'm Test 2"), CompFlags: 0x0000 | 0x8000},
			{Tag: uint16(3), Payload: []byte("Hello I'm Test Comp+10"), CompFlags: 0x0000 | 0x8000},
			{Tag: uint16(192), Payload: a, CompFlags: 0x0000},
		}
	case "heavy":
		return []FieldValue{
			{Tag: uint16(1), Payload: []byte("Hello I'm Test 1"), CompFlags: 0x0000 | 0x8000},
			{Tag: uint16(2), Payload: []byte("Hello I'm Test 2"), CompFlags: 0x0000 | 0x8000},
			{Tag: uint16(10), Payload: []byte("Hello I'm Test Comp 10"), CompFlags: 0x0000 | 0x8000},
			{Tag: uint16(9), Payload: []byte("Hello Testing Heavy"), CompFlags: 0x0000 | 0x8000},
			{Tag: uint16(11), Payload: []byte("Heavy Data Heavy data Heavy Data Heavy Data Heavy data Heavy Data Heavy Data Heavy data Heavy Data Heavy Data Heavy data Heavy Data Heavy Data Heavy data Heavy Data Heavy Data Heavy data Heavy Data Heavy Data Heavy data Heavy Data Heavy Data Heavy data Heavy Data Heavy Data Heavy data Heavy Data"), CompFlags: 0x0000 | 0x8000},
			{Tag: uint16(3), Payload: []byte("Hello I'm Test 3EF"), CompFlags: 0x0000 | 0x8000},
			{Tag: uint16(4), Payload: []byte("Hello I'm Test 4AFE"), CompFlags: 0x0000 | 0x8000},
			{Tag: uint16(5), Payload: []byte("Hello I'm Test 5AFE"), CompFlags: 0x0000 | 0x8000},
			{Tag: uint16(6), Payload: []byte("Hello I'm Test 6 EFE"), CompFlags: 0x0000 | 0x8000},
			{Tag: uint16(7), Payload: []byte("Hello I'm Test 7 DZF"), CompFlags: 0x0000 | 0x8000},
			{Tag: uint16(8), Payload: []byte("Hello I'm Test 8 ABD"), CompFlags: 0x0000 | 0x8000},
		}

	default:
		return nil
	}

}

func BenchmarkEncode_Skinny(b *testing.B) {
	fields := makeTestFields("skinny")
	schemaID := uint64(112)
	hotTags := []uint16{
		uint16(1),
		uint16(2),
		uint16(3),
	}
	b.ReportAllocs()
	buf := make([]byte, 0, 1024)
	var e Encoder
	var out []byte
	for b.Loop() {
		out, _ = e.EncodeRecordFull(schemaID, hotTags, fields)
	}
	buf = buf[:0] // GC-friendly reuse
	buf = append(buf, out...)
	b.SetBytes(int64(len(buf))) // MB/s
}

// Allocs due to unordered fields
// Speed is reduced due to list ordering
func BenchmarkEncodeUnordered_Heavy(b *testing.B) {
	fields := makeTestFields("heavy")
	schemaID := uint64(112)
	hotTags := []uint16{
		uint16(1),
		uint16(2),
		uint16(3),
	}
	b.ReportAllocs()
	buf := make([]byte, 0, 1024)
	var e Encoder
	var out []byte
	for b.Loop() {
		out, _ = e.EncodeRecordFull(schemaID, hotTags, fields)
	}

	buf = buf[:0] // GC-friendly reuse
	buf = append(buf, out...)
	b.SetBytes(int64(len(buf))) // MB/s
}

// 0 allocs ordered fields
func BenchmarkEncodeordered_Heavy(b *testing.B) {
	fields := makeTestFields("heavy")
	schemaID := uint64(112)
	hotTags := []uint16{
		uint16(1),
		uint16(2),
		uint16(3),
	}
	b.ReportAllocs()
	buf := make([]byte, 0, 1024)
	var e Encoder
	var out []byte
	fields = Order(fields)
	for b.Loop() {
		out, _ = e.EncodeRecordFull(schemaID, hotTags, fields)
	}

	buf = buf[:0] // GC-friendly reuse
	buf = append(buf, out...)
	b.SetBytes(int64(len(buf))) // MB/s
}

/*
	func BenchmarkEncode_SkinnyHotVtable(b *testing.B) {
		fields := makeTestFields("skinny")
		schemaID := uint64(112)
		hotTags := []uint16{
			uint16(1),
			uint16(2),
			uint16(3),
			uint16(4),
		}
		b.ReportAllocs()
		buf := make([]byte, 0, 1024)
		var e Encoder
		var out []byte
		for b.Loop() {
			out, _ = e.EncodeRecordHot(schemaID, hotTags, fields)
		}
		buf = buf[:0] // GC-friendly reuse
		buf = append(buf, out...)
		b.SetBytes(int64(len(buf))) // MB/s
	}
*/
func TestDecodeRecordHot(t *testing.T) {
	field := makeTestFields("skinny")
	schemaID := uint64(112)
	hotTags := []uint16{
		uint16(1),
		uint16(2),
	}
	e := &Encoder{headerflag: 0x0001}
	var d Decoder
	enc, err := e.EncodeRecordFull(schemaID, hotTags, field)
	if err != nil {
		t.Fatal(err)
	}
	m, _ := d.DecodeRecordHot(enc)
	for i := range hotTags {
		t.Log(string(m[uint16(i)]))
	}
	/*for i := range hotTags {
		a, err := d.ReadHotField(enc, uint16(i+1), 0)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(a, field[i].Payload) {
			t.Failed()
		}
	}
	t.Logf("size with padding: %d", len(enc))*/
}

func BenchmarkDecode_SkinnyHot(b *testing.B) {
	fields := makeTestFields("skinny")
	schemaID := uint64(112)
	hotTags := []uint16{
		uint16(1),
		uint16(2),
		uint16(3),
		uint16(4),
	}
	var e Encoder
	var d Decoder
	raw, _ := e.EncodeRecordFull(schemaID, hotTags, fields)
	b.ReportAllocs()
	for b.Loop() {
		_, _ = d.ReadHotField(raw, uint16(3), 0)
	}
	b.SetBytes(int64(len(raw)))

}

func BenchmarkDecode_Skinny(b *testing.B) {
	fields := makeTestFields("skinny")
	schemaID := uint64(112)
	hotTags := []uint16{
		uint16(1),
		uint16(2),
		uint16(3),
	}
	e := &Encoder{headerflag: /*0x0001 |*/ 0x0002}
	var d Decoder
	raw, _ := e.EncodeRecordFull(schemaID, hotTags, fields)
	b.ReportAllocs()
	for b.Loop() {
		_, _ = d.DecodeRecord(raw, nil)
	}
	b.SetBytes(int64(len(raw)))
}

func BenchmarkDecode_heavy(b *testing.B) {
	fields := makeTestFields("heavy")
	schemaID := uint64(112)
	hotTags := []uint16{
		uint16(1),
		uint16(2),
		uint16(3),
		uint16(4),
		uint16(5),
		uint16(6),
		uint16(7),
		uint16(8),
	}
	var e Encoder
	var d Decoder
	raw, _ := e.EncodeRecordFull(schemaID, hotTags, fields)
	b.ReportAllocs()
	for b.Loop() {
		//_, _ = d.ReadHotField(raw, uint16(1), 0)
		_, _ = d.DecodeRecord(raw, nil)
	}
	b.SetBytes(int64(len(raw)))
}

func TestEncodeHot(t *testing.T) {
	fields := makeTestFields("heavy")
	schemaID := uint64(112)
	hotTags := []uint16{
		uint16(1),
		uint16(2),
		uint16(3),
		uint16(4),
		uint16(5),
		uint16(6),
		uint16(7),
		uint16(8),
	}
	var e Encoder
	var d Decoder

	raw, _ := e.EncodeRecordHot(schemaID, hotTags, fields)
	fields = Order(fields)
	for i := range hotTags {
		res, _ := d.ReadHotField(raw, uint16(i+1), 0)
		if !bytes.Equal(res, fields[i].Payload) {
			//t.Log(string(fields[i].Payload))
			t.Fail()
		}
	}

}

func TestEncodeTagWalk(t *testing.T) {
	fields := makeTestFields("skinny")
	fields = Order(fields)
	var e Encoder
	var d Decoder
	enc, _ := e.EncodeRecordTagWorK(fields)
	t.Log(enc)
	_, off, _ := d.DecodeRecordTagWalk(enc, 0)
	_, soff, _ := d.DecodeRecordTagWalk(enc, off)
	_, toff, _ := d.DecodeRecordTagWalk(enc, soff)
	r, _, _ := d.DecodeRecordTagWalk(enc, toff)
	t.Log(ReadAny(r[192], TypeUint32))
}

func BenchmarkWrite(b *testing.B) {
	a := uint16(1245)
	b.ReportAllocs()
	for b.Loop() {
		_, _ = Write(a)
	}
}

// ------------------------------------------------------------------------------
// Header Test
// ------------------------------------------------------------------------------

var theader = Header{
	Magic:       MagicV1,
	Version:     VersionV1,
	Flags:       0x03,
	SchemaID:    0xDEADBEEF,
	HotBitmap:   0b00001111,
	VTableSlots: 4,
	DataOffset:  64,
	VTableOff:   24,
}

func BenchmarkEncodeHeader(b *testing.B) {
	dst := make([]byte, 0, HeaderSize)
	for i := 0; i < b.N; i++ {
		dst = encodeHeader(dst[:0], theader)
	}
}
func BenchmarkParseHeader(b *testing.B) {
	data := encodeHeader(make([]byte, 0, HeaderSize), theader)
	for b.Loop() {
		_, _ = ParseHeader(data)
	}
}

// ------------------------------------------------------------------------------
// VarUint  Test
// ------------------------------------------------------------------------------

func BenchmarkWriteVarUint_Small(b *testing.B) {
	buf := make([]byte, 0)
	for b.Loop() {
		writeVarUint(buf, 127)
	}
}

func BenchmarkWriteVarUint_Large(b *testing.B) {
	buf := make([]byte, 0)
	for b.Loop() {
		writeVarUint(buf, 1<<56)
	}
}

// ------------------------------------------------------------------------------
// Inspector Test
// ------------------------------------------------------------------------------
// Finding elements with tags
func TestFindWithTag(t *testing.T) {
	fields := makeTestFields("heavy")
	schemaID := uint64(112)
	hotTags := []uint16{
		uint16(1),
		uint16(2),
		uint16(3),
		uint16(4),
		uint16(5),
		uint16(6),
		uint16(7),
		uint16(8),
	}
	var e Encoder
	var d Decoder
	raw, _ := e.EncodeRecordFull(schemaID, hotTags, fields)
	i, _ := Inspect(raw, &d)
	result, err := i.GetFieldD(uint16(10))
	dec, _ := d.DecodeRecord(raw, nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != string(dec[uint16(10)]) {
		t.Fail()
	}

}

func BenchmarkFindWithTag(b *testing.B) {
	fields := makeTestFields("heavy")
	schemaID := uint64(112)
	hotTags := []uint16{
		uint16(1),
		uint16(2),
		uint16(3),
		uint16(4),
		uint16(5),
		uint16(6),
		uint16(7),
		uint16(8),
	}
	e := &Encoder{headerflag: 0x0001}
	var d Decoder
	raw, _ := e.EncodeRecordFull(schemaID, hotTags, fields)
	i, _ := Inspect(raw, &d)
	result, err := i.GetFieldD(uint16(10))
	dec, _ := d.DecodeRecord(raw, nil)
	if err != nil {
		b.Fatal(err)
	}
	if string(result) != string(dec[uint16(10)]) {
		b.Fail()
	}
	b.ReportAllocs()
	for b.Loop() {
		_, _ = i.GetFieldD(uint16(10))
	}
	b.SetBytes(int64(len(raw)))
}

// ------------------------------------------------------------------------------
// Builder Test
// ------------------------------------------------------------------------------
func TestCommit(t *testing.T) {
	b := NewBuilder(nil)
	s := makeTestFields("skinny")
	s = Order(s)
	for _, i := range s {
		b.AddField(i.Tag, i.CompFlags, i.Payload, true)
	}
	b.Commit(uint64(123), uint16(0x0001))
	i, _ := Inspect(b.out, b.dec)
	for _, val := range s {
		if !(bytes.Equal(i.GetField(val.Tag), val.Payload)) {
			t.Fail()
		}
	}
}
