//go:build js && wasm

package controller

import (
	"fmt"
	"log/slog"

	"github.com/syumai/workers/cloudflare/kv"
)

const kvSessionTTL = 3600 // 1 hour in seconds

// KVSessionProvider stores session state in Cloudflare KV.
// It reads the session on Acquire and writes it back on Release.
//
// Concurrency note: KVSessionProvider has no per-session locking. Two
// concurrent requests for the same session ID will both read the same KV
// state, and the last write wins (lost update). This is acceptable for
// Cloudflare Workers where each isolate handles one request at a time.
type KVSessionProvider[T any] struct {
	ns        *kv.Namespace
	keyPrefix string
	marshal   func(T) ([]byte, error)
	unmarshal func([]byte) (T, error)
}

// NewKVSessionProvider creates a KVSessionProvider.
// nsName is the KV namespace binding name from wrangler.toml.
// keyPrefix is prepended to session IDs to form KV keys.
// marshal/unmarshal convert the interactor to/from bytes.
// Returns an error if the KV namespace cannot be initialised.
func NewKVSessionProvider[T any](
	nsName string,
	keyPrefix string,
	marshal func(T) ([]byte, error),
	unmarshal func([]byte) (T, error),
) (*KVSessionProvider[T], error) {
	ns, err := kv.NewNamespace(nsName)
	if err != nil {
		return nil, fmt.Errorf("failed to initialise KV namespace %q: %w", nsName, err)
	}
	return &KVSessionProvider[T]{
		ns:        ns,
		keyPrefix: keyPrefix,
		marshal:   marshal,
		unmarshal: unmarshal,
	}, nil
}

// Acquire reads the session from KV (or creates via factory if not found),
// and returns a release function that writes the state back to KV.
func (p *KVSessionProvider[T]) Acquire(id string, factory func() T) (T, SessionRelease, bool) {
	if len(id) > SessionMaxIDLen {
		var zero T
		return zero, nil, false
	}
	key := p.keyPrefix + id
	var val T

	if p.ns != nil {
		data, err := p.ns.GetString(key, nil)
		if err == nil && data != "" {
			restored, err := p.unmarshal([]byte(data))
			if err == nil {
				val = restored
				return val, p.releaseFunc(key, &val), true
			}
			slog.Error("KV unmarshal failed, creating new session", "key", key, "error", err)
		}
	}

	val = factory()
	return val, p.releaseFunc(key, &val), true
}

// releaseFunc returns a closure that serialises val and writes it to KV.
func (p *KVSessionProvider[T]) releaseFunc(key string, val *T) SessionRelease {
	return func() {
		if p.ns == nil {
			return
		}
		data, err := p.marshal(*val)
		if err != nil {
			slog.Error("KV marshal failed", "key", key, "error", err)
			return
		}
		if err := p.ns.PutString(key, string(data), &kv.PutOptions{ExpirationTTL: kvSessionTTL}); err != nil {
			slog.Error("KV put failed", "key", key, "error", err)
		}
	}
}

// Stop is a no-op for KV (no background goroutines).
func (p *KVSessionProvider[T]) Stop() {}
