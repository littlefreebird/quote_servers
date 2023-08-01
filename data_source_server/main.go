package main

import (
	"encoding/json"
	"flag"
	"quote/common/etcd"
	"quote/common/kafka"
	"quote/common/log"
	"quote/data_source_server/config"
	"quote/data_source_server/global"
	"quote/data_source_server/pull"
)

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

	global.Locker, err = etcd.CreateLocker(global.GConfig.Etcd.Addr, 30, "/quote/data_source")
	if err != nil {
		log.Fatalf("%+v", err)
		return
	}
	global.Locker.Lock()

	global.Producer, err = kafka.CreateProducer(global.GConfig.Kafka)
	if err != nil {
		log.Fatalf("%+v", err)
		return
	}

	go func() {
		pull.GetStockList()
	}()

	for {
		select {
		case msg := <-global.CHSendMsg2Push:
			data, _ := json.Marshal(msg)
			err1 := global.Producer.SendMessage(data)
			if err1 != nil {
				log.Errorf("%+v", err1)
			}
		}
	}

	global.Locker.Unlock()
}
