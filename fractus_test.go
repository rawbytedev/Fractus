package fractus

import (
	"testing"
	"testing/quick"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type MixedStruct struct {
	Val      string
	Mod      int8
	Data     string
	Integers int16
	Float3   float32
	Float6   float64
}

func FuzzEncodeDecode(f *testing.F) {
	f.Fuzz(fuzzMixedTypes)
}
func fuzzMixedTypes(t *testing.T, Val string, Mod int8, Data string, Integers int16, Float3 float32, Float6 float64) {
	val := MixedStruct{Val: Val, Mod: Mod, Data: Data, Integers: Integers, Float3: Float3, Float6: Float6}
	res := &MixedStruct{}
	f := &Fractus{}
	data, err := f.Encode(val)
	if err != nil {
		t.Log(err)
	}
	err = f.Decode(data, res)
	if err != nil {
		t.Log(err)
	}
	require.EqualExportedValues(t, val, *res)
}
func TestEncodeSimpleTypes(t *testing.T) {
	type NewStruct struct {
		Val      string
		Mod      int8
		Data     string
		Integers int16
		Float3   float32
		Float6   float64
	}
	z := NewStruct{Val: "azerty", Data: "testing",
		Mod: int8(17), Integers: 12,
		Float3: float32(12.3), Float6: float64(1236.2)}
	res := &NewStruct{}
	f := &Fractus{}
	data, err := f.Encode(z)
	if err != nil {
		t.Log(err)
	}
	err = f.Decode(data, res)
	if err != nil {
		t.Log(err)
	}
	require.EqualExportedValues(t, z, *res)
}
func TestConstant(t *testing.T) {
	type NewStructint struct {
		Int1  uint8
		Int2  int8
		Int3  uint16
		Int4  int16
		Int5  uint32
		Int6  int32
		Int7  uint64
		Int9  int64
		Const bool
	}
	f := &Fractus{}
	condition := func(z NewStructint) bool {
		data, err := f.Encode(z)
		if err != nil {
			t.Error(err)
		}
		res := &NewStructint{}
		err = f.Decode(data, res)
		if err != nil {
			t.Error(err)
		}
		require.EqualExportedValues(t, z, *res)
		return true
	}
	err := quick.Check(condition, &quick.Config{})
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}
func TestConstantList(t *testing.T) {
	type NewStructint struct {
		Int1  []uint8
		Int2  []int8
		Int3  []uint16
		Int4  []int16
		Int5  []uint32
		Int6  []int32
		Int7  []uint64
		Int9  []int64
		Const []bool
	}
	f := &Fractus{}
	condition := func(z NewStructint) bool {
		val := z
		data, err := f.Encode(val)
		if err != nil {
			t.Error(err)
		}
		res := &NewStructint{}
		err = f.Decode(data, res)
		if err != nil {
			t.Error(err)
		}
		require.EqualExportedValues(t, val, *res)
		return true
	}
	err := quick.Check(condition, &quick.Config{})
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}
func TestStructPointer(t *testing.T) {
	type StructPtr struct {
		Data string
	}
	val := &StructPtr{Data: "Hello"}
	res := &StructPtr{}
	f := &Fractus{}
	data, err := f.Encode(val)
	if err != nil {
		t.Error(err)
	}
	err = f.Decode(data, res)
	if err != nil {
		t.Error(err)
	}
	require.EqualExportedValues(t, val, res)
}
func TestErrors(t *testing.T) {
	f := &Fractus{}
	dt := "abc"
	data, err := f.Encode(dt)
	if err != ErrNotStruct {
		t.Error(err)
	}
	if data != nil {
		t.Error(err)
	}
	type Eas struct {
		val string // private
	}
	str := Eas{val: "hello"}
	Ptrstr := &Eas{val: "world"}
	data, err = f.Encode(Ptrstr)
	if len(data) > 1 {
		t.Error(err)
	}
	data, err = f.Encode(str)
	if len(data) > 1 {
		t.Error(err)
	}
	err = f.Decode(data, str) // needs pointer
	if err != ErrNotStructPtr {
		t.Error(err)
	}
}
func TestEncodeListOfTypes(t *testing.T) {

	type NewStruct struct {
		Val      []string
		Mod      []int8
		Integers []int16
		Float3   []float32
		Float6   []float64
	}
	f := &Fractus{}
	condition := func(z NewStruct) bool {
		data, err := f.Encode(z)
		if err != nil {
			t.Fatal(err)
		}
		res := &NewStruct{}
		f.Decode(data, res)
		require.EqualExportedValues(t, z, *res)
		return true
	}
	err := quick.Check(condition, &quick.Config{})
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}
func BenchmarkZeroAllocs(b *testing.B) {
	type ZeroAllocs struct {
		Int int8
	}
	z := ZeroAllocs{Int: int8(1)}
	f := &Fractus{Opts: Options{UnsafeStrings: true}}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = f.Encode(z)
	}
}

