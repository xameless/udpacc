package main

import "udpacc/inbound/tun"

func main() {
	inbound := tun.NewInbound()
	inbound.Listen(tun.ListenConfig{
		TunName: "utun123",
		Mtu:     1000,
	})
	select {}
}
