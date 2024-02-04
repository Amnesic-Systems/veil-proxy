package proxy

import (
	"bytes"
	"io"
	"net"
	"sync"
	"testing"
)

func assertEq(t *testing.T, is, should interface{}) {
	t.Helper()
	if should != is {
		t.Fatalf("Expected value\n%v\nbut got\n%v", should, is)
	}
}

func TestAToB(t *testing.T) {
	var (
		wg         sync.WaitGroup
		tun, vsock = net.Pipe()
		ch         = make(chan error)
		send       = []byte("hello world")
		recv       = make([]byte, len(send))
	)

	wg.Add(2)
	go TunToVsock(tun, vsock, ch, &wg)
	go VsockToTun(vsock, tun, ch, &wg)
	defer wg.Wait()

	// Read but ignore errors.
	go func(chan error) {
		for range ch {
		}
	}(ch)

	// Echo data back to sender.
	go func(t *testing.T, expected int) {
		nw, err := io.Copy(vsock, vsock)
		assertEq(t, err, nil)
		assertEq(t, nw, int64(expected))
	}(t, len(send))

	nw, err := tun.Write(send)
	assertEq(t, nw, len(send))
	assertEq(t, err, nil)

	nr, err := tun.Read(recv)
	assertEq(t, err, nil)
	assertEq(t, nr, len(send))

	err = tun.Close()
	assertEq(t, err, nil)

	assertEq(t, bytes.Compare(send, recv), 0)
}
