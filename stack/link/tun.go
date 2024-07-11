package link

import (
	"context"
	"sync"
	"udpacc/tun"

	wtun "golang.zx2c4.com/wireguard/tun"

	"gvisor.dev/gvisor/pkg/buffer"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/link/channel"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

type Tun struct {
	*channel.Endpoint
	tun    wtun.Device
	rbufs  [][]byte
	rsizes []int
	wbufs  [][]byte
	offset int
	mtu    int
	once   sync.Once
	wg     sync.WaitGroup
}

var _ LinkEndpoint = (*Tun)(nil)

const offset = 4

func CreateTunDevice(name string, mtu int) (LinkEndpoint, error) {
	device, err := tun.Open(name, mtu)

	t := &Tun{
		Endpoint: channel.New(1024, uint32(mtu), ""),
		tun:      device,
		rbufs:    make([][]byte, 1),
		rsizes:   make([]int, 1),
		wbufs:    make([][]byte, 1),
		offset:   offset,
		mtu:      mtu,
	}
	return t, err
}

func (t *Tun) Attach(dispatcher stack.NetworkDispatcher) {
	t.Endpoint.Attach(dispatcher)
	t.once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		t.wg.Add(2)
		go func() {
			t.outboundLoop(ctx)
			t.wg.Done()
		}()
		go func() {
			t.dispatchLoop(cancel)
			t.wg.Done()
		}()
	})
}

func (t *Tun) Wait() {
	t.wg.Wait()
}

func (t *Tun) dispatchLoop(cancel context.CancelFunc) {
	defer cancel()

	offset, mtu := t.offset, int(t.mtu)
	data := make([]byte, offset+mtu)
	for {

		n, err := t.Read(data)
		if err != nil {
			break
		}

		if n == 0 || n > mtu {
			continue
		}

		if !t.IsAttached() {
			continue /* unattached, drop packet */
		}

		pkt := stack.NewPacketBuffer(stack.PacketBufferOptions{
			Payload: buffer.MakeWithData(data[offset : offset+n]),
		})

		switch header.IPVersion(data[offset:]) {
		case header.IPv4Version:
			t.InjectInbound(header.IPv4ProtocolNumber, pkt)
		case header.IPv6Version:
			t.InjectInbound(header.IPv6ProtocolNumber, pkt)
		}
		pkt.DecRef()
	}
}

func (t *Tun) outboundLoop(ctx context.Context) {
	for {
		pkt := t.ReadContext(ctx)
		if pkt == nil {
			break
		}
		t.writePacket(pkt)
	}
}

// writePacket writes outbound packets to the io.Writer.
func (t *Tun) writePacket(pkt *stack.PacketBuffer) tcpip.Error {
	defer pkt.DecRef()

	buf := pkt.ToBuffer()
	defer buf.Release()
	if t.offset != 0 {
		v := buffer.NewViewWithData(make([]byte, t.offset))
		_ = buf.Prepend(v)
	}

	if _, err := t.Write(buf.Flatten()); err != nil {
		return &tcpip.ErrInvalidEndpointState{}
	}
	return nil
}

func (t *Tun) Read(buf []byte) (int, error) {
	t.rbufs[0] = buf
	_, err := t.tun.Read(t.rbufs, t.rsizes, t.offset)
	if err != nil {
		return 0, err
	}
	return t.rsizes[0], nil
}

func (t *Tun) Write(buf []byte) (int, error) {
	t.wbufs[0] = buf
	t.tun.Write(t.wbufs, t.offset)
	return 0, nil
}
