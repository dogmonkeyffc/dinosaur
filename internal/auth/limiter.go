package auth

import (
	"sync"
	"time"
)

type rateEntry struct {
	count   int
	expires time.Time
}
type Limit struct {
	Key string
	Max int
}
type Limiter struct {
	mu          sync.Mutex
	entries     map[string]rateEntry
	nextCleanup time.Time
}

func NewLimiter() *Limiter { return &Limiter{entries: make(map[string]rateEntry)} }

// Allow uses fixed 15-minute windows, retaining at most 10000 keys. Callers must
// supply peer addresses rather than trusting client-controlled forwarded headers.
func (l *Limiter) Allow(limits ...Limit) bool { return l.allowAt(time.Now(), limits...) }
func (l *Limiter) allowAt(now time.Time, limits ...Limit) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !now.Before(l.nextCleanup) {
		for key, e := range l.entries {
			if !now.Before(e.expires) {
				delete(l.entries, key)
			}
		}
		l.nextCleanup = now.Add(time.Minute)
	}
	for _, limit := range limits {
		e, exists := l.entries[limit.Key]
		if exists && now.Before(e.expires) && e.count >= limit.Max {
			return false
		}
		if !exists && len(l.entries) >= 10000 {
			return false
		}
	}
	for _, limit := range limits {
		e := l.entries[limit.Key]
		if !now.Before(e.expires) {
			e = rateEntry{expires: now.Add(15 * time.Minute)}
		}
		e.count++
		l.entries[limit.Key] = e
	}
	return true
}
