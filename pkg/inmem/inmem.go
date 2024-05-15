package inmem

import (
	"runtime"
	"sync/atomic"
	"time"

	"github.com/puzpuzpuz/xsync/v3"
	"github.com/razzie/razcache"
	"github.com/razzie/razcache/internal/util"
)

var (
	// special pointers to mark a janitor deleted item
	currentlyDeletingItem = &util.TTLItem[string]{}
	deletedItem           = &util.TTLItem[string]{}
	// special pointer to mark not yet processed TTL data
	dummyTTLData = &util.TTLItem[string]{}
)

type cacheItem struct {
	cacheItemBase
	value string
}

type inMemCache struct {
	items         xsync.MapOf[string, *cacheItem]
	ttlQueue      util.TTLQueue[string]
	ttlUpdateChan chan ttlUpdate
	closed        atomic.Bool
}

func NewInMemCache() razcache.Cache {
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
				c.ttlQueue.Clear()
				return
			}
			ttlData := ttlUpdate.item.ttlData.Load()
			if ttlData != nil && ttlData != dummyTTLData { // TTL for item already exists
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
				key := ttlData.Value()
				// mark the item deleted by the janitor
				if item, _ := c.items.Load(key); item != nil && item.ttlData.CompareAndSwap(ttlData, currentlyDeletingItem) {
					c.items.Delete(key)
					item.ttlData.Store(deletedItem)
				}
			}
		}
	}
}

func (c *inMemCache) Set(key, value string, ttl time.Duration) error {
	item := &cacheItem{value: value}
	old, _ := c.items.LoadAndStore(key, item)
	if ttl != 0 {
		item.ttlData.Store(dummyTTLData)
		c.ttlUpdateChan <- ttlUpdate{
			key:  key,
			item: &item.cacheItemBase,
			exp:  time.Now().Add(ttl),
		}
	}
	// it's possible the janitor has just deleted the new item due to TTL inconsistency,
	// so let's store it again
	if old != nil {
		for {
			switch old.ttlData.Load() {
			case currentlyDeletingItem:
				runtime.Gosched() // wait for janitor to delete previous item
			case deletedItem:
				c.items.Store(key, item)
				return nil
			default:
				return nil
			}
		}
	}
	return nil
}

func (c *inMemCache) Get(key string) (string, error) {
	item, ok := c.items.Load(key)
	if !ok {
		return "", razcache.ErrNotFound
	}
	return item.value, nil
}

func (c *inMemCache) Del(key string) error {
	c.items.Delete(key)
	return nil
}

func (c *inMemCache) GetTTL(key string) (time.Duration, error) {
	item, ok := c.items.Load(key)
	if !ok {
		return 0, razcache.ErrNotFound
	}
	for {
		ttlData := item.ttlData.Load()
		switch ttlData {
		case dummyTTLData:
			if c.closed.Load() {
				return 0, nil
			}
			runtime.Gosched() // wait for janitor to assign ttl data
		case currentlyDeletingItem, deletedItem:
			return 0, razcache.ErrNotFound
		case nil:
			return 0, nil
		default:
			return time.Until(ttlData.Expiration()), nil
		}
	}
}

func (c *inMemCache) SetTTL(key string, ttl time.Duration) error {
	item, ok := c.items.Load(key)
	if !ok {
		return razcache.ErrNotFound
	}
	if ttl == 0 {
		c.ttlUpdateChan <- ttlUpdate{
			key:  key,
			item: &item.cacheItemBase,
			exp:  time.Time{},
		}
	} else {
		c.ttlUpdateChan <- ttlUpdate{
			key:  key,
			item: &item.cacheItemBase,
			exp:  time.Now().Add(ttl),
		}
	}
	return nil
}

func (c *inMemCache) SubCache(prefix string) razcache.Cache {
	return razcache.NewPrefixCache(c, prefix)
}

func (c *inMemCache) Close() error {
	c.closed.Store(true)
	close(c.ttlUpdateChan)
	c.items.Clear()
	return nil
}
