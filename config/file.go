package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type file struct {
	paths []string
}

func (f file) load() (*Config, error) {
	content, err := ioutil.ReadFile("./config.yaml")
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

func (f file) watch() {

}
