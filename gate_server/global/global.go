package global

import (
	"quote/common/etcd"
	"quote/common/redis"
	"quote/gate_server/config"
	"quote/gate_server/relay"
)

var (
	GConfig     *config.GlobalConfig
	RedisClient *redis.Redis
	Clients     *relay.Clients
	Discover    *etcd.ServiceDiscover
	PushServers *relay.PushServers
)
