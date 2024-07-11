package stack

import (
	"log"
	"net"
	"udpacc/handler"
	"udpacc/stack/link"
	"udpacc/transport"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/icmp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
	"gvisor.dev/gvisor/pkg/waiter"
)

func CreateStack(options Options) *stack.Stack {
	s := stack.New(stack.Options{
		NetworkProtocols:   []stack.NetworkProtocolFactory{ipv4.NewProtocol},
		TransportProtocols: []stack.TransportProtocolFactory{udp.NewProtocol, tcp.NewProtocol, icmp.NewProtocol4, icmp.NewProtocol6},
	})

	udpForwarder := udp.NewForwarder(s, func(r *udp.ForwarderRequest) {
		var (
			wq waiter.Queue
		)
		ep, err := r.CreateEndpoint(&wq)
		if err != nil {
			log.Fatal(err)
		}

		udpConn := gonet.NewUDPConn(&wq, ep)
		id := r.ID()

		metadata := transport.Metadata{
			SrcIp:   net.IP(id.RemoteAddress.AsSlice()),
			SrcPort: int(id.RemotePort),
			DstIp:   net.IP(id.LocalAddress.AsSlice()),
			DstPort: int(id.LocalPort),
		}

		udp := transport.Udp{
			Pc: udpConn,
			M:  metadata,
		}
		handler.HandleUdp(udp)

	})
	s.SetTransportProtocolHandler(udp.ProtocolNumber, udpForwarder.HandlePacket)

	tcpForwarder := tcp.NewForwarder(s, 0, 1024, func(r *tcp.ForwarderRequest) {
		var (
			wq  waiter.Queue
			ep  tcpip.Endpoint
			err tcpip.Error
		)
		ep, err = r.CreateEndpoint(&wq)
		if err != nil {
			r.Complete(true)
			return
		}
		defer r.Complete(false)
		conn := gonet.NewTCPConn(&wq, ep)
		id := r.ID()
		metadata := transport.Metadata{
			SrcIp:   net.IP(id.RemoteAddress.AsSlice()),
			SrcPort: int(id.RemotePort),
			DstIp:   net.IP(id.LocalAddress.AsSlice()),
			DstPort: int(id.LocalPort),
		}

		tcp := transport.Tcp{
			Conn: conn,
			M:    metadata,
		}
		handler.HandleTcp(tcp)

	})
	s.SetTransportProtocolHandler(tcp.ProtocolNumber, tcpForwarder.HandlePacket)

	linkEndpoint, err := link.CreateTunDevice(options.TunName, options.Mtu)
	if err != nil {
		panic(err)
	}

	nicId := tcpip.NICID(s.UniqueID())
	if err := s.CreateNIC(nicId, linkEndpoint); err != nil {
		log.Fatal(err)
	}

	s.SetRouteTable([]tcpip.Route{
		{
			Destination: header.IPv4EmptySubnet,
			NIC:         nicId,
		},
	})

	s.SetPromiscuousMode(nicId, true)
	s.SetSpoofing(nicId, true)
	return s
}
