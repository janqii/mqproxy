package server

import (
	"github.com/janqii/mqproxy/global"
	"github.com/janqii/mqproxy/producer/kafka"
	"github.com/janqii/mqproxy/server/router"
	"github.com/janqii/mqproxy/utils"
	"log"
	"net/http"
	"sync"
)

func Startable(cfg *ProxyConfig) error {
	wg := new(sync.WaitGroup)

	var (
		err      error
		zkClient *utils.ZK
	)

	if zkClient, err = utils.NewZK(cfg.ZookeeperAddr, cfg.ZookeeperChroot, cfg.ZookeeperTimeout); err != nil {
		log.Printf("init zkClient error: %v", err)
		return err
	}
	defer zkClient.Close()

	if global.KafkaClient, err = global.NewKafkaClient(zkClient); err != nil {
		log.Printf("create kafka client error: %v", err)
		return err
	}
	defer global.KafkaClient.Close()

	pcfg := &producer.KafkaProducerConfig{
		PartitionerStrategy: cfg.PartitionerStrategy,
		WaitAckStrategy:     cfg.WaitAckStrategy,
		WaitAckTimeoutMs:    cfg.WaitAckTimeoutMs,
		CompressionStrategy: cfg.CompressionStrategy,
		MaxMessageBytes:     cfg.MaxMessageBytes,
		ChannelBufferSize:   cfg.ChannelBufferSize,
	}

	global.ProducerPool, err = global.NewKafkaProducerPool(global.KafkaClient, pcfg, cfg.ProducerPoolSize)
	defer global.DestoryKafkaProducerPool(global.ProducerPool)
	if err != nil {
		log.Printf("create kafka producer pool error: %v", err)
		return err
	}

	var (
		statMux  map[string]func(http.ResponseWriter, *http.Request)
		proxyMux map[string]func(http.ResponseWriter, *http.Request)
	)
	statMux = make(map[string]func(http.ResponseWriter, *http.Request))
	proxyMux = make(map[string]func(http.ResponseWriter, *http.Request))

	statHttpServer := &HttpServer{
		Addr:            ":" + cfg.StatServerPort,
		Handler:         &HttpHandler{Mux: statMux},
		ReadTimeout:     cfg.HttpServerReadTimeout,
		WriteTimeout:    cfg.HttpServerWriteTimeout,
		MaxHeaderBytes:  cfg.HttpServerMaxHeaderBytes,
		KeepAliveEnable: cfg.HttpKeepAliveEnabled,
		RouterFunc:      router.StatServerRouter,
		Wg:              wg,
		Mux:             statMux,
	}
	proxyHttpServer := &HttpServer{
		Addr:            ":" + cfg.HttpServerPort,
		Handler:         &HttpHandler{Mux: proxyMux},
		ReadTimeout:     cfg.HttpServerReadTimeout,
		WriteTimeout:    cfg.HttpServerWriteTimeout,
		MaxHeaderBytes:  cfg.HttpServerMaxHeaderBytes,
		KeepAliveEnable: cfg.HttpKeepAliveEnabled,
		RouterFunc:      router.ProxyServerRouter,
		Wg:              wg,
		Mux:             proxyMux,
	}

	statHttpServer.Startup()
	proxyHttpServer.Startup()

	defer statHttpServer.ShutDown()
	defer proxyHttpServer.ShutDown()

	log.Println("MQ Proxy is running...")
	wg.Wait()
	log.Println("MQ Proxy is exiting...")

	return nil
}
