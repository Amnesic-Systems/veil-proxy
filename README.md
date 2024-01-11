# nitriding-proxy

Nitriding-proxy tunnels network traffic between
[nitriding](https://github.com/Amnesic-Systems/nitriding)
and the Internet by creating a tun-based
virtual network interface that's connected to the enclave's tun interface.
A single TCP connection between nitriding-proxy and nitriding is forwarding traffic back and forth.

<div align="center">
  <img src="https://github.com/Amnesic-Systems/nitriding-proxy/assets/1316283/34a7da90-4a08-4958-8ee8-3d2a5bdf455d" alt="nitriding-proxy's architecture" width="600">
</div>

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
