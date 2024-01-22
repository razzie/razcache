package razcache

import (
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTTL(t *testing.T) {
	cache := NewInMemCache()
	defer cache.(io.Closer).Close()

	// key should be present before expiration and gone afterwards
	assert.Nil(t, cache.Set("key_longrunning", "value", time.Millisecond*2000))
	assert.Nil(t, cache.Set("key", "value", time.Millisecond*500))
	value, err := cache.Get("key")
	assert.Nil(t, err)
	assert.Equal(t, "value", value)
	time.Sleep(time.Second)
	value, err = cache.Get("key")
	assert.Equal(t, ErrNotFound, err)
	assert.NotEqual(t, "value", value)

	// overwritten key with different TTL should make the value persist
	assert.Nil(t, cache.Set("key2", "value2", time.Millisecond*500))
	assert.Nil(t, cache.Set("key2", "newvalue2", 0))
	time.Sleep(time.Second)
	value, err = cache.Get("key2")
	assert.Nil(t, err)
	assert.Equal(t, "newvalue2", value)

	// make sure janitor won't crash if key is removed before expiration
	assert.Nil(t, cache.Set("key3", "value3", time.Millisecond*500))
	assert.Nil(t, cache.Del("key3"))
	time.Sleep(time.Second)
}

func TestLists(t *testing.T) {
	cache := NewInMemCache()
	defer cache.(io.Closer).Close()

	// make a list of 1, 2, 3, 4, 5 using both LPush and RPush
	assert.Nil(t, cache.LPush("list", "3"))
	assert.Nil(t, cache.LPush("list", "1", "2"))
	assert.Nil(t, cache.RPush("list", "4", "5"))
	llen, err := cache.LLen("list")
	assert.Nil(t, err)
	assert.Equal(t, 5, llen)

	// testing LPop and RPop
	result, err := cache.LPop("list")
	assert.Nil(t, err)
	assert.Equal(t, "1", result)
	result, err = cache.RPop("list")
	assert.Nil(t, err)
	assert.Equal(t, "5", result)
	for i := 0; i < 3; i++ {
		_, err = cache.RPop("list")
		assert.Nil(t, err)
	}
	_, err = cache.RPop("list")
	assert.Equal(t, ErrNotFound, err)

	// testing list functions on non-list keys
	assert.Nil(t, cache.Set("non-list", "value", 0))
	_, err = cache.LLen("non-list")
	assert.Equal(t, ErrWrongType, err)
	assert.Equal(t, ErrWrongType, cache.LPush("non-list", "1"))
	_, err = cache.LPop("non-list")
	assert.Equal(t, ErrWrongType, err)
}

func TestSets(t *testing.T) {
	cache := NewInMemCache()
	defer cache.(io.Closer).Close()

	// adding members in multiple steps and asserting correct length
	assert.Nil(t, cache.SAdd("set", "a", "b", "c"))
	assert.Nil(t, cache.SAdd("set", "c", "d"))
	slen, err := cache.SLen("set")
	assert.Nil(t, err)
	assert.Equal(t, 4, slen)

	// removing members and asserting correct length
	assert.Nil(t, cache.SRem("set", "a", "d"))
	slen, err = cache.SLen("set")
	assert.Nil(t, err)
	assert.Equal(t, 2, slen)

	// testing set functions on non-set keys
	assert.Nil(t, cache.Set("non-set", "value", 0))
	_, err = cache.SLen("non-set")
	assert.Equal(t, ErrWrongType, err)
	assert.Equal(t, ErrWrongType, cache.SAdd("non-set", "a"))
	assert.Equal(t, ErrWrongType, cache.SRem("non-set"))
}
