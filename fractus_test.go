package fractus

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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

func TestEncodeListOfTypes(t *testing.T) {
	type NewStruct struct {
		Val      []string
		Mod      []int8
		Integers []int16
		Float3   []float32
		Float6   []float64
	}
	z := NewStruct{Val: []string{"azerty", "hello", "world", "random"},
		Mod: []int8{12, 10, 13, 0}, Integers: []int16{100, 250, 300},
		Float3: []float32{12.13, 16.23, 75.1}, Float6: []float64{100.5, 165.63, 153.5}}
	f := &Fractus{}
	data, err := f.Encode(z)
	if err != nil {
		t.Fatal(err)
	}
	res := &NewStruct{}
	f.Decode(data, res)
	require.EqualExportedValues(t, z, *res)
}

func BenchmarkEncoding(b *testing.B) {
	type NewStruct struct {
		Val      []string
		Mod      []int8
		Integers []int16
		Float3   []float32
		Float6   []float64
	}
	z := NewStruct{Val: []string{"azerty", "hello", "world", "random"},
		Mod: []int8{12, 10, 13, 0}, Integers: []int16{100, 250, 300},
		Float3: []float32{12.13, 16.23, 75.1}, Float6: []float64{100.5, 165.63, 153.5}}
	f := &Fractus{Opts: Options{UnsafeStrings: true}}
	b.ReportAllocs()

	for b.Loop() {
		_, _ = f.Encode(z)
		/*b.StopTimer()
		/*if err != nil {
			b.Fatal(err)
		} /*
			res := &NewStruct{}
			b.StartTimer()
			f.Decode(data, res)
			b.StopTimer()
			require.EqualExportedValues(b, z, *res)*/
	}

}
