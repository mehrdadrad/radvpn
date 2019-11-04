package main

import (
	"context"
	"flag"
	"log"
	"os"

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

	cfg := config.New().File(configFile)
	err := cfg.Load()
	if err != nil {
		log.Fatal(err)
	}

	r := router.New()

	s := server.Server{
		Config: cfg,
		Router: r,
		Logger: log.New(os.Stdout, "", log.Lshortfile),
	}

	s.Run(ctx, 10, 10)
}
