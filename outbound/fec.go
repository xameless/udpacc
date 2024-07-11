package outbound

import (
	"context"
	"net"
	"syscall"
	"udpacc/socket"
	"udpacc/transport"
	"udpacc/transport/protocol"
)

type Fec struct {
	*protocol.Fec
}

func (f *Fec) DialUDP(transport.Metadata) (net.PacketConn, error) {
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			return socket.SetSocketOptions(network, address, "en5", c)
		},
	}

	ln, err := lc.ListenPacket(context.Background(), "udp", "")
	if err != nil {
		return nil, err
	}

	return &Fec{Fec: protocol.NewFec(ln)}, nil
}

func (f *Fec) DialTCP(m transport.Metadata) (net.Conn, error) {
	return nil, nil
}
