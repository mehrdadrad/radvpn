package config

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

var cfgT = `
revision: 1
etcd:
  endpoints:
    - localhost:2379
  timeout: 5

server:
  keepalive: 10
  insecure: false
  mtu: 1300

crypto:
  type: gcm
  key: mykey

nodes:
  - node:
      name: node1
      address: 192.168.55.20
      privateAddresses:
        - 10.0.2.1/24
      privateSubnets:
        - 10.0.2.0/24
`

func TestiFileLoad(t *testing.T) {
	tf, err := ioutil.TempFile("", "")
	if err != nil {
		t.Error(err)
	}

	defer os.Remove(tf.Name())

	tf.WriteString(cfgT)

	f := &file{
		cfile: tf.Name(),
	}

	cfg, err := f.load()
	if err != nil {
		t.Error(err)
	}

	if cfg.Revision != 1 {
		t.Error("expected revision 1 but got,", cfg.Revision)
	}

	if cfg.Crypto.Type != "gcm" {
		t.Error("expected crypto.type gcm but got,", cfg.Crypto.Type)
	}

	if cfg.Server.Keepalive != 10 {
		t.Error("expected server keepalive 10 but got,", cfg.Server.Keepalive)
	}

	if cfg.Nodes[0].Node.Name != "node1" {
		t.Error("expected node.name node1 but got,", cfg.Nodes[0].Node.Name)
	}

	if cfg.Etcd.Timeout != 5 {
		t.Error("expected etcd timeout 5 but got,", cfg.Etcd.Timeout)
	}
}

func TestFileWatch(t *testing.T) {
	tf, err := ioutil.TempFile("", "")
	if err != nil {
		t.Error(err)
	}

	defer os.Remove(tf.Name())

	tf.WriteString(cfgT)

	f := &file{
		cfile:      tf.Name(),
		watchDelay: 1,
	}

	notify := make(chan struct{}, 1)
	f.watch(context.Background(), notify)
	time.Sleep(1 * time.Second)
	tf.WriteString(cfgT)
	time.Sleep(1 * time.Second)

	if len(notify) != 1 {
		t.Error("expected notifcation but got nothing")
	}
}
