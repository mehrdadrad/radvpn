package udp

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"
	"log"
	"net"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/songgao/water"
)

const buffMaxSize = 1518

// UDP represents the udp server
type UDP struct {
	conn       net.PacketConn
	TUNIf      *water.Interface
	MaxThreads int
	RemoteHost string

	bufPool    sync.Pool
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
	}

	u.conn, err = lc.ListenPacket(context.Background(), "udp", ":8085")

	return err
}

// Shutdown stops the server
func (u *UDP) Shutdown() {
	u.conn.Close()
}

// ingress gets the data
func (u *UDP) ingress(passphrase string) {
	for {
		b := u.bufPool.Get().([]byte)

		n, _, err := u.conn.ReadFrom(b)
		if err != nil {
			log.Println(err)
			continue
		}

		u.TUNIf.Write(decrypt(b[:n], passphrase))
		u.bufPool.Put(b)	
	}
}

// egress sends out the data
func (u *UDP) egress(passphrase string) {
	// from tun interface to remote
	rAddress, _ := net.ResolveUDPAddr("udp", u.RemoteHost)
	for {
		b := u.bufPool.Get().([]byte)

		n, err := u.TUNIf.Read(b)
		if err != nil {
			continue
		}

		u.conn.WriteTo(encrypt(b[:n], passphrase), rAddress)
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
		go u.connection()
	}

	select {}
}

func (u UDP) connection() {
	if err := u.connect(); err != nil {
		log.Fatal(err)
	}

	passphrase := "6368616e676520746869732070617373776f726420746f206120736563726574"

	go u.ingress(passphrase)
	go u.egress(passphrase)

}

func encrypt(plainData []byte, passphrase string) []byte {
	key, _ := hex.DecodeString(passphrase)
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatal(err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		log.Fatal(err)
	}

	return gcm.Seal(nonce, nonce, plainData, nil)
}

func decrypt(cipherData []byte, passphrase string) []byte {
	key, _ := hex.DecodeString(passphrase)
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatal(err)
	}

	nonceSize := gcm.NonceSize()
	nonce, cipherData := cipherData[:nonceSize], cipherData[nonceSize:]
	plainData, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		log.Fatal(err)
	}

	return plainData
}
