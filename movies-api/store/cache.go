package store

import "github.com/go-redis/redis"

type Cache interface {
	Get(key string) string
	Set(key string, value string)
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache() *RedisCache {
	return &RedisCache{
		client: redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		}),
	}
}

func (c *RedisCache) Get(key string) string {
	v, err := c.client.Get(key).Result()

	if err != nil {
		return ""
	}

	return v
}

func (c *RedisCache) Set(key string, value string) {
	if err := c.client.Set(key, value, 0).Err(); err != nil {
		panic(err)
	}
}
