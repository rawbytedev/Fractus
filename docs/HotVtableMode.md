# EncodeRecordHot() & DecodeRecordHot()

## EncodeRecordHot()

Efficiently encodes records by isolating up to 8 hotfields into a compact vtable, followed by coldfields in tag-walk format. Enables fast lookups + streaming.

---

### Layout Overview

```go
+----------------------------+
    Header (32–40 bytes)
+----------------------------+
   Hot VTable (≤ 8 slots)
    - Tag (2B)

    - CompFlags (2B)

    - Offset (4B)
+----------------------------+
        Data Section
  ┌── Hotfield Payloads ─┐
  │                      │
  └─Coldfields (tagwalk)─┘

    - Tag (1–2B)

    - Flags (1–2B)

    - VarLen + Payload
+----------------------------+
```

- Hotfields: encoded first; aligned if FlagPadding

- Coldfields: serialized in `[tag][flags][len][data]...` format

---

### Sample

HotFields:

| Tag | Payload   | Offset |
| --- | --------- | ------ |
| 1   | "Hello A" | 0x00   |
| 2   | "Hello B" | 0x08   |

ColdFields:

| Tag | Payload         | Offset |
| --- | --------------- | ------ |
| 10  | "Extra Info X"  | 0x20+  |
| 192 | 300 uint32 (4B) | 0x2F+  |

---

### Hexdump (Annotated)

```go
Header     : 44 42 46 33 ... → Magic + Flags + Offset
HotSlot 1  : 01 00 80 00 00 00 00 00 → Tag=1, Flags=0x8000, Offset=0
HotSlot 2  : 02 00 80 00 08 00 00 00 → Tag=2, Offset=8
Payload:
00–07 : "Hello A"
08–15 : "Hello B"
16–17 : 0A 80 → Tag=10, Flags=0x80
18 : 0B → Length=11
19–29 : "Extra Info X"
30–31 : C0 00 → Tag=192
32–33 : 00 00 → Flags=0x0000
34–37 : 2C 01 00 00 → 300 as uint32
```

---

### Size Comparison

| Config        | Hot VTable | ColdFields | Payload Total |
| ------------- | ---------- | ---------- | ------------- |
| 2 Hot, 2 Cold | 16B        | ~30B       | ~46B          |
| 8 Hot, 5 Cold | 64B        | ~90B       | ~154B         |

With Padding  +8B/field  same  adds ~64B

---

### Encoder API Usage

```go
enc := dbflat.Encoder{}
buf, err := enc.EncodeRecordHot(schemaID, []uint16{1,2}, fields)
```

---

## DecodeRecordHot()

Partial decoder that extracts up to 8 hotfields with O(1) access using slot index math.

---

### How It Works

- Parse header → extract VTable offset + slot count
- Compute SlotOffset = VTableOff + (tag - 1) × SlotSize
- Parse compFlags + offset
- Align pointer if FlagPadding
- Decode field: decompress or raw slice

---

### Field Extraction Logic

```go
ptr := DataOffset + offset
if compressed:
    len, n := readVarUint(buf[ptr:])
    decompress(buf[ptr+n:ptr+len])
else:
    return buf[ptr:ptr+width]
```

---

### Result Map

```go
map[uint16][]byte{
    1: "Hello A",
    2: "Hello B",
    ...
}
```

⚠️**Only fields 1–8 are supported in DecodeRecordHot()!**

---

### Sample Usage

```go
dec := dbflat.Decoder{}
result, err := dec.DecodeRecordHot(raw)
fmt.Println(string(result[1])) // → "Hello A"
```

---

### Performance Characteristics

- Hotfield lookup: direct jump via offset
- Alloc-free decoding
- Compression handled transparently
- Coldfields ignored in this decoder
