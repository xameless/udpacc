package protocol

import (
	"net"
	"time"

	"github.com/klauspost/reedsolomon"
)

type Fec struct {
	pc net.PacketConn

	rs reedsolomon.Encoder
}

func NewFec(pc net.PacketConn) *Fec {
	return &Fec{pc: pc}
}
func (f *Fec) ReadFrom(b []byte) (n int, addr net.Addr, err error) {

	return f.pc.ReadFrom(b)
}

func (f *Fec) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	return f.pc.WriteTo(b, addr)
}

func (f *Fec) Close() error {
	return f.pc.Close()
}

func (f *Fec) LocalAddr() net.Addr {
	return f.pc.LocalAddr()
}

func (f *Fec) SetDeadline(t time.Time) error {
	return f.pc.SetDeadline(t)
}

func (f *Fec) SetReadDeadline(t time.Time) error {
	return f.pc.SetReadDeadline(t)
}

func (f *Fec) SetWriteDeadline(t time.Time) error {
	return f.pc.SetWriteDeadline(t)
}
