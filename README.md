```go
type Cache interface {
	Set(key, value string, ttl time.Duration) error
	Get(key string) (string, error)
	Del(key string) error

	GetTTL(key string) (time.Duration, error)
	SetTTL(key string, ttl time.Duration) error

	SubCache(prefix string) Cache

	Close() error
}

type ExtendedCache interface {
	Cache

	LPush(key string, values ...string) error
	RPush(key string, values ...string) error
	LPop(key string, count int) ([]string, error)
	RPop(key string, count int) ([]string, error)
	LLen(key string) (int, error)
	LRange(key string, start, stop int) ([]string, error)

	SAdd(key string, values ...string) error
	SRem(key string, values ...string) error
	SHas(key, value string) (bool, error)
	SLen(key string) (int, error)

	Incr(key string, increment int64) (int64, error)

	SubExtendedCache(prefix string) ExtendedCache
}

// pkg/inmem
func NewInMemCache() Cache
func NewInMemExtendedCache() ExtendedCache

// pkg/redis
func NewRedisCache(redisDSN string) (ExtendedCache, error)
func NewRedisCacheFromClient(client redis.Cmdable) ExtendedCache

// pkg/badger
func NewBadgerCache(dir string) (Cache, error)
func NewBadgerCacheFromDB(db *badger.DB) Cache
```
