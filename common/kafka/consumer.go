package kafka

import (
	"context"
	"github.com/IBM/sarama"
)

type Consumer struct {
	Topic  string
	Client sarama.ConsumerGroup
	Group  string
}

type funcKafkaConsumeMsgHandler func(msg *sarama.ConsumerMessage) error

type QuoteConsumerGroupHandler struct {
	MsgHandler funcKafkaConsumeMsgHandler
}

func (g QuoteConsumerGroupHandler) SetMsgHandler(h funcKafkaConsumeMsgHandler) {
	g.MsgHandler = h
}

func (g QuoteConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}
func (g QuoteConsumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}
func (g QuoteConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg := <-claim.Messages():
			err := g.MsgHandler(msg)
			if err != nil {
				return err
			}
			session.MarkMessage(msg, "")
		}
	}
	return nil
}

func CreateConsumer(c Config) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	client, err := sarama.NewConsumerGroup([]string{c.Addr}, c.Group, sarama.NewConfig())
	if err != nil {
		return nil, err
	}
	return &Consumer{Client: client, Topic: c.Topic, Group: c.Group}, nil
}

func (c *Consumer) Consume(gh QuoteConsumerGroupHandler) error {
	for {
		err := c.Client.Consume(context.TODO(), []string{c.Topic}, gh)
		if err != nil {
			return err
		}
	}
}
