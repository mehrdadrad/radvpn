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
		Name       string `yaml:"name"`
		Address    string `yaml:"address"`
		MaxWorkers int    `yaml:"maxworkers"`
		Keepalive  int    `yaml:"keepalive"`
		Insecure   bool   `yaml:"insecure"`
		Mtu        int    `yaml:"mtu"`
	} `yaml:"server"`

	Crypto struct {
		Type string `yaml:"type"`
		Key  string `yaml:"key"`
	} `yaml:"crypto"`

	Nodes []struct {
		Node `yaml:"node"`
	} `yaml:"nodes"`

	Etcd struct {
		Endpoints []string `yaml:endpoints`
		Timeout   int      `yaml:timeout`
	} `yaml:"etcd"`

	Revision int `yaml:"revision"`

	source source
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

// FromFile sets the config source to file
func (c *Config) FromFile(cfile string) *Config {
	c.source = &file{
		paths:      []string{"/etc", "/use/local/etc"},
		cfile:      cfile,
		watchDelay: 5,
	}
	return c
}

// FromEtcd sets the config source to etcd
func (c *Config) FromEtcd(cfile string) *Config {
	c.source = &etcd{
		cfile: cfile,
	}
	return c
}

// UpdateEtcd updates etcd from file
func (c Config) UpdateEtcd(cfile string) error {
	e := &etcd{
		cfile: cfile,
	}

	cfg, err := e.loadFromFile()
	if err != nil {
		return err
	}

	e.endpoints = cfg.Etcd.Endpoints

	err = e.connect()
	if err != nil {
		return err
	}

	defer e.close()

	err = e.putConfig(cfg)
	if err != nil {
		return err
	}

	return nil
}

// Load loads configuration from file / etcd
func (c *Config) Load() error {
	cfg, err := c.source.load()
	if err != nil {
		return err
	}

	source := c.source
	*c = *cfg
	c.source = source

	setDefaultConfig(c)

	return nil
}

// Watcher reloads the configuration
func (c *Config) Watcher(ctx context.Context, extNotify chan struct{}) {
	notify := make(chan struct{}, 1)
	go c.source.watch(ctx, notify)
	go func(n chan struct{}) {
		for {
			select {
			case <-notify:
			case <-ctx.Done():
				return
			}

			c.Load()
			extNotify <- struct{}{}
		}
	}(extNotify)
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
func (c *Config) GetIRB() map[string][]string {
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

// UpdateConf updates etcd from file and reverse
func UpdateConf(source string, cfile string) error {

	if source == "etcd" {
		err := New().UpdateEtcd(cfile)
		if err != nil {
			return err
		}
	}

	// TODO update file from etcd

	return nil
}

func setDefaultConfig(c *Config) {
	// set defaults
	if c.Server.Address == "" {
		c.Server.Address = ":8085"
	}

	if c.Server.MaxWorkers == 0 {
		c.Server.MaxWorkers = 10
	}

	if c.Server.Keepalive == 0 {
		c.Server.Keepalive = 10
	}

	if c.Server.Mtu == 0 {
		c.Server.Mtu = 1300
	}
}
