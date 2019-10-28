package udp

import (
	"context"

	"log"
	"net"
	"sync"
	"syscall"
	"time"

	"github.com/mehrdadrad/radvpn/crypto"

	"golang.org/x/sys/unix"

	"github.com/songgao/water"
)

const buffMaxSize = 1518

// UDP represents the udp server
type UDP struct {
	conn       net.PacketConn
	TunIfce    *water.Interface
	MaxThreads int
	RemoteHost string
	KeepAlive  time.Duration
	Cipher     crypto.Cipher

	bufPool sync.Pool
}

func (u *UDP) connect() error {
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

	u.conn, err = lc.ListenPacket(context.Background(), "udp", ":8085")

	return err
}

// Shutdown stops the server
func (u *UDP) Shutdown() {
	u.conn.Close()
}

// ingress gets the data
func (u *UDP) ingress() {
	for {
		b := u.bufPool.Get().([]byte)

		n, _, err := u.conn.ReadFrom(b)
		if err != nil {
			log.Println(err)
			continue
		}

		_, err = u.TunIfce.Write(u.Cipher.Decrypt(b[:n]))
		if err != nil {
			log.Println(err)
		}
		u.bufPool.Put(b)
	}
}

// egress sends out the data
func (u *UDP) egress() {
	// from tun interface to remote
	rAddress, _ := net.ResolveUDPAddr("udp", u.RemoteHost)
	for {
		b := u.bufPool.Get().([]byte)

		n, err := u.TunIfce.Read(b)
		if err != nil {
			continue
		}

		_, err = u.conn.WriteTo(u.Cipher.Encrypt(b[:n]), rAddress)
		if err != nil {
			log.Println(err)
		}
		u.bufPool.Put(b)
	}
}

// Start runs the server
func (u *UDP) Start() {
	u.bufPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, buffMaxSize)
		},
	}

	for i := 0; i < u.MaxThreads; i++ {
		go u.run()
	}

	select {}
}

func (u UDP) run() {
	if err := u.connect(); err != nil {
		log.Fatal(err)
	}

	go u.ingress()
	go u.egress()
}
