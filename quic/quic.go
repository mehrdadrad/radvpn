package quic

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	_"io"
	"log"
	"math/big"
	"time"
	"net"

	quicgo "github.com/lucas-clemente/quic-go"
	"github.com/songgao/water"
)

const buffMaxSize = 1518 

type QUIC struct {
	TUNIf      *water.Interface
	RemoteHost string
}

func (q QUIC) Run() {

	// from remote to tun interface
	go func() {
		config := &quicgo.Config{
			IdleTimeout: 2 * time.Second,
			KeepAlive:   true,
		}
		buff := make([]byte, buffMaxSize)
		for {
			listener, err := quicgo.ListenAddr(":8085", generateTLSConfig(), config)
			if err != nil {
				log.Println("listen",err)
				continue
			}

			session, err := listener.Accept(context.Background())
			if err != nil {
				log.Fatal(err)
			}

			stream, err := session.AcceptStream(context.Background())
			if err != nil {
				log.Fatal(err)
			}

			for {
				n, err := stream.Read(buff)
				if err != nil {
					log.Println(err)
					break	
				}

				q.TUNIf.Write(buff[:n])
			}

			stream.Close()
			session.Close()
			listener.Close()

			log.Println("DONE from remote to tun")
		}
	}()

	// from tun interface to remote
	go func() {
		tlsConf := &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"radvpn"},
		}

		config := &quicgo.Config{
			IdleTimeout: 2 * time.Second,
			KeepAlive:   true,
		}

		rAddress, _ := net.ResolveUDPAddr("udp", q.RemoteHost)

		tunChan := make(chan []byte, 1000)

		go func(){
			buff := make([]byte, buffMaxSize)
			for {
				n, _ := q.TUNIf.Read(buff)
				tunChan <- buff[:n]
			}	
		}()

		for {
			log.Println("start from tun to if")
			conn, err := net.ListenPacket("udp", ":0")
			if err != nil {
				log.Fatal(err)
			}

			session, err := quicgo.Dial(conn, rAddress, q.RemoteHost, tlsConf, config)
			if err != nil {
				log.Println("dial", err)
				continue
			}

			stream, err := session.OpenStreamSync(context.Background())
			if err != nil {
				log.Fatal(err)
			}

			for {
				_, err = stream.Write([]byte(<-tunChan))
				if err != nil {
					log.Println("stream", err)
					break
				}
			}

			stream.Close()
			session.Close()
			conn.Close()
		}
	}()
}

func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, buffMaxSize)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		log.Fatal(err)
	}

	return &tls.Config{
		Certificates:       []tls.Certificate{tlsCert},
		NextProtos:         []string{"radvpn"},
		ClientSessionCache: nil,
	}
}
