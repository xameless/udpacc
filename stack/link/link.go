package link

import (
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

type LinkEndpoint interface {
	stack.LinkEndpoint
}
