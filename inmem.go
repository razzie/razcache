package razcache

import (
	"strconv"
	"sync/atomic"
	"time"

	"github.com/razzie/razcache/internal/util"

	"github.com/puzpuzpuz/xsync/v3"
)

var (
	// special pointer to mark a janitor deleted item
	deletedItem = &util.EQItem{}
)

type cacheItem struct {
	value   atomic.Pointer[any]
	ttlData atomic.Pointer[util.EQItem]
}

func newCacheItem(value any) *cacheItem {
	item := new(cacheItem)
	item.setValue(value)
	return item
}

func (item *cacheItem) setValue(value any) {
	item.value.Store(&value)
}

func (item *cacheItem) getValue() any {
	value := item.value.Load()
	return *value
}

type ttlUpdate struct {
	key  string
	item *cacheItem
	exp  time.Time
}

type inMemCache struct {
	items         xsync.MapOf[string, *cacheItem]
	ttlQueue      util.ExpirationQueue
	ttlUpdateChan chan ttlUpdate
}

func NewInMemCache() Cache {
	cache := &inMemCache{
		items:         *xsync.NewMapOf[string, *cacheItem](),
		ttlUpdateChan: make(chan ttlUpdate, 64),
	}
	go cache.janitor()
	return cache
}

func (c *inMemCache) janitor() {
	timer := time.NewTimer(time.Millisecond)
	defer timer.Stop()
	for {
		select {
		case ttlUpdate, more := <-c.ttlUpdateChan:
			if !more {
				return
			}
			ttlData := ttlUpdate.item.ttlData.Load()
			if ttlData != nil { // TTL for item already exists
				if ttlUpdate.exp.IsZero() { // removing TTL
					ttlUpdate.item.ttlData.Store(nil)
					c.ttlQueue.Delete(ttlData)
				} else { // updating TTL
					c.ttlQueue.Update(ttlData, ttlUpdate.key, ttlUpdate.exp)
				}
			} else { // creating new TTL
				ttlUpdate.item.ttlData.Store(c.ttlQueue.Push(ttlUpdate.key, ttlUpdate.exp))
			}
			if c.ttlQueue.Len() > 0 { // set timer to trigger when the first key expires
				timer.Reset(time.Until(c.ttlQueue.Peek().Expiration()))
			}

		case <-timer.C:
			// check if the next items are about to expire
			for c.ttlQueue.Len() > 0 && c.ttlQueue.Peek().Expiration().Before(time.Now()) {
				ttlData := c.ttlQueue.Pop()
				key := ttlData.Value().(string)
				// mark the item deleted by the janitor
				if item, _ := c.items.Load(key); item != nil && item.ttlData.CompareAndSwap(ttlData, deletedItem) {
					c.items.Delete(key)
				}
			}
		}
	}
}

func (c *inMemCache) Set(key, value string, ttl time.Duration) error {
	item := newCacheItem(value)
	old, _ := c.items.LoadAndStore(key, item)
	if ttl != 0 {
		c.ttlUpdateChan <- ttlUpdate{
			key:  key,
			item: item,
			exp:  time.Now().Add(ttl),
		}
	}
	// it's possible the janitor has just deleted the new item due to TTL inconsistency, so let's store it again
	if old != nil && old.ttlData.CompareAndSwap(deletedItem, nil) {
		c.items.Store(key, item)
	}
	return nil
}

func (c *inMemCache) Get(key string) (string, error) {
	item, ok := c.items.Load(key)
	if !ok {
		return "", ErrNotFound
	}
	switch value := item.getValue().(type) {
	case string:
		return value, nil
	case *int64:
		return strconv.FormatInt(*value, 10), nil
	default:
		return "", ErrWrongType
	}
}

func (c *inMemCache) Del(key string) error {
	c.items.Delete(key)
	return nil
}

func (c *inMemCache) GetTTL(key string) (time.Duration, error) {
	item, ok := c.items.Load(key)
	if !ok {
		return 0, ErrNotFound
	}
	// NOTICE: ttlData could be missing if the TTL hasn't been processed yet by the janitor
	ttlData := item.ttlData.Load()
	if ttlData == nil {
		return 0, nil
	}
	return time.Until(ttlData.Expiration()), nil
}

