package resolver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// ABIResolver fetches the full ABI of a verified contract from Sourcify (default)
// with optional Etherscan-compatible fallback.
type ABIResolver struct {
	httpClient    *http.Client
	sourcifyURL   string
	cache         sync.Map // (chain,addr) -> *abi.ABI or noABISentinel
	negativeCache sync.Map // (chain,addr) -> time.Time when last failed
	negativeTTL   time.Duration
	// Optional Etherscan-compatible providers per chain. Map: chainID -> EtherscanLike.
	etherscan map[uint64]EtherscanLike
}

// EtherscanLike is anything that returns ABI JSON for an address.
type EtherscanLike interface {
	GetContractABI(ctx context.Context, addr common.Address) (string, error)
}

func NewABIResolver() *ABIResolver {
	return &ABIResolver{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		// Sourcify v2 unified endpoint (covers verified contracts on all major chains).
		sourcifyURL: "https://sourcify.dev/server/v2/contract",
		negativeTTL: 5 * time.Minute,
		etherscan:   map[uint64]EtherscanLike{},
	}
}

// WithEtherscan registers an Etherscan-like fallback for a specific chain.
func (r *ABIResolver) WithEtherscan(chain uint64, e EtherscanLike) *ABIResolver {
	r.etherscan[chain] = e
	return r
}

type cacheKey struct {
	Chain uint64
	Addr  common.Address
}

type sourcifyResp struct {
	ABI json.RawMessage `json:"abi"`
}

// FetchABI tries Sourcify, then any Etherscan-like provider for chain.
func (r *ABIResolver) FetchABI(ctx context.Context, chain uint64, addr common.Address) (*abi.ABI, error) {
	k := cacheKey{chain, addr}
	if v, ok := r.cache.Load(k); ok {
		if v == nil {
			return nil, errors.New("ABI cached as not available")
		}
		return v.(*abi.ABI), nil
	}
	if t, ok := r.negativeCache.Load(k); ok {
		if time.Since(t.(time.Time)) < r.negativeTTL {
			return nil, errors.New("ABI not available (negative cache)")
		}
	}

	// Try Sourcify
	if a, err := r.fetchSourcify(ctx, chain, addr); err == nil {
		r.cache.Store(k, a)
		return a, nil
	}

	// Try Etherscan-like
	if es, ok := r.etherscan[chain]; ok && es != nil {
		raw, err := es.GetContractABI(ctx, addr)
		if err == nil && raw != "" && raw != "Contract source code not verified" {
			parsed, err := abi.JSON(stringReader(raw))
			if err == nil {
				r.cache.Store(k, &parsed)
				return &parsed, nil
			}
		}
	}

	r.negativeCache.Store(k, time.Now())
	return nil, errors.New("ABI not found")
}

func (r *ABIResolver) fetchSourcify(ctx context.Context, chain uint64, addr common.Address) (*abi.ABI, error) {
	url := fmt.Sprintf("%s/%d/%s?fields=abi", r.sourcifyURL, chain, addr.Hex())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("not verified on Sourcify")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sourcify status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var sr sourcifyResp
	if err := json.Unmarshal(body, &sr); err != nil {
		return nil, err
	}
	if len(sr.ABI) == 0 || string(sr.ABI) == "null" {
		return nil, errors.New("empty ABI from sourcify")
	}
	parsed, err := abi.JSON(stringReader(string(sr.ABI)))
	if err != nil {
		return nil, fmt.Errorf("parse abi: %w", err)
	}
	return &parsed, nil
}

// stringReader converts string to io.Reader without bringing strings package.
func stringReader(s string) *byteReader { return &byteReader{b: []byte(s)} }

type byteReader struct {
	b   []byte
	pos int
}

func (r *byteReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.b) {
		return 0, io.EOF
	}
	n := copy(p, r.b[r.pos:])
	r.pos += n
	return n, nil
}
