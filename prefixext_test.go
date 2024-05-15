package razcache_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/razzie/razcache"
	"github.com/razzie/razcache/pkg/inmem"
)

func TestPrefixExtendedCache(t *testing.T) {
	cache := inmem.NewInMemCache()

	assert.NoError(t, cache.Set("a", "val_a", 0))
	assert.NoError(t, cache.Set("prefix:b", "val_b", 0))

	subcache := NewPrefixExtendedCache(cache, "prefix:")

	// prefix cache should hide non-prefixed item keys
	_, err := subcache.Get("a")
	assert.Equal(t, ErrNotFound, err)

	// previously created prefixed item key should work
	value, err := subcache.Get("b")
	assert.NoError(t, err)
	assert.Equal(t, "val_b", value)

	// newly created prefixed key should be visible in original cache
	assert.NoError(t, subcache.Set("c", "val_c", 0))
	_, err = cache.Get("c")
	assert.Equal(t, ErrNotFound, err)
	value, err = cache.Get("prefix:c")
	assert.NoError(t, err)
	assert.Equal(t, "val_c", value)
}
