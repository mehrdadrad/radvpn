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

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "", "configuration file")
	flag.Parse()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)

	cfg := config.New().File(configFile)
	err := cfg.Load()
	if err != nil {
		log.Fatal(err)
	}

	r := router.New(ctx)

	s := server.Server{
		Config: cfg,
		Router: r,
		Logger: log.New(os.Stdout, "", log.Lshortfile),
	}

	s.Run(ctx, 10, 10)

	<-sig
}
