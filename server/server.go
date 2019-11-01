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

const (
	maxBufSize  = 1518
	maxChanSize = 1000
)

// Server represents vpn server
type Server struct {
	Cipher      crypto.Cipher
	KeepAlive   time.Duration
	Router      router.Gateway
	Logger      *log.Logger
	Compression bool
	Insecure    bool

	maxWorkers int

	read  chan []byte
	write chan []byte
}

type tun struct {
	logger *log.Logger

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
			return make([]byte, maxBufSize)
		},
	}

	s.Router.Table().Dump()

	s.maxWorkers = maxNetWorkers

	s.read = make(chan []byte, maxChanSize)
	s.write = make(chan []byte, maxChanSize)

	t := &tun{
		maxWorkers: maxTunWorkers,
		logger:     s.Logger,
	}

	t.read = make(chan []byte, maxChanSize)
	t.write = make(chan []byte, maxChanSize)

	go t.run(ctx, bp)
	go s.run(ctx, bp)

	s.cross(ctx, t)

	<-ctx.Done()
}

func (s Server) run(ctx context.Context, bp *sync.Pool) {
	for i := 0; i < s.maxWorkers; i++ {
		conn, err := s.listenPacket(ctx)
		if err != nil {
			s.Logger.Fatal(err)

		}

		go s.reader(ctx, conn, bp)
		go s.writer(ctx, conn, bp)
	}
}

func (s *Server) cross(ctx context.Context, t *tun) {
	go func() {
		for {
			b := <-s.read
			if !s.Insecure {
				b = s.Cipher.Decrypt(b)
			}

			select {
			case t.write <- b:
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		for {
			b := <-t.read
			if !s.Insecure {
				b = s.Cipher.Encrypt(b)
			}

			select {
			case s.write <- b:
			case <-ctx.Done():
				return
			}
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
			s.Logger.Println(err)
			continue
		}

		select {
		case s.read <- b[:n]:
		case <-ctx.Done():
			return
		default:
		}
	}
}

func (s *Server) writer(ctx context.Context, conn net.PacketConn, bp *sync.Pool) {
	for {
		select {
		case b := <-s.write:
			h, err := parseHeader(b)
			if err != nil {
				s.Logger.Println(err)
				continue
			}

			nexthop := s.Router.Table().Get(h.dst)
			if nexthop != nil {
				rAddr, _ := net.ResolveUDPAddr("udp",
					net.JoinHostPort(nexthop.String(), "8085"))

				_, err = conn.WriteTo(b, rAddr)
				if err != nil {
					log.Println(err)
				}
			}

			b = b[:maxBufSize]
			bp.Put(b)

		case <-ctx.Done():
			return
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
			t.logger.Fatal(err)
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
			t.logger.Println(err)
		}

		select {
		case t.read <- b[:n]:
		case <-ctx.Done():
			return

		default:
		}
	}
}

// writer writes to tun interface
func (t *tun) writer(ctx context.Context, ifce *water.Interface, bp *sync.Pool) {
	var b []byte

	for {
		select {
		case b = <-t.write:
			_, err := ifce.Write(b)
			if err != nil {
				t.logger.Println(err)
			}

			b = b[:maxBufSize]
			bp.Put(b)

		case <-ctx.Done():
			return
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
