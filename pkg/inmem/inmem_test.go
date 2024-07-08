package inmem_test

import (
	"testing"
	"time"

	. "github.com/razzie/razcache/pkg/inmem"
	"github.com/razzie/razcache/pkg/testutil"
)

func TestInMemBasic(t *testing.T) {
	cache := NewInMemCache()
	defer cache.Close()
	testutil.TestBasic(t, cache)
}

func TestInMemTTL(t *testing.T) {
	cache := NewInMemCache()
	defer cache.Close()
	testutil.TestTTL(t, cache, time.Millisecond*50)
}
