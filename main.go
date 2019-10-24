package main

import (
	"log"
	"net"
	"os/exec"
	"context"
	"time"
	"flag"

	"github.com/mehrdadrad/radvpn/udp"
	"github.com/mehrdadrad/radvpn/quic"

	"github.com/songgao/water"
)

type Server struct{}

// CreateTAPIfce  creates TAP interface
func(s Server) CreateTUNInterface(ip string) (*water.Interface, error) {
	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = "tun0"

	ifce, err := water.New(config)
	if err != nil {
		return nil, err
	}

	ipCmd("link", "set", "dev", config.Name, "mtu", "1300")
	ipCmd("addr", "add", ip, "dev", config.Name )
	ipCmd("link", "set", "dev", config.Name, "up")

	return ifce, nil
}

func (s Server) UDPServer() (net.PacketConn, error) {
	return net.ListenPacket("udp", ":8085")
}

func ipCmd(args ...string) error{
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return exec.CommandContext(ctx, "ip", args...).Run()
}

var localHost = flag.String("local", "10.0.1.1/24", "IP/Mask")
var remoteHost = flag.String("remote", "192.168.55.10:8085", "IP:Port")
var protoType = flag.String("proto", "udp", "udp or quic")

func main() {
	flag.Parse()

	srv := &Server{}
	tunIf, err := srv.CreateTUNInterface(*localHost)
	if err != nil {
		log.Fatal(err)
	}

	switch *protoType {
		case "udp":
			u := udp.UDP{
				TUNIf: tunIf,
				RemoteHost: *remoteHost,
			}

			u.Run()
		case "quic":	
			q := quic.QUIC{
				TUNIf: tunIf,
				RemoteHost: *remoteHost,
			}

			q.Run()
		default:
			log.Println("not support!")	
	}

	select{}
}