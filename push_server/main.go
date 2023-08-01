package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"quote/common/etcd"
	"quote/common/kafka"
	"quote/common/log"
	"quote/common/model"
	"quote/common/redis"
	"quote/push_server/config"
	"quote/push_server/global"
	"quote/push_server/push"
	"time"
)

var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	var cf string
	flag.StringVar(&cf, "f", "config/config.yaml", "config file")
	flag.Parse()
	log.Infof("config file is :%s\n", cf)

	var err error
	if global.GConfig, err = config.Parse(cf); err != nil {
		log.Fatalf("%+v", err)
		return
	}
	if err = log.SetupWithConfig(global.GConfig.Log); err != nil {
		log.Fatalf("%+v", err)
		return
	}
	global.RedisClient = redis.NewRedisClient(redis.Config{Addr: global.GConfig.Redis.Addr})

	go func() {
		time.Sleep(time.Second)
		a := model.ServerAddr{IP: global.GConfig.IP, Port: global.GConfig.Port}
		data, err1 := json.Marshal(a)
		if err1 != nil {
			log.Fatalf("%+v", err1)
			return
		}
		global.Register, err1 = etcd.NewServiceRegister([]string{global.GConfig.Etcd.Addr}, fmt.Sprintf("%s:%d", global.GConfig.IP, global.GConfig.Port),
			string(data), 10)
		if err1 != nil {
			log.Fatalf("%+v", err1)
			return
		}
		for item := range global.Register.GetLeaseRspChan() {
			log.Infof("etcd response: %+v", item)
		}
	}()

	global.Consumer, err = kafka.CreateConsumer(global.GConfig.Kafka)
	if err != nil {
		log.Fatalf("%+v", err)
		return
	}
	consumeHandler := kafka.QuoteConsumerGroupHandler{
		MsgHandler: push.ConsumeMsgHandler,
	}
	go global.Consumer.Consume(consumeHandler)

	push.GGateServers = push.NewGateServers()
	go push.GGateServers.PushMsg()

	http.HandleFunc("/push", msgHandler)
	http.ListenAndServe(fmt.Sprintf("%s:%d", global.GConfig.IP, global.GConfig.Port), nil)
}

func msgHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("%+v", err)
		return
	}
	defer ws.Close()
	for {
		_, data, err1 := ws.ReadMessage()
		if err != nil {
			log.Errorf("%+v", err1)
			break
		}
		var msg model.MsgStruct
		err1 = json.Unmarshal(data, &msg)
		if err != nil {
			log.Errorf("%+v", err1)
			break
		}
		switch msg.MsgID {
		case model.MsgIDSubscribe:
			var s model.SubscribeReqData
			err1 = json.Unmarshal(msg.Data, &s)
			if err1 != nil {
				log.Errorf("%+v", err1)
				continue
			}
			global.RedisClient.SAdd(context.TODO(), s.StockID, s.ClientID)
		case model.MsgIDUnsubscribe:
			var us model.UnsubscribeReqData
			err1 = json.Unmarshal(msg.Data, &us)
			if err1 != nil {
				log.Errorf("%+v", err1)
				continue
			}
			global.RedisClient.SRem(context.TODO(), us.StockID, us.ClientID)
		default:
			log.Info("unknown msg")
		}
	}
}
