package handler

import (
	"net"
	"udpacc/log"
	"udpacc/outbound"
	"udpacc/transport"
)

func HandleUdp(udp transport.Udp) {

	direct := &outbound.Direct{}

	remote, err := direct.DialUDP(udp.M)

	if err != nil {
		log.Errorf("failed to dial udp: %v", err)
		return
	}

	log.Infof("[DIRECT UDP] [::]:%d <--> %s <--> %s:%d", udp.M.SrcPort, remote.LocalAddr(), udp.M.DstIp, udp.M.DstPort)
	relayUdp(udp.Pc, remote, udp.M)
}

func relayUdp(left, right net.PacketConn, m transport.Metadata) {
	go copyPacket(left, right, nil)
	go copyPacket(right, left, &net.UDPAddr{IP: m.DstIp, Port: m.DstPort})
}

func copyPacket(dst, src net.PacketConn, to net.Addr) {
	buf := make([]byte, 2048)
	for {
		n, _, err := src.ReadFrom(buf)
		if err != nil {
			log.Error(err)
			return
		}
		_, err = dst.WriteTo(buf[:n], to)
		if err != nil {
			log.Error(err)
			return
		}
	}
}
