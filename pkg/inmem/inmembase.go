package inmem

import (
	"errors"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/puzpuzpuz/xsync/v3"
	"github.com/razzie/razcache"
	"github.com/razzie/razcache/internal/util"
)

var (
	// special pointer to mark not yet processed TTL data
	dummyTTLData = &util.TTLItem[string]{}

	ErrCacheClosed = errors.New("cache is closed")
)

type ttlDataConstraint interface {
	comparable
	ttlData
}

type ttlData interface {
	LoadTTLData() *util.TTLItem[string]
	StoreTTLData(*util.TTLItem[string])
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

type ttlUpdate struct {
	key  string
	item ttlData
	exp  time.Time
}

type inMemCacheBase[T ttlDataConstraint] struct {
	items         xsync.MapOf[string, T]
	ttlQueue      util.TTLQueue[string]
	ttlUpdateChan chan ttlUpdate
	closedChan    chan struct{}
	isClosed      atomic.Bool
}

func (cache *inMemCacheBase[T]) init() {
	cache.items = *xsync.NewMapOf[string, T]()
	cache.ttlUpdateChan = make(chan ttlUpdate, 64)
	cache.closedChan = make(chan struct{})
	go cache.janitor()
}

func (c *inMemCacheBase[T]) janitor() {
	timer := time.NewTimer(time.Millisecond)
	defer timer.Stop()
	for {
		select {
		case <-c.closedChan:
			c.ttlQueue.Clear()
			return

		case ttlUpdate := <-c.ttlUpdateChan:
			ttlData := ttlUpdate.item.LoadTTLData()
			if ttlData == dummyTTLData { // creating new TTL
				ttlUpdate.item.StoreTTLData(c.ttlQueue.Push(ttlUpdate.key, ttlUpdate.exp))
			} else { // updating TTL for existing item
				if ttlUpdate.exp.IsZero() { // removing TTL
					c.ttlQueue.Delete(ttlData)
				} else { // updating TTL
					c.ttlQueue.Update(ttlData, ttlUpdate.key, ttlUpdate.exp)
				}
			}
			if c.ttlQueue.Len() > 0 { // set timer to trigger when the first key expires
				timer.Reset(time.Until(c.ttlQueue.Peek().Expiration()))
			}

		case <-timer.C:
			// check if the next items are about to expire
			for c.ttlQueue.Len() > 0 && c.ttlQueue.Peek().Expiration().Before(time.Now()) {
				ttlData := c.ttlQueue.Pop()
				key := ttlData.Value()
				c.items.Compute(key, func(oldValue T, loaded bool) (newValue T, delete bool) {
					delete = !loaded || (loaded && oldValue.LoadTTLData() == ttlData)
					newValue = oldValue
					return
				})
			}
		}
	}
}

func (c *inMemCacheBase[T]) set(key string, item T, ttl time.Duration) error {
	if c.isClosed.Load() {
		return ErrCacheClosed
	}
	c.items.Store(key, item)
	if ttl != 0 {
		item.StoreTTLData(dummyTTLData)
		return c.sendTTLUpdate(key, item, ttl)
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
			if c.isClosed.Load() {
				return 0, nil
			}
			runtime.Gosched() // wait for janitor to assign ttl data
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
		item.StoreTTLData(nil)
	}
	return c.sendTTLUpdate(key, item, ttl)
}

func (c *inMemCacheBase[T]) sendTTLUpdate(key string, item T, ttl time.Duration) error {
	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	select {
	case <-c.closedChan:
		return ErrCacheClosed
	case c.ttlUpdateChan <- ttlUpdate{
		key:  key,
		item: item,
		exp:  exp,
	}:
		return nil
	}
}

func (c *inMemCacheBase[T]) Close() error {
	c.isClosed.Store(true)
	close(c.closedChan)
	c.items.Clear()
	return nil
}
