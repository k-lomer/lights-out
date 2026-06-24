package cache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_KvStore_SetGet(t *testing.T) {
	kv := MakeKvStore(time.Minute)

	kv.Set("a", "b")
	v, err := kv.Get("a")

	assert.NoError(t, err)
	assert.Equal(t, "b", v)
}

func Test_KvStore_SetMultiple(t *testing.T) {
	kv := MakeKvStore(time.Minute)

	kv.Set("a", "b")
	v, err := kv.Get("a")

	assert.NoError(t, err)
	assert.Equal(t, "b", v)

	kv.Set("a", "c")
	v, err = kv.Get("a")

	assert.NoError(t, err)
	assert.Equal(t, "c", v)
}

func Test_KvStore_GetEmpty(t *testing.T) {
	kv := MakeKvStore(time.Minute)

	_, err := kv.Get("a")

	assert.ErrorIs(t, err, ErrMissingKey{"a"})
}

func Test_KvStore_GetExpired(t *testing.T) {
	kv := MakeKvStore(5 * time.Millisecond)

	kv.Set("a", "b")

	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		_, err := kv.Get("a")
		assert.ErrorIs(c, err, ErrExpiredValue{"a"})
	}, 100*time.Millisecond, 1*time.Millisecond)
}

func Test_KvStore_Concurrency(t *testing.T) {
	runners := 5
	runTime := 500 * time.Millisecond
	var wg sync.WaitGroup
	wg.Add(runners)

	kv := MakeKvStore(time.Minute)

	for i := 0; i < runners; i++ {
		go func() {
			defer wg.Done()
			keys := []int{}
			startTime := time.Now()
			r := rand.New(rand.NewSource(startTime.UnixMicro()))

			for time.Since(startTime) < runTime {
				// Set new value
				kInt := r.Int()
				k := strconv.Itoa(kInt)
				kv.Set(k, k)
				keys = append(keys, kInt)

				// Get existing value
				kInt = keys[r.Intn(len(keys))]
				k = strconv.Itoa(kInt)
				v, err := kv.Get(k)

				assert.NoError(t, err)
				assert.Equal(t, k, v)
			}
		}()
	}
	wg.Wait()
}
