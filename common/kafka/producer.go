package kafka

import (
	"github.com/IBM/sarama"
	"quote/common/log"
)

type Producer struct {
	Client sarama.SyncProducer
	Topic  string
}

type Config struct {
	Addr  string `yaml:"addr"`
	Topic string `yaml:"topic"`
	Group string `yaml:"group"`
}

func CreateProducer(c Config) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	p, err := sarama.NewSyncProducer([]string{c.Addr}, config)
	if err != nil {
		return nil, err
	}
	return &Producer{
		Client: p,
		Topic:  c.Topic,
	}, nil
}

func (p *Producer) Close() {
	p.Close()
}

func (p *Producer) SendMessage(data []byte) error {
	msg := &sarama.ProducerMessage{}
	msg.Topic = p.Topic
	msg.Value = sarama.ByteEncoder(data)
	partition, offset, err := p.Client.SendMessage(msg)
	if err == nil {
		log.Infof("produce kafka, topic:%s, partition:%d, offset:%d ", msg.Topic, partition, offset)
	}
	return err
}
