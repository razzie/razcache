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

type ttlData interface {
	LoadTTLData() *util.TTLItem[string]
	StoreTTLData(*util.TTLItem[string])
	CompareAndSwapTTLData(old, new *util.TTLItem[string]) (swapped bool)
}

type cacheItemBase struct {
	ttlData atomic.Pointer[util.TTLItem[string]]
}

func (item *cacheItemBase) LoadTTLData() *util.TTLItem[string] {
	return item.ttlData.Load()
}

func (item *cacheItemBase) StoreTTLData(val *util.TTLItem[string]) {
	item.ttlData.Store(val)
}

func (item *cacheItemBase) CompareAndSwapTTLData(old, new *util.TTLItem[string]) (swapped bool) {
	return item.ttlData.CompareAndSwap(old, new)
}

type ttlUpdate struct {
	key     string
	ttlData ttlData
	exp     time.Time
}

type inMemCacheBase[T ttlData] struct {
	items         xsync.MapOf[string, T]
	ttlQueue      util.TTLQueue[string]
	ttlUpdateChan chan ttlUpdate
	closed        atomic.Bool
}

func (cache *inMemCacheBase[T]) init() {
	cache.items = *xsync.NewMapOf[string, T]()
	cache.ttlUpdateChan = make(chan ttlUpdate, 64)
	go cache.janitor()
}

func (c *inMemCacheBase[T]) janitor() {
	timer := time.NewTimer(time.Millisecond)
	defer timer.Stop()
	for {
		select {
		case ttlUpdate, more := <-c.ttlUpdateChan:
			if !more {
				c.ttlQueue.Clear()
				return
			}
			ttlData := ttlUpdate.ttlData.LoadTTLData()
			if ttlData != nil && ttlData != dummyTTLData { // TTL for item already exists
				if ttlUpdate.exp.IsZero() { // removing TTL
					ttlUpdate.ttlData.StoreTTLData(nil)
					c.ttlQueue.Delete(ttlData)
				} else { // updating TTL
					c.ttlQueue.Update(ttlData, ttlUpdate.key, ttlUpdate.exp)
				}
			} else { // creating new TTL
				ttlUpdate.ttlData.StoreTTLData(c.ttlQueue.Push(ttlUpdate.key, ttlUpdate.exp))
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
				if item, ok := c.items.Load(key); ok && item.CompareAndSwapTTLData(ttlData, currentlyDeletingItem) {
					c.items.Delete(key)
					item.StoreTTLData(deletedItem)
				}
			}
		}
	}
}

func (c *inMemCacheBase[T]) set(key string, item T, ttl time.Duration) error {
	old, oldExists := c.items.LoadAndStore(key, item)
	if ttl != 0 {
		item.StoreTTLData(dummyTTLData)
		c.ttlUpdateChan <- ttlUpdate{
			key:     key,
			ttlData: item,
			exp:     time.Now().Add(ttl),
		}
	}
	// it's possible the janitor has just deleted the new item due to TTL inconsistency,
	// so let's store it again
	if oldExists {
		for {
			switch old.LoadTTLData() {
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

func (c *inMemCacheBase[T]) get(key string) (item T, err error) {
	var ok bool
	item, ok = c.items.Load(key)
	if !ok {
		err = razcache.ErrNotFound
	}
	return
}

func (c *inMemCacheBase[T]) Del(key string) error {
	c.items.Delete(key)
	return nil
}

func (c *inMemCacheBase[T]) GetTTL(key string) (time.Duration, error) {
	item, ok := c.items.Load(key)
	if !ok {
		return 0, razcache.ErrNotFound
	}
	for {
		ttlData := item.LoadTTLData()
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

func (c *inMemCacheBase[T]) SetTTL(key string, ttl time.Duration) error {
	item, ok := c.items.Load(key)
	if !ok {
		return razcache.ErrNotFound
	}
	if ttl == 0 {
		c.ttlUpdateChan <- ttlUpdate{
			key:     key,
			ttlData: item,
			exp:     time.Time{},
		}
	} else {
		c.ttlUpdateChan <- ttlUpdate{
			key:     key,
			ttlData: item,
			exp:     time.Now().Add(ttl),
		}
	}
	return nil
}

func (c *inMemCacheBase[T]) Close() error {
	c.closed.Store(true)
	close(c.ttlUpdateChan)
	c.items.Clear()
	return nil
}
