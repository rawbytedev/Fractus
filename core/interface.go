package core

// Main interface
type IFractus interface {
	Encode(interface{}) ([]byte, error)
	Decode([]byte, interface{}) error
}
