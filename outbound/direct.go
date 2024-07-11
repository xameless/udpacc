package outbound

import (
	"context"
	"fmt"
	"net"
	"syscall"
	"udpacc/socket"
	"udpacc/transport"
)

type Direct struct {
}

var _ Outbound = (*Direct)(nil)

func (d *Direct) DialUDP(transport.Metadata) (net.PacketConn, error) {
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			return socket.SetSocketOptions(network, address, "en5", c)
		},
	}

	ln, err := lc.ListenPacket(context.Background(), "udp", "")

	if err != nil {
		return nil, err
	}

	return ln, nil

}

func (d *Direct) DialTCP(m transport.Metadata) (net.Conn, error) {
	dialer := net.Dialer{
		Control: func(network, address string, c syscall.RawConn) error {
			return socket.SetSocketOptions(network, address, "en5", c)
		},
	}
	conn, err := dialer.Dial("tcp", fmt.Sprintf("%s:%d", m.DstIp.String(), m.DstPort))

	return conn, err
}
