package proxy

import (
	"bytes"
	"crypto/rand"
	"errors"
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

// buffer implements io.ReadWriteCloser.
type buffer struct {
	*bytes.Buffer
}

func (b *buffer) Close() error {
	return nil
}

func TestAToB(t *testing.T) {
	var (
		err          error
		wg           sync.WaitGroup
		ch           = make(chan error)
		conn1, conn2 = net.Pipe()
		sendBuf      = make([]byte, tunMTU*2)
		recvBuf      = &buffer{
			Buffer: new(bytes.Buffer),
		}
	)

	// We only expect to see errors containing io.EOF.
	go func() {
		for err := range ch {
			assertEq(t, errors.Is(err, io.EOF), true)
		}
	}()

	// Fill sendBuf with random data.
	_, err = rand.Read(sendBuf)
	assertEq(t, err, nil)

	wg.Add(2)
	go TunToVsock(bytes.NewReader(sendBuf), conn1, ch, &wg)
	go VsockToTun(conn2, recvBuf, ch, &wg)
	wg.Wait()

	assertEq(t, bytes.Equal(
		sendBuf,
		recvBuf.Bytes(),
	), true)
}
