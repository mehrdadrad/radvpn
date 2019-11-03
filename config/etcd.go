package config

type etcd struct {
	url      string
	insecure bool
}

func (e etcd) load() (*Config, error) {

	return nil, nil
}

func (e etcd) watch() {

}
