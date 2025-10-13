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
			if inf.Kind != reflect.String {
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
	tmp, err := utils.ListStructElem(&val)
	if err != nil {
		return err
	}
	info := utils.BuildInfo(tmp)
	cursor := 0
	for i, inf := range info {
		size := utils.GetLenght(inf.Kind)
		// string
		if size == 0 {
			a, n := utils.ReadVarUint(data[cursor:])
			cursor = cursor + n
			inf.Val, err = utils.ReadAny(data[cursor:cursor+int(a)], utils.TypeString)
			//fmt.Print(string(data[cursor : cursor+int(a)]))
			if err != nil {
				return err
			}
			utils.SetField(tmp[i], i, data[cursor:cursor+int(a)])
			cursor = cursor + int(a)
		} else {
			inf.Val, err = utils.ReadAny(data[cursor:cursor+size], utils.TypeInt8)
			if err != nil {
				return err
			}
			utils.SetField(tmp[i], i, data[cursor:cursor+size])
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
