package pull

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"quote/common/log"
	"quote/common/model"
	"quote/data_source_server/global"
	"strconv"
	"time"
)

const (
	GetStockListUrl = "http://web.juhe.cn/finance/stock/hkall"
)

var headers = map[string]string{
	"Content-Type": "application/x-www-form-urlencoded",
}

func GetStockList() {
	/*file, _ := os.OpenFile("data", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0777)
	defer file.Close()
	writer := bufio.NewWriter(file)*/
	page := 1
	for {
		if true {
			break
		}
		params := map[string]string{
			"key":  "9cbaf6d5eb8336c8e1a1b53e4b2824da",
			"type": "4",
		}
		flag := false
		params["page"] = fmt.Sprintf("%d", page)
		page++
		response, err := HttpRequest("GET", GetStockListUrl, params, headers, 15)
		if err != nil {
			log.Errorf("%+v", err)
		} else {
			var rsp model.StockListRsp
			err = json.Unmarshal(response, &rsp)
			if err != nil {
				log.Errorf("%+v", err)
				continue
			}
			if rsp.ErrorCode != 0 {
				log.Error(rsp.Reason)
				continue
			}
			for _, item := range rsp.Result.Data {
				global.CHSendMsg2Push <- item
				/*bytes, _ := json.Marshal(item)
				_, err1 := writer.WriteString(string(bytes) + "\n")
				if err1 != nil {
					log.Errorf("%+v", err1)
				}
				writer.Flush()*/
			}
			num, _ := strconv.Atoi(rsp.Result.Num)
			if num == 0 {
				flag = true
			}
		}
		if flag {
			time.Sleep(time.Second * 5)
			page = 1
		}
	}
	file, _ := os.Open("data1")
	defer file.Close()
	reader := bufio.NewReader(file)
	version := 0
	count := 0
	for {
		count++
		bytes, err := reader.ReadBytes('\n')
		if err == io.EOF {
			file.Seek(0, io.SeekStart)
			count = 0
			version++
			continue
		}
		var stock model.StockDetail
		err = json.Unmarshal(bytes, &stock)
		stock.Version = version
		if err == nil {
			global.CHSendMsg2Push <- stock
		}
		if count%20 == 0 {
			time.Sleep(time.Second)
		}
	}
}
