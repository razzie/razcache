package inmem

import (
	"strconv"
	"sync/atomic"
	"time"

	"github.com/razzie/razcache"
	"github.com/razzie/razcache/internal/util"

	"github.com/puzpuzpuz/xsync/v3"
)

type extCacheItem struct {
	cacheItemBase
	value atomic.Pointer[any]
}

func newExtCacheItem(value any) *extCacheItem {
	item := new(extCacheItem)
	item.setValue(value)
	return item
}

func (item *extCacheItem) setValue(value any) {
	item.value.Store(&value)
}

func (item *extCacheItem) getValue() any {
	value := item.value.Load()
	return *value
}

type inMemExtCache struct {
	inMemCacheBase[*extCacheItem]
}

func NewInMemExtendedCache() razcache.ExtendedCache {
	cache := new(inMemExtCache)
	cache.init()
	return cache
}

func (c *inMemExtCache) Set(key, value string, ttl time.Duration) error {
	item := newExtCacheItem(value)
	return c.set(key, item, ttl)
}

func (c *inMemExtCache) Get(key string) (string, error) {
	item, err := c.get(key)
	if err != nil {
		return "", err
	}
	switch value := item.getValue().(type) {
	case string:
		return value, nil
	case *int64:
		return strconv.FormatInt(*value, 10), nil
	default:
		return "", razcache.ErrWrongType
	}
}

func (c *inMemExtCache) getList(key string) (*util.List[string], error) {
	item, _, err := c.getOrCompute(key, func() *extCacheItem {
		return newExtCacheItem(new(util.List[string]))
	})
	if err != nil {
		return nil, err
	}
	if value, ok := item.getValue().(*util.List[string]); ok {
		return value, nil
	}
	return nil, razcache.ErrWrongType
}

func (c *inMemExtCache) LPush(key string, values ...string) error {
	list, err := c.getList(key)
	if err != nil {
		return err
	}
	list.PushFront(values...)
	return nil
}

func (c *inMemExtCache) RPush(key string, values ...string) error {
	list, err := c.getList(key)
	if err != nil {
		return err
	}
	list.PushBack(values...)
	return nil
}

func (c *inMemExtCache) LPop(key string, count int) ([]string, error) {
	list, err := c.getList(key)
	if err != nil {
		return nil, err
	}
	return list.PopFront(count), nil
}

func (c *inMemExtCache) RPop(key string, count int) ([]string, error) {
	list, err := c.getList(key)
	if err != nil {
		return nil, err
	}
	return list.PopBack(count), nil
}

func (c *inMemExtCache) LLen(key string) (int, error) {
	list, err := c.getList(key)
	if err != nil {
		return 0, err
	}
	return list.Len(), nil
}

func (c *inMemExtCache) LRange(key string, start, stop int) ([]string, error) {
	list, err := c.getList(key)
	if err != nil {
		return nil, err
	}
	return list.Range(start, stop), nil
}

func (c *inMemExtCache) getSet(key string) (*xsync.Map, error) {
	item, _, err := c.getOrCompute(key, func() *extCacheItem {
		return newExtCacheItem(xsync.NewMap())
	})
	if err != nil {
		return nil, err
	}
	if value, ok := item.getValue().(*xsync.Map); ok {
		return value, nil
	}
	return nil, razcache.ErrWrongType
}

func (c *inMemExtCache) SAdd(key string, values ...string) error {
	set, err := c.getSet(key)
	if err != nil {
		return err
	}
	for _, value := range values {
		set.Store(value, true)
	}
	return nil
}

func (c *inMemExtCache) SRem(key string, values ...string) error {
	set, err := c.getSet(key)
	if err != nil {
		return err
	}
	for _, value := range values {
		set.Delete(value)
	}
	return nil
}

func (c *inMemExtCache) SHas(key, value string) (bool, error) {
	set, err := c.getSet(key)
	if err != nil {
		return false, err
	}
	_, found := set.Load(value)
	return found, nil
}

func (c *inMemExtCache) SLen(key string) (int, error) {
	set, err := c.getSet(key)
	if err != nil {
		return 0, err
	}
	return set.Size(), nil
}

func (c *inMemExtCache) Incr(key string, increment int64) (int64, error) {
	item, loaded, err := c.getOrCompute(key, func() *extCacheItem {
		return newExtCacheItem(&increment)
	})
	if err != nil {
		return 0, err
	}
	if !loaded {
		return increment, nil
	}
	for {
		oldVal := item.value.Load()
		switch value := (*oldVal).(type) {
		case string:
			i, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return 0, razcache.ErrWrongType
			}
			i += increment
			var newVal any = &i
			if !item.value.CompareAndSwap(oldVal, &newVal) {
				continue
			}
			return i, nil
		case *int64:
			return atomic.AddInt64(value, increment), nil
		default:
			return 0, razcache.ErrWrongType
		}
	}
}

func (c *inMemExtCache) SubCache(prefix string) razcache.Cache {
	return razcache.NewPrefixCache(c, prefix)
}

func (c *inMemExtCache) SubExtendedCache(prefix string) razcache.ExtendedCache {
	return razcache.NewPrefixExtendedCache(c, prefix)
}
