package producer

import (
	"github.com/Shopify/sarama"
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
	producer *sarama.Producer
	m        sync.Mutex
}

type KafkaProducerConfig struct {
	PartitionerStrategy string
	WaitAckStrategy     string
	WaitAckTimeoutMs    time.Duration
	CompressionStrategy string
	MaxMessageBytes     int
	ChannelBufferSize   int
}

func NewKafkaProducer(client *sarama.Client, config *KafkaProducerConfig) (*KafkaProducer, error) {
	var err error

	kp := new(KafkaProducer)
	pcfg := NewProducerConfig(config)
	if kp.producer, err = sarama.NewProducer(client, pcfg); err != nil {
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

	kp.producer.Input() <- &sarama.MessageToSend{
		Topic: req.Topic,
		Key:   sarama.StringEncoder(req.PartitionKey),
		Value: sarama.ByteEncoder(b),
	}

	select {
	case msg := <-kp.producer.Errors():
		return Response{-1, msg.Err.Error(), make([]MessageLocation, 0)}, msg.Err
	case msg := <-kp.producer.Successes():
		//        arr := [1]MessageLocation{}
		return Response{0, "ok", []MessageLocation{{
			Partition: msg.Partition(),
			Offset:    msg.Offset(),
		}}}, nil
	}
}

func (kp *KafkaProducer) Close() error {
	return kp.producer.Close()
}

func NewProducerConfig(cfg *KafkaProducerConfig) *sarama.ProducerConfig {
	producerConfig := new(sarama.ProducerConfig)

	if cfg.PartitionerStrategy == "Random" {
		producerConfig.Partitioner = sarama.NewRandomPartitioner
	} else if cfg.PartitionerStrategy == "RoundRobin" {
		producerConfig.Partitioner = sarama.NewRoundRobinPartitioner
	} else {
		producerConfig.Partitioner = sarama.NewHashPartitioner
	}

	if cfg.WaitAckStrategy == "NoRespond" {
		producerConfig.RequiredAcks = sarama.NoResponse
	} else if cfg.WaitAckStrategy == "WaitForAll" {
		producerConfig.RequiredAcks = sarama.WaitForAll
	} else {
		producerConfig.RequiredAcks = sarama.WaitForLocal
	}

	producerConfig.Timeout = cfg.WaitAckTimeoutMs

	if cfg.CompressionStrategy == "None" {
		producerConfig.Compression = sarama.CompressionNone
	} else if cfg.CompressionStrategy == "Gzip" {
		producerConfig.Compression = sarama.CompressionGZIP
	} else {
		producerConfig.Compression = sarama.CompressionSnappy
	}

	producerConfig.MaxMessageBytes = cfg.MaxMessageBytes
	producerConfig.ChannelBufferSize = cfg.ChannelBufferSize
	producerConfig.AckSuccesses = true

	return producerConfig
}
