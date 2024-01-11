package proxy

import (
	"fmt"
	"net"
	"os"
	"unsafe"

	"github.com/milosgajdos/tenus"
	"golang.org/x/sys/unix"
)

type ifReq struct {
	Name  [0x10]byte
	Flags uint16
	pad   [0x28 - 0x10 - 2]byte
}

// configureTun configures our tun device.  The function assigns an IP address
// and sets the link MTU, after which the device is ready to work.
func configureTun() error {
	link, err := tenus.NewLinkFrom(TunName)
	if err != nil {
		return fmt.Errorf("failed to retrieve link: %w", err)
	}

	addr, network, err := net.ParseCIDR("10.0.0.1/24")
	if err != nil {
		return fmt.Errorf("failed to parse CIDR: %w", err)
	}
	if err = link.SetLinkIp(addr, network); err != nil {
		return fmt.Errorf("failed to set link address: %w", err)
	}
	if err := link.SetLinkMTU(TunMTU); err != nil {
		return fmt.Errorf("failed to set link MTU: %w", err)
	}
	if err := link.SetLinkUp(); err != nil {
		return fmt.Errorf("failed to bring up link: %w", err)
	}

	return nil
}

// openTun returns a ready-to-use file descriptor for our tun interface.  This
// code was taken in part from:
// https://github.com/golang/go/issues/30426#issuecomment-470335255
func openTun() (*os.File, error) {
	tunfd, err := unix.Open("/dev/net/tun", os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	var ifr ifReq
	// We want a tun interface and we want multiqueue support, i.e., we
	// want multiple file descriptors to parallelize packet processing.
	//
	// Note that packet information is enabled, i.e., the tun driver is
	// going to prepend the network protocol number.  It is crucial that
	// packet information is enabled for both this proxy *and* the
	// cooperating proxy inside the enclave.  A mismatch is going to break
	// the communication channel.
	ifr.Flags = unix.IFF_TUN | unix.IFF_NO_PI
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

// CreateTun creates a new tun device and returns its file descriptor.
func CreateTun() *os.File {
	tun, err := openTun()
	if err != nil {
		l.Fatalf("Error opening tun device: %v", err)
	}
	l.Println("Opened tun file descriptor.")

	if err := configureTun(); err != nil {
		l.Fatalf("Error configuring tun device: %v", err)
	}
	l.Println("Configured tun device.")

	return tun
}
