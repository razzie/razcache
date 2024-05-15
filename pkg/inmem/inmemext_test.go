package inmem_test

import (
	"io"
	"testing"

	"github.com/razzie/razcache/internal/testutil"
	. "github.com/razzie/razcache/pkg/inmem"
)

func TestInMemExtTTL(t *testing.T) {
	cache := NewInMemExtendedCache()
	defer cache.(io.Closer).Close()
	testutil.TestTTL(t, cache)
}

func TestInMemExtLists(t *testing.T) {
	cache := NewInMemExtendedCache()
	defer cache.(io.Closer).Close()
	testutil.TestLists(t, cache)
}

func TestInMemExtSets(t *testing.T) {
	cache := NewInMemExtendedCache()
	defer cache.(io.Closer).Close()
	testutil.TestSets(t, cache)
}

func TestInMemExtIncr(t *testing.T) {
	cache := NewInMemExtendedCache()
	defer cache.(io.Closer).Close()
	testutil.TestIncr(t, cache)
}
