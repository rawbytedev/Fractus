# Fractus

**Fractus** is a zero-allocation, introspectable encoding framework built for schema-evolving protocols and decentralized systems.
Fractus supports multiple encoding layouts, dynamic header logic, and streaming field inspection—all while keeping performance flat and GC-free.

> Because protocol design deserves as much care as protocol logic.

---

## ✨ Features

- ⚡ **Zero-Allocation Encoding & Decoding**  
  Fractus encodes directly into preallocated buffers—no heap allocations, no GC pressure.

- 🧠 **Schema Evolution Ready**  
  Forward-compatible layouts, introspection, and optional schema ID fields.

- 🧩 **Multi-Mode Encoding**  
  Choose between full vtable, hotfield-indexed, or streamable tag-walk formats.

- 🔍 **Fast Field Lookup**  
  Hot fields decoded with O(1) access; cold fields scanned via streaming or schema-driven offsets.

- 🛠 **Developer Ergonomics First**  
  Built-in `Builder`, `Inspector`,`Encoder` and `Decoder` APIs with reusable scratch buffers.

- 📦 **Compression & Encoding Flags**  
  Supports Zstd, Huffman, RLE, and raw payload modes via `CompFlags`.

- 📡 **MTU-aware Fragmentation (Planned)**  
  Future framing module (`CompactWire`) will support segmented frames and CRC.

---

## 📦 Core Modules

| Component       | Purpose                                                  |
| --------------- | -------------------------------------------------------- |
| `dbflat`        | Layout-aware record encoder/decoder                      |
| `Encoder`       | Compresses + encodes fields into chosen format           |
| `Decoder`       | Efficient field access, hotfield reads, tagwalk scanner  |
| `Inspector`     | Field introspection, partial decoding, lazy scanning     |
| `Builder`       | API for appending structured fields into reusable buffer |
| `ControlFrames` | Runtime signal layer for coordination & recovery (WIP)   |

---

## 🚀 Getting Started

```bash
go get github.com/rawbytedev/Fractus
```

### 🔐 Define Fields

```go
fields := []dbflat.FieldValue{
    {Tag: 1, CompFlags:0x8000, Payload: []byte("hello")},
    {Tag: 2, CompFlags:0x8000, Payload: []byte("world")},
}
```

### 🔧 Encode in Full VTable Mode

```go
enc := dbflat.NewEncoder()
record, err := enc.EncodeRecordFull(0xDEAD, []uint16{1}, fields)
```

### 🧠 Decode with Schema-Awareness

```go
dec := dbflat.NewDecoder()
parsed, err := dec.DecodeRecord(record, nil)
fmt.Println(string(parsed[1])) // → "hello"
```

### 🔍 Inspect Tag-Walk Field Stream

```go
ins, _ := dbflat.Inspect(record, dec)
value := ins.GetFieldD(2)
```

---

### 🔁 Encoding Modes

| Mode          | API Function          | Lookup        | Notes                                    |
|:-------------:|:---------------------:|:-------------:|:----------------------------------------:|
| `Full VTable` | EncodeRecordFull()    | O(1)/O(log n) | Best for introspection + large schemas   |
| `Hot VTable`  | EncodeRecordHot()     | O(1)+stream   | Up to 8 hotfields with fast field jumps  |
| `Tag-Walk`    | EncodeRecordTagWalk() | O(n)          | Streamable; great for logs and telemetry |

Each mode adapts layout and header flags automatically.

---

📄 Header Flags

Fractus uses bitwise flags to signal layout configuration:

```go
const (
  FlagPadding       = 0x0001 // Align payload to 8B
  FlagNoSchemaID    = 0x0002 // Schema ID omitted
  FlagModeHotVtable = 0x0004
  FlagModeNoVtable  = 0x0008
  FlagModeTagWalk   = 0x0010
)
```

Use combinations like:

```go
flags := FlagModeHotVtable | FlagPadding
```

---

🧪 Benchmarks & Testing

Run performance benchmarks via:

```bash
cd pkg/dbflat
go test -bench=.
```

Highlights:

- 💨 0 allocs per encode/decode
- ⚙️ Hotfield lookup ~O(1)
- 🎯 Streaming decode available via Inspector.Next()
- 🧪 Verified against compressed + fixed-width payloads

---

🧰 Debug & Utilities

- Builder: Append + commit fields into reusable buffer
- Inspector: Scan, peek, and retrieve fields dynamically
- ReadHotField: Fast-path access for hotfields
- DecodeRecordTagWalk: Line-by-line decode for streamable records
- WriteUint24, writeVarUint, ReadAny: Payload converters

---

📜 Protocol Philosophy

Fractus is built on the belief that:

- Traceability and reversibility are trust primitives
- Schema shouldn't be metadata—it should be accessible, introspectable data
- Decoding should be partial, lazy, and composable
- Layouts should reveal intention—not obscure it

---

📈 Roadmap Highlights

- [x] Hotfield + TagWalk decoding synergy
- [x] Header-mode routing via flag bits
- [x] Compression and array-length support
- [x] Inspector with tag-based scanning
- [x] Benchmarks with heavy & skinny payloads
- [ ] CompactWire framing with CRC + chunking
- [ ] Runtime schema negotiation tools
- [ ] CLI tool: fractus inspect, fractus encode, etc.

---

🤝 Contributing

Pull requests are welcome. Features, flags, format diagrams, and spec clarifications are even better. You can help Fractus become the Rosetta Stone of wire formats.

---

🧑‍🚀 Author

Crafted with surgical care by @rawbytedev

---

📄 License

MIT