package global

import (
	"github.com/google/uuid"
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
	ClientID    int
)

func init() {
	ClientID = int(uuid.New().ID())
}
