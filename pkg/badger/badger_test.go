package badger_test

import (
	"testing"

	"github.com/razzie/razcache/internal/testutil"
	. "github.com/razzie/razcache/pkg/badger"
	"github.com/stretchr/testify/require"
)

func TestBadgerCache(t *testing.T) {
	cache, err := NewBadgerCache("")
	require.NoError(t, err)
	defer cache.Close()

	testutil.TestBasic(t, cache)
}
