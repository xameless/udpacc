package handler

import (
	"io"
	"net"
	"udpacc/log"
	"udpacc/outbound"
	"udpacc/transport"
)

func HandleTcp(tcp transport.Tcp) {
	defer tcp.Conn.Close()
	d := &outbound.Direct{}

	remote, err := d.DialTCP(tcp.M)
	if err != nil {
		log.Errorf("failed to dial tcp: %v", err)
		return
	}
	log.Infof("[DIRECT TCP] [::]:%d <--> %s <--> %s:%d", tcp.M.SrcPort, remote.LocalAddr(), tcp.M.DstIp, tcp.M.DstPort)

	relay(tcp.Conn, remote)
}

func relay(left, right net.Conn) {
	defer left.Close()
	defer right.Close()

	go io.Copy(left, right)
	io.Copy(right, left)
}
