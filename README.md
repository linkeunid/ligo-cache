# ligo-cache

NestJS-inspired caching for [Ligo](https://github.com/linkeunid/ligo), with pluggable store backends, per-entry TTL, and automatic HTTP response caching via interceptors.

[![Go Version](https://img.shields.io/badge/go-1.25+-blue)](https://go.dev/dl)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Tests](https://img.shields.io/badge/tests-28%20passing-brightgreen)](https://github.com/linkeunid/ligo-cache)

## Install

```bash
go get github.com/linkeunid/ligo-cache
```

## Quick start

Register the cache module with a default TTL:

```go
import (
    "time"

    "github.com/linkeunid/ligo"
    ligo_cache "github.com/linkeunid/ligo-cache"
)

app.Register(ligo_cache.Module(ligo_cache.Config{
    DefaultTTL: 5 * time.Minute,
}))
```

Inject `*ligo_cache.Cache` into constructors:

```go
func NewUserService(cache *ligo_cache.Cache) *UserService {
    return &UserService{cache: cache}
}
```

## Manual caching

```go
// Set with default TTL
cache.Set(ctx, "user:123", user)

// Set with per-call TTL
cache.Set(ctx, "user:123", user, 30*time.Second)

// Get
val, err := cache.Get(ctx, "user:123")

// Get-or-compute
user, err := cache.Wrap(ctx, "user:123", func() (any, error) {
    return db.FindUser(123)
})

// Delete
cache.Del(ctx, "user:123")

// Clear all
cache.Reset(ctx)
```

## Auto-caching with interceptor

Apply `CacheInterceptor` to route groups for automatic GET response caching:

```go
func NewUserController(cache *ligo_cache.Cache) ligo.Controller {
    return &UserController{
        cache:       cache,
        interceptor: ligo_cache.NewCacheInterceptor(cache, 30*time.Second),
    }
}

func (c *UserController) Routes(r ligo.Router) {
    cr := ligo.NewChainRouter(r.Group("/users"))
    cr.GET("", c.List).Intercept(c.interceptor).Handle()
    cr.GET("/:id", c.Get).Intercept(c.interceptor).Handle()
    cr.POST("", c.Create).Handle() // POST bypasses cache automatically
}
```

Cache keys are deterministic: `GET:/users:page=1&sort=name` (query params sorted alphabetically).

## Pluggable stores

The default `InMemoryStore` ships with the package. Implement the `Store` interface for custom backends (Redis, memcached, etc.):

```go
type Store interface {
    Get(ctx context.Context, key string) (value any, found bool, err error)
    Set(ctx context.Context, key string, value any, ttl time.Duration) error
    Del(ctx context.Context, key string) error
    Reset(ctx context.Context) error
}
```

Register with a custom store:

```go
app.Register(ligo_cache.Module(ligo_cache.Config{
    Store:      myRedisStore,
    DefaultTTL: 10 * time.Minute,
}))
```

## License

MIT
