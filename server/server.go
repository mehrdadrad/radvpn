package server

import (
	"context"
	"errors"
	"log"
	"net"
	"sync"
	"syscall"
	"time"

	"github.com/mehrdadrad/radvpn/crypto"
	"github.com/mehrdadrad/radvpn/router"

	"golang.org/x/sys/unix"

	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
)

const maxBufsize = 1518

// Server represents vpn server
type Server struct {
	Cipher      crypto.Cipher
	KeepAlive   time.Duration
	Router      router.Gateway
	Compression bool
	Insecure    bool

	maxWorkers int

	read  chan []byte
	write chan []byte
}

type tun struct {
	maxWorkers int

	read  chan []byte
	write chan []byte
}

type header struct {
	version int
	src     net.IP
	dst     net.IP
}

// Run stars workers
func (s Server) Run(ctx context.Context, maxTunWorkers, maxNetWorkers int) {
	bp := &sync.Pool{
		New: func() interface{} {
			return make([]byte, maxBufsize)
		},
	}

	s.Router.Table().Dump()

	s.maxWorkers = maxNetWorkers

	s.read = make(chan []byte, 1000)
	s.write = make(chan []byte, 1000)

	t := new(tun)
	t.maxWorkers = maxTunWorkers

	t.read = make(chan []byte, 1000)
	t.write = make(chan []byte, 1000)

	go t.run(ctx, bp)
	go s.run(ctx, bp)

	s.cross(t)

	<-ctx.Done()
}

func (s Server) run(ctx context.Context, bp *sync.Pool) {
	for i := 0; i < s.maxWorkers; i++ {
		conn, err := s.listenPacket(ctx)
		if err != nil {
			log.Fatal(err)

		}

		go s.reader(ctx, conn, bp)
		go s.writer(ctx, conn, bp)
	}
}

func (s *Server) cross(t *tun) {
	go func() {
		for {
			b := <-s.read
			if !s.Insecure {
				b = s.Cipher.Decrypt(b)
			}
			t.write <- b
		}
	}()

	go func() {
		for {
			b := <-t.read
			if !s.Insecure {
				b = s.Cipher.Encrypt(b)
			}
			s.write <- b
		}
	}()
}

func (s Server) listenPacket(ctx context.Context) (net.PacketConn, error) {
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var sockoptErr error
			err := c.Control(func(fd uintptr) {
				sockoptErr = unix.SetsockoptInt(
					int(fd),
					unix.SOL_SOCKET,
					unix.SO_REUSEPORT,
					1,
				)
			})

			if err != nil {
				return err
			}
			return sockoptErr
		},
		KeepAlive: s.KeepAlive,
	}

	return lc.ListenPacket(ctx, "udp", ":8085")
}

func (s *Server) reader(ctx context.Context, conn net.PacketConn, bp *sync.Pool) {
	for {
		b := bp.Get().([]byte)
		n, _, err := conn.ReadFrom(b)
		if err != nil {
			log.Println(err)
			continue
		}

		select {
		case s.read <- b[:n]:
		default:
		}
	}
}

func (s *Server) writer(ctx context.Context, conn net.PacketConn, bp *sync.Pool) {
	for {
		b := <-s.write

		h, err := parseHeader(b)
		if err != nil {
			log.Println(err)
			continue
		}

		nexthop := s.Router.Table().Get(h.dst)
		rAddr, _ := net.ResolveUDPAddr("udp", nexthop.String()+":8085")

		_, err = conn.WriteTo(b, rAddr)
		if err != nil {
			log.Println(err)
		}
	}
}

// run stars workers to read/write from tunnel
func (t tun) run(ctx context.Context, bp *sync.Pool) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i := 0; i < t.maxWorkers; i++ {
		ifce, err := createTunInterface()
		if err != nil {
			log.Fatal(err)
		}
		go t.reader(ctx, ifce, bp)
		go t.writer(ctx, ifce, bp)
	}

	<-ctx.Done()
}

// reader reads from tun interface
func (t *tun) reader(ctx context.Context, ifce *water.Interface, bp *sync.Pool) {
	for {
		b := bp.Get().([]byte)
		n, err := ifce.Read(b)
		if err != nil {
			log.Println(err)
		}

		select {
		case t.read <- b[:n]:
		default:
		}
	}
}

// writer writes to tun interface
func (t *tun) writer(ctx context.Context, ifce *water.Interface, bp *sync.Pool) {
	//ifce := <- t.ifces

	for {
		b := <-t.write
		_, err := ifce.Write(b)
		if err != nil {
			log.Println(err)
		}
	}
}

// SetupTunInterface creates and sets tun interface attributes
func SetupTunInterface(ipaddrs []string, mtu int) error {

	ifname := "radvpn"
	config := water.Config{
		DeviceType: water.TUN,
		PlatformSpecificParams: water.PlatformSpecificParams{
			Name:       ifname,
			MultiQueue: true,
		},
	}

	_, err := water.New(config)
	if err != nil {
		return err
	}

	ifce, _ := netlink.LinkByName(ifname)
	netlink.LinkSetMTU(ifce, mtu)
	netlink.LinkSetTxQLen(ifce, 1000)
	netlink.LinkSetUp(ifce)

	for _, ipnet := range ipaddrs {
		addr, err := netlink.ParseAddr(ipnet)
		if err != nil {
			return err
		}
		err = netlink.AddrAdd(ifce, addr)
		if err != nil {
			return err
		}
	}

	return nil
}

// createTunInterface creates a cloned tun interface
func createTunInterface() (*water.Interface, error) {
	ifname := "radvpn"
	config := water.Config{
		DeviceType: water.TUN,
		PlatformSpecificParams: water.PlatformSpecificParams{
			Name:       ifname,
			MultiQueue: true,
		},
	}

	return water.New(config)
}

func parseHeader(b []byte) (*header, error) {
	if len(b) < net.IPv4len {
		return nil, errors.New("small packet")
	}

	h := new(header)

	h.version = int(b[0] >> 4)

	if h.version == 4 {
		h.src = make(net.IP, net.IPv4len)
		copy(h.src, b[12:16])
		h.dst = make(net.IP, net.IPv4len)
		copy(h.dst, b[16:20])

		return h, nil
	}

	h.src = make(net.IP, net.IPv6len)
	copy(h.src, b[8:24])
	h.dst = make(net.IP, net.IPv6len)
	copy(h.dst, b[24:40])

	return h, nil
}
