package global

import (
	"fmt"
	"github.com/janqii/mqproxy/utils"
	"gitlab.baidu.com/go/sarama"
	"os"
)

func NewKafkaClient(zkClient *utils.ZK) (*sarama.Client, error) {
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

	var client *sarama.Client
	if client, err = sarama.NewClient(clientName, brokerList, nil); err != nil {
		return nil, err
	}

	KafkaClient = client

	return client, nil
}
