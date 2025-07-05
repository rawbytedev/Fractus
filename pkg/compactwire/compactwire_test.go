package compactwire

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"reflect"
	"testing"
)

// minimal helper to corrupt a single byte
func flipByte(b []byte, idx int) []byte {
	out := make([]byte, len(b))
	copy(out, b)
	out[idx] ^= 0xFF
	return out
}

func TestDataFrameRoundTripNoOffset(t *testing.T) {
	payload := []byte("hello, CW!")
	flags := FlagEndOfMessage
	d := &DataFrame{}
	frame, err := d.EncodeDataFrame(payload, flags, nil)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	gotPayload, gotOffsets, gotFlags, err := d.DecodeDataFrame(frame)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !bytes.Equal(gotPayload, payload) {
		t.Errorf("payload mismatch; want %q, got %q", payload, gotPayload)
	}
	if len(gotOffsets) != 0 {
		t.Errorf("expected no offsets, got %v", gotOffsets)
	}
	if gotFlags != flags {
		t.Errorf("flags mismatch; want %02x, got %02x", flags, gotFlags)
	}
}

func TestDataFrameRoundTripWithOffset(t *testing.T) {
	payload := []byte("ABCDEFGH")
	offsets := []uint32{0, 4}
	flags := byte(FlagHasOffsetTable | FlagEndOfMessage)
	d := &DataFrame{}
	frame, err := d.EncodeDataFrame(payload, flags, offsets)
	if err != nil {
		t.Fatalf("encode with offsets failed: %v", err)
	}
	gotPayload, gotOffsets, gotFlags, err := d.DecodeDataFrame(frame)
	if err != nil {
		t.Fatalf("decode with offsets failed: %v", err)
	}
	if !bytes.Equal(gotPayload, payload) {
		t.Errorf("Got frame: %v", frame)
		t.Errorf("payload mismatch; want %v, got %v", payload, gotPayload)
	}
	if !reflect.DeepEqual(gotOffsets, offsets) {
		t.Errorf("offsets mismatch; want %v, got %v", offsets, gotOffsets)
	}
	if gotFlags != flags {
		t.Errorf("flags mismatch; want %02x, got %02x", flags, gotFlags)
	}
}

func TestDataFrame_InvalidCRC(t *testing.T) {
	payload := []byte("badcrc")
	flags := FlagEndOfMessage
	d := &DataFrame{}
	frame, err := d.EncodeDataFrame(payload, flags, nil)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	// flip a byte in the body (before the CRC) to break the checksum
	bad := flipByte(frame, len(frame)/2)
	if _, _, _, err := d.DecodeDataFrame(bad); err == nil {
		t.Error("expected CRC error, got nil")
	}
}

func TestErrorFrame_RoundTrip(t *testing.T) {
	code := byte(0x05)
	data := []byte("FAIL")
	e := &ErrorFrame{}
	frame, err := e.EncodeErrorFrame(code, data)
	if err != nil {
		t.Fatalf("EncodeErrorFrame: %v", err)
	}

	gotCode, gotData, err := e.DecodeErrorFrame(frame)
	if err != nil {
		t.Fatalf("DecodeErrorFrame: %v", err)
	}
	if gotCode != code {
		t.Errorf("error code mismatch; want %02x, got %02x", code, gotCode)
	}
	if !bytes.Equal(gotData, data) {
		t.Errorf("error data mismatch; want %q, got %q", data, gotData)
	}
}

func TestErrorFrame_InvalidMagic(t *testing.T) {
	code := byte(0x01)
	data := []byte("X")
	e := &ErrorFrame{}
	frame, _ := e.EncodeErrorFrame(code, data)
	// corrupt magic
	frame[0] = 0x00
	if _, _, err := e.DecodeErrorFrame(frame); err == nil {
		t.Error("expected invalid magic error, got nil")
	}
}

func TestHandshakeFrame_RoundTrip(t *testing.T) {
	h0 := HandshakeFrame{
		VersionMask: 0xdeadbeef,
		MTU:         1500,
		TimeoutMS:   250,
		AlgCodes:    []byte{1, 2, 3},
	}
	h := &HandshakeFrame{}
	frame, err := h.EncodeHandshake(h0)
	if err != nil {
		t.Fatalf("EncodeHandshake: %v", err)
	}

	h1, err := h.DecodeHandshake(frame)
	if err != nil {
		t.Fatalf("DecodeHandshake: %v", err)
	}
	if h1.VersionMask != h0.VersionMask {
		t.Errorf("VersionMask mismatch; want %08x, got %08x", h0.VersionMask, h1.VersionMask)
	}
	if h1.MTU != h0.MTU {
		t.Errorf("MTU mismatch; want %d, got %d", h0.MTU, h1.MTU)
	}
	if h1.TimeoutMS != h0.TimeoutMS {
		t.Errorf("TimeoutMS mismatch; want %d, got %d", h0.TimeoutMS, h1.TimeoutMS)
	}
	if !bytes.Equal(h1.AlgCodes, h0.AlgCodes) {
		t.Errorf("AlgCodes mismatch; want %v, got %v", h0.AlgCodes, h1.AlgCodes)
	}
}

func TestHandshakeFrame_InvalidCRC(t *testing.T) {
	h0 := HandshakeFrame{
		VersionMask: 0x01020304,
		MTU:         1200,
		TimeoutMS:   100,
		AlgCodes:    []byte{0xAA},
	}
	h := &HandshakeFrame{}
	frame, _ := h.EncodeHandshake(h0)
	// break CRC
	frame[len(frame)-1] ^= 0xFF
	if _, err := h.DecodeHandshake(frame); err == nil {
		t.Error("expected CRC error on handshake, got nil")
	}
}

func TestPreamble_ReadWrite(t *testing.T) {
	buf := new(bytes.Buffer)
	writePreamble(buf, TypeControl)
	// read back
	rd := bytes.NewReader(buf.Bytes())
	ft, err := readPreamble(rd)
	if err != nil {
		t.Fatalf("readPreamble: %v", err)
	}
	if ft != TypeControl {
		t.Errorf("frame type mismatch; want %02x, got %02x", TypeControl, ft)
	}
}

// ensure length placeholder is correctly written
func TestDataFrame_LengthField(t *testing.T) {
	payload := []byte("lencheck")
	flags := byte(0)
	d := &DataFrame{}
	frame, err := d.EncodeDataFrame(payload, flags, nil)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	// length is at offset 3–6
	length := binary.LittleEndian.Uint32(frame[3:7])
	// total frame includes 2-byte magic, 1-byte type, 4-byte length field, 1-byte flags, payload, 4-byte CRC
	want := uint32(2 + 1 + 4 + 1 + len(payload) + 4)
	if length != want {
		t.Errorf("length field wrong; want %d, got %d", want, length)
	}
	// CRC should match body
	crcStart := int(length) - 4
	wantCRC := binary.LittleEndian.Uint32(frame[crcStart:])
	if wantCRC != crc32.ChecksumIEEE(frame[4:crcStart]) {
		t.Error("CRC field does not match computed checksum")
	}
}
