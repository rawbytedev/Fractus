package utils

import "testing"

func TestRe(t *testing.T) {
	type test struct {
		data string
		id   int
		op   string
	}
	var val test
	val.data = "Hello"
	val.id = 1
	ListStructElem(val)
}
