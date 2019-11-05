package config

import (
	"time"
	"context"

	"go.etcd.io/etcd/clientv3"
)

type etcd struct {
	endpoints []string
	file      string
	insecure  bool
	client    *clientv3.Client
}

func (e *etcd) connect() error {
	var err error

	e.client, err = clientv3.New(clientv3.Config{
		Endpoints:   e.endpoints,
		DialTimeout: 2 * time.Second,
	})

	if err != nil {
		return err
	}

	return nil
}

func (e *etcd) close() {
	e.client.Close()
}

func (e etcd) load() (*Config, error) {

	return nil, nil
}

func (e etcd) loadFromFile(cfile string) error {
	cf := &file{
		paths: []string{"/etc", "/use/local/etc"},
		cfile: cfile,
	}

	cfg, err := cf.load()
	if err != nil {
		return err
	}

	// TODO
	_ = cfg

	return nil
}

// update updates etcd from yaml file
func (e etcd) update() error {

	return nil
}

func (e etcd) watch(ctx context.Context, notify chan struct{}) {

}
