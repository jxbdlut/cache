package cache

import (
	"github.com/go-redis/redis"
	"time"
)

// 新建 Redis 连接
func NewRedisCache(redisHost, password string, db int) RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr: redisHost,
		Password: password,
		DB: db,
		PoolSize: 10,
	})
	return RedisCache{client}
}

type RedisCache struct {
	*redis.Client
}

func (c RedisCache) Set(key string, v interface{}) (err error) {
	defer c.Close()
	_, err = c.Client.Set(key, v, 0).Result()
	return err
}

func (c RedisCache) Expire(key string, expiration time.Duration) (bool, error) {
	return c.Client.Expire(key, expiration).Result()
}

func (c RedisCache) Get(key string) ([]byte, error) {
	str, err := c.Client.Get(key).Result()
	return []byte(str), err
}

func (c RedisCache) Keys(key string) ([]string, error) {
	result, err := c.Client.Keys(key).Result()
	return result, err
}

func (c RedisCache) Exists(key string) (bool, error) {
	result, err := c.Client.Exists(key).Result()
	log.Debugf("result:%v, err:%v", result, err)
	if result == 1 {
		return true, err
	}
	return false, err
}

func (c RedisCache) HGetAll(key string) (ret interface{}, err error) {
	return c.Client.HGetAll(key).Result()
}

func (c RedisCache) HSet(key string, field string, v interface{}) (bool, error) {
	return c.Client.HSet(key, field, v).Result()
}

func (c RedisCache) HMSet(key string, fields map[string]interface{}) (string, error) {
	return c.Client.HMSet(key, fields).Result()
}

func (c RedisCache) Del(key string) (int64, error) {
	return c.Client.Del(key).Result()
}

func (c RedisCache) Close() error {
	return c.Client.Close()
}