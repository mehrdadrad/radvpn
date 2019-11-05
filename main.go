package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/mehrdadrad/radvpn/config"
	"github.com/mehrdadrad/radvpn/router"
	"github.com/mehrdadrad/radvpn/server"
)

var (
	configFile string
	etcd       bool
	cfg        *config.Config
)

func init() {
	flag.StringVar(&configFile, "config", "", "configuration file")
	flag.BoolVar(&etcd, "etcd", false, "enable etcd")
	flag.Parse()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)

	if etcd {
		cfg = config.New().FromEtcd(configFile)
	} else {
		cfg = config.New().FromFile(configFile)
	}

	err := cfg.Load()
	if err != nil {
		log.Fatal(err)
	}

	notify := make(chan struct{})
	cfg.Watcher(notify)

	r := router.New(ctx)

	s := server.Server{
		Config: cfg,
		Router: r,
		Notify: notify,
		Logger: log.New(os.Stdout, "", log.Lshortfile),
	}

	s.Run(ctx, 10, 10)

	<-sig
}
