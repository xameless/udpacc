package socket

import (
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

func SetSocketOptions(network, address, nic string, c syscall.RawConn) (err error) {

	var innerErr error
	err = c.Control(func(fd uintptr) {
		host, _, _ := net.SplitHostPort(address)

		if ip := net.ParseIP(host); ip != nil && !ip.IsGlobalUnicast() {
			return
		}
		idx := 0
		if nic != "" {
			if iface, err := net.InterfaceByName(nic); err == nil {
				idx = iface.Index
			}
		}
		if idx != 0 {
			switch network {
			case "tcp4", "udp4":
				innerErr = unix.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_BOUND_IF, idx)
			case "tcp6", "udp6":
				innerErr = unix.SetsockoptInt(int(fd), syscall.IPPROTO_IPV6, syscall.IPV6_BOUND_IF, idx)
			}
			if innerErr != nil {
				return
			}
		}
	})

	if innerErr != nil {
		err = innerErr
	}
	return
}
