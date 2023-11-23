package razcache

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/razzie/razcache/internal/util"

	"github.com/puzpuzpuz/xsync"
)

// special pointer to mark a janitor deleted item
var deletedItem = &util.EQItem{}

type item struct {
	value   any
	ttlData atomic.Pointer[util.EQItem]
}

type ttlUpdate struct {
	key  string
	item *item
	exp  time.Time
}

type inMemCache struct {
	items         xsync.MapOf[string, *item]
	ttlQueue      util.ExpirationQueue
	ttlUpdateChan chan ttlUpdate
}

func NewInMemCache(tickerInterval time.Duration) (Cache, error) {
	cache := &inMemCache{
		items:         *xsync.NewMapOf[*item](),
		ttlUpdateChan: make(chan ttlUpdate, 64),
	}
	go cache.janitor(tickerInterval)
	return cache, nil
}

func (c *inMemCache) janitor(tickerInterval time.Duration) {
	ticker := time.NewTicker(tickerInterval)
	for {
		select {
		case ttlUpdate := <-c.ttlUpdateChan:
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

		case <-ticker.C:
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
	item := &item{value: value}
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
		return "", fmt.Errorf("not found: %s", key)
	}
	if value, ok := item.value.(string); ok {
		return value, nil
	}
	return "", fmt.Errorf("type mismatch")
}

func (c *inMemCache) Del(key string) error {
	c.items.Delete(key)
	return nil
}

func (c *inMemCache) GetTTL(key string) (time.Duration, error) {
	item, ok := c.items.Load(key)
	if !ok {
		return 0, fmt.Errorf("not found: %s", key)
	}
	// NOTICE: ttlData could be missing if the TTL hasn't been processed yet by the janitor
	ttlData := item.ttlData.Load()
	if ttlData == nil {
		return 0, nil
	}
	return ttlData.Expiration().Sub(time.Now()), nil
}

func (c *inMemCache) SetTTL(key string, ttl time.Duration) error {
	item, ok := c.items.Load(key)
	if !ok {
		return fmt.Errorf("not found: %s", key)
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
	item, _ := c.items.LoadOrCompute(key, func() *item {
		return &item{value: &util.List{}}
	})
	if value, ok := item.value.(*util.List); ok {
		return value, nil
	}
	return nil, fmt.Errorf("not a list")
}

func (c *inMemCache) LPush(key, value string) error {
	list, err := c.getList(key)
	if err != nil {
		return err
	}
	list.PushFront(value)
	return nil
}

func (c *inMemCache) RPush(key, value string) error {
	list, err := c.getList(key)
	if err != nil {
		return err
	}
	list.PushBack(value)
	return nil
}

func (c *inMemCache) LPop(key string) (string, error) {
	list, err := c.getList(key)
	if err != nil {
		return "", err
	}
	value := list.PopFront()
	if value == nil {
		return "", fmt.Errorf("empty list")
	}
	return value.(string), nil
}

func (c *inMemCache) RPop(key string) (string, error) {
	list, err := c.getList(key)
	if err != nil {
		return "", err
	}
	value := list.PopBack()
	if value == nil {
		return "", fmt.Errorf("empty list")
	}
	return value.(string), nil
}

func (c *inMemCache) LLen(key string) (int, error) {
	list, err := c.getList(key)
	if err != nil {
		return 0, err
	}
	return list.Len(), nil
}

func (c *inMemCache) getSet(key string) (*xsync.Map, error) {
	item, _ := c.items.LoadOrCompute(key, func() *item {
		return &item{value: xsync.NewMap()}
	})
	if value, ok := item.value.(*xsync.Map); ok {
		return value, nil
	}
	return nil, fmt.Errorf("not a set")
}

func (c *inMemCache) SAdd(key, value string) error {
	set, err := c.getSet(key)
	if err != nil {
		return err
	}
	set.Store(value, true)
	return nil
}

func (c *inMemCache) SRem(key, value string) error {
	set, err := c.getSet(key)
	if err != nil {
		return err
	}
	set.Delete(value)
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
