package ligo_cache

import (
	"context"
	"sync"
	"time"
)

// Store is the pluggable cache backend interface. Implementations handle
// storage, retrieval, and TTL-based expiration of cache entries.
type Store interface {
	Get(ctx context.Context, key string) (value any, found bool, err error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Del(ctx context.Context, key string) error
	Reset(ctx context.Context) error
}

type entry struct {
	value     any
	expiresAt time.Time
}

func (e entry) isZero() bool {
	return e.expiresAt.IsZero()
}

func (e entry) isExpired() bool {
	return !e.isZero() && time.Now().After(e.expiresAt)
}

// InMemoryStore is a thread-safe in-memory Store implementation with
// per-entry TTL and background janitor goroutine for expired entry eviction.
type InMemoryStore struct {
	mu              sync.RWMutex
	data            map[string]entry
	janitorInterval time.Duration
	stopJanitor     chan struct{}
}

// NewInMemoryStore creates a new InMemoryStore. If janitorInterval is 0,
// expired entries are cleaned up every second.
func NewInMemoryStore(janitorInterval ...time.Duration) *InMemoryStore {
	interval := time.Second
	if len(janitorInterval) > 0 && janitorInterval[0] > 0 {
		interval = janitorInterval[0]
	}
	s := &InMemoryStore{
		data:            make(map[string]entry),
		janitorInterval: interval,
		stopJanitor:     make(chan struct{}),
	}
	go s.janitor()
	return s
}

func (s *InMemoryStore) janitor() {
	ticker := time.NewTicker(s.janitorInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.evictExpired()
		case <-s.stopJanitor:
			return
		}
	}
}

func (s *InMemoryStore) evictExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for k, e := range s.data {
		if !e.isZero() && now.After(e.expiresAt) {
			delete(s.data, k)
		}
	}
}

func (s *InMemoryStore) Get(_ context.Context, key string) (any, bool, error) {
	s.mu.RLock()
	e, ok := s.data[key]
	s.mu.RUnlock()
	if !ok || e.isExpired() {
		return nil, false, nil
	}
	return e.value, true, nil
}

func (s *InMemoryStore) Set(_ context.Context, key string, value any, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}
	s.data[key] = entry{value: value, expiresAt: expiresAt}
	return nil
}

func (s *InMemoryStore) Del(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}

func (s *InMemoryStore) Reset(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]entry)
	return nil
}

// Close stops the background janitor goroutine.
func (s *InMemoryStore) Close() {
	close(s.stopJanitor)
}
