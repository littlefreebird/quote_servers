package discover

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"quote/common/log"
	"quote/common/model"
	"quote/gate_server/global"
)

const (
	PushServerPrefix = "/quote/push/"
)

func PutHandler(k, v string) error {
	go func() {
		c, _, err := websocket.DefaultDialer.Dial(v, nil)
		if err != nil {
			log.Fatalf("%+v", err)
			return
		}
		defer c.Close()

		// receive message from push
		for {
			_, data, err1 := c.ReadMessage()
			if err1 != nil {
				log.Errorf("%+v", err1)
				return
			}
			if len(data) == 0 {
				break
			}
			var msg model.PushMsgStruct
			err1 = json.Unmarshal(data, &msg)
			if err1 != nil {
				log.Errorf("%+v", err1)
				continue
			}
			if msg.MsgID == model.MsgIDClientAndPush {
				global.Clients.PutMsg(msg.Receiver, msg.Data)
			} else {
				log.Info("unknown message")
			}
		}
	}()
	return nil
}

func DelHandler(k, v string) error {
	var a model.ServerAddr
	err := json.Unmarshal([]byte(v), &a)
	if err != nil {
		log.Errorf("%+v", err)
		return err
	}
	global.PushServers.Del(fmt.Sprintf("%s:%d", a.IP, a.Port))
	return nil
}
