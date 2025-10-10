package core

import (
	"errors"
	"fractus/utils"
	"reflect"
)

var (
	ErrInvalidElem = errors.New("invalid Element")
)

// core struct
type Fractus struct {
	types []InterType
}

// usually name of field aren't important to us
// so we just use an Id for each field and expose field names as extension
type InterType struct {
	id   int
	kind reflect.Kind //need for type assertion
	val  any
}

func (f *Fractus) Encode(val interface{}) ([]byte, error) {
	tmp, err := utils.ListStructElem(val)
	if err != nil {
		return nil, err
	}
	var res []byte
	for i, dt := range tmp {
		f.types = append(f.types, InterType{
			val:  utils.ReturnConverted(dt),
			kind: dt.Kind(),
			id:   i,
		})
		if a, err := utils.Write(f.types[i].val); err != nil {
			return nil, err
		} else {
			if f.types[i].kind != reflect.String {
				res = append(res, a...)
			} else {
				res = append(res, utils.WriteVarUint(make([]byte, 0), uint64(len(a)))...)
				res = append(res, a...)
			}
		}
	}
	return res, nil
}
func (f *Fractus) Decode(data []byte, val interface{}) error {
	return nil
}
func NewFractus() IFractus {
	return &Fractus{types: make([]InterType, 0)}
}
func main() {
	NewFractus()
}
