package main

import (
	"flag"

	"github.com/mehrdadrad/radvpn/udp"
	"github.com/mehrdadrad/radvpn/netdev"
)

type server interface {
	Start()
}

var localHost = flag.String("local", "10.0.1.1/24", "IP/Mask")
var remoteHost = flag.String("remote", "192.168.55.10:8085", "IP:Port")

func main() {
	var srv server
	flag.Parse()

	srv = udp.UDP{
		TUNIf:      netdev.New([]string{*localHost}, 1300),
		RemoteHost: *remoteHost,
	}

	srv.Start()

	select {}
}
