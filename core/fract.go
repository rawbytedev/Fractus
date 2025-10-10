package core

import "errors"

var (
	ErrInvalidElem = errors.New("invalid Element")
)

// core struct
type Fractus struct {
}

func (f *Fractus) Encode(val interface{}) ([]byte, error) {
	return []byte{}, nil
}
func (f *Fractus) Decode(data []byte, val interface{}) error {
	return nil
}
func NewFractus() IFractus {
	return &Fractus{}
}
func main() {
	NewFractus()
}
