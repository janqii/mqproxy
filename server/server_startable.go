package server

import (
	"github.com/janqii/mqproxy/global"
	"github.com/janqii/mqproxy/producer/kafka"
	"github.com/janqii/mqproxy/server/router"
	"github.com/wvanbergen/kazoo-go"
	"log"
	"net/http"
	"sync"
)

func Startable(cfg *ProxyConfig) error {
	var err error
	wg := new(sync.WaitGroup)

	zkNodes, chroot := kazoo.ParseConnectionString(cfg.ZookeeperAddr)
	zkConfig := &kazoo.Config{
		Chroot:  chroot,
		Timeout: cfg.ZookeeperTimeout,
	}

	zkProxy, err := kazoo.NewKazoo(zkNodes, zkConfig)
	if err != nil {
		log.Printf("NewKazoo error: %v", err)
		return err
	}
	defer zkProxy.Close()

	brokerList, err := zkProxy.BrokerList()
	if err != nil {
		log.Printf("get broker list error: %v", err)
		return err
	}

	pcfg := &producer.KafkaProducerConfig{
		Addrs:               brokerList,
		PartitionerStrategy: cfg.PartitionerStrategy,
		WaitAckStrategy:     cfg.WaitAckStrategy,
		WaitAckTimeoutMs:    cfg.WaitAckTimeoutMs,
		CompressionStrategy: cfg.CompressionStrategy,
		MaxMessageBytes:     cfg.MaxMessageBytes,
		ChannelBufferSize:   cfg.ChannelBufferSize,
	}

	global.ProducerPool, err = global.NewKafkaProducerPool(pcfg, cfg.ProducerPoolSize)
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

	defer statHttpServer.Shutdown()
	defer proxyHttpServer.Shutdown()

	log.Println("MQ Proxy is running...")
	wg.Wait()
	log.Println("MQ Proxy is exiting...")

	return nil
}
