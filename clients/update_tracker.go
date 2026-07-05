package clients

import (
	"sync"
	"time"
)

// UpdateTracker provides caching synchronization for a DnoClient.
// Embed it as a pointer (*UpdateTracker) in a DnoClient.
type UpdateTracker struct {
	mu         sync.Mutex
	lastUpdate *time.Time
}

// LastUpdate returns the time of the last successful update, or nil if never updated.
// Callers read it while holding the update lock.
func (u *UpdateTracker) LastUpdate() *time.Time {
	return u.lastUpdate
}

// SetUpdated sets the time of the most recent update.
// Callers read it while holding the update lock.
func (u *UpdateTracker) SetUpdated() time.Time {
	now := time.Now()
	u.lastUpdate = &now
	return now
}

func (u *UpdateTracker) UpdateLock() {
	u.mu.Lock()
}

func (u *UpdateTracker) UpdateUnlock() {
	u.mu.Unlock()
}
