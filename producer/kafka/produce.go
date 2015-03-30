package producer

import (
	"github.com/janqii/mqproxy/global"
	"gitlab.baidu.com/go/sarama"
	"gopkg.in/vmihailenco/msgpack.v2"
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
	Partition int
	Offset    int
}

func SendMessage(req Request) (Response, error) {
	producer, err := sarama.NewSimpleProducer(global.KafkaClient, req.Topic, sarama.NewHashPartitioner)
	if err != nil {
		return Response{-1, err.Error(), make([]MessageLocation, 0)}, err
	}
	defer producer.Close()

	b, err := msgpack.Marshal(map[string]interface{}{"Data": req.Data})
	if err != nil {
		return Response{-2, err.Error(), make([]MessageLocation, 0)}, err
	}

	err = producer.SendMessage(sarama.StringEncoder(req.PartitionKey), sarama.ByteEncoder(b))
	if err != nil {
		return Response{-2, err.Error(), make([]MessageLocation, 0)}, err
	}

	return Response{0, "", make([]MessageLocation, 0)}, nil
}
