package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"math/rand"
	"os"
	"os/signal"
	"quote/common/log"
	"quote/common/model"
	"sort"
	"sync"
	"time"
)

const (
	gateSvrAddr = "ws://127.0.0.1:8080/gate"
)

var stockChoices = []string{"00700", "01024", "02318", "01070", "00763", "03690", "03888", "01810", "09988", "00941",
	"00939", "00386", "01398", "03988", "01357", "02007", "02202", "02628", "02238", "00489"}

var subStocks = make(map[string]*model.StockDetail)
var unsubStocks = make(map[string]*model.StockDetail)

var lock sync.RWMutex
var uidGenerator uuid.UUID
var random *rand.Rand

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	chMsg2Gate := make(chan []byte)
	uidGenerator = uuid.New()
	s := rand.NewSource(time.Now().UnixNano())
	random = rand.New(s)

	for _, item := range stockChoices {
		unsubStocks[item] = nil
	}

	c, _, err := websocket.DefaultDialer.Dial(gateSvrAddr, nil)
	if err != nil {
		log.Fatalf("%+v", err)
		return
	}
	defer c.Close()

	// receive message from gate
	go func() {
		for {
			_, data, err1 := c.ReadMessage()
			if err1 != nil {
				log.Fatalf("%+v", err1)
				return
			}
			err1 = msgHandler(data)
			if err1 != nil {
				log.Fatalf("%+v", err1)
				return
			}
		}
	}()

	// send message to gate
	go func() {
		for {
			select {
			case data := <-chMsg2Gate:
				if err2 := c.WriteMessage(websocket.TextMessage, data); err2 != nil {
					log.Fatalf("%+v", err2)
					return
				}
			}
		}
	}()

	// send heartbeat to gate
	go func() {
		ticker := time.NewTicker(time.Second * 10)
		for {
			select {
			case <-ticker.C:
				hb := model.MsgStruct{
					MsgID: model.MsgIDHeartBeat,
				}
				hbData, _ := json.Marshal(hb)
				chMsg2Gate <- hbData
			}
		}
		ticker.Stop()
	}()

	// subscribe
	go func() {
		ticker := time.NewTicker(time.Second * 15)
		for {
			select {
			case <-ticker.C:
				stock, flag := getSubscribeStock()
				var msg model.MsgStruct
				if flag {
					msg.Data, _ = json.Marshal(model.SubscribeReqData{ClientID: fmt.Sprintf("%d", uidGenerator.ID()),
						StockID: stock})
					msg.MsgID = model.MsgIDSubscribe
				} else {
					msg.Data, _ = json.Marshal(model.UnsubscribeReqData{ClientID: fmt.Sprintf("%d", uidGenerator.ID()),
						StockID: stock})
					msg.MsgID = model.MsgIDUnsubscribe
				}
				var wrapperMsg model.MsgStruct
				wrapperMsg.MsgID = model.MsgIDClientAndPush
				wrapperMsg.Data, _ = json.Marshal(msg)
				data, _ := json.Marshal(wrapperMsg)
				chMsg2Gate <- data
			}
		}
		ticker.Stop()
	}()

	// print stock
	go func() {
		ticker := time.NewTicker(time.Second * 10)
		for {
			select {
			case <-ticker.C:
				var arrStocks []string
				stocksCopy := cloneSubStocks()
				for k, _ := range stocksCopy {
					arrStocks = append(arrStocks, k)
				}
				sort.Strings(arrStocks)
				var content string
				for _, item := range arrStocks {
					content += fmt.Sprintf("%s\t\t%s\t\t%s\t\t%s\n", stocksCopy[item].Symbol, stocksCopy[item].Name,
						stocksCopy[item].LastTrade, stocksCopy[item].ChangePercent)
				}
				if content != "" {
					log.Info(content)
				}
			}
		}
		ticker.Stop()
	}()

	select {
	case <-interrupt:
		log.Infof("process quit")
	}
}

func cloneSubStocks() map[string]*model.StockDetail {
	lock.RLock()
	defer lock.RUnlock()
	ret := make(map[string]*model.StockDetail)
	for k, v := range subStocks {
		if v != nil {
			ret[k] = v
		}
	}
	return ret
}

func msgHandler(data []byte) error {
	var msg model.MsgStruct
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Errorf("%+v", err)
		return err
	}
	switch msg.MsgID {
	case model.MsgIDHeartBeat:
		log.Info("heartbeat response")
	case model.MsgIDStock:
		updateStock(msg.Data)
	default:
		log.Info("unknown msg")
	}
	return nil
}

func updateStock(data []byte) {
	var stockDetail model.StockDetail
	err := json.Unmarshal(data, &stockDetail)
	if err != nil {
		log.Errorf("%+v", err)
		return
	}
	lock.Lock()
	defer lock.Unlock()
	subStocks[stockDetail.Symbol] = &stockDetail
}

func getSubscribeStock() (string, bool) {
	lock.RLock()
	defer lock.RUnlock()
	if len(subStocks) >= 10 {
		idx := random.Intn(len(subStocks))
		i := 0
		for k, _ := range subStocks {
			if i == idx {
				delete(subStocks, k)
				unsubStocks[k] = nil
				return k, false
			}
			i++
		}
	} else if len(subStocks) == 0 {
		idx := random.Intn(len(unsubStocks))
		i := 0
		for k, _ := range unsubStocks {
			if i == idx {
				delete(unsubStocks, k)
				subStocks[k] = nil
				return k, true
			}
			i++
		}
	} else {
		idx := random.Intn(1)
		if idx == 1 {
			idx = random.Intn(len(unsubStocks))
			i := 0
			for k, _ := range unsubStocks {
				if i == idx {
					delete(unsubStocks, k)
					subStocks[k] = nil
					return k, true
				}
				i++
			}
		} else {
			idx = random.Intn(len(subStocks))
			i := 0
			for k, _ := range subStocks {
				if i == idx {
					delete(subStocks, k)
					unsubStocks[k] = nil
					return k, false
				}
				i++
			}
		}
	}
	return "", false
}
