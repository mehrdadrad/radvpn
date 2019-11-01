package netdev

import (
	"log"

	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
)

type netdev struct {
	ifce    *water.Interface
	ipaddrs []string
	mtu     int
}

func (n *netdev) create() error {
	var err error

	config := water.Config{
		DeviceType: water.TUN,
		PlatformSpecificParams: water.PlatformSpecificParams{
			Name: "radvpn",
			MultiQueue: true,
		},
	}

	n.ifce, err = water.New(config)
	if err != nil {
		return err
	}

	return nil
}

func (n netdev) setip() error {
	ifce, _ := netlink.LinkByName(n.ifce.Name())
	for _, ipnet := range n.ipaddrs {
		addr, err := netlink.ParseAddr(ipnet)
		if err != nil {
			return err
		}
		err = netlink.AddrAdd(ifce, addr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n netdev) setmtu() error {
	ifce, _ := netlink.LinkByName(n.ifce.Name())
	return netlink.LinkSetMTU(ifce, n.mtu)
}

func (n netdev) setup() error {
	ifce, _ := netlink.LinkByName(n.ifce.Name())
	return netlink.LinkSetUp(ifce)
}

// New constructs a new netdev / tun interface
// it sets ip addresses, mtu and turns it up
func New(ipaddrs []string, mtu int) *water.Interface {
	var err error
	nd := &netdev{
		ipaddrs: ipaddrs,
		mtu: mtu,
	}

	err = nd.create()
	if err != nil {
		log.Fatal(err)
	}

	err = nd.setip()
	if err != nil {
		log.Fatal(err)
	}

	err = nd.setmtu()
	if err != nil {
		log.Fatal(err)
	}

	err = nd.setup()
	if err != nil {
		log.Fatal(err)
	}

	return nd.ifce
}

func GetTunIfceHandler() (*water.Interface, error) {
	config := water.Config{
		DeviceType: water.TUN,
		PlatformSpecificParams: water.PlatformSpecificParams{
			Name: "radvpn",
			MultiQueue: true,
		},
	}

	return water.New(config)
}
