package utils

import (
	"fmt"
	"testing"
)

func TestBuildinfo(t *testing.T) {
	type test struct {
		daz string 
		id  int    
		op  string
	}
	var val test
	store, err := ListStructElem(val) // simple test for utils
	if err != nil {
		t.Fatal(err)
	}
	info := BuildInfo(store)
	for i, d := range info {
		fmt.Printf("id: %d, val: %s\n", i, d.Kind)
	}
}
