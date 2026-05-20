package ligo_cache_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/linkeunid/ligo"
	"github.com/linkeunid/ligo/adapters/echo"

	ligo_cache "github.com/linkeunid/ligo-cache"
)

func startApp(t *testing.T, modules ...ligo.Module) *ligo.App {
	t.Helper()
	app := ligo.New(
		ligo.WithRouter(echo.NewAdapter()),
		ligo.WithAddr(":0"),
	)
	app.Register(modules...)
	go func() { _ = app.Run() }()
	time.Sleep(200 * time.Millisecond)
	return app
}

func containsType(types []reflect.Type, want reflect.Type) bool {
	for _, typ := range types {
		if typ == want {
			return true
		}
	}
	return false
}

func TestModuleIntegratesWithApp(t *testing.T) {
	app := startApp(t, ligo_cache.Module(ligo_cache.Config{
		DefaultTTL: 5 * time.Minute,
	}))

	want := reflect.TypeFor[*ligo_cache.Cache]()
	if !containsType(app.Container().Types(), want) {
		t.Fatalf("*Cache not found in DI container; registered types: %v", app.Container().Types())
	}
}

func TestProviderInjectsCache(t *testing.T) {
	var injected *ligo_cache.Cache

	mod := ligo.NewModule(
		"test",
		ligo.Providers(
			ligo_cache.Provider(ligo_cache.Config{DefaultTTL: 5 * time.Minute}),
			ligo.Factory[*string](func(c *ligo_cache.Cache) *string {
				injected = c
				s := "injected"
				return &s
			}),
		),
	)
	app := startApp(t, mod)

	// Resolve *string to trigger the factory chain (factories are lazy).
	if _, err := ligo.Resolve[*string](app); err != nil {
		t.Fatalf("resolve *string: %v", err)
	}

	if injected == nil {
		t.Fatal("expected *Cache to be injected into factory")
	}

	ctx := context.Background()
	_ = injected.Set(ctx, "key", "value")
	val, err := injected.Get(ctx, "key")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "value" {
		t.Fatalf("expected %q, got %q", "value", val)
	}
}

func TestCacheInterceptor_WrapIntegration(t *testing.T) {
	cache := ligo_cache.NewCache(ligo_cache.NewInMemoryStore(), 0)
	ic := ligo_cache.NewCacheInterceptor(cache, time.Minute)

	ctx := context.Background()
	key := ic.BuildKey("GET", "/users", "")

	val, err := cache.Wrap(ctx, key, func() (any, error) {
		return map[string]string{"hello": "world"}, nil
	})
	if err != nil {
		t.Fatalf("Wrap: %v", err)
	}

	data, ok := val.(map[string]string)
	if !ok || data["hello"] != "world" {
		t.Fatalf("unexpected wrap result: %v", val)
	}

	// Second call hits cache
	val2, err := cache.Wrap(ctx, key, func() (any, error) {
		t.Fatal("should not be called on cache hit")
		return nil, nil
	})
	if err != nil {
		t.Fatalf("Wrap hit: %v", err)
	}
	if val2.(map[string]string)["hello"] != "world" {
		t.Fatalf("unexpected cache hit result: %v", val2)
	}
}
