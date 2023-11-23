package razcache

import (
	"time"
)

type Cache interface {
	Set(key, value string, ttl time.Duration) error
	Get(key string) (string, error)
	Del(key string) error
	GetTTL(key string) (time.Duration, error)
	SetTTL(key string, ttl time.Duration) error

	LPush(key, value string) error
	RPush(key, value string) error
	LPop(key string) (string, error)
	RPop(key string) (string, error)
	LLen(key string) (int, error)

	SAdd(key, value string) error
	SRem(key, value string) error
	SHas(key, value string) (bool, error)
	SLen(key string) (int, error)
}
