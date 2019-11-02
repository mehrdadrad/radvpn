package config

import (
	"errors"

	"github.com/vishvananda/netlink"
)

// Config represents configuration
type Config struct {
	Server struct {
		Keepalive int  `yaml:"keepalive"`
		Insecure  bool `yaml:"insecure"`
		Mtu       int  `yaml:"mtu"`
	} `yaml:"server"`
	Crypto struct {
		Type string `yaml:"type"`
	} `yaml:"crypto"`
	Nodes []struct {
		Node `yaml:"node"`
	} `yaml:"nodes"`

	source source

	file file
	etcd etcd
}

// Node represents node / host IP configuration
type Node struct {
	Name           string   `yaml:"name"`
	Address        string   `yaml:"address"`
	PrivateAddress []string `yaml:"privateAddress"`
	PrivateSubnets []string `yaml:"privateSubnets"`
}

type source interface {
	load() (*Config, error)
	watch()
}

// New constructs new empty configuration
func New() *Config {
	return &Config{}
}

func (c *Config) File() *Config {
	c.source = &file{}
	return c
}

func (c *Config) Etcd() *Config {
	c.source = &etcd{}
	return c
}


// Load loads configuration from file
func (c *Config) Load() (Node, error) {
	cfg, err := c.source.load()
	if err != nil {
		return Node{}, err
	}

	*c = *cfg

	return c.whoiam()
}

func (c Config) GetNodesPrivateSubnets() []string {
	var subnets []string
	for _, nodes := range c.Nodes {
		subnets = append(subnets, nodes.Node.PrivateSubnets...)
	}
	return subnets
}

func (c Config) whoiam() (Node, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return Node{}, err
	}
	var ipList []string

	for _, link := range links {
		addrs, err := netlink.AddrList(link, netlink.FAMILY_ALL)
		if err != nil {
			return Node{}, err
		}
		for _, addr := range addrs {
			ipList = append(ipList, addr.IP.String())
		}
	}

	for _, nodes := range c.Nodes {
		for _, ip := range ipList {
			if nodes.Node.Address == ip {
				return nodes.Node, nil
			}
		}
	}

	return Node{}, errors.New("whoami error: can not find node")
}

func (n Node) GetPrivateSubnets() []string {
	return n.PrivateSubnets
}

func (n Node) GetPrivateAddress() []string {
	return n.PrivateSubnets
}
