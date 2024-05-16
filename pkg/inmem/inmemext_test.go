package inmem_test

import (
	"testing"

	"github.com/razzie/razcache/internal/testutil"
	. "github.com/razzie/razcache/pkg/inmem"
)

func TestInMemExtBasic(t *testing.T) {
	cache := NewInMemExtendedCache()
	defer cache.Close()
	testutil.TestBasic(t, cache)
}

func TestInMemExtTTL(t *testing.T) {
	cache := NewInMemExtendedCache()
	defer cache.Close()
	testutil.TestTTL(t, cache)
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
