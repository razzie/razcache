package razcache

import (
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
)

var (
	ErrNotFound  = fmt.Errorf("not found")
	ErrWrongType = fmt.Errorf("wrong type")
)

func translateRedisError(err error) error {
	switch {
	case err == nil:
		return nil
	case err == redis.Nil:
		return ErrNotFound
	case strings.HasPrefix(err.Error(), "WRONGTYPE"):
		return ErrWrongType
	default:
		return err
	}
}
