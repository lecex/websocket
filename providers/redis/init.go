package redis

import (
	"github.com/go-redis/redis"
	"github.com/lecex/user/core/env"
	"github.com/micro/go-micro/v2/util/log"
)

// NewClient 创建新的 redis 连接
func NewClient() *redis.Client {
	addr := env.Getenv("REDIS_HOST", "127.0.0.1") + ":" + env.Getenv("REDIS_PORT", "6379")
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: env.Getenv("REDIS_PASSWORD", ""),
		DB:       0, // use default DB
	})
	pong, err := client.Ping().Result()
	if err != nil {
		log.Log(pong, err)
	}
	return client
}
