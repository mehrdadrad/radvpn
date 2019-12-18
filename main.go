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
	update     string
	etcd       bool
	cfg        *config.Config
)

func init() {
	flag.StringVar(&configFile, "config", "", "configuration file")
	flag.StringVar(&update, "update", "", "update etc / file")
	flag.BoolVar(&etcd, "etcd", false, "enable etcd")
	flag.Parse()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if update != "" {
		err := config.UpdateConf(update, configFile)
		if err != nil {
			log.Println(err)
		}
		os.Exit(0)
	}

	sig := make(chan os.Signal, 1)
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

	notify := make(chan struct{}, 1)
	cfg.Watcher(ctx, notify)

	r := router.New(ctx)

	s := server.Server{
		Config: cfg,
		Router: r,
		Notify: notify,
	}

	s.Run(ctx)

	<-sig
}
