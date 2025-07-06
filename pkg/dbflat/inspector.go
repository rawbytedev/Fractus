package dbflat

// still building not error free
type Inspector struct {
	buf     []byte            // raw dbflat message
	offsets map[uint16]uint32 // tag → offset if offset table present
	fields  map[uint16][]byte // tag → raw field data
	hotset  map[uint16]bool   // optional: hot/cold hints
}

func Inspect(buf []byte) (*Inspector, error) {
	return &Inspector{buf: buf}, nil

}
