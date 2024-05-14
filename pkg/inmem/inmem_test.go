package inmem_test

import (
	"io"
	"testing"

	"github.com/razzie/razcache/internal/testutil"
	. "github.com/razzie/razcache/pkg/inmem"
)

func TestInMemTTL(t *testing.T) {
	cache := NewInMemCache()
	defer cache.(io.Closer).Close()
	testutil.TestTTL(t, cache)
}

func TestInMemLists(t *testing.T) {
	cache := NewInMemCache()
	defer cache.(io.Closer).Close()
	testutil.TestLists(t, cache)
}

func TestInMemSets(t *testing.T) {
	cache := NewInMemCache()
	defer cache.(io.Closer).Close()
	testutil.TestSets(t, cache)
}

func TestInMemIncr(t *testing.T) {
	cache := NewInMemCache()
	defer cache.(io.Closer).Close()
	testutil.TestIncr(t, cache)
}
