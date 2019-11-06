package config

import (
	"context"
	"errors"
	"log"
	"os"
	"path"
	"strconv"
	"time"

	"go.etcd.io/etcd/clientv3"
	yaml "gopkg.in/yaml.v2"
)

type etcd struct {
	endpoints []string
	cfile     string
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

func (e *etcd) load() (*Config, error) {
	cfg, err := e.loadFromFile()
	if err != nil {
		return nil, err
	}

	e.endpoints = cfg.Etcd.Endpoints

	if err := e.connect(); err != nil {
		return nil, err
	}
	defer e.close()

	etcdRev, err := e.getKey("/radvpn/revision")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	// revision not exist at etcd / fresh etcd server
	if err != nil && errors.Is(err, os.ErrNotExist) {
		err := e.putConfig(cfg)
		if err != nil {
			return nil, err
		}

		return cfg, nil
	}

	// file has been updated and not yet sync w/ etcd
	rev, _ := strconv.Atoi(string(etcdRev))
	if rev < cfg.Revision {
		err := e.putConfig(cfg)
		if err != nil {
			return nil, err
		}

		return cfg, nil
	}

	cfgEtcd, err := e.getConfig()
	if err != nil {
		return nil, err
	}

	return cfgEtcd, nil
}

func (e etcd) putConfig(cfg *Config) error {
	base := "/radvpn/"
	config, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	ops := []clientv3.Op{
		clientv3.OpPut(path.Join(base, "revision"), strconv.Itoa(cfg.Revision)),
		clientv3.OpPut(path.Join(base, "config"), string(config)),
	}

	for _, op := range ops {
		if _, err := e.client.Do(context.TODO(), op); err != nil {
			return err
		}
	}

	return nil
}

func (e etcd) getConfig() (*Config, error) {
	cfg := &Config{}
	base := "/radvpn/"

	b, err := e.getKey(path.Join(base, "config"))
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(b, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (e etcd) loadFromFile() (*Config, error) {
	cf := &file{
		paths: []string{"/etc", "/use/local/etc"},
		cfile: e.cfile,
	}

	cfg, err := cf.load()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (e etcd) getKey(key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

	resp, err := e.client.Get(ctx, key)
	cancel()
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, os.ErrNotExist
	}

	return resp.Kvs[0].Value, nil
}

func (e etcd) watch(ctx context.Context, notify chan struct{}) {
	var (
		pRev, cRev int
		ticker     = time.NewTicker(5 * time.Second)
	)

	for {
		e.close()

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}

		if err := e.connect(); err != nil {
			log.Println(err)
			time.Sleep(2 * time.Second)
			continue
		}

		revStr, err := e.getKey("/radvpn/revision")
		if err != nil {
			log.Println(err)
			time.Sleep(5 * time.Second)
			continue
		}

		if pRev == 0 && cRev == 0 {
			revInt, err := strconv.Atoi(string(revStr))
			if err != nil {
				log.Println(err)
				continue
			}

			pRev = revInt
			cRev = revInt
		}

		cRev, err = strconv.Atoi(string(revStr))
		if err != nil {
			log.Println(err)
			continue
		}

		if cRev > pRev {
			select {
			case notify <- struct{}{}:
				pRev = cRev
			default:
			}
		}

	}
}
