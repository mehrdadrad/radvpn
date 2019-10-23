package main

import (
	"log"
	"net"
	"os/exec"
	"context"
	"time"

	"github.com/songgao/water"
)

type Server struct{}

// CreateTAPIfce  creates TAP interface
func(s Server) CreateTAPIfce() (*water.Interface, error) {
	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = "tun0"

	ifce, err := water.New(config)
	if err != nil {
		return nil, err
	}

	ipCmd("link", "set", "dev", config.Name, "mtu", "1300")
	ipCmd("addr", "add", "10.10.55.1/24", "dev", config.Name )
	ipCmd("link", "set", "dev", config.Name, "up")

	return ifce, nil
}

func (s Server) UDPServer() (net.PacketConn, error) {
	return net.ListenPacket("udp", ":8055")
}

func ipCmd(args ...string) error{
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return exec.CommandContext(ctx, "ip", args...).Run()
}

func main() {
	srv := &Server{}
	ifce, err := srv.CreateTAPIfce()
	if err != nil {
		log.Fatal(err)
	}

	pConn, err := srv.UDPServer()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		buff := make([]byte, 1024)	
		for {
			n, _, err := pConn.ReadFrom(buff)	
			if err != nil {
				continue
			}

			ifce.Write(buff[:n])
		}
	}()

	go func() {
		buff := make([]byte, 1024)
		rAddress, _ := net.ResolveUDPAddr("udp", "192.168.55.10:8085")
		for { 
			n, err := ifce.Read(buff)
			if err != nil {
				continue
			}

			pConn.WriteTo(buff[:n], rAddress)
		}
	}()

	select{}
}