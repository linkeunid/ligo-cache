package ligo_cache

import (
	"sort"
	"strings"
	"time"

	"github.com/linkeunid/ligo"
)

// CacheInterceptor is a ligo Interceptor that automatically caches GET
// responses. Non-GET requests pass through uncached.
type CacheInterceptor struct {
	cache *Cache
	ttl   time.Duration
}

// NewCacheInterceptor creates an interceptor that caches GET responses
// for the given ttl using the provided cache.
func NewCacheInterceptor(cache *Cache, ttl time.Duration) *CacheInterceptor {
	return &CacheInterceptor{cache: cache, ttl: ttl}
}

// Intercept implements ligo.Interceptor. It caches GET responses by
// route key and skips all other methods.
func (ci *CacheInterceptor) Intercept(ctx *ligo.Context, next ligo.HandlerFunc) error {
	req := ctx.Request()
	if req.Method != "GET" {
		return next(ctx)
	}

	key := ci.BuildKey(req.Method, req.URL.Path, req.URL.RawQuery)

	val, err := ci.cache.Get(req.Context(), key)
	if err == nil && val != nil {
		if cached, ok := val.(cachedResponse); ok {
			return ctx.JSON(cached.StatusCode, cached.Body)
		}
	}

	if err := next(ctx); err != nil {
		return err
	}

	return nil
}

type cachedResponse struct {
	StatusCode int
	Body       any
}

// BuildKey produces a deterministic cache key from method, path, and
// sorted query string: "METHOD:path:sorted_params".
func (ci *CacheInterceptor) BuildKey(method, path, rawQuery string) string {
	if rawQuery == "" {
		return method + ":" + path + ":"
	}
	params := strings.Split(rawQuery, "&")
	sort.Strings(params)
	return method + ":" + path + ":" + strings.Join(params, "&")
}
