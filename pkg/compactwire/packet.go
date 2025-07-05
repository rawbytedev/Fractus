package compactwire

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// Frame magic and types
const (
	FrameMagic    uint16 = 0xC5A3
	TypeData      byte   = 0x01
	TypeHandshake byte   = 0x02
	TypeError     byte   = 0x03
	TypeControl   byte   = 0x04
)

// Flag bits for Data Frames
const (
	FlagHasOffsetTable  byte = 1 << 0
	FlagEndOfMessage    byte = 1 << 1
	FlagFragmented      byte = 1 << 2
	FlagChecksumCRC32C  byte = 1 << 3 // example
	FlagOffsetTableComp byte = 1 << 5
)

// DataFrame represent buffer used
type DataFrame struct {
	buf *bytes.Buffer
	rdr *bytes.Reader
}

// Ta
type ControlFrame struct {
	buf *bytes.Buffer
	rdr *bytes.Reader
}

// ErrorFrame reprensent buffer used
type ErrorFrame struct {
	buf *bytes.Buffer
	rdr *bytes.Reader
}

type Compactwire struct {
	d             *DataFrame
	e             *ErrorFrame
	h             *HandshakeFrame
	c             *ControlFrame
	HandskakeConf *Configs
}

func (c *Compactwire) NewFrame(Conf *Configs) {
	c.d = &DataFrame{}
	c.e = &ErrorFrame{}
	c.h = &HandshakeFrame{VersionMask: Conf.Version, MTU: Conf.MTU, TimeoutMS: Conf.TimeoutMS}
	c.c = &ControlFrame{}
	c.HandskakeConf = Conf
}

type Configs struct {
	Version   uint32
	MTU       uint16
	TimeoutMS uint32
}

// HandshakeFrame represents a client/server handshake payload.
type HandshakeFrame struct {
	VersionMask uint32
	MTU         uint16
	TimeoutMS   uint32
	AlgCodes    []byte
}

// writePreamble writes magic (2 bytes) + frame type (1 byte)
func writePreamble(buf *bytes.Buffer, frameType byte) {
	binary.Write(buf, binary.LittleEndian, FrameMagic)
	buf.WriteByte(frameType)
}

// readPreamble reads and validates magic and returns frameType
func readPreamble(buf *bytes.Reader) (byte, error) {
	var magic uint16
	if err := binary.Read(buf, binary.LittleEndian, &magic); err != nil {
		return 0, err
	}
	if magic != FrameMagic {
		return 0, errors.New("invalid frame magic")
	}
	t, err := buf.ReadByte()
	return t, err
}
