package main

import (
	"fractus"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	f, err := os.Create("mem.prof")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	runtime.MemProfileRate = 1
	type NewStruct struct {
		Val      []string
		Mod      []int8
		Integers []int16
		Float3   []float32
		Float6   []float64
		str      string
		in       int8
		Data     string
		Inte     int16
		Float    float32
		Float64  float64
	}
	z := NewStruct{Val: []string{"azerty", "hello", "world", "random"},
		Mod: []int8{12, 10, 13, 0}, Integers: []int16{100, 250, 300},
		Float3: []float32{12.13, 16.23, 75.1}, Float6: []float64{100.5, 165.63, 153.5},
		str: "azerty", Data: "testing",
		in: int8(17), Inte: 12,
		Float: float32(12.3), Float64: float64(1236.2)}
	y := &fractus.Fractus{Opts: fractus.Options{UnsafeStrings: true}}
	y2 := &fractus.Fractus{Opts: fractus.Options{UnsafeStrings: false}}
	for i := 0; i < 10000; i++ {
		data, _ := y.Encode(z)
		_, _ = y2.Encode(z)
		res := &NewStruct{}
		_ = y.Decode(data, res)
	}
	pprof.WriteHeapProfile(f)
	time.Sleep(5 * time.Minute)
}
