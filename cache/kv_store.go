package cache

import (
	"fmt"
	"time"
)

type ValueExt struct {
	v       string
	updated time.Time
}

type KvStore struct {
	store map[string]ValueExt
	ttl   time.Duration
}

func MakeKvStore(ttl time.Duration) KvStore {
	return KvStore{
		store: make(map[string]ValueExt),
		ttl:   ttl,
	}
}

func (kv *KvStore) Set(k string, v string) {
	kv.store[k] = ValueExt{
		v:       v,
		updated: time.Now(),
	}
}

func (kv *KvStore) Get(k string) (string, error) {
	vExt, found := kv.store[k]
	if !found {
		return "", fmt.Errorf("no value found for key: %s", k)
	}

	elapsed := time.Since(vExt.updated)
	if elapsed > kv.ttl {
		return vExt.v, fmt.Errorf("value has expired for key: %s", k)
	}

	return vExt.v, nil
}
