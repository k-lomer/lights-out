package cache

import (
	"fmt"
	"sync"
	"time"
)

type ErrMissingKey struct {
	k string
}

type ErrExpiredValue struct {
	k string
}

func (e ErrMissingKey) Error() string {
	return fmt.Sprintf("no value found for key: %s", e.k)
}

func (e ErrExpiredValue) Error() string {
	return fmt.Sprintf("value has expired for key: %s", e.k)
}

type ValueExt struct {
	v       string
	updated time.Time
}

type KvStore struct {
	store map[string]ValueExt
	ttl   time.Duration
	lock  sync.RWMutex
}

func MakeKvStore(ttl time.Duration) KvStore {
	return KvStore{
		store: make(map[string]ValueExt),
		ttl:   ttl,
	}
}

func (kv *KvStore) Set(k string, v string) {
	kv.lock.Lock()
	defer kv.lock.Unlock()

	kv.store[k] = ValueExt{
		v:       v,
		updated: time.Now(),
	}
}

func (kv *KvStore) Get(k string) (string, error) {
	kv.lock.RLock()
	defer kv.lock.RUnlock()

	vExt, found := kv.store[k]
	if !found {
		return "", ErrMissingKey{k}
	}

	elapsed := time.Since(vExt.updated)
	if elapsed > kv.ttl {
		return vExt.v, ErrExpiredValue{k}
	}

	return vExt.v, nil
}
