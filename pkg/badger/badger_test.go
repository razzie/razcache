package badger_test

import (
	"testing"
	"time"

	. "github.com/razzie/razcache/pkg/badger"
	"github.com/razzie/razcache/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func TestBadgerCache(t *testing.T) {
	cache, err := NewBadgerCache("")
	require.NoError(t, err)
	defer cache.Close()

	testutil.TestBasic(t, cache)
}

func TestBadgerCacheTTL(t *testing.T) {
	cache, err := NewBadgerCache("")
	require.NoError(t, err)
	defer cache.Close()

	testutil.TestTTL(t, cache, time.Second)
}
