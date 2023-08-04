package push

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/IBM/sarama"
	"quote/common/log"
	"quote/common/model"
	"quote/push_server/global"
)

func ConsumeMsgHandler(msg *sarama.ConsumerMessage) error {
	var stock model.StockDetail
	if msg == nil || msg.Value == nil {
		log.Errorf("wrong message")
		return errors.New("wrong message")
	}
	err := json.Unmarshal(msg.Value, &stock)
	if err != nil {
		log.Errorf("%+v", err)
		return err
	}
	// for debug
	return nil
	clients, err := global.RedisClient.GetSet(context.TODO(), stock.Symbol)
	if err != nil {
		log.Errorf("%+v", err)
		return err
	}
	go func() {
		var arrDel []string
		for _, item := range clients {
			var err1 error
			gate, err1 := global.RedisClient.Get(context.TODO(), item)
			if err1 != nil {
				log.Errorf("%+v", err1)
				arrDel = append(arrDel, item)
				continue
			}
			GGateServers.PutMsg(gate, item, msg.Value)
		}
		for _, item := range arrDel {
			global.RedisClient.SRem(context.TODO(), stock.Symbol, item)
		}
	}()
	return nil
}
