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
