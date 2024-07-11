package tun

import (
	wtun "golang.zx2c4.com/wireguard/tun"
)

func Open(name string, mtu int) (wtun.Device, error) {
	return wtun.CreateTUN(name, mtu)
}
