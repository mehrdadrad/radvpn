package config

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	yaml "gopkg.in/yaml.v2"
)

type file struct {
	paths      []string
	cfile      string
	watchDelay int
}

func (f *file) load() (*Config, error) {
	configFile := ""

	if f.cfile != "" {
		configFile = f.cfile
	} else {
		for _, path := range f.paths {
			cf := filepath.Join(path, "radvpn.yaml")
			if _, err := os.Stat(cf); err == nil {
				configFile = cf
			}
		}
	}

	if configFile == "" {
		return nil, errors.New("config file not found")
	}

	f.cfile = configFile

	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	c := &Config{}

	err = yaml.Unmarshal(content, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (f file) watch(ctx context.Context, notify chan struct{}) {
	go func() {
		stat, err := os.Stat(f.cfile)
		if err != nil {
			log.Fatal(err)
		}

		if f.watchDelay == 0 {
			f.watchDelay = 5
		}

		modTime := stat.ModTime()
		ticker := time.NewTicker(time.Duration(f.watchDelay) * time.Second)

		for {

			select {
			case <-ticker.C:
			case <-ctx.Done():
				return
			}

			stat, err := os.Stat(f.cfile)
			if err != nil {
				log.Fatal(err)
			}

			if ok := modTime.Equal(stat.ModTime()); !ok {
				select {
				case notify <- struct{}{}:
				default:
				}

				modTime = stat.ModTime()
			}
		}
	}()
}
