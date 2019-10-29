package main

import (
	"context"
	"flag"
	"net"
	"time"

	"github.com/mehrdadrad/radvpn/crypto"
	"github.com/mehrdadrad/radvpn/netdev"
	"github.com/mehrdadrad/radvpn/router"
	"github.com/mehrdadrad/radvpn/udp"
)

type server interface {
	Start(context.Context)
}

var localHost = flag.String("local", "10.0.1.1/24", "IP/Mask")
var remoteHost = flag.String("remote", "192.168.55.10:8085", "IP:Port")

func main() {
	var srv server

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	flag.Parse()

	crp := crypto.GCM{
		Passphrase: "6368616e676520746869732070617373776f726420746f206120736563726574",
	}

	r := router.New()
	table := r.Table()

	_, dst, _ := net.ParseCIDR("10.0.1.0/24")
	nh := net.ParseIP("192.168.55.10")
	table.Add(dst, nh)

	_, dst, _ = net.ParseCIDR("10.0.2.0/24")
	nh = net.ParseIP("192.168.55.20")
	table.Add(dst, nh)


	srv = &udp.UDP{
		TunIfce:     netdev.New([]string{*localHost}, 1300),
		RemoteHosts: []string{*remoteHost},
		MaxThreads:  10,
		KeepAlive:   10 * time.Second,
		Cipher:      crp,
		Router:      r,
	}

	srv.Start(ctx)

	select {}
}
