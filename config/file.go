package config

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"
)

type file struct {
	paths []string
	cfile string
}

func (f file) load() (*Config, error) {
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

func (f file) watch(notify chan struct{}) {
	go func() {
		stat, err := os.Stat(f.cfile)
		if err != nil {
			log.Fatal(err)
		}
		modTime := stat.ModTime()
		ticker := time.NewTicker(5 * time.Second)
		for {
			<-ticker.C

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
