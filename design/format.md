- **Header:** Varint-encoded field count N.
- **Presence bitmap:** N/8 bytes, bit i indicates if field i is present.
- **Offsets table:** Varint-encoded relative offsets for variable-size fields only; fixed-size fields are implicit.
- **Body:** Fields serialized in declaration order, fixed-size in-place, variable-size as length-prefixed blobs.

- **Random access:** Compute fixed offsets; read the varint offset for variable fields; jump directly.
- **Zero-copy:** Return slices into the body for bytes and strings (unsafe for strings) without realloc.
- **Tiny size:** No per-field tags; presence + selective offsets is very compact.

## Field classes and how to encode

Define three classes by kind:

- **Fixed-size scalars:** int8/16/32/64, uint8/16/32/64, float32/64, bool.
  - **Encoding:** Little-endian, in-place.
  - **Decoding:** Offset computed from prior fixed-size fields and present variable fields’ lengths.

- **Variable-size blobs:** string, []byte.
  - **Encoding:** Varint length + payload.
  - **Decoding:** Use variable field offset to the start of the varint, read length, slice payload.

- **Lists:** []T
  - **Preferred encoding:** Varint count C, then either:
    - **Fixed T:** C elements back-to-back.
    - **Variable T:** C varint lengths + payloads.
  - **Random access:** Optionally include a compact per-list element offset table (varints) after count to avoid scanning. If you must stay minimal, scanning within a list is acceptable, but add an opt-in “indexed list” tag for hot paths.