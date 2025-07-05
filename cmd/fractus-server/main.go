package fractus

import (
	"fractus/pkg/compactwire"
	"fractus/pkg/dbflat"
	"reflect"
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
	for i := range rv.NumField(){
		rv.Field(i)
		_ =t
	}
}
