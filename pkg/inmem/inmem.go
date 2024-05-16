package inmem

import (
	"time"

	"github.com/razzie/razcache"
)

type cacheItem struct {
	cacheItemBase
	value string
}

type inMemCache struct {
	inMemCacheBase[*cacheItem]
}

func NewInMemCache() razcache.Cache {
	cache := new(inMemCache)
	cache.init()
	return cache
}

func (c *inMemCache) Set(key, value string, ttl time.Duration) error {
	item := &cacheItem{value: value}
	return c.set(key, item, ttl)
}

func (c *inMemCache) Get(key string) (string, error) {
	item, err := c.get(key)
	if err != nil {
		return "", err
	}
	return item.value, nil
}

func (c *inMemCache) SubCache(prefix string) razcache.Cache {
	return razcache.NewPrefixCache(c, prefix)
}
