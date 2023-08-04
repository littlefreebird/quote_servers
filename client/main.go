package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"math/rand"
	"net/http"
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

var stockChoices = []string{
	"00001", "00002", "00003", "00004", "00005", "00006", "00007", "00008", "00009", "00010", "00011", "00012", "00013", "00014",
	"00016", "00017", "00018", "00019", "00020", "00021", "00022", "00023", "00025", "00026", "00027", "00028", "00029", "00030",
	"00031", "00032", "00033", "00034", "00035", "00036", "00037", "00038", "00039", "00040", "00041", "00042", "00045", "00046",
	"00048", "00050", "00051", "00052", "00053", "00055", "00057", "00058", "00059", "00060", "00061", "00062", "00063", "00064",
	"00065", "00066", "00069", "00070", "00071", "00072", "00073", "00075", "00076", "00077", "00078", "00079", "00080", "00081",
	"00082", "00083", "00084", "00085", "00086", "00087", "00088", "00089", "00090", "00091", "00092", "00093", "00094", "00095",
	"00096", "00097", "00098", "00099", "00101", "00102", "00103", "00104", "00105", "00106", "00107", "00108", "00110", "00111",
	"00113", "00114"}

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
	header := http.Header{}
	header.Add("ClientID", fmt.Sprintf("%d", ClientID))
	c, _, err := websocket.DefaultDialer.Dial(gateSvrAddr, header)
	if err != nil {
		log.Fatalf("%+v", err)
		return
	}
	log.Infof("connected with %s", c.RemoteAddr().String())
	defer c.Close()

	// receive message from gate
	go func() {
		for {
			_, data, err1 := c.ReadMessage()
			if err1 != nil {
				log.Fatalf("%+v", err1)
				return
			}
			if len(data) == 0 {
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
		hbTicker := time.NewTicker(time.Second * 30)
		subTicket := time.NewTicker(time.Second * 10)
		printTicker := time.NewTicker(time.Second * 5)
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
		content += fmt.Sprintf("\n%s\t\t%s\t\t%s\t\t%s\t\t%d\n", subStocks[item].Symbol, subStocks[item].Name,
			subStocks[item].LastTrade, subStocks[item].ChangePercent, subStocks[item].Version)
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
	hb := model.MsgStruct{
		MsgID: model.MsgIDHeartBeat,
		Data:  []byte(fmt.Sprintf("%d", ClientID)),
	}
	hbData, _ := json.Marshal(hb)
	err := ws.WriteMessage(websocket.TextMessage, hbData)
	if err != nil {
		log.Errorf("%+v", err)
	}
}

func subscribe(ws *websocket.Conn) {
	var msg model.MsgStruct
	flag := true
	if rand.Float64() > 0.7 {
		flag = false
	}
	if len(subStocks) > 10 {
		flag = false
	}
	if len(subStocks) == 0 {
		flag = true
	}
	if flag {
		stock := stockChoices[rand.Intn(100)]
		msg.Data, _ = json.Marshal(model.SubscribeReqData{ClientID: ClientID, StockID: stock})
		msg.MsgID = model.MsgIDSubscribe
		log.Infof("subscribe %s", stock)
	} else {
		i := 0
		r := rand.Intn(len(subStocks))
		var stock string
		for k, _ := range subStocks {
			if i == r {
				stock = k
				break
			}
			i++
		}
		delete(subStocks, stock)
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
