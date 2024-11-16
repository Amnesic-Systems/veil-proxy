package main

import (
	"flag"
	"log"
	"math"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sync"

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

func acceptLoop(ln net.Listener) {
	var (
		err error
		wg  sync.WaitGroup
		tun *os.File
		ch  = make(chan error)
	)
	defer close(ch)
	defer tun.Close()

	// Print errors that occur while forwarding packets.
	go func(ch chan error) {
		for err := range ch {
			l.Print(err)
		}
	}(ch)

	// Listen for connections from the enclave and begin forwarding packets
	// once a new connection is established. At any given point, we only expect
	// to have a single TCP-over-VSOCK connection with the enclave.
	for {
		tun, err = proxy.SetupTunAsProxy()
		if err != nil {
			l.Printf("Error creating tun device: %v", err)
			continue
		}
		l.Print("Created tun device.")

		l.Println("Waiting for new connection from enclave.")
		vm, err := ln.Accept()
		if err != nil {
			l.Printf("Error accepting connection: %v", err)
			continue
		}
		l.Printf("Accepted new connection from %s.", vm.RemoteAddr())

		wg.Add(2)
		go proxy.VsockToTun(vm, tun, ch, &wg)
		go proxy.TunToVsock(tun, vm, ch, &wg)
		wg.Wait()
	}
}

func main() {
	var (
		profile bool
		port    uint64
		ln      net.Listener
	)

	flag.BoolVar(
		&profile, "profile",
		false,
		"Enable profiling.",
	)
	flag.Uint64Var(
		&port, "port",
		proxy.DefaultPort,
		"VSOCK port that the enclave connects to.",
	)
	flag.Parse()

	if port < 1 || port > math.MaxUint32 {
		l.Fatalf("Flag -port must be in interval [1, %d].", math.MaxUint32)
	}

	l.Print("Enabling NAT.")
	if err := proxy.ToggleNAT(proxy.On); err != nil {
		l.Fatalf("Error enabling NAT: %v", err)
	}
	defer func() {
		if err := proxy.ToggleNAT(proxy.Off); err != nil {
			l.Printf("Error disabling NAT: %v", err)
		}
	}()

	ln = listenVSOCK(uint32(port))
	defer ln.Close()

	// If desired, set up a Web server for the profiler.
	if profile {
		go func() {
			const hostPort = "localhost:6060"
			l.Printf("Starting profiling Web server at: http://%s", hostPort)
			http.ListenAndServe(hostPort, nil)
		}()
	}

	acceptLoop(ln)
}
