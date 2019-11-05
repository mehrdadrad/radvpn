package config

import (
	"context"
	"errors"
	"os"

	"github.com/vishvananda/netlink"
)

// Config represents configuration
type Config struct {
	Server struct {
		Name      string `yaml:"name"`
		Keepalive int    `yaml:"keepalive"`
		Insecure  bool   `yaml:"insecure"`
		Mtu       int    `yaml:"mtu"`
	} `yaml:"server"`
	Crypto struct {
		Type string `yaml:"type"`
		Key  string `yaml:"key"`
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
	Name             string   `yaml:"name"`
	Address          string   `yaml:"address"`
	PrivateAddresses []string `yaml:"privateAddresses"`
	PrivateSubnets   []string `yaml:"privateSubnets"`
}

type source interface {
	load() (*Config, error)
	watch(context.Context, chan struct{})
}

// New constructs new empty configuration
func New() *Config {
	return &Config{}
}

// File sets the config source to file
func (c *Config) File(cfile string) *Config {
	c.source = &file{
		paths: []string{"/etc", "/use/local/etc"},
		cfile: cfile,
	}
	return c
}

// Etcd sets the config source to etcd
func (c *Config) Etcd() *Config {
	c.source = &etcd{}
	return c
}

// Load loads configuration from file
func (c *Config) Load() error {
	cfg, err := c.source.load()
	if err != nil {
		return err
	}

	*c = *cfg

	return nil
}

// GetNodesPrivateSubnets returns all nodes private subnets
func (c Config) GetNodesPrivateSubnets() []string {
	var subnets []string
	for _, nodes := range c.Nodes {
		subnets = append(subnets, nodes.Node.PrivateSubnets...)
	}
	return subnets
}

// GetIRB returns information route base
func (c Config) GetIRB() map[string][]string {
	irb := make(map[string][]string)
	for _, nodes := range c.Nodes {
		irb[nodes.Node.Address] = nodes.Node.PrivateSubnets
	}

	return irb
}

// Whoami returns current node config
func (c Config) Whoami() (Node, error) {
	// if the server name exist at env
	nodeName := os.Getenv("RADVPN_NODE_NAME")
	if nodeName != "" {
		for _, nodes := range c.Nodes {
			if nodes.Node.Name == nodeName {
				return nodes.Node, nil
			}
		}
	}

	// find node based on the external ip address
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

// GetPrivateSubnets gets the node's private subnets
func (n Node) GetPrivateSubnets() []string {
	return n.PrivateSubnets
}

// GetPrivateAddresses gets the node's private addresses
func (n Node) GetPrivateAddresses() []string {
	return n.PrivateAddresses
}
