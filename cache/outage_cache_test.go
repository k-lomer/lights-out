package cache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/k-lomer/lights-out/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// outages builds a single-element outage slice identified by id, so cached
// values can be constructed and compared succinctly.
func outages(id string) []model.Outage {
	return []model.Outage{{ID: id}}
}

// Test that a value can be set and read back.
func Test_OutageCache_SetGet(t *testing.T) {
	c := MakeOutageCache(time.Minute)

	c.Set("a", outages("b"))
	v, err := c.Get("a")

	assert.NoError(t, err)
	assert.Equal(t, outages("b"), v)
}

// Test that setting an existing key overwrites the previous value.
func Test_OutageCache_SetMultiple(t *testing.T) {
	c := MakeOutageCache(time.Minute)

	c.Set("a", outages("b"))
	v, err := c.Get("a")

	require.NoError(t, err)
	require.Equal(t, outages("b"), v)

	c.Set("a", outages("c"))
	v, err = c.Get("a")

	assert.NoError(t, err)
	assert.Equal(t, outages("c"), v)
}

// Test that getting a missing key returns ErrMissingKey.
func Test_OutageCache_GetEmpty(t *testing.T) {
	c := MakeOutageCache(time.Minute)

	_, err := c.Get("a")

	assert.ErrorIs(t, err, ErrMissingKey{"a"})
}

// Test that getting an expired value returns ErrExpiredValue.
func Test_OutageCache_GetExpired(t *testing.T) {
	c := MakeOutageCache(5 * time.Millisecond)

	c.Set("a", outages("b"))

	assert.EventuallyWithT(t, func(ct *assert.CollectT) {
		_, err := c.Get("a")
		assert.ErrorIs(ct, err, ErrExpiredValue{"a"})
	}, 100*time.Millisecond, 1*time.Millisecond)
}

// Test that a value can be deleted.
func Test_OutageCache_Delete(t *testing.T) {
	c := MakeOutageCache(time.Minute)

	c.Set("a", outages("b"))
	v, err := c.Get("a")

	assert.NoError(t, err)
	assert.Equal(t, outages("b"), v)

	c.Delete("a")

	_, err = c.Get("a")
	assert.ErrorIs(t, err, ErrMissingKey{"a"})

}

// Test that concurrent sets and gets are safe under many goroutines.
func Test_OutageCache_Concurrency(t *testing.T) {
	runners := 5
	runTime := 500 * time.Millisecond
	var wg sync.WaitGroup
	wg.Add(runners)

	c := MakeOutageCache(time.Minute)

	for i := 0; i < runners; i++ {
		go func() {
			defer wg.Done()
			keys := []int{}
			startTime := time.Now()
			r := rand.New(rand.NewSource(startTime.UnixMicro()))

			for time.Since(startTime) < runTime {
				// Set new value.
				kInt := r.Int()
				k := strconv.Itoa(kInt)
				c.Set(k, outages(k))
				keys = append(keys, kInt)

				// Get existing value.
				kInt = keys[r.Intn(len(keys))]
				k = strconv.Itoa(kInt)
				v, err := c.Get(k)

				require.NoError(t, err)
				require.Equal(t, outages(k), v)
			}
		}()
	}
	wg.Wait()
}
