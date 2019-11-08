package server

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/mehrdadrad/radvpn/config"

	"github.com/vishvananda/netlink"
)

func TestInitCrypto(t *testing.T) {
	cfg := &config.Config{
		Crypto: struct {
			Type string `yaml:"type"`
			Key  string `yaml:"key"`
		}{"gcm", "mykey"},
	}

	s := &Server{
		Config: cfg,
	}

	err := s.initCrypto()
	if err != nil {
		t.Error("expected err nil but got,", err)
	}

	cfg = &config.Config{
		Crypto: struct {
			Type string `yaml:"type"`
			Key  string `yaml:"key"`
		}{"unknown", "mykey"},
	}

	s = &Server{
		Config: cfg,
	}

	err = s.initCrypto()
	if err == nil {
		t.Error("expected to have err but got nil")
	}
}

func TestListenPacket(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			Name       string `yaml:"name"`
			Address    string `yaml:"address"`
			MaxWorkers int    `yaml:"maxworkers"`
			Keepalive  int    `yaml:"keepalive"`
			Insecure   bool   `yaml:"insecure"`
			Mtu        int    `yaml:"mtu"`
		}{
			Keepalive: 5,
			Address:   ":8085",
		},
	}

	s := &Server{
		Config: cfg,
	}

	msg := "hello radvpn"
	l, err := s.listenPacket(context.Background())
	if err == nil {
		_ = l
		conn, err := net.Dial("udp", "localhost:8085")
		if err == nil {
			fmt.Fprintf(conn, msg)
		}

		buff := make([]byte, 20)
		n, _, _ := l.ReadFrom(buff)
		if string(buff[:n]) != msg {
			t.Errorf("expect to have %s but got, %s", msg, string(buff[:n]))
		}
	} else {
		t.Error(err)
	}
}

func TestDiffStrSlice(t *testing.T) {
	a := []string{"a", "c"}
	b := []string{"a", "b", "c"}

	c := diffStrSlice(a, b)
	if len(c) > 0 {
		if c[0] != "b" {
			t.Error("expected c but got, ", c[0])
		}
	} else {
		t.Error("expected having an ellemnt but nothing")
	}
}

func TestParseHeader(t *testing.T) {
	// ipv4 header
	b := []byte{
		0x45, 0x0, 0x0, 0x4b, 0x8,
		0xf8, 0x0, 0x0, 0x3e, 0x11,
		0x82, 0x91, 0xc0, 0xe5, 0xd8,
		0x8f, 0xc0, 0xe5, 0x96, 0xbe,
		0x64, 0x9b, 0x0, 0x35, 0x0,
	}

	h, err := parseHeader(b)
	if err != nil {
		t.Error("unexpected error", err)
	}

	if h.version != 4 {
		t.Error("expected version 4 but got,", h.version)
	}

	if h.src.String() != "192.229.216.143" {
		t.Error("expected 192.229.216.143 but got,", h.src)
	}

	if h.dst.String() != "192.229.150.190" {
		t.Error("expected 192.229.150.190 but got,", h.src)
	}
}

func TestCross(t *testing.T) {
	s := &Server{
		read:  make(chan []byte, 2),
		write: make(chan []byte, 2),
	}

	tu := &tun{
		read:  make(chan []byte, 2),
		write: make(chan []byte, 2),
	}

	s.cross(context.Background(), tu)

	var a []byte
	s.read <- []byte("vpn")
	select {
	case a = <-tu.write:
	case <-time.After(1 * time.Millisecond):
	}

	if string(a) != "vpn" {
		t.Error("expected to have vpn but got,", string(a))
	}

	tu.read <- []byte("decentralized")
	select {
	case a = <-s.write:
	case <-time.After(1 * time.Millisecond):
	}

	if string(a) != "decentralized" {
		t.Error("expected to have decentralized but got,", string(a))
	}
}

func testCreateTunInterface(t *testing.T) {
	_, err := createTunInterface()
	if err != nil {
		t.Error("unexpected error:", err)
	}

	_, err = netlink.LinkByName("radvpn")
	if err != nil {
		t.Error("unexpected error", err)
	}
}

func testSetupTunInterface(t *testing.T) {
	createTunInterface()
	setupTunInterface([]string{"10.0.1.1/24"}, 1400)

	ifce, err := netlink.LinkByName("radvpn")
	if err != nil {
		t.Error("unexpected error", err)
	}

	addrs, err := netlink.AddrList(ifce, netlink.FAMILY_ALL)
	if err != nil {
		t.Error("unexpected error", err)
	}

	if len(addrs) > 0 {
		if addrs[0].IP.String() != "10.0.1.1" {
			t.Error("expected 10.0.1.1 but got,", addrs[0].IP.String())
		}
		if addrs[0].Mask.String() != "ffffff00" {
			t.Error("expected ffffff00 but got,", addrs[0].Mask.String())
		}
	} else {
		t.Error("expected having ip address but got nothing")
	}
}
