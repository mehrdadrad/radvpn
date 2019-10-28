package main

import (
	"flag"
	"time"

	"github.com/mehrdadrad/radvpn/udp"
	"github.com/mehrdadrad/radvpn/netdev"
	"github.com/mehrdadrad/radvpn/crypto"
)

type server interface {
	Start()
}

var localHost = flag.String("local", "10.0.1.1/24", "IP/Mask")
var remoteHost = flag.String("remote", "192.168.55.10:8085", "IP:Port")

func main() {
	var srv server
	flag.Parse()

	crp := crypto.GCM{
		Passphrase: "6368616e676520746869732070617373776f726420746f206120736563726574",
	}

	srv = &udp.UDP{
		TUNIf:      netdev.New([]string{*localHost}, 1300),
		RemoteHost: *remoteHost,
		MaxThreads: 10,
		KeepAlive: 10 * time.Second,
		Cipher: crp,
	}

	srv.Start()

	select {}
}
