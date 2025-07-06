# 🧬 Fractus

**Fractus** is a zero-allocation, schema-evolving, developer-first framework for encoding structured data in decentralized systems. It combines high-performance encoding with introspective design, MTU-aware layouts, and reversible traceability, making it ideal for protocols that demand clarity, compactness, and control.

> Because protocol design deserves as much care as protocol logic.

✨ Features

- ⚡ **Zero-Allocation Encoding & Decoding**  
  Fractus’ `dbflat` engine encodes data directly to preallocated buffers—no GC pressure, ever.

- 🧠 **Schema Evolution with Version Tolerance**  
  Dynamic field introspection, offset tables, and forward-compatible layouts.

- 🧩 **Composable ControlFrames**  
  Meta-communication channels for runtime signaling and fault recovery.

- 📡 **MTU-aware Fragmentation**  
  Frame construction respects network boundaries for efficient transport.

- 🛠 **Developer Ergonomics First**  
  Clear APIs, reusable scratch buffers, and debugging visibility built-in.

## 📦 Core Components

Component  Purpose
dbflat  Schema-structured encoding/decoding with sorted field layout
CompactWire  Framing layer with CRC, offset tables, and payload segmentation
ControlFrames  Runtime signals, behavior toggles, and error recovery commands
Encoder / Decoder  Allocation-free core for building and parsing frames

## 🚀 Getting Started

```bash
go get github.com/rawbytedev/Fractus

```

Use dbflat to define and encode structured records:

```go
fields := []dbflat.Field{
    {Tag: 1, CompFlags: 0, Payload: []byte("hello")},
    {Tag: 2, CompFlags: 0, Payload: []byte("world")},
}

enc := dbflat.NewEncoder()
frame := enc.EncodeRecord(recordID, metadata, fields)
```

Use dbflat.DecodeRecord to parse with schema hints and introspect fields on the fly.

---

## 🧪 Performance

| Metric         | Result           |
| -------------- | ---------------- |
| Encode Alloc   | 0 allocations/op |
| Decode Alloc   | 0 allocations/op |
| Encode Speed   | Blazing fast™    |
| MTU Compliance | Built-in         |

Benchmarks were profiled with go test -bench and pprof to ensure memory and latency flatlines across large volumes.

---

## 📜 Philosophy

Fractus is rooted in a belief that:

- Traceability and reversibility are not luxuries—they’re trust primitives
- Protocols should evolve like well-versioned code, not opaque blobs
- Schema is not metadata—it is the data
- Tooling should reveal what the wire obscures

---

## 🧭 Roadmap Highlights

- [x] Zero-allocation encode/decode path
- [x] Sorted field layout with introspection
- [] MTU fragmentation and field grouping
- [x] ControlFrames design
- [] Runtime schema converters
- [] Schema negotiation over protocol

---

🤝 Contributing

Pull requests are welcome! Ideas, use cases, and design critiques are even better. The more we evolve together, the more expressive the protocol becomes.

---

🧑‍🚀 Author

Developed with precision by @rawbytedev 

---

📄 License

MIT
