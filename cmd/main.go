package main

import (
	"flag"
	"log"
	"math"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"

	proxy "github.com/Amnesic-Systems/nitriding-proxy"
	"github.com/mdlayher/vsock"
)

var l = log.New(
	os.Stderr,
	"nitriding-proxy-cmd: ",
	log.Ldate|log.Ltime|log.LUTC|log.Lshortfile,
)

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

		go proxy.VsockToTun(vm, tun, ch)
		go proxy.TunToVsock(tun, vm, ch)
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
	tun = proxy.CreateTun()
	defer tun.Close()

	if err := proxy.ToggleNAT(proxy.On); err != nil {
		l.Fatalf("Error setting up NAT: %v", err)
	}
	defer proxy.ToggleNAT(proxy.Off)

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
