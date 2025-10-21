package core

import (
	"fmt"
	"testing"
)

type Test struct {
	Data string
	Id   int8
	Op   string
}

func TestFractusEncode(t *testing.T) {

	var val Test
	val.Data = "Hello"
	val.Id = 1
	val.Op = "zero"
	f := NewFractus()
	data, err := f.Encode(val)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Print(data)
	//*
	var dt Test
	err = f.Decode(data, &dt)
	if err != nil {
		t.Fatal(err)
	} //*/
	fmt.Print(dt.Id)
	fmt.Print(dt.Data)
	fmt.Print(dt.Op)
}
