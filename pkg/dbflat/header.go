package dbflat

const (
	MagicV1   = 0x44424633 // "DBF3"
	VersionV1 = 1

	// compFlags & 0x7F == compressor ID
	CompressionMask = 0x000F
	CompRaw         = 0x0000
	CompRLE         = 0x0001
	CompHuffman     = 0x0002
	CompLZ4         = 0x0003
	CompZstd        = 0x0004
	ArrayMask       = 0x8000 // MSB signals variable-length
	HeaderSize      = 40
	SlotSize        = 8
)

type FieldType int

const (
	TypeBool FieldType = iota
	TypeInt8
	TypeUint8
	TypeInt16
	TypeUint16
	TypeInt32
	TypeUint32
	TypeInt64
	TypeUint64
	TypeFloat32
	TypeFloat64
	TypeString
	TypeBytes
)

type Dbflat struct {
	e *Encoder
	d *Decoder
}

type Header struct {
	Magic       uint32   // 4B
	Version     uint16   // 2B
	Flags       uint16   // 2B: bit0=align8,bit1=schemaID,bit2=index
	SchemaID    uint64   // 8B
	HotBitmap   byte     // 1B: presence map for tags 1–8
	VTableSlots byte     // number of slot in VTable
	DataOffset  uint16   // offset to start of Data section (from header start) 2B
	VTableOff   uint32   // offset to start VTable (from header start) 4B
	_           [16]byte // reserved for upgrade
}

// VTableSlot is 8B
type VTableSlot struct {
	Tag        uint16 // 2B
	DataOffset uint16 // 2B
	CompFlags  uint32 // 4B
}

type FieldValue struct {
	Tag       uint16 // 2B
	CompFlags uint16 // 2B
	Payload   []byte // raw or already-compressed bytes
}
