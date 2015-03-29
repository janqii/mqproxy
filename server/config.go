package server

import (
	"errors"
	"flag"
	"strings"
	"time"
)

type ProxyConfig struct {
	ID               int
	HttpServerPort   string
	NsheadServerPort string
	StatServerPort   string

	HttpServerReadTimeout    time.Duration
	HttpServerWriteTimeout   time.Duration
	HttpServerMaxHeaderBytes int
	HttpKeepAliveEnabled     bool

	NsheadServerReadTimeout  time.Duration
	NsheadServerWriteTimeout time.Duration

	ZookeeperAddr    []string
	ZookeeperChroot  string
	ZookeeperTimeout time.Duration
}

func (cfg *ProxyConfig) Parse() error {
	id := flag.Int("id", 0, "proxy id")

	httpServerPort := flag.String("http_port", "", "http server port")
	nsheadServerPort := flag.String("nshead_port", "", "nshead server port")
	statServerPort := flag.String("stat_port", "", "stat server port")

	httpServerReadTimeout := flag.Int64("http_server_read_timeout", 5000, "http server read timeout")
	httpServerWriteTimeout := flag.Int64("http_server_write_timeout", 5000, "http server write timeout")
	httpServerMaxHeaderBytes := flag.Int("http_server_max_header_bytes", 1<<20, "http server max header bytes")
	httpKeepAliveEnabled := flag.Int("http_keep_alive", 0, "http keep alive enable")

	nsheadServerReadTimeout := flag.Int64("nshead_server_read_timeout", 5000, "nshead server read timeout")
	nsheadServerWriteTimeout := flag.Int64("nshead_server_write_timeout", 5000, "nshead server write timeout")

	zookeeperAddr := flag.String("zookeeper_addr", "", "zookeeper address")
	zookeeperChroot := flag.String("zookeeper_chroot", "", "zookeeper chroot")
	zookeeperTimeout := flag.Int64("zookeeper_timeout", 1, "zookeeper connect timeout")

	flag.Parse()

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
	if *zookeeperChroot == "" {
		return errors.New("zookeeper_chroot nil")
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

	cfg.ZookeeperAddr = strings.Split(*zookeeperAddr, ",")
	cfg.ZookeeperChroot = *zookeeperChroot
	cfg.ZookeeperTimeout = time.Duration(*zookeeperTimeout) * time.Second

	return nil
}
