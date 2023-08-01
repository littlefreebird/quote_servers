package relay

import (
	"github.com/gorilla/websocket"
	"quote/common/log"
	"sync"
)

type ChanMsg struct {
	Client string
	Data   []byte
}

type Clients struct {
	clients      map[string]*websocket.Conn
	CHMsg2Client chan ChanMsg
	lock         sync.RWMutex
}

func (cs *Clients) Put(c string, conn *websocket.Conn) {
	cs.lock.Lock()
	defer cs.lock.Unlock()
	cs.clients[c] = conn
}

func (cs *Clients) Get(c string) *websocket.Conn {
	cs.lock.RLock()
	cs.lock.RUnlock()
	if _, ok := cs.clients[c]; !ok {
		return nil
	}
	return cs.clients[c]
}

func (cs *Clients) Del(c string) {
	cs.lock.Lock()
	defer cs.lock.Unlock()
	delete(cs.clients, c)
}

func (cs *Clients) PutMsg(client string, data []byte) {
	cs.CHMsg2Client <- ChanMsg{
		Client: client,
		Data:   data,
	}
}

func (cs *Clients) RelayDownMsg() {
	for {
		select {
		case msg := <-cs.CHMsg2Client:
			c := cs.Get(msg.Client)
			if c == nil {
				log.Infof("%s disconnected", msg.Client)
			} else {
				err := c.WriteMessage(websocket.TextMessage, msg.Data)
				if err != nil {
					log.Errorf("%+v", err)
				}
			}
		}
	}
}

func NewClients() *Clients {
	return &Clients{
		clients:      make(map[string]*websocket.Conn),
		CHMsg2Client: make(chan ChanMsg, 1024),
	}
}
