package fractus

import (
	"errors"
	"reflect"
	"unsafe"
)
var (
	ErrNotStruct    = errors.New("expected struct")
	ErrNotStructPtr = errors.New("expected pointer to struct")
	ErrUnsupported  = errors.New("unsupported type")
)
type Options struct {
	UnsafeStrings bool // zero-copy strings via unsafe; caller must ensure buf lifetime
}

type Fractus struct {
	Opts Options
	buf  []byte
	pres []byte
	body []byte
}

// Header: varint N
// Presence: (N/8) bytes
// VarOffsets: varint per variable present field
// Body: fields in declaration order

func (f *Fractus) Encode(val any) ([]byte, error) {
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, ErrNotStruct
	}

	t := v.Type()
	N := t.NumField()

	// gather fields (exported only)
	type fld struct {
		idx   int
		kind  reflect.Kind
		val   reflect.Value
		isVar bool
	}
	fields := make([]fld, 0, N)
	for i := 0; i < N; i++ {
		sf := t.Field(i)
		if sf.PkgPath != "" && !sf.Anonymous {
			continue // skip unexported
		}
		k := sf.Type.Kind()
		fields = append(fields, fld{
			idx:   i,
			kind:  k,
			val:   v.Field(i),
			isVar: !isFixedKind(k),
		})
	}

	n := len(fields)
	f.buf = make([]byte, 0, n*2+32)

	// write N
	f.buf = writeVarUint(f.buf, uint64(n))

	// presence bitmap
	f.pres = make([]byte, (n+7)/8)
	for i := range fields {
		// zero values of fixed fields as "present" (carry zeros), variable fields absent
		// For now: everything present.
		f.pres[i/8] |= 1 << (uint(i) % 8)
	}
	f.buf = append(f.buf, f.pres...)

	// collect variable field offsets while building body
	f.body = make([]byte, 0, 64)
	varOffsets := make([]int, 0, n)
	curr := 0

	for _, fi := range fields {
		if fi.isVar {
			// record offset to start
			varOffsets = append(varOffsets, curr)
			switch fi.kind {
			case reflect.String:
				s := fi.val.String()
				f.body = writeVarUint(f.body, uint64(len(s)))
				f.body = append(f.body, s...)
				curr += varintLen(uint64(len(s))) + len(s)
			case reflect.Slice:
				if fi.val.Type().Elem().Kind() == reflect.Uint8 {
					b := fi.val.Bytes()
					f.body = writeVarUint(f.body, uint64(len(b)))
					f.body = append(f.body, b...)
					curr += varintLen(uint64(len(b))) + len(b)
				} else {
					// lists: encode count then elements
					l := fi.val.Len()
					f.body = writeVarUint(f.body, uint64(l))
					curr += varintLen(uint64(l))
					for j := 0; j < l; j++ {
						elem := fi.val.Index(j)
						k := elem.Kind()
						if isFixedKind(k) {
							f.body = writeFixed(f.body, elem)
							curr += fixedSize(k)
						} else if k == reflect.String {
							s := elem.String()
							f.body = writeVarUint(f.body, uint64(len(s)))
							f.body = append(f.body, s...)
							curr += varintLen(uint64(len(s))) + len(s)
						} else if k == reflect.Slice && elem.Type().Elem().Kind() == reflect.Uint8 {
							b := elem.Bytes()
							f.body = writeVarUint(f.body, uint64(len(b)))
							f.body = append(f.body, b...)
							curr += varintLen(uint64(len(b))) + len(b)
						} else {
							return nil, ErrUnsupported
						}
					}
				}
			default:
				return nil, ErrUnsupported
			}
		} else {
			f.body = writeFixed(f.body, fi.val)
			curr += fixedSize(fi.kind)
		}
	}

	// write varOffsets as varints
	for _, off := range varOffsets {
		f.buf = writeVarUint(f.buf, uint64(off))
	}
	// append body
	f.buf = append(f.buf, f.body...)
	return f.buf, nil
}

// Decode: compute fixed offsets; read varOffsets; slice body.
// Unsafe string mode returns string without copy.
func (f *Fractus) Decode(data []byte, out any) error {
	v := reflect.ValueOf(out)
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		return ErrNotStructPtr
	}
	dst := v.Elem()
	t := dst.Type()

	// read N
	N, nHdr := readVarUint(data)
	if N == 0 {
		return nil
	}
	cursor := nHdr
	// presence
	pLen := int((N + 7) / 8)
	pres := data[cursor : cursor+pLen]
	cursor += pLen

	// build field meta
	type fld struct {
		idx   int
		kind  reflect.Kind
		isVar bool
	}
	fields := make([]fld, 0, int(N))
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.PkgPath != "" && !sf.Anonymous {
			continue
		}
		k := sf.Type.Kind()
		fields = append(fields, fld{i, k, !isFixedKind(k)})
		if len(fields) == int(N) {
			break
		}
	}

	// read varOffsets
	var varOffsets []int
	for _, fi := range fields {
		if fi.isVar {
			off, n := readVarUint(data[cursor:])
			cursor += n
			varOffsets = append(varOffsets, int(off))
		}
	}

	body := data[cursor:]
	// pass 1: compute fixed positions by walking body
	bodyPos := 0
	var varIdx int
	for _, fi := range fields {
		if !bitPresent(pres, fi.idx) {
			continue
		}
		fv := dst.Field(fi.idx)
		if fi.isVar {
			start := varOffsets[varIdx]
			varIdx++
			// read len
			lv, n := readVarUint(body[start:])
			payload := body[start+n : start+n+int(lv)]
			bodyPos += int(lv) + n
			switch fi.kind {
			case reflect.String:
				if f.Opts.UnsafeStrings {
					str := *(*string)(unsafe.Pointer(&payload))
					fv.SetString(str)
				} else {
					fv.SetString(string(payload))
				}
			case reflect.Slice:
				if fv.Type().Elem().Kind() == reflect.Uint8 {
					fv.SetBytes(payload)
				} else {
					// list decoding (simple mode)
					// re-run from start: first varint is count, then elements
					cnt, n2 := readVarUint(body[start:])
					pos := start + n2
					elemK := fv.Type().Elem().Kind()
					slice := reflect.MakeSlice(fv.Type(), int(cnt), int(cnt))
					for i := 0; i < int(cnt); i++ {
						ev := slice.Index(i)
						if isFixedKind(elemK) {
							sz := fixedSize(elemK)
							setFixed(ev, body[pos:pos+sz], elemK)
							pos += sz
						} else if elemK == reflect.String {
							ll, ln := readVarUint(body[pos:])
							pos += ln
							ev.SetString(string(body[pos : pos+int(ll)]))
							pos += int(ll)
						} else if elemK == reflect.Uint8 {
							ll, ln := readVarUint(body[pos:])
							pos += ln
							ev.SetBytes(body[pos : pos+int(ll)])
							pos += int(ll)
						} else {
							return ErrUnsupported
						}
					}
					fv.Set(slice)
				}
			}
		} else {
			sz := fixedSize(fi.kind)
			setFixed(fv, body[bodyPos:bodyPos+sz], fi.kind)
			bodyPos += sz
		}
	}
	return nil
}
