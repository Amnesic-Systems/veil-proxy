# nitriding-proxy

Nitriding-proxy tunnels network traffic between
[nitriding](https://github.com/Amnesic-Systems/nitriding)
and the Internet by creating a tun-based
virtual network interface that's connected to the enclave's tun interface.
A single TCP connection between nitriding-proxy and nitriding is forwarding traffic back and forth.

<div align="center">
  <img src="https://github.com/Amnesic-Systems/nitriding-proxy/assets/1316283/98667cb2-c85b-471a-8b34-e44c2d4a4ded" alt="nitriding-proxy's architecture" width="700">
</div>

The diagram above illustrates the architecture. The yellow components are under your control: clients, the enclave application, and the network traffic between clients and the enclave application. Nitriding-proxy tunnels your network traffic over a VSOCK-based point-to-point TCP connection between nitriding-proxy and nitriding. The diagram above shows a client making an HTTPS request to the enclave.

## Usage

Compile nitriding-proxy by running:
```
make
```
Start nitriding-proxy on the EC2 host by running:
```
sudo ./nitriding-proxy
```

## Performance

Take a look at
[this wiki page](https://github.com/Amnesic-Systems/nitriding-proxy/wiki/Performance-measurements)
to learn more about traffic throughput.
