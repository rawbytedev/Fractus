# EncodeRecordTagWalk() & DecodeRecordTagWalk()

The TagWalk mode is Fractus’s most stream-friendly layout. Each field is encoded as a self-contained unit, allowing progressive, linear decoding without metadata or vtable support.

Perfect for logs, telemetry, field-by-field scanning, or when schema introspection isn’t needed upfront.

## Documentation

---

### Layout Structure

```go
for array:
[Tag][CompFlags][Len][Payload][Tag][CompFlags][Len][Payload]
for fixedwidth:
[Tag][CompFlags][Payload]
```

| Component | Size      | Description                  |
| --------- | --------- | ---------------------------- |
| Tag       | 2B        | Field identifier (uint16)    |
| CompFlags | 2B        | Compression & array flags    |
| Len       | Varint    | Length of payload            |
| Payload   | Len bytes | Raw or compressed field data |

---

### Encoding Workflow

Each field is written in order. If compression is active, payload is compressed and prefixed by varint-encoded length.

```go
enc := dbflat.NewEncoder()
record, _ := enc.EncodeRecordTagWorK(fields)
```

---

### Hexdump Example (3 fields)

```go
00–01  : 01 00        → Tag = 1
02–03  : 80 00        → CompFlags = 0x8000 (array)
04     : 0E           → Length = 14 bytes
05–18  : "Hello Field 1"
19–20 : 02 00 → Tag = 2
21–22 : 80 00 → Flags
23 : 0E → Length
24–37 : "Hello Field 2"
38–39 : C0 00 → Tag = 192
40–41 : 00 00 → Raw
42–45 : 2C 01 00 00 → 300 (uint32)
```

---

### Progressive Scanning Diagram

Here's how decoding advances field-by-field:

```go
+------------------------------------------+
Buf Offset = 0
→ Read Tag (2B)
→ Read CompFlags (2B)
→ Read Length (Varint)
→ Skip Payload [Length]
→ Next Field: buf[offset + 4 + length]
+------------------------------------------+
```

```go
dec := dbflat.NewDecoder()
offset := 0
for {
 fieldMap, nextOff, err := dec.DecodeRecordTagWalk(buf, offset)
 if err != nil { break }
 fmt.Println("Tag:", keys(fieldMap)[0], "Data:", fieldMap)
 offset = nextOff
}
```

---

### Offset Strategy

Offsets in TagWalk are incremental, computed as:

`NextOffset = CurrentOffset + 4 + PayloadLength + VarintSize`

- No centralized index
- Decoder walks forward linearly
- Stops when buffer exhausted or tag matched

---

### When to Use TagWalk

| Use Case        | Why It Shines**                        |
| --------------- | -------------------------------------- |
| Log Records     | Decode as entries stream in            |
| Field Filtering | Stop decoding once target tag found    |
| MTU-fragmented  | Works across partial buffers           |
| Minimal Layouts | No vtable, no bitmap, no schema needed |

---

### Size Comparison

| Config             | Header | VTable | Payload | Total |
| ------------------ | ------ | ------ | ------- | ----- |
| TagWalk (3 fields) | 0B     | 0B     | ~48B    | ~48B  |
| HotVtable (same)   | 32B    | 24B    | ~48B    | ~104B |

---

### Decoder Output

```go
{
  1: []byte("Hello Field 1"),
  2: []byte("Hello Field 2"),
  192: []byte{0x2C, 0x01, 0x00, 0x00}
}
```
