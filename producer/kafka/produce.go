package producer

import (
	"gopkg.in/Shopify/sarama.v1"
	"gopkg.in/vmihailenco/msgpack.v2"
	"sync"
	"time"
)

type Request struct {
	Topic        string
	PartitionKey string
	Data         interface{}
}

type Response struct {
	Errno  int
	Errmsg string
	Data   []MessageLocation
}

type MessageLocation struct {
	Partition int32
	Offset    int64
}

type KafkaProducer struct {
	producer sarama.SyncProducer
	m        sync.Mutex
}

type KafkaProducerConfig struct {
	Addrs               []string
	PartitionerStrategy string
	WaitAckStrategy     string
	WaitAckTimeoutMs    time.Duration
	CompressionStrategy string
	MaxMessageBytes     int
	ChannelBufferSize   int
}

func NewKafkaProducer(config *KafkaProducerConfig) (*KafkaProducer, error) {
	var err error

	kp := new(KafkaProducer)
	pcfg := NewProducerConfig(config)
	if kp.producer, err = sarama.NewSyncProducer(config.Addrs, pcfg); err != nil {
		return nil, err
	}

	return kp, nil
}

func (kp *KafkaProducer) SendMessage(req Request) (Response, error) {
	var err error

	kp.m.Lock()
	defer kp.m.Unlock()

	b, err := msgpack.Marshal(map[string]interface{}{"Data": req.Data})
	if err != nil {
		return Response{-1, err.Error(), make([]MessageLocation, 0)}, err
	}

	msg := &sarama.ProducerMessage{
		Topic: req.Topic,
		Key:   sarama.StringEncoder(req.PartitionKey),
		Value: sarama.ByteEncoder(b),
	}

	partition, offset, err := kp.producer.SendMessage(msg)
	if err != nil {
		return Response{-1, err.Error(), make([]MessageLocation, 0)}, err
	}

	return Response{0, "ok", []MessageLocation{{
		Partition: partition,
		Offset:    offset,
	}}}, nil
}

func (kp *KafkaProducer) Close() error {
	return kp.producer.Close()
}

func NewProducerConfig(cfg *KafkaProducerConfig) *sarama.Config {
	config := new(sarama.Config)

	if cfg.PartitionerStrategy == "Random" {
		config.Producer.Partitioner = sarama.NewRandomPartitioner
	} else if cfg.PartitionerStrategy == "RoundRobin" {
		config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	} else {
		config.Producer.Partitioner = sarama.NewHashPartitioner
	}

	if cfg.WaitAckStrategy == "NoRespond" {
		config.Producer.RequiredAcks = sarama.NoResponse
	} else if cfg.WaitAckStrategy == "WaitForAll" {
		config.Producer.RequiredAcks = sarama.WaitForAll
	} else {
		config.Producer.RequiredAcks = sarama.WaitForLocal
	}

	config.Producer.Timeout = cfg.WaitAckTimeoutMs

	if cfg.CompressionStrategy == "None" {
		config.Producer.Compression = sarama.CompressionNone
	} else if cfg.CompressionStrategy == "Gzip" {
		config.Producer.Compression = sarama.CompressionGZIP
	} else {
		config.Producer.Compression = sarama.CompressionSnappy
	}

	config.Producer.MaxMessageBytes = cfg.MaxMessageBytes
	config.ChannelBufferSize = cfg.ChannelBufferSize
	//	producerConfig.AckSuccesses = true

	return config
}
