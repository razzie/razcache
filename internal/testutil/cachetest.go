package testutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "github.com/razzie/razcache"
)

func TestTTL(t *testing.T, cache Cache) {
	// key should be present before expiration and gone afterwards
	assert.NoError(t, cache.Set("key_longrunning", "value", time.Millisecond*2000))
	assert.NoError(t, cache.Set("key", "value", time.Millisecond*500))
	value, err := cache.Get("key")
	assert.NoError(t, err)
	assert.Equal(t, "value", value)
	time.Sleep(time.Second)
	value, err = cache.Get("key")
	assert.Equal(t, ErrNotFound, err)
	assert.NotEqual(t, "value", value)

	// overwritten key with different TTL should make the value persist
	assert.NoError(t, cache.Set("key2", "value2", time.Millisecond*500))
	assert.NoError(t, cache.Set("key2", "newvalue2", 0))
	time.Sleep(time.Second)
	value, err = cache.Get("key2")
	assert.NoError(t, err)
	assert.Equal(t, "newvalue2", value)

	// make sure janitor won't crash if key is removed before expiration
	assert.NoError(t, cache.Set("key3", "value3", time.Millisecond*500))
	assert.NoError(t, cache.Del("key3"))
	time.Sleep(time.Second)
}

func TestLists(t *testing.T, cache ExtendedCache) {
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
	assert.Equal(t, ErrWrongType, err)
	assert.Equal(t, ErrWrongType, cache.LPush("non-list", "1"))
	_, err = cache.LPop("non-list", 1)
	assert.Equal(t, ErrWrongType, err)
}

func TestSets(t *testing.T, cache ExtendedCache) {
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
	assert.Equal(t, ErrWrongType, err)
	assert.Equal(t, ErrWrongType, cache.SAdd("non-set", "a"))
	assert.Equal(t, ErrWrongType, cache.SRem("non-set"))
}

func TestIncr(t *testing.T, cache ExtendedCache) {
	// strings that cannot be converted to int should fail with wrong type
	assert.NoError(t, cache.Set("non-int", "a", 0))
	_, err := cache.Incr("non-int", 1)
	assert.Equal(t, ErrWrongType, err)

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
