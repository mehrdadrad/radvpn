package udp

import(
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net"
	"io"

	"github.com/songgao/water"
)

const buffMaxSize = 1518 

type UDP struct {
	conn	   net.PacketConn
	TUNIf	   *water.Interface
	RemoteHost string
}

func (u *UDP) connect() error {
	var err error
	u.conn, err = net.ListenPacket("udp", ":8085")

	return err
}

func (u UDP) Shutdown() {
	u.conn.Close()
}

func (u UDP) Run() {
	if err := u.connect(); err != nil {
		log.Fatal(err)
	}

	passphrase := "6368616e676520746869732070617373776f726420746f206120736563726574"

	// from remote to tun interface
	go func() {
		buff := make([]byte, buffMaxSize)	
		for {
			n, _, err := u.conn.ReadFrom(buff)	
			if err != nil {
				log.Println(err)
				continue
			}

			u.TUNIf.Write(decrypt(buff[:n], passphrase))
		}
	}()

	// from tun interface to remote
	go func() {
		buff := make([]byte, buffMaxSize)
		rAddress, _ := net.ResolveUDPAddr("udp", u.RemoteHost)
		for {
			n, err := u.TUNIf.Read(buff)
			if err != nil {
				continue
			}

			u.conn.WriteTo(encrypt(buff[:n], passphrase), rAddress)
		}
	}()
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

func decrypt(cipherData []byte, passphrase string) []byte{
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