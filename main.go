package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/mehrdadrad/radvpn/config"
	"github.com/mehrdadrad/radvpn/crypto"
	"github.com/mehrdadrad/radvpn/router"
	"github.com/mehrdadrad/radvpn/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.New().File()
	err := cfg.Load()
	if err != nil {
		log.Fatal(err)
	}

	crp := crypto.GCM{
		Passphrase: cfg.Crypto.Key,
	}

	log.Println(cfg.Crypto.Key)
	log.Println("keylen", len(cfg.Crypto.Key))

	r := router.New()

	s := server.Server{
		KeepAlive: 10 * time.Second,
		Insecure:  cfg.Server.Insecure,
		Cipher:    crp,
		Config:    cfg,
		Router:    r,
		Logger:    log.New(os.Stdout, "", log.Lshortfile),
	}

	s.Run(ctx, 10, 10)
}
