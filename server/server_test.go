package server

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/mehrdadrad/radvpn/config"
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
		t.Error("expect err nil but got,", err)
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
		t.Error("expect to have err but got nil")
	}
}

func TestListenPacket(t *testing.T) {
	cfg := &config.Config{
		Server: struct {
			Name      string `yaml:"name"`
			Keepalive int    `yaml:"keepalive"`
			Insecure  bool   `yaml:"insecure"`
			Mtu       int    `yaml:"mtu"`
		}{
			Keepalive: 5,
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
