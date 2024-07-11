package tun

import (
	"testing"
)

func TestTun(t *testing.T) {
	device, err := Open("utun123", 1000)
	if err != nil {
		panic(err)
	}

	for {
		// log.Printf("????")
		t.Logf("????")
		bufs := make([][]byte, 1)
		sizes := make([]int, 1)
		bufs[0] = make([]byte, 2048)
		_, err := device.Read(bufs, sizes, 4)
		if err != nil {
			panic(err)
		}
		t.Logf("read %d bytes: %s", sizes[0], bufs[0][:sizes[0]])
	}
}
