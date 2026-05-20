package ligo_cache

import (
	"time"

	"github.com/linkeunid/ligo"
)

// Config configures the cache module.
type Config struct {
	Store           Store         // nil = InMemoryStore
	DefaultTTL      time.Duration // 0 = no expiry
	JanitorInterval time.Duration // 0 = 1s default, only for InMemoryStore
}

// Provider returns a ligo.Provider that registers a [*Cache] as a
// singleton in the DI container.
func Provider(cfg Config) ligo.Provider {
	return ligo.Factory[*Cache](func() *Cache {
		store := cfg.Store
		if store == nil {
			store = NewInMemoryStore(cfg.JanitorInterval)
		}
		return NewCache(store, cfg.DefaultTTL)
	})
}

// Module returns a Ligo module that registers a [*Cache] as a singleton
// provider via DI.
func Module(cfg Config) ligo.Module {
	return ligo.NewModule(
		"ligo-cache",
		ligo.Providers(
			Provider(cfg),
		),
	)
}