func (c *inMemCache) SetTTL(key string, ttl time.Duration) error {
	item, ok := c.items.Load(key)
	if !ok {
		return ErrNotFound
	}
	if ttl == 0 {
		c.ttlUpdateChan <- ttlUpdate{
			key:  key,
			item: item,
			exp:  time.Time{},
		}
	} else {
		c.ttlUpdateChan <- ttlUpdate{
			key:  key,
			item: item,
			exp:  time.Now().Add(ttl),
		}
	}
	return nil
}

func (c *inMemCache) getList(key string) (*util.List, error) {
	item, _ := c.items.LoadOrCompute(key, func() *cacheItem {
		return newCacheItem(new(util.List))
	})
	if value, ok := item.getValue().(*util.List); ok {
		return value, nil
	}
	return nil, ErrWrongType
}

func (c *inMemCache) LPush(key string, values ...string) error {
	list, err := c.getList(key)
	if err != nil {
		return err
	}
	list.PushFront(values...)
	return nil
}

func (c *inMemCache) RPush(key string, values ...string) error {
	list, err := c.getList(key)
	if err != nil {
		return err
	}
	list.PushBack(values...)
	return nil
}

func (c *inMemCache) LPop(key string, count int) ([]string, error) {
	list, err := c.getList(key)
	if err != nil {
		return nil, err
	}
	return list.PopFront(count), nil
}

func (c *inMemCache) RPop(key string, count int) ([]string, error) {
	list, err := c.getList(key)
	if err != nil {
		return nil, err
	}
	return list.PopBack(count), nil
}

func (c *inMemCache) LLen(key string) (int, error) {
	list, err := c.getList(key)
	if err != nil {
		return 0, err
	}
	return list.Len(), nil
}

func (c *inMemCache) LRange(key string, start, stop int) ([]string, error) {
	list, err := c.getList(key)
	if err != nil {
		return nil, err
	}
	return list.Range(start, stop), nil
}

func (c *inMemCache) getSet(key string) (*xsync.Map, error) {
	item, _ := c.items.LoadOrCompute(key, func() *cacheItem {
		return newCacheItem(xsync.NewMap())
	})
	if value, ok := item.getValue().(*xsync.Map); ok {
		return value, nil
	}
	return nil, ErrWrongType
}

func (c *inMemCache) SAdd(key string, values ...string) error {
	set, err := c.getSet(key)
	if err != nil {
		return err
	}
	for _, value := range values {
		set.Store(value, true)
	}
	return nil
}

func (c *inMemCache) SRem(key string, values ...string) error {
	set, err := c.getSet(key)
	if err != nil {
		return err
	}
	for _, value := range values {
		set.Delete(value)
	}
	return nil
}

func (c *inMemCache) SHas(key, value string) (bool, error) {
	set, err := c.getSet(key)
	if err != nil {
		return false, err
	}
	_, found := set.Load(value)
	return found, nil
}

func (c *inMemCache) SLen(key string) (int, error) {
	set, err := c.getSet(key)
	if err != nil {
		return 0, err
	}
	return set.Size(), nil
}

func (c *inMemCache) Incr(key string, increment int64) (int64, error) {
	item, loaded := c.items.LoadOrCompute(key, func() *cacheItem {
		value := int64(1)
		return newCacheItem(&value)
	})
	if !loaded {
		return 1, nil
	}
	switch value := item.getValue().(type) {
	case string:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return 0, ErrWrongType
		}
		i += increment
		item.setValue(&i) // TODO: figure out how to do the whole operation atomically when the underlying type is string
		return i, nil
	case *int64:
		return atomic.AddInt64(value, increment), nil
	default:
		return 0, ErrWrongType
	}
}

func (c *inMemCache) SubCache(prefix string) Cache {
	return NewPrefixCache(c, prefix)
}

func (c *inMemCache) Close() error {
	close(c.ttlUpdateChan)
	c.items.Clear()
	return nil
}
