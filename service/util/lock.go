package util

import (
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

// Lock Redis锁
type Lock struct {
	Redis *redis.Client
}

// Set设置锁
func (srv *Lock) Set(key string, expiration time.Duration) bool {
	var mutex sync.Mutex
	mutex.Lock()
	defer mutex.Unlock()
	bool, err := srv.Redis.SetNX(key, 1, expiration).Result()
	if err != nil {
		log.Println(err.Error())
	}
	return bool
}

// Del 删除锁
func (srv *Lock) Del(key string) int64 {
	nums, err := srv.Redis.Del(key).Result()
	if err != nil {
		log.Println(err.Error())
		return 0
	}
	return nums
}
