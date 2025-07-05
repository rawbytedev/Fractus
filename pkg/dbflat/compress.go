package dbflat

import (
	"errors"

	"github.com/klauspost/compress/zstd"
	//"github.com/pierrec/lz4"
)

func compressData(compFlags uint16, raw []byte) ([]byte, error) {
	switch compFlags &^ ArrayMask {
	case CompRaw:
		return raw, nil
	case CompRLE:
		return rleEncode(raw), nil
	case CompHuffman:
		return huffmanEncode(raw)
	/*case CompLZ4:
	var buf bytes.Buffer
	w := lz4.NewWriter(&buf)
	if _, err := w.Write(raw); err != nil {
		return nil, err
	}
	w.Close()
	return buf.Bytes(), nil
	*/
	case CompZstd:
		bestLevel := zstd.WithEncoderLevel(zstd.SpeedBetterCompression)
		enc, err := zstd.NewWriter(nil, bestLevel)
		if err != nil {
			return nil, err
		}
		return enc.EncodeAll(raw, nil), nil
	default:
		return nil, errors.New("unknown compFlags")
	}
}

func decompressData(compFlags uint16, blob []byte, uncompressedSize int) ([]byte, error) {
	switch compFlags &^ ArrayMask {
	case CompRaw:
		return blob, nil
	case CompRLE:
		return rleDecode(blob), nil
	case CompHuffman:
		return huffmanDecode(blob)
	/*case CompLZ4:
	out := make([]byte, uncompressedSize)
	if err := lz4.UncompressBlock(blob, out); err != nil {
		return nil, err
	}
	return out, nil
	*/
	case CompZstd:
		dec, err := zstd.NewReader(nil)
		if err != nil {
			return nil, err
		}
		return dec.DecodeAll(blob, make([]byte, 0, uncompressedSize))

	default:
		return nil, errors.New("unknown compFlags")
	}
}

// simple RLE (for demo)
func rleEncode(src []byte) []byte              { return make([]byte, 10) }
func rleDecode(src []byte) []byte              { return make([]byte, 10) }
func huffmanEncode(src []byte) ([]byte, error) { return make([]byte, 10), nil }
func huffmanDecode(src []byte) ([]byte, error) { return make([]byte, 10), nil }
