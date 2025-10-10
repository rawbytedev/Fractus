package utils

import (
	"fmt"
	"reflect"
)

// works only on structs
// only list Elem from struct
func ListStructElem(v interface{}) ([]reflect.Value, error) {
	data := reflect.ValueOf(v)
	if data.Kind() != reflect.Struct {
		return nil, &reflect.ValueError{} // change to proper error message later
	}
	var elems []reflect.Value
	for i := range data.NumField() {
		elems = append(elems, data.Field(i))
	}
	return elems, nil
}
func SecListStructElem(v interface{}) ([]reflect.Value, error) {
	data := reflect.ValueOf(v)
	if data.Kind() != reflect.Struct {
		return nil, &reflect.ValueError{} // change to proper error message later
	}
	f := data.Type()
	field := f.Field(1)
	fmt.Print(field.Tag.Get("fractus"))

	return nil, nil
}

// extend it
// should never return an error
func ReturnConverted(v reflect.Value) any {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int:
		return v.Int()
	case reflect.Int8:
		return int8(v.Int())
	case reflect.Uint8:
		return uint8(v.Uint())
	case reflect.Uint64:
		return v.Uint()
	default:
		return nil
	}
}
