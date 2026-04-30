// Package resolver provides ABI / signature / contract metadata resolution layered on
// multiple sources with caching:
//
//   - go-ethereum embedded 4byte database (offline)
//   - openchain.xyz signature API (via dbadoy/signature/openchain) — for unknown selectors/topics
//   - Sourcify v2 contract endpoint — for full ABI of verified contracts (any chain)
//   - Etherscan-compatible APIs — pluggable, optional
//
// All lookups are cached in memory; consumers can wrap their own persistent cache.
package resolver

import (
	"context"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/dbadoy/signature/openchain"
	gethfourbyte "github.com/ethereum/go-ethereum/signer/fourbyte"
)

// SignatureResolver resolves a 4-byte selector or 32-byte event topic to its
// human-readable signature. It is concurrency-safe.
type SignatureResolver struct {
	openchain *openchain.Client
	fourbyte  *gethfourbyte.Database
	cache     sync.Map // hex string -> []string
	timeout   time.Duration
}

// NewSignatureResolver constructs a resolver with sane defaults.
func NewSignatureResolver() *SignatureResolver {
	cli, _ := openchain.New(openchain.DefaultConfig())
	db, _ := gethfourbyte.New() // embedded 4byte snapshot
	return &SignatureResolver{
		openchain: cli,
		fourbyte:  db,
		timeout:   5 * time.Second,
	}
}

// Selector resolves a 4-byte function selector to one or more candidate signatures.
// Order: in-memory cache → embedded 4byte → openchain online.
func (r *SignatureResolver) Selector(ctx context.Context, sel [4]byte) ([]string, error) {
	key := "f:" + hex.EncodeToString(sel[:])
	if v, ok := r.cache.Load(key); ok {
		return v.([]string), nil
	}

	// Embedded 4byte: only resolves *known* selectors that are in the snapshot.
	if r.fourbyte != nil {
		if name, err := r.fourbyte.Selector(sel[:]); err == nil && name != "" {
			out := []string{name}
			r.cache.Store(key, out)
			return out, nil
		}
	}

	// Online openchain
	if r.openchain != nil {
		sigs, err := r.openchain.Signature("0x" + hex.EncodeToString(sel[:]))
		if err == nil && len(sigs) > 0 {
			r.cache.Store(key, sigs)
			return sigs, nil
		}
	}
	return nil, errors.New("no signature found")
}

// EventTopic resolves a 32-byte event topic0 to one or more candidate signatures.
func (r *SignatureResolver) EventTopic(ctx context.Context, topic0 [32]byte) ([]string, error) {
	key := "e:" + hex.EncodeToString(topic0[:])
	if v, ok := r.cache.Load(key); ok {
		return v.([]string), nil
	}
	if r.openchain != nil {
		sigs, err := r.openchain.Signature("0x" + hex.EncodeToString(topic0[:]))
		if err == nil && len(sigs) > 0 {
			r.cache.Store(key, sigs)
			return sigs, nil
		}
	}
	return nil, errors.New("no event signature found")
}

// PrimeCache pre-populates the cache with a known mapping (useful for tests).
func (r *SignatureResolver) PrimeCache(hexKey string, sigs []string) {
	r.cache.Store(hexKey, sigs)
}
