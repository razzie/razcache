package razcache

import (
	"time"
)

type prefixExtCache struct {
	cache  ExtendedCache
	prefix string
}

func NewPrefixExtendedCache(cache ExtendedCache, prefix string) ExtendedCache {
	return &prefixExtCache{
		cache:  cache,
		prefix: prefix,
	}
}

func (c *prefixExtCache) Set(key, value string, ttl time.Duration) error {
	return c.cache.Set(c.prefix+key, value, ttl)
}

func (c *prefixExtCache) Get(key string) (string, error) {
	return c.cache.Get(c.prefix + key)
}

func (c *prefixExtCache) Del(key string) error {
	return c.cache.Del(c.prefix + key)
}

func (c *prefixExtCache) GetTTL(key string) (time.Duration, error) {
	return c.cache.GetTTL(c.prefix + key)
}

func (c *prefixExtCache) SetTTL(key string, ttl time.Duration) error {
	return c.cache.SetTTL(c.prefix+key, ttl)
}

func (c *prefixExtCache) LPush(key string, values ...string) error {
	return c.cache.LPush(c.prefix+key, values...)
}

func (c *prefixExtCache) RPush(key string, values ...string) error {
	return c.cache.RPush(c.prefix+key, values...)
}

func (c *prefixExtCache) LPop(key string, count int) ([]string, error) {
	return c.cache.LPop(c.prefix+key, count)
}

func (c *prefixExtCache) RPop(key string, count int) ([]string, error) {
	return c.cache.RPop(c.prefix+key, count)
}

func (c *prefixExtCache) LLen(key string) (int, error) {
	return c.cache.LLen(c.prefix + key)
}

func (c *prefixExtCache) LRange(key string, start, stop int) ([]string, error) {
	return c.cache.LRange(c.prefix+key, start, stop)
}

func (c *prefixExtCache) SAdd(key string, values ...string) error {
	return c.cache.SAdd(c.prefix+key, values...)
}

func (c *prefixExtCache) SRem(key string, values ...string) error {
	return c.cache.SRem(c.prefix+key, values...)
}

func (c *prefixExtCache) SHas(key, value string) (bool, error) {
	return c.cache.SHas(c.prefix+key, value)
}

func (c *prefixExtCache) SLen(key string) (int, error) {
	return c.cache.SLen(c.prefix + key)
}

func (c *prefixExtCache) Incr(key string, increment int64) (int64, error) {
	return c.cache.Incr(c.prefix+key, increment)
}

func (c *prefixExtCache) SubCache(prefix string) Cache {
	return NewPrefixCache(c, prefix)
}

func (c *prefixExtCache) SubExtendedCache(prefix string) ExtendedCache {
	return NewPrefixExtendedCache(c, prefix)
}
