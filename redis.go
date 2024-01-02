package razcache

import (
	"context"
	"time"

	"github.com/razzie/razcache/internal/util"
	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	client *redis.Client
}

func NewRedisCache(redisDSN string) (Cache, error) {
	opts, err := redis.ParseURL(redisDSN)
	if err != nil {
		return nil, err
	}
	return &redisCache{
		client: redis.NewClient(opts),
	}, nil
}

func (c *redisCache) Set(key, value string, ttl time.Duration) error {
	return c.client.Set(context.Background(), key, value, ttl).Err()
}

func (c *redisCache) Get(key string) (string, error) {
	return c.client.Get(context.Background(), key).Result()
}

func (c *redisCache) Del(key string) error {
	return c.client.Del(context.Background(), key).Err()
}

func (c *redisCache) GetTTL(key string) (time.Duration, error) {
	return c.client.TTL(context.Background(), key).Result()
}

func (c *redisCache) SetTTL(key string, ttl time.Duration) error {
	return c.client.Expire(context.Background(), key, ttl).Err()
}

func (c *redisCache) LPush(key string, values ...string) error {
	return c.client.LPush(context.Background(), key, util.StringToAnySlice(values)...).Err()
}

func (c *redisCache) RPush(key string, values ...string) error {
	return c.client.RPush(context.Background(), key, util.StringToAnySlice(values)...).Err()
}

func (c *redisCache) LPop(key string) (string, error) {
	return c.client.LPop(context.Background(), key).Result()
}

func (c *redisCache) RPop(key string) (string, error) {
	return c.client.RPop(context.Background(), key).Result()
}

func (c *redisCache) LLen(key string) (int, error) {
	result, err := c.client.LLen(context.Background(), key).Result()
	return int(result), err
}

func (c *redisCache) SAdd(key string, values ...string) error {
	return c.client.SAdd(context.Background(), key, util.StringToAnySlice(values)...).Err()
}

func (c *redisCache) SRem(key string, values ...string) error {
	return c.client.SRem(context.Background(), key, util.StringToAnySlice(values)...).Err()
}

func (c *redisCache) SHas(key, value string) (bool, error) {
	return c.client.SIsMember(context.Background(), key, value).Result()
}

func (c *redisCache) SLen(key string) (int, error) {
	result, err := c.client.SCard(context.Background(), key).Result()
	return int(result), err
}

func (c *redisCache) SubCache(prefix string) Cache {
	return NewPrefixCache(c, prefix)
}
