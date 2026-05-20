package ligo_cache

import (
	"testing"
	"time"
)

func TestBuildKey_Format(t *testing.T) {
	cache := NewCache(NewInMemoryStore(), 0)
	ic := NewCacheInterceptor(cache, time.Minute)

	tests := []struct {
		method, path, query, want string
	}{
		{"GET", "/api/users", "page=1&sort=name", "GET:/api/users:page=1&sort=name"},
		{"GET", "/api/users", "", "GET:/api/users:"},
		{"GET", "/api/users", "z=1&a=2", "GET:/api/users:a=2&z=1"},
	}
	for _, tt := range tests {
		got := ic.BuildKey(tt.method, tt.path, tt.query)
		if got != tt.want {
			t.Errorf("buildKey(%q,%q,%q) = %q, want %q", tt.method, tt.path, tt.query, got, tt.want)
		}
	}
}

func TestBuildKey_SortsParams(t *testing.T) {
	cache := NewCache(NewInMemoryStore(), 0)
	ic := NewCacheInterceptor(cache, time.Minute)

	a := ic.BuildKey("GET", "/path", "c=3&a=1&b=2")
	b := ic.BuildKey("GET", "/path", "b=2&c=3&a=1")
	if a != b {
		t.Fatalf("expected same key for reordered params: %q != %q", a, b)
	}
}
