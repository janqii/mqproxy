package global

import "github.com/Shopify/sarama"

var (
	KafkaClient  *sarama.Client
	ProducerPool *KafkaProducerPool
)
