package transport

import "net"

type Metadata struct {
	SrcIp   net.IP
	SrcPort int
	DstIp   net.IP
	DstPort int
}
