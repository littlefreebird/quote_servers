package kafka

import (
	"github.com/IBM/sarama"
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
	p, err := sarama.NewSyncProducer([]string{c.Addr}, sarama.NewConfig())
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
	_, _, err := p.Client.SendMessage(msg)
	return err
}
