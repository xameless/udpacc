package main

import (
	"log"
	"net"
	"time"
)

func main() {

	ln, _ := net.ListenUDP("udp", nil)

	go func() {
		buf := make([]byte, 2048)
		for {
			n, addr, err := ln.ReadFrom(buf)
			if err != nil {
				panic(err)
			}
			log.Printf("received %s from %s\n", string(buf[:n]), addr.String())
		}
	}()
	log.Printf("listen at %v", ln.LocalAddr())
	for range time.Tick(time.Second) {
		ln.WriteTo([]byte("hello"), &net.UDPAddr{IP: net.IPv4(123, 56, 16, 26), Port: 31500})
	}
}
