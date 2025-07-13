package dbflat

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

// Helper functions to encode int into byte slices
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

// read byte and turns into corresponding type
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

func WriteUint24(v uint32) []byte {
	if v > 0xFFFFFF {
		panic("Value too large for uint24")
	}
	return []byte{
		byte(v),
		byte(v >> 8),
		byte(v >> 16),
	}
}

// writeVarUint writes a u64 as LEB128 varint.
func writeVarUint(buf []byte, x uint64) []byte {
	for x >= 0x80 {
		buf = append(buf, byte(x)|0x80)
		x >>= 7
	}
	buf = append(buf, byte(x))
	return buf
}

// readVarUint reads a varint from buf, returns value and bytes read.
func readVarUint(b []byte) (uint64, int) {
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

// uint24 helpers
func writeUint24(buf *bytes.Buffer, v uint32) {
	buf.WriteByte(byte(v))
	buf.WriteByte(byte(v >> 8))
	buf.WriteByte(byte(v >> 16))
}

func readUint24(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16
}

func align(off, a int) int {
	return off + ((a - (off % a)) % a)
}

// zigzag helper
