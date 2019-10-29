package udp

import (
	"context"

	"log"
	"net"
	"sync"
	"syscall"
	"time"
	"errors"

	"github.com/mehrdadrad/radvpn/crypto"
	"github.com/mehrdadrad/radvpn/router"

	"golang.org/x/sys/unix"

	"github.com/songgao/water"
)

const buffMaxSize = 1518

// header represents ip v4/v6 header
type header struct {
	version int
	src net.IP
	dst net.IP
}

// UDP represents the udp server
type UDP struct {
	conn        net.PacketConn
	TunIfce     *water.Interface
	MaxThreads  int
	RemoteHosts []string
	KeepAlive   time.Duration
	Cipher      crypto.Cipher
	Router      *router.Router
}

func (u *UDP) connect(ctx context.Context) error {
	var err error

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
		KeepAlive: u.KeepAlive,
	}

	u.conn, err = lc.ListenPacket(ctx, "udp", ":8085")

	return err
}

// Shutdown stops the server
func (u *UDP) Shutdown() {
	u.conn.Close()
}

// ingress gets the data
func (u *UDP) ingress(bufPool *sync.Pool) {
	for {
		b := bufPool.Get().([]byte)

		n, _, err := u.conn.ReadFrom(b)
		if err != nil {
			log.Println(err)
			continue
		}

		_, err = u.TunIfce.Write(u.Cipher.Decrypt(b[:n]))
		if err != nil {
			log.Println(err)
		}
		bufPool.Put(b)
	}
}

// egress sends out the data
// from tun interface to remote
func (u *UDP) egress(bufPool *sync.Pool) {
	for {
		b := bufPool.Get().([]byte)

		n, err := u.TunIfce.Read(b)
		if err != nil {
			continue
		}

		h, _ := parseHeader(b)
		table := u.Router.Table()
		nexthop := table.Get(h.dst)
		rAddr, _ := net.ResolveUDPAddr("udp", nexthop.String() + ":8085")

		_, err = u.conn.WriteTo(u.Cipher.Encrypt(b[:n]), rAddr)
		if err != nil {
			log.Println(err)
		}
		bufPool.Put(b)
	}
}

// Start runs the server
func (u *UDP) Start(ctx context.Context) {
	bufPool := &sync.Pool{
		New: func() interface{} {
			return make([]byte, buffMaxSize)
		},
	}

	for i := 0; i < u.MaxThreads; i++ {
		go u.thread(ctx, bufPool)
	}
}

func (u UDP) thread(ctx context.Context, bufPool *sync.Pool) {
	if err := u.connect(ctx); err != nil {
		log.Fatal(err)
	}

	go u.ingress(bufPool)
	go u.egress(bufPool)
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
