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

func (c *prefixCache) SubCache(prefix string) Cache {
	return NewPrefixCache(c, prefix)
}
