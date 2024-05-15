package redis

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/razzie/razcache"
	"github.com/razzie/razcache/internal/util"
	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	client redis.Cmdable
}

func NewRedisCache(redisDSN string) (razcache.ExtendedCache, error) {
	opts, err := redis.ParseURL(redisDSN)
	if err != nil {
		return nil, err
	}
	return &redisCache{
		client: redis.NewClient(opts),
	}, nil
}

func NewRedisCacheFromClient(client redis.Cmdable) razcache.ExtendedCache {
	return &redisCache{
		client: client,
	}
}

func (c *redisCache) Set(key, value string, ttl time.Duration) error {
	err := c.client.Set(context.Background(), key, value, ttl).Err()
	return translateRedisError(err)
}

func (c *redisCache) Get(key string) (string, error) {
	result, err := c.client.Get(context.Background(), key).Result()
	return result, translateRedisError(err)
}

func (c *redisCache) Del(key string) error {
	err := c.client.Del(context.Background(), key).Err()
	return translateRedisError(err)
}

func (c *redisCache) GetTTL(key string) (time.Duration, error) {
	result, err := c.client.TTL(context.Background(), key).Result()
	return result, translateRedisError(err)
}

func (c *redisCache) SetTTL(key string, ttl time.Duration) error {
	err := c.client.Expire(context.Background(), key, ttl).Err()
	return translateRedisError(err)
}

func (c *redisCache) LPush(key string, values ...string) error {
	err := c.client.LPush(context.Background(), key, util.StringToAnySlice(values)...).Err()
	return translateRedisError(err)
}

func (c *redisCache) RPush(key string, values ...string) error {
	err := c.client.RPush(context.Background(), key, util.StringToAnySlice(values)...).Err()
	return translateRedisError(err)
}

func (c *redisCache) LPop(key string, count int) ([]string, error) {
	result, err := c.client.LPopCount(context.Background(), key, count).Result()
	return result, translateRedisError(err)
}

func (c *redisCache) RPop(key string, count int) ([]string, error) {
	result, err := c.client.RPopCount(context.Background(), key, count).Result()
	return result, translateRedisError(err)
}

func (c *redisCache) LLen(key string) (int, error) {
	result, err := c.client.LLen(context.Background(), key).Result()
	return int(result), translateRedisError(err)
}

func (c *redisCache) LRange(key string, start, stop int) ([]string, error) {
	result, err := c.client.LRange(context.Background(), key, int64(start), int64(stop)).Result()
	return result, translateRedisError(err)
}

func (c *redisCache) SAdd(key string, values ...string) error {
	err := c.client.SAdd(context.Background(), key, util.StringToAnySlice(values)...).Err()
	return translateRedisError(err)
}

func (c *redisCache) SRem(key string, values ...string) error {
	err := c.client.SRem(context.Background(), key, util.StringToAnySlice(values)...).Err()
	return translateRedisError(err)
}

func (c *redisCache) SHas(key, value string) (bool, error) {
	result, err := c.client.SIsMember(context.Background(), key, value).Result()
	return result, translateRedisError(err)
}

func (c *redisCache) SLen(key string) (int, error) {
	result, err := c.client.SCard(context.Background(), key).Result()
	return int(result), translateRedisError(err)
}

func (c *redisCache) Incr(key string, increment int64) (int64, error) {
	result, err := c.client.IncrBy(context.Background(), key, increment).Result()
	return result, translateRedisError(err)
}

func (c *redisCache) SubCache(prefix string) razcache.Cache {
	return razcache.NewPrefixCache(c, prefix)
}

func (c *redisCache) SubExtendedCache(prefix string) razcache.ExtendedCache {
	return razcache.NewPrefixExtendedCache(c, prefix)
}

func (c *redisCache) Close() error {
	if client, ok := c.client.(io.Closer); ok {
		return client.Close()
	}
	return nil
}

func translateRedisError(err error) error {
	switch {
	case err == nil:
		return nil
	case err == redis.Nil:
		return razcache.ErrNotFound
	case strings.HasPrefix(err.Error(), "WRONGTYPE"):
		return razcache.ErrWrongType
	default:
		return err
	}
}
