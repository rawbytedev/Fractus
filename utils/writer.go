package utils

import (
	"encoding/binary"
	"fmt"
	"math"
)

type FieldType int

const (
	TypeBool FieldType = iota
	TypeBytes
	TypeFloat32
	TypeFloat64
	TypeInt8
	TypeInt16
	TypeInt32
	TypeInt64
	TypeString
	TypeUint8
	TypeUint16
	TypeUint32
	TypeUint64
)

// Helper functions to turn any value to bytes
func Write(value any) ([]byte, error) {
	switch v := value.(type) {
	case uint8:
		return []byte{v}, nil
	case uint16:
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, v)
		return buf, nil
	case uint32:
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, v)
		return buf, nil
	case uint64:
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, v)
		return buf, nil
	case int8:
		return []byte{byte(v)}, nil
	case int16:
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, uint16(v))
		return buf, nil
	case int32:
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(v))
		return buf, nil
	case int64:
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, uint64(v))
		return buf, nil
	case float32:
		bits := math.Float32bits(v)
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, bits)
		return buf, nil
	case float64:
		bits := math.Float64bits(v)
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, bits)
		return buf, nil
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", value)
	}
}

// From bytes to any
func ReadAny(buf []byte, typ FieldType) (any, error) {
	switch typ {
	case TypeBool:
		return buf[0] != 0, nil
	case TypeUint8:
		return buf[0], nil
	case TypeUint16:
		return binary.LittleEndian.Uint16(buf), nil
	case TypeUint32:
		return binary.LittleEndian.Uint32(buf), nil
	case TypeUint64:
		return binary.LittleEndian.Uint64(buf), nil
	case TypeInt8:
		return int8(buf[0]), nil
	case TypeInt16:
		return int16(binary.LittleEndian.Uint16(buf)), nil
	case TypeInt32:
		return int32(binary.LittleEndian.Uint32(buf)), nil
	case TypeInt64:
		return int64(binary.LittleEndian.Uint64(buf)), nil
	case TypeFloat32:
		bits := binary.LittleEndian.Uint32(buf)
		return math.Float32frombits(bits), nil
	case TypeFloat64:
		bits := binary.LittleEndian.Uint64(buf)
		return math.Float64frombits(bits), nil
	case TypeString:
		return string(buf), nil
	case TypeBytes:
		return buf, nil
	default:
		return nil, fmt.Errorf("unknown field type: %d", typ)
	}
}

// writeVarUint writes a u8 as LEB128 varint.
func WriteVarUint(buf []byte, x uint64) []byte {
	for x >= 0x80 {
		buf = append(buf, byte(x)|0x80)
		x >>= 7
	}
	buf = append(buf, byte(x))
	return buf
}

// readVarUint reads a varint from buf, returns value and bytes read.
func ReadVarUint(b []byte) (uint64, int) {
	var x uint64
	var s uint
	for i, c := range b {
		x |= uint64(c&0x7F) << s
		if c&0x80 == 0 {
			return x, i + 1
		}
		s += 7
	}
	return 0, 0 // truncated
}
func align(off, a int) int {
	return off + ((a - (off % a)) % a)
}
