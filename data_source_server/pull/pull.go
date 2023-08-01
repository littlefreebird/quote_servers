package pull

import (
	"encoding/json"
	"fmt"
	"quote/common/log"
	"quote/common/model"
	"quote/data_source_server/global"
	"time"
)

const (
	GetStockListUrl = "http://web.juhe.cn/finance/stock/hkall"
)

var headers = map[string]string{
	"Content-Type": "application/x-www-form-urlencoded",
}

func GetStockList() {
	for {
		params := map[string]string{
			"key":  "",
			"type": "4",
		}
		flag := false
		page := 1
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
			}
			if rsp.Result.Num == 0 {
				flag = true
			}
		}
		if flag {
			time.Sleep(time.Second * 5)
			page = 1
		}
	}
}
