package ligo_cache

import (
	"context"
	"fmt"
	"time"
)

// Cache wraps a Store and provides high-level caching operations with
// a configurable default TTL and get-or-compute semantics via Wrap.
type Cache struct {
	store      Store
	defaultTTL time.Duration
}

// NewCache creates a Cache backed by store. defaultTTL is used when no
// per-call TTL is provided to Set or Wrap. A zero defaultTTL means entries
// never expire unless a per-call TTL is given.
func NewCache(store Store, defaultTTL time.Duration) *Cache {
	return &Cache{store: store, defaultTTL: defaultTTL}
}

// Get retrieves a cached value by key. Returns (nil, nil) on cache miss.
func (c *Cache) Get(ctx context.Context, key string) (any, error) {
	val, _, err := c.store.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("cache get %q: %w", key, err)
	}
	return val, nil
}

// Set stores a value with the optional per-call TTL. When no ttl is given,
// the cache's defaultTTL is used.
func (c *Cache) Set(ctx context.Context, key string, value any, ttl ...time.Duration) error {
	d := c.defaultTTL
	if len(ttl) > 0 {
		d = ttl[0]
	}
	if err := c.store.Set(ctx, key, value, d); err != nil {
		return fmt.Errorf("cache set %q: %w", key, err)
	}
	return nil
}

// Del removes a single entry from the cache.
func (c *Cache) Del(ctx context.Context, key string) error {
	if err := c.store.Del(ctx, key); err != nil {
		return fmt.Errorf("cache del %q: %w", key, err)
	}
	return nil
}

// Reset clears all entries from the cache.
func (c *Cache) Reset(ctx context.Context) error {
	if err := c.store.Reset(ctx); err != nil {
		return fmt.Errorf("cache reset: %w", err)
	}
	return nil
}

// Wrap implements get-or-compute: returns the cached value for key if it
// exists, otherwise calls fn, caches the result, and returns it. If fn
// returns an error, the result is NOT cached.
func (c *Cache) Wrap(ctx context.Context, key string, fn func() (any, error), ttl ...time.Duration) (any, error) {
	val, found, err := c.store.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("cache wrap get %q: %w", key, err)
	}
	if found {
		return val, nil
	}

	result, fnErr := fn()
	if fnErr != nil {
		return nil, fnErr
	}

	if setErr := c.Set(ctx, key, result, ttl...); setErr != nil {
		return nil, setErr
	}

	return result, nil
}
