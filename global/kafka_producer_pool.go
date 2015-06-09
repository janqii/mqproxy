package global

import (
	"github.com/Shopify/sarama"
	"github.com/janqii/mqproxy/producer/kafka"
)

type KafkaProducerPool struct {
	producers []*producer.KafkaProducer
	size      int
	curr      int
}

func NewKafkaProducerPool(client *sarama.Client, config *producer.KafkaProducerConfig, poolSize int) (*KafkaProducerPool, error) {
	var err error

	pool := &KafkaProducerPool{
		producers: make([]*producer.KafkaProducer, poolSize),
		size:      0,
		curr:      -1,
	}

	for i := 0; i < poolSize; i++ {
		if pool.producers[i], err = producer.NewKafkaProducer(client, config); err != nil {
			return pool, err
		}
		pool.size++
	}
	ProducerPool = pool

	return pool, nil
}

func DestoryKafkaProducerPool(pool *KafkaProducerPool) error {
	for i := 0; i < pool.size; i++ {
		if err := pool.producers[i].Close(); err != nil {
			return err
		}
	}

	return nil
}

func (pool *KafkaProducerPool) GetProducer() *producer.KafkaProducer {
	return pool.producers[(pool.curr+1)%pool.size]
}
