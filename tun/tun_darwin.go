package tun

import (
	"golang.zx2c4.com/wireguard/tun"
)

func Open(name string, mtu int) (tun.Device, error) {
	return tun.CreateTUN(name, mtu)
}
