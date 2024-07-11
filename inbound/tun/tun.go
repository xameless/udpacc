package tun

import (
	"udpacc/inbound"
	"udpacc/stack"

	gstack "gvisor.dev/gvisor/pkg/tcpip/stack"
)

type tun struct {
	s *gstack.Stack
}

type ListenConfig struct {
	TunName string
	Mtu     int
}

func (ListenConfig) Config() {}

func NewInbound() inbound.Inbound {
	return &tun{}
}

func (t *tun) Listen(lc inbound.ListenConfig) {
	c, ok := lc.(ListenConfig)
	if !ok {
		panic("invalid config")
	}
	_ = c
	s := stack.CreateStack(stack.Options{
		TunName: c.TunName,
		Mtu:     c.Mtu,
	})
	t.s = s
}

func (t *tun) Close() {
	t.s.Close()
}
