package dbflat_test

import (
	//"encoding/binary"
	"fractus/pkg/dbflat"

	"testing"

	"github.com/klauspost/compress/zstd"
)

func TestDecodeHotField(t *testing.T) {
	field := makeTestFields("skinny")
	schemaID := uint64(112)
	hotTags := []uint16{
		uint16(1),
		uint16(2),
	}
	var e dbflat.Encoder
	var d dbflat.Decoder
	enc, err := e.EncodeRecord(schemaID, hotTags, field)
	if err != nil {
		t.Fatal(err)
	}
	a, err := d.DecodeRecord(enc, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(dbflat.ReadAny(a[2], dbflat.TypeString))
	/*
		for i := range 8 {
			if i != 0 {
				result, _ := d.ReadHotField(enc, uint16(i), 0)
				t.Log(dbflat.ReadAny(result, dbflat.TypeString))
				//t.Log(dbflat.ReadAny(result, dbflat.TypeUint32))

			}
		}*/

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
	res, err := dec.DecodeAll(s, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(res))
}

func TestWriter(t *testing.T) {
	a, _ := dbflat.Write(uint32(1000))
	t.Log(dbflat.ReadAny(a, dbflat.TypeUint32))
}

func makeTestFields(shape string) []dbflat.FieldValue {
	switch shape {
	case "skinny":
		a, _ := dbflat.Write(uint32(300))
		return []dbflat.FieldValue{
			{Tag: uint16(1), Payload: []byte("Hello I'm Test 1"), CompFlags: 0x0000 | 0x8000},
			{Tag: uint16(2), Payload: []byte("Hello I'm Test 2"), CompFlags: 0x0000 | 0x8000},
			{Tag: uint16(3), Payload: []byte("Hello I'm Test Comp+10"), CompFlags: 0x0000 | 0x8000},
			{Tag: uint16(4), Payload: a, CompFlags: 0x0000 | 0x8000},
		}
	case "heavy":
		return []dbflat.FieldValue{
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
	var e dbflat.Encoder
	var out []byte
	for b.Loop() {
		out, _ = e.EncodeRecord(schemaID, hotTags, fields)
	}
	buf = buf[:0] // GC-friendly reuse
	buf = append(buf, out...)
	b.SetBytes(int64(len(buf))) // MB/s
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
	var e dbflat.Encoder
	var d dbflat.Decoder
	raw, _ := e.EncodeRecord(schemaID, hotTags, fields)
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
	var e dbflat.Encoder
	var d dbflat.Decoder
	raw, _ := e.EncodeRecord(schemaID, hotTags, fields)
	b.ReportAllocs()
	for b.Loop() {
		//_, _ = d.ReadHotField(raw, uint16(3), 0)
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
	var e dbflat.Encoder
	var d dbflat.Decoder
	raw, _ := e.EncodeRecord(schemaID, hotTags, fields)
	b.ReportAllocs()
	for b.Loop() {
		//_, _ = d.ReadHotField(raw, uint16(1), 0)
		_, _ = d.DecodeRecord(raw, nil)
	}
	b.SetBytes(int64(len(raw)))
}
