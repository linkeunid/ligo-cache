package ligo_cache

import (
	"reflect"
	"testing"
	"time"

	"github.com/linkeunid/ligo"
)

func TestProviderType(t *testing.T) {
	p := Provider(Config{DefaultTTL: 5 * time.Minute})
	want := reflect.TypeFor[*Cache]()
	if p.Type() != want {
		t.Fatalf("expected type %v, got %v", want, p.Type())
	}
}

func TestModuleName(t *testing.T) {
	m := Module(Config{DefaultTTL: 5 * time.Minute})
	if m.Name != "ligo-cache" {
		t.Fatalf("expected module name %q, got %q", "ligo-cache", m.Name)
	}
}

func TestModuleRegistersCache(t *testing.T) {
	m := Module(Config{DefaultTTL: 5 * time.Minute})
	want := reflect.TypeFor[*Cache]()
	for _, raw := range m.Providers {
		if p, ok := raw.(ligo.Provider); ok && p.Type() == want {
			return
		}
	}
	t.Fatalf("Module must register *Cache; providers: %v", m.Providers)
}

func TestModuleProviderType(t *testing.T) {
	m := Module(Config{DefaultTTL: 5 * time.Minute})
	p := m.Providers[0].(ligo.Provider)
	want := reflect.TypeFor[*Cache]()
	if p.Type() != want {
		t.Fatalf("Module provider type: expected %v, got %v", want, p.Type())
	}
}
