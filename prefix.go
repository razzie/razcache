package razcache

import (
	"time"
)

type prefixCache struct {
	cache  Cache
	prefix string
}

func NewPrefixCache(cache Cache, prefix string) Cache {
	return &prefixCache{
		cache:  cache,
		prefix: prefix,
	}
}

func (c *prefixCache) Set(key, value string, ttl time.Duration) error {
	return c.cache.Set(c.prefix+key, value, ttl)
}

func (c *prefixCache) Get(key string) (string, error) {
	return c.cache.Get(c.prefix + key)
}

func (c *prefixCache) Del(key string) error {
	return c.cache.Del(c.prefix + key)
}

func (c *prefixCache) GetTTL(key string) (time.Duration, error) {
	return c.cache.GetTTL(c.prefix + key)
}

func (c *prefixCache) SetTTL(key string, ttl time.Duration) error {
	return c.cache.SetTTL(c.prefix+key, ttl)
}

func (c *prefixCache) LPush(key string, values ...string) error {
	return c.cache.LPush(c.prefix+key, values...)
}

func (c *prefixCache) RPush(key string, values ...string) error {
	return c.cache.RPush(c.prefix+key, values...)
}

func (c *prefixCache) LPop(key string, count int) ([]string, error) {
	return c.cache.LPop(c.prefix+key, count)
}

func (c *prefixCache) RPop(key string, count int) ([]string, error) {
	return c.cache.RPop(c.prefix+key, count)
}

func (c *prefixCache) LLen(key string) (int, error) {
	return c.cache.LLen(c.prefix + key)
}

func (c *prefixCache) LRange(key string, start, stop int) ([]string, error) {
	return c.cache.LRange(c.prefix+key, start, stop)
}

func (c *prefixCache) SAdd(key string, values ...string) error {
	return c.cache.SAdd(c.prefix+key, values...)
}

func (c *prefixCache) SRem(key string, values ...string) error {
	return c.cache.SRem(c.prefix+key, values...)
}

func (c *prefixCache) SHas(key, value string) (bool, error) {
	return c.cache.SHas(c.prefix+key, value)
}

func (c *prefixCache) SLen(key string) (int, error) {
	return c.cache.SLen(c.prefix + key)
}

func (c *prefixCache) Incr(key string, increment int64) (int64, error) {
	return c.cache.Incr(c.prefix+key, increment)
}

func (c *prefixCache) SubCache(prefix string) Cache {
	return NewPrefixCache(c, prefix)
}
