package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/mdlayher/vsock"
)

const (
	LenBufSize = 2
	TunMTU     = 65535 // The maximum-allowed MTU for the tun interface.
	TunName    = "tun0"
)

var (
	l = log.New(
		os.Stderr,
		"nitriding-proxy: ",
		log.Ldate|log.Ltime|log.LUTC|log.Lshortfile,
	)
)

// TunToVsock forwards network packets from the tun device to our
// TCP-over-VSOCK connection. The function keeps on forwarding packets until we
// encounter an error or EOF. Errors (including EOF) are written to the given
// channel.
func TunToVsock(from io.Reader, to io.WriteCloser, ch chan error) {
	defer to.Close()
	var (
		err       error
		pktLenBuf = make([]byte, LenBufSize)
		pktBuf    = make([]byte, TunMTU)
	)

	for {
		// Read a network packet from the tun interface.
		nr, rerr := from.Read(pktBuf)
		if nr > 0 {
			// Forward the network packet to our TCP-over-VSOCK connection.
			binary.BigEndian.PutUint16(pktLenBuf, uint16(nr))
			if _, werr := to.Write(append(pktLenBuf, pktBuf[:nr]...)); err != nil {
				err = werr
				break
			}
		}
		if rerr != nil {
			err = rerr
			break
		}
	}
	ch <- fmt.Errorf("stopped tun-to-vsock forwarding because: %v", err)
}

// VsockToTun forwards network packets from our TCP-over-VSOCK connection to
// the tun interface. The function keeps on forwarding packets until we
// encounter an error or EOF. Errors (including EOF) are written to the given
// channel.
func VsockToTun(from io.Reader, to io.Writer, ch chan error) {
	var (
		err       error
		pktLen    uint16
		pktLenBuf = make([]byte, LenBufSize)
		pktBuf    = make([]byte, TunMTU)
	)

	for {
		// Read the length prefix that tells us the size of the subsequent
		// packet.
		if _, err = io.ReadFull(from, pktLenBuf); err != nil {
			break
		}
		pktLen = binary.BigEndian.Uint16(pktLenBuf)

		// Read the packet.
		if _, err = io.ReadFull(from, pktBuf[:pktLen]); err != nil {
			break
		}

		// Forward the packet to the tun interface.
		if _, err := to.Write(pktBuf[:pktLen]); err != nil {
			break
		}
	}
	ch <- fmt.Errorf("stopped vsock-to-tun forwarding because: %v", err)
}

func listenVSOCK(port uint32) net.Listener {
	var (
		ln  net.Listener
		cid uint32
		err error
	)

	if cid, err = vsock.ContextID(); err != nil {
		l.Fatalf("Error retrieving VSOCK context ID: %v", err)
	}

	ln, err = vsock.ListenContextID(cid, port, nil)
	if err != nil {
		l.Fatalf("Error listening on VSOCK port: %v", err)
	}
	l.Printf("Listening on VSOCK iface at %s.", ln.Addr())

	return ln
}

func acceptLoop(ln net.Listener, tun *os.File) {
	ch := make(chan error)
	defer close(ch)

	// Print errors that occur while forwarding packets.
	go func(ch chan error) {
		for err := range ch {
			l.Print(err)
		}
	}(ch)

	// Listen for connections from the enclave and begin forwarding packets
	// once a new connection is established. At any given point, we only expect
	// to have a single TCP-over-VSOCK connection with the enclave
	for {
		l.Println("Waiting for new connection from enclave.")
		vm, err := ln.Accept()
		if err != nil {
			l.Printf("Failed to accept connection: %v", err)
			break
		}
		l.Printf("Accepted new connection from %s.", vm.RemoteAddr())

		go VsockToTun(vm, tun, ch)
		go TunToVsock(tun, vm, ch)
	}
}

func main() {
	var (
		profile   bool
		vsockPort uint
		ln        net.Listener
		tun       *os.File
	)

	flag.BoolVar(&profile, "profile", false, "Enable profiling.")
	flag.UintVar(&vsockPort, "port", 1024, "VSOCK forwarding port that the enclave connects to.")
	flag.Parse()

	if vsockPort < 1 || vsockPort > math.MaxUint32 {
		l.Fatalf("Flag -port must be in interval [1, %d].", math.MaxUint32)
	}

	// Set up a VSOCK listener for the "right" side of the proxy, i.e., the
	// side facing the enclave.
	ln = listenVSOCK(uint32(vsockPort))
	defer ln.Close()

	// Set up a file descriptor for the "left" side of the proxy, i.e., a tun
	// device that handles the enclave's traffic.
	tun = createTun()
	defer tun.Close()
	defer toggleNAT(off)

	// If desired: set up a Web server for the profiler.
	if profile {
		go func() {
			const hostPort = "localhost:6060"
			l.Printf("Starting profiling Web server at: http://%s", hostPort)
			http.ListenAndServe(hostPort, nil)
		}()
	}

	acceptLoop(ln, tun)
}