func BenchmarkEncoding(b *testing.B) {
	type NewStruct struct {
		Val      []string
		Mod      []int8
		Integers []int16
		Float3   []float32
		Float6   []float64
	}
	Val := []string{"azerty", "hello", "world", "random"}
	z := NewStruct{Val: Val,
		Mod: []int8{12, 10, 13, 0}, Integers: []int16{100, 250, 300},
		Float3: []float32{12.13, 16.23, 75.1}, Float6: []float64{100.5, 165.63, 153.5}}
	f := &Fractus{Opts: Options{UnsafeStrings: true}}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = f.Encode(z)
	}

}
func BenchmarkDecoding(b *testing.B) {
	type NewStruct struct {
		Val      []string
		Mod      []int8
		Integers []int16
		Float3   []float32
		Float6   []float64
	}
	Val := []string{"azerty", "hello", "world", "random"}
	z := NewStruct{Val: Val,
		Mod: []int8{12, 10, 13, 0}, Integers: []int16{100, 250, 300},
		Float3: []float32{12.13, 16.23, 75.1}, Float6: []float64{100.5, 165.63, 153.5}}
	y := &NewStruct{}
	f := &Fractus{Opts: Options{UnsafeStrings: true}}
	b.ReportAllocs()
	res, _ := f.Encode(z)
	for i := 0; i < b.N; i++ {
		f.Decode(res, y)
	}
	require.EqualValues(b, z, *y)
}
func BenchmarkFractus(b *testing.B) {
	type NewStructint struct {
		Int1 uint8
		Int2 int8
		Int3 uint16
		Int4 int16
		Int5 uint32
		Int6 int32
		Int7 uint64
		Int9 int64
	}
	z := NewStructint{Int1: 1, Int2: 2, Int3: 16, Int4: 18, Int5: 1586, Int6: 15262, Int7: 1547544565, Int9: 15484565656}
	y := &NewStructint{}
	f := &Fractus{}
	res := []byte{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		res, _ = f.Encode(z) // 3allocs
		f.Decode(res, y)     // 1allocs
	}
	require.EqualValues(b, z, *y)
}
func BenchmarkDoubleFractus(b *testing.B) {
	type NewStructint struct {
		Int1 uint8
		Int2 int8
		Int3 uint16
		Int4 int16
		Int5 uint32
		Int6 int32
		Int7 uint64
		Int9 int64
	}
	z := NewStructint{Int1: 1, Int2: 2, Int3: 16, Int4: 18, Int5: 1586, Int6: 15262, Int7: 1547544565, Int9: 15484565656}
	v := z
	y := &NewStructint{}
	f := &Fractus{}
	res := []byte{}
	b.ReportAllocs()
	_, _ = f.Encode(z)
	for i := 0; i < b.N; i++ {
		res, _ = f.Encode(v)
	}
	f.Decode(res, y)
	require.EqualValues(b, v, *y)
}
func BenchmarkYaml(b *testing.B) {
	type NewStructint struct {
		Int1 uint8
		Int2 int8
		Int3 uint16
		Int4 int16
		Int5 uint32
		Int6 int32
		Int7 uint64
		Int9 int64
	}
	z := NewStructint{Int1: 1, Int2: 2, Int3: 16, Int4: 18, Int5: 1586, Int6: 15262, Int7: 1547544565, Int9: 15484565656}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = yaml.Marshal(z)
	}
}
