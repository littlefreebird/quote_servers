package relay

import (
	"github.com/gorilla/websocket"
	"math/rand"
	"quote/common/log"
	"sync"
)

type PushServers struct {
	Servers    map[string]*websocket.Conn
	CHMsg2Push chan []byte
	lock       sync.RWMutex
	rand       *rand.Rand
}

func NewPushServers() *PushServers {
	ret := &PushServers{
		Servers:    make(map[string]*websocket.Conn),
		CHMsg2Push: make(chan []byte, 1024),
	}
	ret.rand.Seed(0)
	return ret
}

func (s *PushServers) Put(a string, c *websocket.Conn) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Servers[a] = c
}

func (s *PushServers) Del(a string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.Servers, a)
}

func (s *PushServers) Get() *websocket.Conn {
	s.lock.RLock()
	defer s.lock.RUnlock()
	idx := s.rand.Intn(len(s.Servers))
	i := 0
	for _, v := range s.Servers {
		if i == idx {
			return v
		}
		i++
	}
	return nil
}

func (s *PushServers) PutMsg(data []byte) {
	s.CHMsg2Push <- data
}

func (s *PushServers) RelayUpMsg() {
	for {
		select {
		case data := <-s.CHMsg2Push:
			c := s.Get()
			if c == nil {
				log.Info("no push server connected")
			} else {
				err := c.WriteMessage(websocket.TextMessage, data)
				if err != nil {
					log.Errorf("%+v", err)
				}
			}
		}
	}
}
