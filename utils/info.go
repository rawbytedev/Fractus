package utils

import "reflect"

type InterType struct {
	Id   int
	Kind reflect.Kind //needed for type assertion
	Val  any          // data itself
}

func BuildInfo(fields []reflect.Value) []InterType {
	var info []InterType
	for i, dt := range fields {
		info = append(info, InterType{
			Id:   i,
			Val:  ReturnConverted(dt),
			Kind: dt.Kind(),
		})
	}
	return info
}

func GetLenght(kd reflect.Kind) int {
	switch kd {
	case reflect.String:
		return 0
	case reflect.Int8:
		return 1
	default:
		return 0
	}
}
