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
	"time"
)

const (
	gateSvrAddr = "ws://127.0.0.1:9700/gate"
)

var stockChoices = []string{"00700", "01024", "02318", "01070", "00763", "03690", "03888", "01810", "09988", "00941",
	"00939", "00386", "01398", "03988", "01357", "02007", "02202", "02628", "02238", "00489"}

var subStocks = make(map[string]*model.StockDetail)
var unsubStocks = make(map[string]*model.StockDetail)
var chUpdate = make(chan []byte)

var random *rand.Rand
var ClientID int

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	ClientID = int(uuid.New().ID())
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

	go func() {
		hbTicker := time.NewTicker(time.Second * 10)
		subTicket := time.NewTicker(time.Second * 15)
		printTicker := time.NewTicker(time.Second * 20)
		for {
			select {
			case <-hbTicker.C:
				sendHeartbeat(c)
			case <-subTicket.C:
				subscribe(c)
			case <-printTicker.C:
				printStock()
			case data := <-chUpdate:
				updateStock(data)
			}
		}
	}()

	select {
	case <-interrupt:
		log.Infof("process quit")
	}
}

func printStock() {
	log.Info("print stocks")
	var arrStocks []string
	for k, _ := range subStocks {
		arrStocks = append(arrStocks, k)
	}
	sort.Strings(arrStocks)
	var content string
	for _, item := range arrStocks {
		if subStocks[item] == nil {
			continue
		}
		content += fmt.Sprintf("%s\t\t%s\t\t%s\t\t%s\n", subStocks[item].Symbol, subStocks[item].Name,
			subStocks[item].LastTrade, subStocks[item].ChangePercent)
	}
	if content != "" {
		log.Info(content)
	}
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
		chUpdate <- msg.Data
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
	subStocks[stockDetail.Symbol] = &stockDetail
}

func sendHeartbeat(ws *websocket.Conn) {
	log.Info("send heartbeat")
	hb := model.MsgStruct{
		MsgID: model.MsgIDHeartBeat,
	}
	hbData, _ := json.Marshal(hb)
	err := ws.WriteMessage(websocket.TextMessage, hbData)
	if err != nil {
		log.Errorf("%+v", err)
	}
}

func subscribe(ws *websocket.Conn) {
	stock, flag := getSubscribeStock()
	var msg model.MsgStruct
	if flag {
		msg.Data, _ = json.Marshal(model.SubscribeReqData{ClientID: ClientID, StockID: stock})
		msg.MsgID = model.MsgIDSubscribe
		log.Infof("subscribe %s", stock)
	} else {
		msg.Data, _ = json.Marshal(model.UnsubscribeReqData{ClientID: ClientID, StockID: stock})
		msg.MsgID = model.MsgIDUnsubscribe
		log.Infof("unsubscribe %s", stock)
	}
	var wrapperMsg model.MsgStruct
	wrapperMsg.MsgID = model.MsgIDClientAndPush
	wrapperMsg.Data, _ = json.Marshal(msg)
	data, _ := json.Marshal(wrapperMsg)
	err := ws.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.Errorf("%+v", err)
	}
}

func getSubscribeStock() (string, bool) {
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
		f := random.Float64()
		if f < 0.7 {
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
		}
	}
	return "", false
}
