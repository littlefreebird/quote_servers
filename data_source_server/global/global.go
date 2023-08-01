package global

import (
	"quote/common/etcd"
	"quote/common/kafka"
	"quote/common/model"
	"quote/data_source_server/config"
)

var (
	GConfig        *config.GlobalConfig
	Producer       *kafka.Producer
	Locker         *etcd.QuoteLock
	CHSendMsg2Push chan model.StockDetail
)

func init() {
	CHSendMsg2Push = make(chan model.StockDetail, 1024)
}
