package proxy

import (
	"fmt"
	"net"
	"os"
	"unsafe"

	"github.com/milosgajdos/tenus"
	"golang.org/x/sys/unix"
)

const (
	isEnclave = iota
	isProxy
)

type ifReq struct {
	Name  [0x10]byte
	Flags uint16
	pad   [0x28 - 0x10 - 2]byte
}

// SetupTunAsProxy configures the tun device to be used as a proxy.
func SetupTunAsProxy() error {
	return configureTun(isProxy)
}

// SetupTunAsEnclave configures the tun device to be used as an enclave.
func SetupTunAsEnclave() error {
	return configureTun(isEnclave)
}

// setupTun configures our tun device. The function assigns an IP address and
// sets the link MTU, after which the device is ready to work.
func setupTun(typ int) error {
	cidrStr := "10.0.0.1/24"
	if typ == isEnclave {
		cidrStr = "10.0.0.2/24"
	}

	link, err := tenus.NewLinkFrom(TunName)
	if err != nil {
		return fmt.Errorf("failed to retrieve link: %w", err)
	}
	cidr, network, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return fmt.Errorf("failed to parse CIDR: %w", err)
	}
	if err = link.SetLinkIp(cidr, network); err != nil {
		return fmt.Errorf("failed to set link address: %w", err)
	}
	if err := link.SetLinkMTU(TunMTU); err != nil {
		return fmt.Errorf("failed to set link MTU: %w", err)
	}
	// Set the enclave's default gateway to the proxy's IP address.
	if typ == isEnclave {
		gw := net.ParseIP("10.0.0.1")
		if err := link.SetLinkDefaultGw(&gw); err != nil {
			return fmt.Errorf("failed to set default gateway: %w", err)
		}
	}
	if err := link.SetLinkUp(); err != nil {
		return fmt.Errorf("failed to bring up link: %w", err)
	}

	return nil
}

// CreateTun returns a ready-to-use file descriptor for our tun interface.
func CreateTun() (*os.File, error) {
	tunfd, err := unix.Open("/dev/net/tun", os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	var ifr ifReq
	ifr := ifReq{
		Flags: unix.IFF_TUN | unix.IFF_NO_PI,
	}
	copy(ifr.Name[:], TunName)

	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(tunfd),
		uintptr(unix.TUNSETIFF),
		uintptr(unsafe.Pointer(&ifr)),
	)
	if errno != 0 {
		return nil, errno
	}
	unix.SetNonblock(tunfd, true)

	return os.NewFile(uintptr(tunfd), "/dev/net/tun"), nil
}
