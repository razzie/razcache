package inmem_test

import (
	"testing"
	"time"

	. "github.com/razzie/razcache/pkg/inmem"
	"github.com/razzie/razcache/pkg/testutil"
)

func TestInMemExtBasic(t *testing.T) {
	cache := NewInMemExtendedCache()
	defer cache.Close()
	testutil.TestBasic(t, cache)
}

func TestInMemExtTTL(t *testing.T) {
	cache := NewInMemExtendedCache()
	defer cache.Close()
	testutil.TestTTL(t, cache, time.Millisecond*50)
}

func TestInMemExtLists(t *testing.T) {
	cache := NewInMemExtendedCache()
	defer cache.Close()
	testutil.TestLists(t, cache)
}

func TestInMemExtSets(t *testing.T) {
	cache := NewInMemExtendedCache()
	defer cache.Close()
	testutil.TestSets(t, cache)
}

func TestInMemExtIncr(t *testing.T) {
	cache := NewInMemExtendedCache()
	defer cache.Close()
	testutil.TestIncr(t, cache)
}
