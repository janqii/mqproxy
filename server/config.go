package server

import (
	"errors"
	"flag"
	"fmt"
	"github.com/janqii/mqproxy/internal/version"
	"os"
	//	"strings"
	"time"
)

type ProxyConfig struct {
	ID               int
	HttpServerPort   string
	NsheadServerPort string
	StatServerPort   string

	PrintVersion bool

	HttpServerReadTimeout    time.Duration
	HttpServerWriteTimeout   time.Duration
	HttpServerMaxHeaderBytes int
	HttpKeepAliveEnabled     bool

	NsheadServerReadTimeout  time.Duration
	NsheadServerWriteTimeout time.Duration

	ZookeeperAddr    string
	ZookeeperTimeout time.Duration

	//Producer Configure
	PartitionerStrategy string //Hash, Random, RoundRobin
	WaitAckStrategy     string // NoRespond, WaitForLocal, WaitForAll
	WaitAckTimeoutMs    time.Duration
	CompressionStrategy string //None, Gzip, Snappy
	MaxMessageBytes     int
	ChannelBufferSize   int

	ProducerPoolSize int
}

func NewProxyConfig() (*ProxyConfig, error) {
	cfg := new(ProxyConfig)
	if err := cfg.Parse(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *ProxyConfig) Parse() error {
	id := flag.Int("id", 0, "proxy id")

	httpServerPort := flag.String("http_port", "", "http server port")
	nsheadServerPort := flag.String("nshead_port", "", "nshead server port")
	statServerPort := flag.String("stat_port", "", "stat server port")

	needPrintVersion := flag.Int("version", 0, "print version")

	httpServerReadTimeout := flag.Int64("http_server_read_timeout", 5000, "http server read timeout")
	httpServerWriteTimeout := flag.Int64("http_server_write_timeout", 5000, "http server write timeout")
	httpServerMaxHeaderBytes := flag.Int("http_server_max_header_bytes", 1<<20, "http server max header bytes")
	httpKeepAliveEnabled := flag.Int("http_keep_alive", 0, "http keep alive enable")

	nsheadServerReadTimeout := flag.Int64("nshead_server_read_timeout", 5000, "nshead server read timeout")
	nsheadServerWriteTimeout := flag.Int64("nshead_server_write_timeout", 5000, "nshead server write timeout")

	zookeeperAddr := flag.String("zookeeper_addr", "", "zookeeper address")
	zookeeperTimeout := flag.Int64("zookeeper_timeout", 1, "zookeeper connect timeout")

	partitionerStrategy := flag.String("partitioner_strategy", "Hash", "partitioner strategy")
	waitAckStrategy := flag.String("wait_ack_strategy", "WaitForLocal", "The level of acknowledgement reliability needed from the broker")
	waitAckTimeoutMs := flag.Int64("wait_ack_timeout_ms", 0, "The maximum duration the broker will wait the receipt of the number of RequiredAcks.")
	compressionStrategy := flag.String("compression_strategy", "None", "The type of compression to use on messages")
	maxMessageBytes := flag.Int("max_message_bytes", 1000000, "The maximum permitted size of a message")
	channelBufferSize := flag.Int("channel_buffer_size", 0, "The size of the buffers of the channels between the different goroutines")
	producerPoolSize := flag.Int("producer_pool_size", 256, "The size of producer pool")

	flag.Parse()

	cfg.PrintVersion = (*needPrintVersion > 0)
	if cfg.PrintVersion {
		fmt.Println(version.String("mqproxy"))
		os.Exit(1)
	}

	if *id <= 0 {
		return errors.New("id nil")
	}
	if *httpServerPort == "" {
		return errors.New("http_port nil")
	}
	if *nsheadServerPort == "" {
		return errors.New("nshead_port nil")
	}
	if *statServerPort == "" {
		return errors.New("stat_port nil")
	}
	if *zookeeperAddr == "" {
		return errors.New("zookeeper_addr nil")
	}

	cfg.ID = *id
	cfg.HttpServerPort = *httpServerPort
	cfg.NsheadServerPort = *nsheadServerPort
	cfg.StatServerPort = *statServerPort

	cfg.HttpServerReadTimeout = time.Duration(*httpServerReadTimeout) * time.Millisecond
	cfg.HttpServerWriteTimeout = time.Duration(*httpServerWriteTimeout) * time.Millisecond
	cfg.HttpServerMaxHeaderBytes = *httpServerMaxHeaderBytes
	cfg.HttpKeepAliveEnabled = (*httpKeepAliveEnabled > 0)

	cfg.NsheadServerReadTimeout = time.Duration(*nsheadServerReadTimeout) * time.Millisecond
	cfg.NsheadServerWriteTimeout = time.Duration(*nsheadServerWriteTimeout) * time.Millisecond

	cfg.ZookeeperAddr = *zookeeperAddr
	cfg.ZookeeperTimeout = time.Duration(*zookeeperTimeout) * time.Second

	cfg.PartitionerStrategy = *partitionerStrategy
	cfg.WaitAckStrategy = *waitAckStrategy
	cfg.WaitAckTimeoutMs = time.Duration(*waitAckTimeoutMs) * time.Millisecond
	cfg.CompressionStrategy = *compressionStrategy
	cfg.MaxMessageBytes = *maxMessageBytes
	cfg.ChannelBufferSize = *channelBufferSize

	cfg.ProducerPoolSize = *producerPoolSize

	return nil
}
