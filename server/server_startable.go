package server

import (
	"fmt"
	"gitlab.baidu.com/go/sarama"
	"gitlab.baidu.com/hanjianqi/mqproxy/server/router"
	"gitlab.baidu.com/hanjianqi/mqproxy/utils"
	"log"
	"net/http"
	"os"
	"sync"
)

var (
	gKafkaClient *sarama.Client
)

func Startable(cfg *ProxyConfig) error {
	wg := new(sync.WaitGroup)

	var zkClient *utils.ZK
	zkClient, err := utils.NewZK(cfg.ZookeeperAddr, cfg.ZookeeperChroot, cfg.ZookeeperTimeout)
	if err != nil {
		log.Printf("init zkClient error: %v", err)
		return err
	}

	gKafkaClient, err = newKafkaClient(zkClient)
	if err != nil {
		log.Printf("create kafka client error")
		return err
	}

	var (
		statMux  map[string]func(http.ResponseWriter, *http.Request)
		proxyMux map[string]func(http.ResponseWriter, *http.Request)
	)
	statMux = make(map[string]func(http.ResponseWriter, *http.Request))
	proxyMux = make(map[string]func(http.ResponseWriter, *http.Request))

	controller := &ProxyController{
		ZkClient:     zkClient,
		IsRunning:    false,
		IsController: false,
		ID:           cfg.ID,
	}
	controller.Startup()

	healthChecker := &ZookeeperHealthChecker{
		ZkClient:  zkClient,
		IsRunning: false,
	}

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
		ZkClient:        zkClient,
		Controller:      controller,
		HealthChecker:   healthChecker,
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
		ZkClient:        zkClient,
		Controller:      controller,
		HealthChecker:   healthChecker,
	}

	statHttpServer.Startup()
	proxyHttpServer.Startup()

	/*tell everyone we are alive*/
	healthChecker.Startup()

	log.Println("MQ Proxy is running...")
	wg.Wait()
	log.Println("MQ Proxy is exiting...")

	return nil
}

func newKafkaClient(zkClient *utils.ZK) (*sarama.Client, error) {
	brokers, err := zkClient.Brokers()
	if err != nil {
		return nil, err
	}

	brokerList := make([]string, 0, len(brokers))
	for _, broker := range brokers {
		brokerList = append(brokerList, broker)
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	pid := os.Getpid()
	clientName := fmt.Sprintf("%s:%d", hostname, pid)

	fmt.Println(clientName)
	var client *sarama.Client
	if client, err = sarama.NewClient(clientName, brokerList, nil); err != nil {
		return nil, err
	}

	return client, nil
}
