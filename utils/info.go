package utils

import (
	"fmt"
	"reflect"
)

type InterType struct {
	Id   int
	Kind reflect.Type //needed for type assertion
	Val  any          // data itself
}

func BuildInfo(fields []reflect.Value) []InterType {
	var info []InterType
	for i, dt := range fields {
		t := dt.Type()
		// unwrap pointers so Kind is the underlying type
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		fmt.Println("field", i, "type:", dt.Type(), "-> stored as:", t)

		info = append(info, InterType{
			Id:   i,
			Val:  ReturnConverted(dt),
			Kind: t, // store the normalized type
		})
	}
	return info
}
func Info(strct reflect.Value) []InterType {
    t := strct.Type()
    var info []InterType
    for i := 0; i < t.NumField(); i++ {
        f := t.Field(i)        // reflect.StructField
        fv := strct.Field(i)   // reflect.Value of the field

        // unwrap pointer types if you want the underlying element
        ft := f.Type
        for ft.Kind() == reflect.Ptr {
            ft = ft.Elem()
        }

        fmt.Printf("field %d declared: %s -> stored as: %s\n", i, f.Type, ft)

        info = append(info, InterType{
            Id:   i,
            Kind: ft,          // normalized type
            Val:  fv.Interface(),
        })
    }
    return info
}


func GetLength(kd reflect.Kind) int {
	switch kd {
	case reflect.String:
		// variable length, handled separately
		fmt.Print("String detected")
		return 0
	case reflect.Int8, reflect.Uint8:
		return 1
	case reflect.Int16, reflect.Uint16:
		return 2
	case reflect.Int32, reflect.Uint32, reflect.Float32:
		return 4
	case reflect.Int64, reflect.Uint64, reflect.Float64:
		return 8
	default:
		fmt.Print("Unsupported type detected")
		return -1 // unsupported type
	}
}

func SetField(val reflect.Value, idx int, data []byte) {
	v := val.Elem().Field(idx)
	if !v.CanSet() {
		return
	}
	if v.CanSet() {
		//fmt.Print("Yeah")
		switch v.Kind() {
		case reflect.String:
			v.SetString(string(data))
		case reflect.Int8:
			res, err := ReadAny(data, TypeInt8)
			if err != nil {
				return
			}
			a, ok := res.(int8)
			if ok {
				v.SetInt(int64(a))
			}
		default:
			panic("unsupported")

		}
	}
}
