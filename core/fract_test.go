package core

import (
	"fmt"
	"testing"
)

func TestFractusEncode(t *testing.T) {
	type test struct {
		Data string
		Id   int8
		Op   string
	}
	var val test
	val.Data = "Hello"
	val.Id = 1
	val.Op = "zero"
	f := NewFractus()
	data, err := f.Encode(val)
	if err != nil {
		t.Fatal(err)
	}
	var dt test
	f.Decode(data, &dt)
	fmt.Print(dt.Id)
}
