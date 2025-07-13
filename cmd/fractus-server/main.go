package main

import (
	"fractus/pkg/compactwire"
	"fractus/pkg/dbflat"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"reflect"
	"runtime"
	"runtime/pprof"
	"time"
)

type Fractus struct {
	d *dbflat.Dbflat
	n *compactwire.Compactwire
	b *Buffer
}

type Buffer struct {
	writer []*byte
	reader []*byte
}

func (b *Buffer) ResetBuffer() {
	b.writer = b.writer[:0]
	b.reader = b.reader[:0]
}

// Encode data for local storage
func (f *Fractus) LocalMarshal(data any) {

}

// Decode data from local storage
func (f *Fractus) LocalUnmarshal(data any) {

}

func (f *Fractus) ParseStruct(data any) {
	rv := reflect.ValueOf(data).Elem()
	t := rv.Type()
	for i := range rv.NumField() {
		rv.Field(i)
		_ = t
	}
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

// Debugging
func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	fields := makeTestFields("skinny")
	schemaID := uint64(112)
	hotTags := []uint16{
		uint16(1),
		uint16(2),
		uint16(3),
	}
	var e dbflat.Encoder
	var d dbflat.Decoder
	f, err := os.Create("mem.prof")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	payload := []byte("hello, CW!")
	flags := compactwire.FlagEndOfMessage
	da := &compactwire.DataFrame{}

	runtime.MemProfileRate = 1
	for i := 0; i < 10000; i++ {
		a, _ := e.EncodeRecordFull(schemaID, hotTags, fields)
		_, _ = d.DecodeRecord(a, nil)

		frame, _ := da.EncodeDataFrame(payload, flags, nil)
		_, _, _, _ = da.DecodeDataFrame(frame)
	}
	pprof.WriteHeapProfile(f)
	time.Sleep(5 * time.Minute)
}
