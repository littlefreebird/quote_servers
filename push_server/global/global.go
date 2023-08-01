package global

import (
	"quote/common/etcd"
	"quote/common/kafka"
	"quote/common/redis"
	"quote/push_server/config"
)

var (
	GConfig     *config.GlobalConfig
	RedisClient *redis.Redis
	Register    *etcd.ServiceRegister
	Consumer    *kafka.Consumer
)
