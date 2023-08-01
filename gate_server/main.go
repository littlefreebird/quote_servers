package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"quote/common/etcd"
	"quote/common/log"
	"quote/common/model"
	"quote/common/redis"
	"quote/gate_server/config"
	"quote/gate_server/discover"
	"quote/gate_server/global"
	"quote/gate_server/relay"
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
	// relay msg to clients
	global.Clients = relay.NewClients()
	global.Clients.RelayDownMsg()
	// relay msg to push server
	global.PushServers = relay.NewPushServers()
	global.PushServers.RelayUpMsg()
	// push servers discover
	global.Discover, err = etcd.NewServiceDiscover([]string{global.GConfig.Etcd.Addr})
	if err != nil {
		log.Fatalf("%+v", err)
		return
	}
	global.Discover.WatchService(discover.PushServerPrefix, discover.PutHandler, discover.DelHandler)

	// listen relay
	http.HandleFunc("/gate", clientMsgHandler)
	http.ListenAndServe(fmt.Sprintf("%s:%d", global.GConfig.IP, global.GConfig.Port), nil)
}

func clientMsgHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("%+v", err)
		return
	}
	key := fmt.Sprintf("%v", &ws)
	val := fmt.Sprintf("%v", ws.LocalAddr().String())
	global.RedisClient.Set(context.TODO(), key, val, time.Second*60)
	global.Clients.Put(key, ws)
	defer ws.Close()
	for {
		mt, data, err1 := ws.ReadMessage()
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
		case model.MsgIDHeartBeat:
			log.Infof("heartbeat from %v", &ws)
			key = fmt.Sprintf("%v", &ws)
			global.RedisClient.Expire(context.TODO(), key, time.Second*60)
			data, _ = json.Marshal(model.MsgStruct{
				MsgID: model.MsgIDHeartBeat,
			})
			err1 = ws.WriteMessage(mt, data)
			if err1 != nil {
				log.Errorf("%+v", err1)
			}
		case model.MsgIDClientAndPush:
			log.Info("relay message to push server")
			global.PushServers.PutMsg(msg.Data)
		default:
			log.Info("unknown msg")
		}
	}
	global.Clients.Del(key)
}
