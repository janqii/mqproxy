package global

import "gitlab.baidu.com/go/sarama"

var (
	Config      *ProxyConfig
	KafkaClient *sarama.Client
)
