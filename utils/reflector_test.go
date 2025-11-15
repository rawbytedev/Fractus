package utils

import (
	"fmt"
	"testing"
)

func TestUtilsListStructElem(t *testing.T) {
	type test struct {
		daz string `fractus:"data"`
		id  int8    `fractus:"datad"`
		op  string
	}
	var val test
	store, err := ListStructElem(val) // simple test for utils
	if err != nil {
		t.Fatal(err)
	}
	for i := range store {
		fmt.Print(ReturnConverted(store[i]))
	}
}

func BenchmarkUtilsListStructElem(b *testing.B) {
	type test struct {
		data string
		id   int
		op   string
	}
	var val test
	b.ReportAllocs()
	store, err := ListStructElem(val) // simple test for utils

	if err != nil {
		b.Fatal(err)
	}
	for i := range store {
		ReturnConverted(store[i])
	}

}
