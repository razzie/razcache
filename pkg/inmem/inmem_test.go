package inmem_test

import (
	"testing"

	"github.com/razzie/razcache/internal/testutil"
	. "github.com/razzie/razcache/pkg/inmem"
)

func TestInMemBasic(t *testing.T) {
	cache := NewInMemCache()
	defer cache.Close()
	testutil.TestBasic(t, cache)
}

func TestInMemTTL(t *testing.T) {
	cache := NewInMemCache()
	defer cache.Close()
	testutil.TestTTL(t, cache)
}
