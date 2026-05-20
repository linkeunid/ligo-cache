package ligo_cache

import (
	"context"
	"testing"
	"time"
)

func TestInMemoryStore_SetAndGet(t *testing.T) {
	s := NewInMemoryStore()
	ctx := context.Background()

	err := s.Set(ctx, "foo", "bar", 0)
	if err != nil {
		t.Fatalf("Set: %v", err)
	}

	val, found, err := s.Get(ctx, "foo")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found {
		t.Fatal("expected found=true")
	}
	if val != "bar" {
		t.Fatalf("expected %q, got %q", "bar", val)
	}
}

func TestInMemoryStore_GetMissing(t *testing.T) {
	s := NewInMemoryStore()
	ctx := context.Background()

	val, found, err := s.Get(ctx, "nope")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if found {
		t.Fatal("expected found=false for missing key")
	}
	if val != nil {
		t.Fatalf("expected nil, got %v", val)
	}
}

func TestInMemoryStore_Del(t *testing.T) {
	s := NewInMemoryStore()
	ctx := context.Background()

	_ = s.Set(ctx, "foo", "bar", 0)

	err := s.Del(ctx, "foo")
	if err != nil {
		t.Fatalf("Del: %v", err)
	}

	_, found, _ := s.Get(ctx, "foo")
	if found {
		t.Fatal("expected key to be deleted")
	}
}

func TestInMemoryStore_Reset(t *testing.T) {
	s := NewInMemoryStore()
	ctx := context.Background()

	_ = s.Set(ctx, "a", 1, 0)
	_ = s.Set(ctx, "b", 2, 0)

	err := s.Reset(ctx)
	if err != nil {
		t.Fatalf("Reset: %v", err)
	}

	for _, key := range []string{"a", "b"} {
		_, found, _ := s.Get(ctx, key)
		if found {
			t.Fatalf("expected %q to be cleared", key)
		}
	}
}

func TestInMemoryStore_TTLExpiry(t *testing.T) {
	s := NewInMemoryStore(50 * time.Millisecond)
	ctx := context.Background()

	_ = s.Set(ctx, "short", "data", 100*time.Millisecond)

	val, found, _ := s.Get(ctx, "short")
	if !found || val != "data" {
		t.Fatal("expected entry to exist immediately after Set")
	}

	time.Sleep(200 * time.Millisecond)

	_, found, _ = s.Get(ctx, "short")
	if found {
		t.Fatal("expected entry to expire after TTL")
	}
}

func TestInMemoryStore_TTLZero_NoExpiry(t *testing.T) {
	s := NewInMemoryStore(50 * time.Millisecond)
	ctx := context.Background()

	_ = s.Set(ctx, "permanent", "data", 0)

	time.Sleep(150 * time.Millisecond)

	val, found, _ := s.Get(ctx, "permanent")
	if !found || val != "data" {
		t.Fatal("expected entry with TTL=0 to persist")
	}
}

func TestInMemoryStore_ConcurrentReadWrite(t *testing.T) {
	s := NewInMemoryStore()
	ctx := context.Background()

	const n = 100
	done := make(chan struct{})

	go func() {
		defer close(done)
		for i := range n {
			_ = s.Set(ctx, "key", i, 0)
		}
	}()

	for range n {
		s.Get(ctx, "key")
	}
	<-done
}

func TestInMemoryStore_SetOverwrite(t *testing.T) {
	s := NewInMemoryStore()
	ctx := context.Background()

	_ = s.Set(ctx, "key", "first", 0)
	_ = s.Set(ctx, "key", "second", 0)

	val, found, _ := s.Get(ctx, "key")
	if !found || val != "second" {
		t.Fatalf("expected %q, got %q", "second", val)
	}
}

func TestInMemoryStore_NilValue(t *testing.T) {
	s := NewInMemoryStore()
	ctx := context.Background()

	_ = s.Set(ctx, "nil", nil, 0)

	val, found, _ := s.Get(ctx, "nil")
	if !found {
		t.Fatal("expected found=true for nil value")
	}
	if val != nil {
		t.Fatalf("expected nil, got %v", val)
	}
}
