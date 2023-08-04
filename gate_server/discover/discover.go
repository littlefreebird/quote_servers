package discover

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"quote/common/log"
	"quote/common/model"
	"quote/gate_server/global"
)

const (
	PushServerPrefix = "/quote/push/"
)

func PutHandler(k, v string) error {
	go func() {
		header := http.Header{}
		header.Add("ClientID", fmt.Sprintf("%d", global.ClientID))
		c, _, err := websocket.DefaultDialer.Dial(v, header)
		if err != nil {
			log.Fatalf("%+v", err)
			return
		}
		log.Infof("connected with %s", c.RemoteAddr().String())
		global.PushServers.Put(v, c)
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
	global.PushServers.Del(v)
	return nil
}
