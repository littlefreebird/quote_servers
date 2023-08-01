package push

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"quote/common/log"
	"quote/common/model"
	"sync"
)

type ChanMsg struct {
	Receiver string
	Client   string
	Data     []byte
}

type GateServers struct {
	Gates      map[string]*websocket.Conn
	CHMsg2Gate chan ChanMsg
	lock       sync.RWMutex
}

var GGateServers *GateServers

func NewGateServers() *GateServers {
	return &GateServers{
		Gates:      make(map[string]*websocket.Conn),
		CHMsg2Gate: make(chan ChanMsg, 1024),
	}
}

func (s *GateServers) Put(gate string, conn *websocket.Conn) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Gates[gate] = conn
}

func (s *GateServers) Del(gate string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.Gates, gate)
}

func (s *GateServers) Get(gate string) *websocket.Conn {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if _, ok := s.Gates[gate]; !ok {
		return nil
	}
	return s.Gates[gate]
}

func (s *GateServers) PutMsg(receiver string, client string, data []byte) {
	s.CHMsg2Gate <- ChanMsg{
		Receiver: receiver,
		Client:   client,
		Data:     data,
	}
}

func (s *GateServers) PushMsg() {
	for {
		select {
		case msg := <-s.CHMsg2Gate:
			g := s.Get(msg.Receiver)
			if g == nil {
				log.Infof("disconnected with gate : %s", msg.Receiver)
			} else {
				wrapperMsg := model.MsgStruct{MsgID: model.MsgIDStock, Data: msg.Data}
				wrapperData, _ := json.Marshal(wrapperMsg)
				pushData := model.PushMsgStruct{
					MsgID:    model.MsgIDClientAndPush,
					Receiver: msg.Client,
					Data:     wrapperData,
				}
				data, _ := json.Marshal(pushData)
				err := g.WriteMessage(websocket.TextMessage, data)
				if err != nil {
					log.Errorf("%+v", err)
				}
			}
		}
	}
}
