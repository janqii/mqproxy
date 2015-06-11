package main

import (
	"github.com/janqii/mqproxy/server"
	"log"
)

func main() {
	cfg, err := server.NewProxyConfig()

	if err != nil {
		log.Fatalf("parse config error, %v", err)
	}

	if err = server.Startable(cfg); err != nil {
		log.Fatalf("server startable error, %v", err)
	}
}
