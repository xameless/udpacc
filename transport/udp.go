package transport

import "net"

type Udp struct {
	Pc net.PacketConn
	M  Metadata
}
