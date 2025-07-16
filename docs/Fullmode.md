# Full Mode Documentation

## 📄 EncodeRecordFull Documentation

Fractus's EncodeRecordFull() is the default mode for encoding structured records using a complete vtable. It supports introspection, field sorting, hotfield bitmap, and optional padding or schema inclusion—all driven by header flags.

---

### 🧱 Structure Overview

```plaintext
> +------------+-------------+----------------+
> |  Header    |  VTable     |   Payload      |
> +------------+-------------+----------------+
> |  40B       |  N × 8B     | Encoded Fields |
> +------------+-------------+----------------+
```

---

### 🔑 Header Fields

| Field          | Size | Notes                               |
| -------------- | ---- | ----------------------------------- |
| `Magic`        | 4B   | Constant: 0x44424633 ("DBF3")       |
| `Version`      | 2B   | Current: 0x0001                     |
| `Flags`        | 2B   | Mode, padding, schema control       |
| `SchemaID`     | 8B   | Optional; omitted if FlagNoSchemaID |
| `HotBitmap`    | 1B   | Bits 1–8 → hotfield presence        |
| `VTableSlots`  | 1B   | Number of slots (fields)            |
| `DataOffset`   | 2B   | Start of payload                    |
| `VTableOffset` | 4B   | Start of VTable                     |
| `Reserved`     | 16B  | Future extension                    |

> Header total: 40B or 32B (if schema ID omitted)

---

### 🧩 VTable Format

Each slot is 8 bytes:

`Slot = [Tag:2B][CompFlags:2B][Offset:4B]
`

- Tag: Field ID
- CompFlags: Compression or array flag
- Offset: Position from payload start

---

### 🧬 Hexdump Example

For record with 3 fields: tag 1, tag 2, tag 192

```bash
Header       : 44 42 46 33 01 00 01 00 ... → Magic, Version, Flags
VTable Slot1 : 01 00 80 00 00 00 00 00 → Tag=1, Flags=0x8000, Offset=0
VTable Slot2 : 02 00 80 00 1A 00 00 00 → Tag=2, Flags=0x8000,Offset=26
VTable Slot3 : C0 00 00 00 32 00 00 00 → Tag=192, Offset=50
Payload      : [0x00–0x19] = "Hello I'm Test 1"
               [0x1A–0x31] = "Hello I'm Test 2"
               [0x32–0x35] = 300 as uint32
```

---

### 🔍 Padding & Size Comparison

| Config                  | Header Size | VTable | Payload | Total |
| ----------------------- | ----------- | ------ | ------- | ----- |
| `No Schema, No Padding` | 32B         | 24B    | ~40B    | ~96B  |
| `Schema, No Padding`    | 40B         | 24B    | ~40B    | ~104B |
| `Schema + Padding`      | 40B         | 24B    | ~56B    | ~120B |
| `Heavy Payload`         | 40B         | 88B    | ~128B   | ~256B |

---

### 🧪 Sample Code (Encoding)

```go
enc := &dbflat.Encoder{headerflag: dbflat.FlagPadding}
record, err := enc.EncodeRecordFull(0xDEADBEEF, []uint16{1,2}, fields)
```

---

## 📄 DecodeRecordFull Documentation

DecodeRecordFull() walks through the vtable and payload using the schema or fixed-width guesses.

### 🧠 Decode Strategy

```mermaid
ParseHeader(buf) → h
Loop slots:
    tag, compFlags, offset := slot[i]
    ptr := h.DataOffset + offset
    Decode payload based on flags
```

---

### 🔍 Field Width Table (Fallback)

Used when schema isn’t present:

| CompFlag Range | Type    | Width |
| -------------- | ------- | ----- |
| ≤ 15           | bool    | 1B    |
| 16–31          | int8    | 1B    |
| 32–63          | uint8   | 1B    |
| 64–127         | int16   | 2B    |
| 128–191        | uint16  | 2B    |
| 192–255        | int32   | 4B    |
| 256-319        | uint32  | 4B    |
| 320-383        | int64   | 8B    |
| 384-447        | uint64  | 8B    |
| 448-511        | float32 | 4B    |
| 512-575        | float64 | 8B    |

---

### 🧪 Sample Code (Decoding)

```go
dec := dbflat.NewDecoder()
fields, err := dec.DecodeRecord(rawBuf, nil)
fmt.Println(string(fields[1])) // → "Hello I'm Test 1"
```

---

### ⚡ Performance Notes

- 0 alloc decoding
- Hotfield reads in O(1)
- Cold fields decoded via vtable loop
- Flags control padding, schema presence, and layout
