package outbound

import (
	"net"
	"udpacc/transport"
)

type Outbound interface {
	DialUDP(transport.Metadata) (net.PacketConn, error)
	DialTCP(transport.Metadata) (net.Conn, error)
}
