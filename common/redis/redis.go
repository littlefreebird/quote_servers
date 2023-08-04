package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

type Config struct {
	Addr     string `yaml:"addr""`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type Redis struct {
	Client *redis.Client
}

func NewRedisClient(c Config) *Redis {
	return &Redis{
		Client: redis.NewClient(&redis.Options{
			Addr:     c.Addr,
			Password: c.Password,
			DB:       c.DB,
		}),
	}
}

func (r *Redis) Set(ctx context.Context, key, val string, expire time.Duration) error {
	return r.Client.Set(ctx, key, val, expire).Err()
}

func (r *Redis) Expire(ctx context.Context, key string, expire time.Duration) error {
	return r.Client.ExpireNX(ctx, key, expire).Err()
}

func (r *Redis) SAdd(ctx context.Context, key, val string) error {
	return r.Client.SAdd(ctx, key, val).Err()
}

func (r *Redis) GetSet(ctx context.Context, key string) ([]string, error) {
	var ret []string
	ret, err := r.Client.SMembers(ctx, key).Result()
	return ret, err
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

func (r *Redis) SRem(ctx context.Context, key, val string) error {
	return r.Client.SRem(ctx, key, val).Err()
}
