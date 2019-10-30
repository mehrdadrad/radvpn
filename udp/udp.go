package udp

import (
	"context"

	"log"
	"net"
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

// MGT represents in-bound management data
type MGT struct {
	code int
	data []byte
}

// UDP represents the udp server
type UDP struct {
	conn        net.PacketConn
	TunIfce     *water.Interface
	MaxThreads  int
	RemoteHosts []string
	KeepAlive   time.Duration
	Cipher      crypto.Cipher
	Router      router.Gateway

	Local		chan *MGT
}

func (u *UDP) connect(ctx context.Context) (net.PacketConn ,error) {
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

	return lc.ListenPacket(ctx, "udp", ":8085")
}

// Shutdown stops the server
func (u *UDP) Shutdown() {
	u.conn.Close()
}

// ingress gets the data
func (u *UDP) ingress(conn net.PacketConn) {
	var (
		dec []byte
		buf = make([]byte, buffMaxSize)
	)
	for {
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			log.Println(err)
			continue
		}

		dec = u.Cipher.Decrypt(buf[:n])
		//h, _ := parseHeader(dec)
		//log.Println(h.dst)

		_, err = u.TunIfce.Write(dec)
		if err != nil {
			log.Println(err)
		}
	}
}

// egress sends out the data
// from tun interface to remote
func (u *UDP) egress(conn net.PacketConn) {
	var (
		buf = make([]byte, buffMaxSize)
		nexthop net.IP
		rAddr *net.UDPAddr
		err error
		h *header
		n int
	)

	for {
		n, err = u.TunIfce.Read(buf)
		if err != nil {
			continue
		}

		h, err = parseHeader(buf)
		if err != nil {
			log.Println(err)
			continue
		}

		nexthop = u.Router.Table().Get(h.dst)
		rAddr, _ = net.ResolveUDPAddr("udp", nexthop.String() + ":8085")

		_, err = conn.WriteTo(u.Cipher.Encrypt(buf[:n]), rAddr)
		if err != nil {
			log.Println(err)
		}
	}
}

// Start runs the server
func (u *UDP) Start(ctx context.Context) {
	for i := 0; i < u.MaxThreads; i++ {
		go u.thread(ctx)
	}
}

func (u UDP) thread(ctx context.Context) {
	conn, err := u.connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	go u.ingress(conn)
	go u.egress(conn)

	select{}
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
