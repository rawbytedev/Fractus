# Fractus

![Test and Benchmark](https://github.com/rawbytedev/Fractus/actions/workflows/test-and-bench.yml/badge.svg)

Fractus is a lightweight, serialization library for Go.  
It encodes and decodes structs into a compact binary format with support for:

- Fixed-size primitive types (`int8`, `int16`, `int32`, `int64`, `uint*`, `float32`, `float64`, `bool`)
- Variable-size types (`string`, `[]byte`, slices of primitives/strings/byte slices)
- Struct pointers
- Presence bitmaps for optional fields
- Unsafe zero-copy string decoding (optional)

The goal is **fast, reproducible encoding/decoding** with fewer allocations(zero-allocs).

---

## Features

- **Struct encoding/decoding**: Works with exported fields of Go structs.
- **Slices and strings**: Handles variable-length data with varint length prefixes.
- **Presence bitmap**: Marks which fields are present.
- **Unsafe string mode**: Zero-copy decoding of strings (caller must ensure buffer lifetime).
- **Fuzz & property-based tests**: Ensures round-trip correctness.

---

## Installation

```bash
go get github.com/rawbytedev/fractus
```

---

## Usage

### Encode / Decode

```go
package main

import (
    "fmt"
    "github.com/rawbytedev/fractus"
)

type Example struct {
    Name   string
    Age    int32
    Scores []float64
}

func main() {
    f := &fractus{}

    val := Example{Name: "Alice", Age: 30, Scores: []float64{95.5, 88.0}}
    data, err := f.Encode(val)
    if err != nil {
        panic(err)
    }

    var out Example
    if err := f.Decode(data, &out); err != nil {
        panic(err)
    }

    fmt.Printf("Decoded: %+v\n", out)
}
```

---

## Benchmarks

Fractus aims to minimize allocations and improve throughput:

```bash
go test -bench=. -benchmem
```

With buffer pooling and unsafe string mode, allocations can be reduced further.

---

## Testing

Fractus includes fuzz and property-based tests:

```bash
go test ./...
```

- `FuzzEncodeDecode` ensures correctness across mixed types.
- `quick.Check` encoding/decoding for random structs.
- Error cases are tested (non-structs, unexported fields, wrong pointer types).

---

## Important

- **UnsafeStrings**: When enabled, decoded strings reference the original buffer.  
  Ensure the buffer outlives the string usage, or disable this option for safe copies.
- **Unexported fields**: Skipped during encoding.
- **Unsupported types**: Maps, interfaces, complex numbers, and nested slices (except `[]byte`) are not supported.
- Dev version achieve 0 allocs encoding/decoding(still experimental): [https://rawbytedev/fractus/dev](https://github.com/rawbytedev/Fractus/tree/dev)

---

## License

MIT License. See [LICENSE](LICENSE) for details.
