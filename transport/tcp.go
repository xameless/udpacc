package transport

import (
	"net"
)

type Tcp struct {
	Conn net.Conn
	M    Metadata
}
