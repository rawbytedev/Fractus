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
}

func (f *Fractus) Encode(val interface{}) ([]byte, error) {
	tmp, err := utils.ListStructElem(val)
	if err != nil {
		return nil, err
	}
	var res []byte
	infos := utils.BuildInfo(tmp)
	for _, inf := range infos {
		if a, err := utils.Write(inf.Val); err != nil {
			return nil, err
		} else {
			if inf.Kind != reflect.TypeFor[string]() {
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
	v := reflect.ValueOf(val)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		panic("expectedpointer to struct")
	}
	v = v.Elem()
	//t := v.Type()
	cursor := 0
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		size := utils.GetLength(field.Kind())
		// string
		if size == 0 {
			a, n := utils.ReadVarUint(data[cursor:])
			cursor = cursor + n
			if field.CanSet() {
				field.SetString(string(data[cursor : cursor+int(a)]))
			}
			cursor = cursor + int(a)
		} else {
			Val, err := utils.ReadAny(data[cursor:cursor+size], utils.TypeInt8)
			if err != nil {
				return err
			}
			if field.CanSet() {
				field.SetInt(int64(Val.(int8)))
			}
			//fmt.Print(inf.Val)
			cursor = cursor + size
		}

	}
	//panic("can't assign to val yet")
	return nil
}

func NewFractus() IFractus {
	return &Fractus{}
}
func main() {
	NewFractus()
}
