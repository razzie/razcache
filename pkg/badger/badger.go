package badger

import (
	"time"
	"unsafe"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/razzie/razcache"
)

type badgerCache badger.DB

func NewBadgerCache(dir string) (razcache.Cache, error) {
	opts := badger.DefaultOptions(dir)
	if len(dir) == 0 {
		opts = opts.WithInMemory(true)
	}
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return (*badgerCache)(db), nil
}

func NewBadgerCacheFromDB(db *badger.DB) razcache.Cache {
	return (*badgerCache)(db)
}

func (c *badgerCache) Get(key string) (val string, err error) {
	err = translateBadgerError((*badger.DB)(c).View(func(txn *badger.Txn) error {
		item, err := txn.Get(yoloBytes(key))
		if err != nil {
			return err
		}
		return item.Value(func(raw []byte) error {
			val = string(raw)
			return nil
		})
	}))
	return
}

func (c *badgerCache) Set(key string, value string, ttl time.Duration) error {
	e := badger.NewEntry(yoloBytes(key), yoloBytes(value))
	if ttl > 0 {
		e = e.WithTTL(ttl)
	}
	return translateBadgerError((*badger.DB)(c).Update(func(txn *badger.Txn) error {
		return txn.SetEntry(e)
	}))
}

func (c *badgerCache) Del(key string) error {
	return translateBadgerError((*badger.DB)(c).Update(func(txn *badger.Txn) error {
		return txn.Delete(yoloBytes(key))
	}))
}

func (c *badgerCache) GetTTL(key string) (ttl time.Duration, err error) {
	err = translateBadgerError((*badger.DB)(c).View(func(txn *badger.Txn) error {
		item, err := txn.Get(yoloBytes(key))
		if err != nil {
			return err
		}
		exp := item.ExpiresAt()
		if exp != 0 {
			ttl = time.Until(time.Unix(int64(exp), 0))
		}
		return nil
	}))
	return
}

func (c *badgerCache) SetTTL(key string, ttl time.Duration) error {
	return translateBadgerError((*badger.DB)(c).Update(func(txn *badger.Txn) error {
		item, err := txn.Get(yoloBytes(key))
		if err != nil {
			return err
		}
		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		e := badger.NewEntry(yoloBytes(key), val)
		if ttl > 0 {
			e = e.WithTTL(ttl)
		}
		return txn.SetEntry(e)
	}))
}

func (c *badgerCache) SubCache(prefix string) razcache.Cache {
	return razcache.NewPrefixCache(c, prefix)
}

func (c *badgerCache) Close() error {
	return (*badger.DB)(c).Close()
}

func translateBadgerError(err error) error {
	if err == badger.ErrKeyNotFound {
		err = razcache.ErrNotFound
	}
	return err
}

func yoloBytes(s string) []byte {
	if s == "" {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
