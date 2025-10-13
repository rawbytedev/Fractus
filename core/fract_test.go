package core

import (
	"fmt"
	"testing"
)

func TestFractusEncode(t *testing.T) {
	type test struct {
		data string
		id   int8
		op   string
	}
	var val test
	val.data = "Hello"
	val.id = 1
	val.op = "zero"
	f := NewFractus()
	data, err := f.Encode(val)
	if err != nil {
		t.Fatal(err)
	}
	var dt test
	f.Decode(data, dt)
	fmt.Print(dt.data)
}
