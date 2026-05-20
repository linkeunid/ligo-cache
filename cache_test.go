package ligo_cache

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockStore struct {
	data map[string]any
	err  error
}

func newMockStore() *mockStore {
	return &mockStore{data: make(map[string]any)}
}

func (m *mockStore) Get(_ context.Context, key string) (any, bool, error) {
	if m.err != nil {
		return nil, false, m.err
	}
	v, ok := m.data[key]
	return v, ok, nil
}

func (m *mockStore) Set(_ context.Context, key string, value any, _ time.Duration) error {
	if m.err != nil {
		return m.err
	}
	m.data[key] = value
	return nil
}

func (m *mockStore) Del(_ context.Context, key string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.data, key)
	return nil
}

func (m *mockStore) Reset(_ context.Context) error {
	if m.err != nil {
		return m.err
	}
	m.data = make(map[string]any)
	return nil
}

func TestCache_Get(t *testing.T) {
	store := newMockStore()
	store.data["key"] = "value"
	c := NewCache(store, 0)

	val, err := c.Get(context.Background(), "key")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "value" {
		t.Fatalf("expected %q, got %q", "value", val)
	}
}

func TestCache_GetMiss(t *testing.T) {
	c := NewCache(newMockStore(), 0)

	val, err := c.Get(context.Background(), "missing")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != nil {
		t.Fatalf("expected nil, got %v", val)
	}
}

func TestCache_Set_UsesDefaultTTL(t *testing.T) {
	store := newMockStore()
	c := NewCache(store, 5*time.Minute)

	err := c.Set(context.Background(), "key", "val")
	if err != nil {
		t.Fatalf("Set: %v", err)
	}
	if store.data["key"] != "val" {
		t.Fatal("expected value in store")
	}
}

func TestCache_Set_OverrideTTL(t *testing.T) {
	store := newMockStore()
	c := NewCache(store, 5*time.Minute)

	err := c.Set(context.Background(), "key", "val", 10*time.Minute)
	if err != nil {
		t.Fatalf("Set: %v", err)
	}
}

func TestCache_Del(t *testing.T) {
	store := newMockStore()
	store.data["key"] = "val"
	c := NewCache(store, 0)

	err := c.Del(context.Background(), "key")
	if err != nil {
		t.Fatalf("Del: %v", err)
	}
	if _, ok := store.data["key"]; ok {
		t.Fatal("expected key deleted")
	}
}

func TestCache_Reset(t *testing.T) {
	store := newMockStore()
	store.data["a"] = 1
	store.data["b"] = 2
	c := NewCache(store, 0)

	err := c.Reset(context.Background())
	if err != nil {
		t.Fatalf("Reset: %v", err)
	}
	if len(store.data) != 0 {
		t.Fatal("expected empty store")
	}
}

func TestCache_Wrap_CacheMiss(t *testing.T) {
	store := newMockStore()
	c := NewCache(store, 0)
	called := false

	val, err := c.Wrap(context.Background(), "key", func() (any, error) {
		called = true
		return "computed", nil
	})
	if err != nil {
		t.Fatalf("Wrap: %v", err)
	}
	if !called {
		t.Fatal("expected fn to be called on miss")
	}
	if val != "computed" {
		t.Fatalf("expected %q, got %q", "computed", val)
	}
	if store.data["key"] != "computed" {
		t.Fatal("expected value cached in store")
	}
}

func TestCache_Wrap_CacheHit(t *testing.T) {
	store := newMockStore()
	store.data["key"] = "cached"
	c := NewCache(store, 0)
	called := false

	val, err := c.Wrap(context.Background(), "key", func() (any, error) {
		called = true
		return "should not be called", nil
	})
	if err != nil {
		t.Fatalf("Wrap: %v", err)
	}
	if called {
		t.Fatal("expected fn NOT called on hit")
	}
	if val != "cached" {
		t.Fatalf("expected %q, got %q", "cached", val)
	}
}

func TestCache_Wrap_FnError_NotCached(t *testing.T) {
	store := newMockStore()
	c := NewCache(store, 0)
	fnErr := errors.New("boom")

	_, err := c.Wrap(context.Background(), "key", func() (any, error) {
		return nil, fnErr
	})
	if !errors.Is(err, fnErr) {
		t.Fatalf("expected fn error, got %v", err)
	}
	if _, ok := store.data["key"]; ok {
		t.Fatal("expected value NOT cached on fn error")
	}
}

func TestCache_StoreError(t *testing.T) {
	store := newMockStore()
	store.err = errors.New("store fail")
	c := NewCache(store, 0)

	_, err := c.Get(context.Background(), "key")
	if !errors.Is(err, store.err) {
		t.Fatalf("expected store error, got %v", err)
	}
}
