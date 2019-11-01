package main

import (
	"context"
	"flag"
	"log"
	"net"
	"time"

	"github.com/mehrdadrad/radvpn/crypto"
	"github.com/mehrdadrad/radvpn/router"
	"github.com/mehrdadrad/radvpn/server"
)

var localHost = flag.String("local", "10.0.2.1/24", "IP/Mask")
var remoteHost = flag.String("remote", "192.168.55.10:8085", "IP:Port")

func main() {
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server.SetupTunInterface([]string{*localHost}, 1300)

	crp := crypto.GCM{
		Passphrase: "6368616e676520746869732070617373776f726420746f206120736563726574",
	}

	r := router.New()

	_, dst, _ := net.ParseCIDR("10.0.1.0/24")
	nexthop := net.ParseIP("192.168.55.10")
	err := r.Table().Add(dst, nexthop)
	if err != nil {
		log.Println(err)
	}

	_, dst, _ = net.ParseCIDR("10.0.2.0/24")
	nexthop = net.ParseIP("192.168.55.20")
	err = r.Table().Add(dst, nexthop)
	if err != nil {
		log.Println(err)
	}

	s := server.Server{
		KeepAlive: 10 * time.Second,
		Insecure:  true,
		Cipher:    crp,
		Router:    r,
	}

	s.Run(ctx, 10, 10)
}

/*
func main() {
	var srv server

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	flag.Parse()

	crp := crypto.GCM{
		Passphrase: "6368616e676520746869732070617373776f726420746f206120736563726574",
	}

	ifce := netdev.New([]string{*localHost}, 1300)
	//ifce1 := netdev.New([]string{*localHost}, 1300)
	//_ = ifce1
	r := router.New()

	_, dst, _ := net.ParseCIDR("10.0.1.0/24")
	nexthop := net.ParseIP("192.168.55.10")
	err := r.Table().Add(dst, nexthop)
	if err != nil {
		log.Println(err)
	}

	_, dst, _ = net.ParseCIDR("10.0.2.0/24")
	nexthop = net.ParseIP("192.168.55.20")
	err = r.Table().Add(dst, nexthop)
	if err != nil {
		log.Println(err)
	}



	srv = &udp.UDP{
		TunIfce:     ifce,
		RemoteHosts: []string{*remoteHost},
		MaxThreads:  10,
		KeepAlive:   10 * time.Second,
		Cipher:      crp,
		Router:      r,
	}

	srv.Start(ctx)
	r.Table().Dump()

	select {}
}
*/
