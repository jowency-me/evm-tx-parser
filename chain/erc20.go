package chain

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// Minimal ERC-20 ABI for symbol() and decimals(). Some non-compliant tokens (MKR,
// SAI) return bytes32 for symbol — we attempt string first and fall back.
const erc20MetaABI = `[
{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"stateMutability":"view","type":"function"},
{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"stateMutability":"view","type":"function"},
{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"stateMutability":"view","type":"function"}
]`

const erc20MetaABIBytes32 = `[
{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"bytes32"}],"stateMutability":"view","type":"function"}
]`

// TokenMeta holds ERC-20 token metadata.
type TokenMeta struct {
	Address  common.Address
	Symbol   string
	Decimals uint8
	Name     string
}

// TokenCache resolves and caches ERC-20 token metadata via on-chain calls.
// Lookups are concurrency-safe and best-effort: if a token does not implement
// symbol()/decimals(), we cache a placeholder rather than failing.
type TokenCache struct {
	client     *Client
	cache      sync.Map // common.Address -> *TokenMeta
	abiStr     abi.ABI
	abiBytes32 abi.ABI
}

// NewTokenCache constructs the cache. It panics if internal ABIs fail to parse
// (which would only happen if the constants above are corrupted).
func NewTokenCache(c *Client) *TokenCache {
	a1, err := abi.JSON(strings.NewReader(erc20MetaABI))
	if err != nil {
		panic(err)
	}
	a2, err := abi.JSON(strings.NewReader(erc20MetaABIBytes32))
	if err != nil {
		panic(err)
	}
	return &TokenCache{client: c, abiStr: a1, abiBytes32: a2}
}

// Get returns metadata for token, fetching it on-chain if not cached.
// Always returns a non-nil meta (with at least the address); errors are logged into the meta.
func (t *TokenCache) Get(ctx context.Context, token common.Address) *TokenMeta {
	if v, ok := t.cache.Load(token); ok {
		return v.(*TokenMeta)
	}
	m := &TokenMeta{Address: token}
	t.fillSymbol(ctx, token, m)
	t.fillDecimals(ctx, token, m)
	t.fillName(ctx, token, m)
	t.cache.Store(token, m)
	return m
}

// Preload populates the cache concurrently for the given tokens.
func (t *TokenCache) Preload(ctx context.Context, tokens []common.Address) {
	var wg sync.WaitGroup
	for _, tok := range tokens {
		if _, ok := t.cache.Load(tok); ok {
			continue
		}
		wg.Add(1)
		go func(a common.Address) {
			defer wg.Done()
			t.Get(ctx, a)
		}(tok)
	}
	wg.Wait()
}

func (t *TokenCache) call(ctx context.Context, abiDef abi.ABI, addr common.Address, method string) ([]any, error) {
	bound := bind.NewBoundContract(addr, abiDef, t.client.eth, t.client.eth, t.client.eth)
	var out []any
	if err := bound.Call(&bind.CallOpts{Context: ctx}, &out, method); err != nil {
		return nil, err
	}
	return out, nil
}

func (t *TokenCache) fillSymbol(ctx context.Context, addr common.Address, m *TokenMeta) {
	out, err := t.call(ctx, t.abiStr, addr, "symbol")
	if err == nil && len(out) == 1 {
		if s, ok := out[0].(string); ok && s != "" {
			m.Symbol = s
			return
		}
	}
	// Fallback bytes32
	out, err = t.call(ctx, t.abiBytes32, addr, "symbol")
	if err == nil && len(out) == 1 {
		if b, ok := out[0].([32]byte); ok {
			m.Symbol = strings.TrimRight(string(b[:]), "\x00")
		}
	}
}

func (t *TokenCache) fillDecimals(ctx context.Context, addr common.Address, m *TokenMeta) {
	out, err := t.call(ctx, t.abiStr, addr, "decimals")
	if err == nil && len(out) == 1 {
		if d, ok := out[0].(uint8); ok {
			m.Decimals = d
		}
	}
}

func (t *TokenCache) fillName(ctx context.Context, addr common.Address, m *TokenMeta) {
	out, err := t.call(ctx, t.abiStr, addr, "name")
	if err == nil && len(out) == 1 {
		if s, ok := out[0].(string); ok {
			m.Name = s
		}
	}
}

// Format converts a raw token amount to a human-readable string with at most maxDp decimal places.
// E.g. (1000000, 6, 4) → "1.0000".
func Format(raw *big.Int, decimals uint8, maxDp int) string {
	return formatRaw(raw, decimals, maxDp)
}

// We declare big.Int as math/big.Int but only via this unexported helper to avoid an extra import in the public sig.
func formatRaw(raw *big.Int, decimals uint8, maxDp int) string {
	if raw == nil {
		return "0"
	}
	if decimals == 0 {
		return raw.String()
	}
	// Build divisor 10^decimals
	div := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	q, r := new(big.Int).QuoRem(raw, div, new(big.Int))
	if r.Sign() == 0 {
		return q.String()
	}
	// Frac, padded
	fracStr := r.String()
	for len(fracStr) < int(decimals) {
		fracStr = "0" + fracStr
	}
	if maxDp > 0 && len(fracStr) > maxDp {
		fracStr = fracStr[:maxDp]
	}
	// strip trailing zeros
	fracStr = strings.TrimRight(fracStr, "0")
	if fracStr == "" {
		return q.String()
	}
	return fmt.Sprintf("%s.%s", q.String(), fracStr)
}
