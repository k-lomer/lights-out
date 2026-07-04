package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/k-lomer/lights-out/model"
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
	v       []model.Outage
	updated time.Time
}

type OutageCache struct {
	store map[string]ValueExt
	ttl   time.Duration
	lock  sync.RWMutex
}

func MakeOutageCache(ttl time.Duration) *OutageCache {
	outageCache := OutageCache{
		store: make(map[string]ValueExt),
		ttl:   ttl,
	}

	return &outageCache
}

func (c *OutageCache) GetTtl() time.Duration {
	return c.ttl
}

func (c *OutageCache) Set(k string, v []model.Outage) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.store[k] = ValueExt{
		v:       v,
		updated: time.Now(),
	}
}

func (c *OutageCache) Get(k string) ([]model.Outage, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	vExt, found := c.store[k]
	if !found {
		return nil, ErrMissingKey{k}
	}

	elapsed := time.Since(vExt.updated)
	if elapsed > c.ttl {
		return vExt.v, ErrExpiredValue{k}
	}

	return vExt.v, nil
}

func (c *OutageCache) Delete(k string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.store, k)
}
