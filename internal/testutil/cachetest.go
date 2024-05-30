package testutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/razzie/razcache"
)

func TestBasic(t *testing.T, cache razcache.Cache) {
	_, err := cache.Get("key")
	assert.Equal(t, razcache.ErrNotFound, err)

	assert.NoError(t, cache.Set("key", "value1", 0))

	assert.NoError(t, cache.Set("key", "value2", 0))

	value, err := cache.Get("key")
	assert.NoError(t, err)
	assert.Equal(t, "value2", value)
}

func TestTTL(t *testing.T, cache razcache.Cache, ttlGran time.Duration) {
	// key should be present before expiration and gone afterwards
	assert.NoError(t, cache.Set("key1", "value1", ttlGran*3))
	assert.NoError(t, cache.Set("key2", "value2", ttlGran))

	value, err := cache.Get("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value)
	value, err = cache.Get("key2")
	assert.NoError(t, err)
	assert.Equal(t, "value2", value)

	time.Sleep(ttlGran * 2)

	value, err = cache.Get("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value)
	_, err = cache.Get("key2")
	assert.Equal(t, razcache.ErrNotFound, err)

	time.Sleep(ttlGran * 2)

	_, err = cache.Get("key1")
	assert.Equal(t, razcache.ErrNotFound, err)

	// overwritten key with different TTL should make the value persist
	assert.NoError(t, cache.Set("key2", "value2", ttlGran))
	assert.NoError(t, cache.Set("key2", "newvalue2", 0))

	time.Sleep(ttlGran * 2)

	value, err = cache.Get("key2")
	assert.NoError(t, err)
	assert.Equal(t, "newvalue2", value)

	// make sure janitor won't crash if key is removed before expiration
	assert.NoError(t, cache.Set("key3", "value3", ttlGran))
	assert.NoError(t, cache.Del("key3"))
	time.Sleep(ttlGran * 2)
}

func TestLists(t *testing.T, cache razcache.ExtendedCache) {
	// make a list of 1, 2, 3, 4, 5 using both LPush and RPush
	assert.NoError(t, cache.LPush("list", "3"))
	assert.NoError(t, cache.LPush("list", "1", "2"))
	assert.NoError(t, cache.RPush("list", "4", "5"))
	llen, err := cache.LLen("list")
	assert.NoError(t, err)
	assert.Equal(t, 5, llen)

	// testing LRange
	result, err := cache.LRange("list", 0, -1)
	assert.NoError(t, err)
	assert.Equal(t, []string{"1", "2", "3", "4", "5"}, result)
	result, err = cache.LRange("list", -1, 99999)
	assert.NoError(t, err)
	assert.Equal(t, []string{"5"}, result)
	result, err = cache.LRange("list", 99999, -1)
	assert.NoError(t, err)
	assert.Nil(t, result)

	// testing LPop and RPop
	result, err = cache.LPop("list", 1)
	assert.NoError(t, err)
	assert.Equal(t, []string{"1"}, result)
	result, err = cache.RPop("list", 1)
	assert.NoError(t, err)
	assert.Equal(t, []string{"5"}, result)
	result, err = cache.RPop("list", 3)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(result))
	result, err = cache.RPop("list", 1)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))

	// testing list functions on non-list keys
	assert.NoError(t, cache.Set("non-list", "value", 0))
	_, err = cache.LLen("non-list")
	assert.Equal(t, razcache.ErrWrongType, err)
	assert.Equal(t, razcache.ErrWrongType, cache.LPush("non-list", "1"))
	_, err = cache.LPop("non-list", 1)
	assert.Equal(t, razcache.ErrWrongType, err)
}

func TestSets(t *testing.T, cache razcache.ExtendedCache) {
	// adding members in multiple steps and asserting correct length
	assert.NoError(t, cache.SAdd("set", "a", "b", "c"))
	assert.NoError(t, cache.SAdd("set", "c", "d"))
	slen, err := cache.SLen("set")
	assert.NoError(t, err)
	assert.Equal(t, 4, slen)

	// removing members and asserting correct length
	assert.NoError(t, cache.SRem("set", "a", "d"))
	slen, err = cache.SLen("set")
	assert.NoError(t, err)
	assert.Equal(t, 2, slen)

	// testing set functions on non-set keys
	assert.NoError(t, cache.Set("non-set", "value", 0))
	_, err = cache.SLen("non-set")
	assert.Equal(t, razcache.ErrWrongType, err)
	assert.Equal(t, razcache.ErrWrongType, cache.SAdd("non-set", "a"))
	assert.Equal(t, razcache.ErrWrongType, cache.SRem("non-set"))
}

func TestIncr(t *testing.T, cache razcache.ExtendedCache) {
	// strings that cannot be converted to int should fail with wrong type
	assert.NoError(t, cache.Set("non-int", "a", 0))
	_, err := cache.Incr("non-int", 1)
	assert.Equal(t, razcache.ErrWrongType, err)

	// strings that can be converted to int should work
	assert.NoError(t, cache.Set("int", "2", 0))
	value, err := cache.Incr("int", 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(4), value)

	// further increments should work too
	value, err = cache.Incr("int", 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), value)

	// the internal type became int64 after increment(s)
	// let's make sure Get still works on it
	str, err := cache.Get("int")
	assert.NoError(t, err)
	assert.Equal(t, "5", str)

	// non-existing key acts like a 0
	value, err = cache.Incr("new", 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), value)
}
