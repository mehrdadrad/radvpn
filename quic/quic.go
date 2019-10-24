package quic

import(
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"context"
	"log"
	"io"

	quicgo "github.com/lucas-clemente/quic-go"
	"github.com/songgao/water"
)


type QUIC struct {
	TUNIf	   *water.Interface
	RemoteHost string
}

func (q QUIC) Run() {

	// from remote to tun interface
	go func(){
		config := &quicgo.Config{
			KeepAlive: true,
		}
		listener, err := quicgo.ListenAddr(":8085", generateTLSConfig(), config)
		if err != nil {
			log.Fatal(err)
		}

		session, err := listener.Accept(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		stream, err := session.AcceptStream(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		io.Copy(q.TUNIf, stream)
	}()

	// from tun interface to remote
	go func(){
		buff := make([]byte, 1024)
		tlsConf := &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"radvpn"},
		}

		session, err := quicgo.DialAddr(q.RemoteHost, tlsConf, nil)
		if err != nil {
			log.Fatal(err)
		}

		stream, err := session.OpenStreamSync(context.Background())
		if err != nil {
			log.Fatal(err)	
		}

		for{ 
			n, err := q.TUNIf.Read(buff)
			if err != nil {
				continue	
			}

			stream.Write([]byte(buff[:n]))
		}	
	}()
}

func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type: "RSA PRIVATE KEY", 
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		log.Fatal(err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"radvpn"},
		ClientSessionCache: nil,
	}
}