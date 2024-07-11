package fec

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"udpacc/fec/header"
	"udpacc/mempool"

	"github.com/klauspost/reedsolomon"
)

func init() {

}

type FecCodec struct {
	net.PacketConn
	rs reedsolomon.Encoder

	dataShards   int
	parityShards int

	seq atomic.Uint32

	rChan chan addrPacket
	wChan chan []byte

	rawData chan addrPacket

	packetswChan chan [][]byte
	rtimeout     chan struct{}
	wtimeout     chan struct{}
	quit         chan struct{}

	doneSeq [65535]bool

	recvPacketsM sync.RWMutex
	recvPackets  map[uint16]packetGroup
}

type packetGroup struct {
	m           *sync.Mutex
	serverAddr  net.Addr
	seq         uint16
	recvData    int
	recvParity  int
	recvPadding int
	packets     [][]byte
	notRecv     []bool
}

type addrPacket struct {
	data []byte
	addr net.Addr
}

var _ net.PacketConn = (*FecCodec)(nil)

func NewFec(dataShards, parityShards int, pc net.PacketConn) *FecCodec {

	rs, err := reedsolomon.New(dataShards, parityShards)
	if err != nil {
		panic(err)
	}
	fec := &FecCodec{
		PacketConn:   pc,
		rChan:        make(chan addrPacket, 1024),
		wChan:        make(chan []byte, 1024),
		rs:           rs,
		dataShards:   dataShards,
		parityShards: parityShards,
		rawData:      make(chan addrPacket, 1024),
		packetswChan: make(chan [][]byte, 1024),
		rtimeout:     make(chan struct{}),
		quit:         make(chan struct{}),
		recvPackets:  make(map[uint16]packetGroup),
	}
	go fec.encodeloop()
	go fec.processPacketGroup()
	go fec.recvloop()
	go fec.dispacherLoop()
	return fec
}

func (f *FecCodec) recvloop() {
	pkt := make([]byte, 65535)
	for {
		n, addr, err := f.PacketConn.ReadFrom(pkt)
		if err != nil {
			return
		}
		f.rawData <- addrPacket{
			data: pkt[:n],
			addr: addr,
		}
	}
}

func (f *FecCodec) dispacherLoop() {
	for pkt := range f.rawData {
		if header.Type(pkt.data) != header.Data && header.Type(pkt.data) != header.Parity && header.Type(pkt.data) != header.Padding {
			continue
		}
		if header.Type(pkt.data) == header.Data {
			f.rChan <- pkt
		}
		go f.handlePkt(pkt)
	}
}

func (f *FecCodec) getPacketGroup(seq uint16) packetGroup {
	f.recvPacketsM.RLock()
	pg, ok := f.recvPackets[seq]
	f.recvPacketsM.RUnlock()
	if ok {
		return pg
	}
	pg = packetGroup{
		seq:     seq,
		packets: make([][]byte, f.dataShards+f.parityShards),
		notRecv: make([]bool, f.dataShards+f.parityShards),
	}
	for i := 0; i < f.dataShards+f.parityShards; i++ {
		pg.notRecv[i] = true
	}
	f.recvPacketsM.Lock()
	f.recvPackets[seq] = pg
	f.recvPacketsM.Unlock()
	return pg
}

func (f *FecCodec) handlePkt(pkt addrPacket) {

	seq := uint16(header.Seq(pkt.data))
	if f.doneSeq[seq] {
		return
	}

	pg := f.getPacketGroup(seq)
	pg.m.Lock()
	defer pg.m.Unlock()

	if f.doneSeq[seq] {
		return
	}

	idx := header.Idx(pkt.data)
	if !pg.notRecv[idx] {
		return
	}
	pg.packets[idx] = pkt.data
	pg.notRecv[idx] = false

	switch header.Type(pkt.data) {
	case header.Data:
		pg.recvData++
	case header.Parity:
		pg.recvParity++
	case header.Padding:
		pg.recvPadding++
	}

	if pg.recvData+pg.recvPadding+pg.recvParity >= f.dataShards && pg.recvData < f.dataShards {
		// for i := f.dataShards; i < f.dataShards+f.parityShards; i++ {
		// 	pg.notRecv[i] = false
		// }
		if err := f.rs.ReconstructSome(pg.packets, pg.notRecv); err != nil {
			return
		}
		for i := 0; i < f.dataShards; i++ {
			if pg.notRecv[i] {
				f.rChan <- addrPacket{
					data: pg.packets[i],
					addr: pkt.addr,
				}
			}
		}
		f.doneSeq[seq] = true
	}

}

func (f *FecCodec) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	select {
	case pkt := <-f.rChan:
		return copy(b, pkt.data), pkt.addr, nil
	case <-f.rtimeout:
		return 0, nil, fmt.Errorf("timeout")
	case <-f.quit:
		return 0, nil, fmt.Errorf("quit")
	}
}

func (f *FecCodec) encodeloop() {
	idx := uint8(0)
	encodeGroup := make([][]byte, f.dataShards+f.parityShards)
	for b := range f.wChan {
		encodeGroup[idx] = b
		idx++
		if idx == uint8(f.dataShards) {
			idx = 0
			f.packetswChan <- encodeGroup
			encodeGroup = make([][]byte, f.dataShards+f.parityShards)

		}
	}
}

func (f *FecCodec) processPacketGroup() {
	for pg := range f.packetswChan {
		go f.processPg(pg)
	}
}

func (f *FecCodec) processPg(pg [][]byte) {
	seq := uint16(f.seq.Add(1))
	maxLen := 0
	for i := 0; i < f.dataShards; i++ {
		maxLen = max(maxLen, len(pg[i]))
	}
	for i := 0; i < f.dataShards; i++ {
		pg[i] = header.NewPacket(header.Data, seq, uint8(i), maxLen-len(pg[i]), pg[i])
	}
	for i := f.dataShards; i < f.dataShards+f.parityShards; i++ {
		pg[i] = header.NewPacket(header.Parity, seq, uint8(i), maxLen, nil)
	}
	f.rs.Encode(pg)
	for i := 0; i < f.dataShards+f.parityShards; i++ {
		go func(idx int) {
			// f.PacketConn.WriteTo(pg[idx], f.serverAddr)
			mempool.Put(pg[idx])
		}(i)
	}
}

func (f *FecCodec) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	select {
	case f.wChan <- b:
		return len(b), nil
	case <-f.wtimeout:
		return 0, fmt.Errorf("timeout")
	}
}
