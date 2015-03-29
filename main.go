package main

import (
	"github.com/janqii/mqproxy/server"
	"log"
)

func main() {
	cfg := new(server.ProxyConfig)
	err := cfg.Parse()
	if err != nil {
		log.Fatalf("parse config error, %v", err)
	}

	err = server.Startable(cfg)
	if err != nil {
		log.Fatalf("server startable error, %v", err)
	}
}
