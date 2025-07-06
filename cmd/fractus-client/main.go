package fractus

import wire "fractus/pkg/compactwire"

// Work in Progress
type client struct {
	wire    *wire.Compactwire
	configs *wire.Configs
}

func main() {

}

func NewWire() *wire.Compactwire {
	return &wire.Compactwire{}
}
func DefaultConfigs() *wire.Configs {
	return &wire.Configs{Version: 1, MTU: 1500, TimeoutMS: 10}
}
func NewClient() *client {
	conf := DefaultConfigs()
	pkt := NewWire()
	pkt.NewFrame(conf)
	return &client{wire: pkt, configs: conf}
}
func (c *client) CreateHanshake() {

}
