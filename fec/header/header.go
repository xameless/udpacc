package header

import (
	"encoding/binary"
	"udpacc/mempool"
)

const HeaderSize = 8

/*
+------+-----+-----+-----+-------------+---------+---------+
| TYPE | SEQ | IDX | LEN | PADDING LEN | PAYLOAD | PADDING |
+------+-----+-----+-----+-------------+---------+---------+
|  1   |  2  |  1  |  2  |      2      |   var   |   var   |
+------+-----+-----+-----+-------------+---------+---------+
*/

type PacketType byte

const (
	Data    PacketType = 0
	Parity  PacketType = 1
	Padding PacketType = 2
)

func Type(p []byte) PacketType {
	return PacketType(p[0])
}

func Seq(p []byte) int {
	return int(binary.BigEndian.Uint16(p[1:3]))
}

func Idx(p []byte) byte {
	return p[3]
}

func Len(p []byte) int {
	return int(binary.BigEndian.Uint16(p[4:6]))
}

func PaddingLen(p []byte) int {
	return int(binary.BigEndian.Uint16(p[6:8]))
}

func Payload(p []byte) []byte {
	return p[HeaderSize : len(p)-PaddingLen(p)]
}

func NewPacket(packetType PacketType, seq uint16, idx uint8, padding int, payload []byte) []byte {
	payload = append(make([]byte, HeaderSize), payload...)
	p := mempool.Get(HeaderSize + len(payload) + padding)
	copy(p[HeaderSize:], payload)
	p[0] = uint8(packetType)
	binary.BigEndian.PutUint16(p[1:3], seq)
	p[3] = idx
	binary.BigEndian.PutUint16(p[4:6], uint16(len(payload)))
	binary.BigEndian.PutUint16(p[6:8], uint16(padding))
	return p
}
