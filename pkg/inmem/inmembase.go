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
	ttlNotYetProcessed = &util.TTLItem[string]{}

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
	items         atomic.Pointer[xsync.MapOf[string, T]]
	ttlQueue      util.TTLQueue[string]
	ttlUpdateChan chan ttlUpdate
	closedChan    chan struct{}
}

func (cache *inMemCacheBase[T]) init() {
	cache.items.Store(xsync.NewMapOf[string, T]())
	cache.ttlUpdateChan = make(chan ttlUpdate, 64)
	cache.closedChan = make(chan struct{})
	go cache.janitor()
}

func (c *inMemCacheBase[T]) janitor() {
	timer := time.NewTimer(0)
	timerRunning := true
	var nextExp time.Time

	defer func() {
		close(c.ttlUpdateChan)

		items := c.items.Swap(nil)
		items.Clear()

		if !timer.Stop() {
			<-timer.C
		}
		timer = nil
	}()

	for {
		select {
		case <-c.closedChan:
			c.ttlQueue.Clear()
			return

		case ttlUpdate := <-c.ttlUpdateChan:
			ttlData := ttlUpdate.item.LoadTTLData()
			if ttlData == ttlNotYetProcessed { // creating new TTL
				ttlUpdate.item.StoreTTLData(c.ttlQueue.Push(ttlUpdate.key, ttlUpdate.exp))
			} else { // updating TTL for existing item
				if ttlUpdate.exp.IsZero() { // removing TTL
					c.ttlQueue.Delete(ttlData)
				} else { // updating TTL
					c.ttlQueue.Update(ttlData, ttlUpdate.key, ttlUpdate.exp)
				}
			}
			// update timer to trigger when the first key expires
			if c.ttlQueue.Len() > 0 {
				prevExp := nextExp
				nextExp = c.ttlQueue.Peek().Expiration()
				if timerRunning {
					if nextExp.Before(prevExp) {
						if !timer.Stop() {
							<-timer.C
						}
						timer.Reset(time.Until(nextExp))
						timerRunning = true
					}
				} else {
					timer.Reset(time.Until(nextExp))
					timerRunning = true
				}
			}

		case <-timer.C:
			timerRunning = false
			now := time.Now()
			// check if the next items are about to expire
			for c.ttlQueue.Len() > 0 && c.ttlQueue.Peek().Expiration().Before(now) {
				ttlData := c.ttlQueue.Pop()
				key := ttlData.Value()
				c.items.Load().Compute(key, func(oldValue T, loaded bool) (newValue T, delete bool) {
					delete = !loaded || (loaded && oldValue.LoadTTLData() == ttlData)
					newValue = oldValue
					return
				})
			}
			// reset timer for next item in queue
			if c.ttlQueue.Len() > 0 {
				nextExp = c.ttlQueue.Peek().Expiration()
				timer.Reset(time.Until(nextExp))
				timerRunning = true
			}
		}
	}
}

func (c *inMemCacheBase[T]) set(key string, item T, ttl time.Duration) error {
	items := c.items.Load()
	if items == nil {
		return ErrCacheClosed
	}
	items.Store(key, item)
	if ttl != 0 {
		item.StoreTTLData(ttlNotYetProcessed)
		return c.sendTTLUpdate(key, item, ttl)
	}
	return nil
}

func (c *inMemCacheBase[T]) get(key string) (item T, err error) {
	items := c.items.Load()
	if items == nil {
		err = ErrCacheClosed
		return
	}
	var ok bool
	item, ok = items.Load(key)
	if !ok {
		err = razcache.ErrNotFound
	}
	return
}

func (c *inMemCacheBase[T]) getOrCompute(key string, compute func() T) (item T, loaded bool, err error) {
	items := c.items.Load()
	if items == nil {
		err = ErrCacheClosed
		return
	}
	item, loaded = items.LoadOrCompute(key, compute)
	return
}

func (c *inMemCacheBase[T]) Del(key string) error {
	items := c.items.Load()
	if items == nil {
		return ErrCacheClosed
	}
	items.Delete(key)
	return nil
}

func (c *inMemCacheBase[T]) GetTTL(key string) (time.Duration, error) {
	item, err := c.get(key)
	if err != nil {
		return 0, err
	}
	for {
		ttlData := item.LoadTTLData()
		switch ttlData {
		case ttlNotYetProcessed:
			if c.items.Load() == nil { // closed
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
	item, err := c.get(key)
	if err != nil {
		return err
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
	close(c.closedChan)
	return nil
}
