package utils

import (
	"fmt"
	"reflect"
)

// works only on structs
func ListStructElem(v interface{}) error {
	data := reflect.ValueOf(v)
	if data.Kind() != reflect.Struct {
		return &reflect.ValueError{} // change to proper error message later
	}
	maxitem := data.NumField()
	var val []reflect.Value
	for i := range maxitem {
		val = append(val, data.Field(i))
	}
	for _, dt := range val {
		fmt.Print(ReturnConverted(dt))
	}
	return nil
}

// extend it
func ReturnConverted(v reflect.Value) any {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int:
		return v.Int()
	default:
		return nil
	}
}
