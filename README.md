# veil-proxy

[![GoDoc](https://pkg.go.dev/badge/github.com/Amnesic-Systems/veil-proxy)](https://pkg.go.dev/github.com/Amnesic-Systems/veil-proxy)
[![Go Report Card](https://goreportcard.com/badge/github.com/Amnesic-Systems/veil-proxy)](https://goreportcard.com/report/github.com/Amnesic-Systems/veil-proxy)

veil-proxy tunnels network traffic between
[veil](https://github.com/Amnesic-Systems/veil)
and the Internet by creating a tun-based
virtual network interface that's connected to the enclave's tun interface.
A single TCP connection between veil-proxy and veil is forwarding traffic back and forth.

<div align="center">
  <img src="https://github.com/Amnesic-Systems/veil-proxy/assets/1316283/10504730-d5a1-4432-925e-b9e4bdad1478" alt="veil-proxy's architecture" width="700">
</div>

The diagram above illustrates the architecture. The yellow components are under your control: clients, the enclave application, and the network traffic between clients and the enclave application. veil-proxy tunnels your network traffic over a VSOCK-based point-to-point TCP connection between veil-proxy and veil. The diagram above shows a client making an HTTPS request to the enclave.

## Usage

Compile and run veil-proxy by running:
```
make run
```

## Performance

Take a look at
[this wiki page](https://github.com/Amnesic-Systems/veil-proxy/wiki/Performance-measurements)
to learn more about traffic throughput.
