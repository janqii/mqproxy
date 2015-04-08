package main

import (
	"github.com/janqii/mqproxy/global"
	"github.com/janqii/mqproxy/server"
	"log"
)

func main() {
	cfg, err := global.NewProxyConfig()
	if err != nil {
		log.Fatalf("parse config error, %v", err)
	}

	if err = server.Startable(cfg); err != nil {
		log.Fatalf("server startable error, %v", err)
	}
}
