package chain

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

// CodeCache caches eth_getCode(addr) results so we can cheaply tell whether an
// address is an EOA (no code) or a contract.
type CodeCache struct {
	client *Client
	cache  sync.Map // common.Address -> bool (true = is contract)
}

// NewCodeCache constructs the cache.
func NewCodeCache(c *Client) *CodeCache {
	return &CodeCache{client: c}
}

// IsContract returns whether `addr` has bytecode at the latest block.
// Returns false on RPC error (which is the safe default for our classifier).
func (c *CodeCache) IsContract(ctx context.Context, addr common.Address) bool {
	if v, ok := c.cache.Load(addr); ok {
		return v.(bool)
	}
	code, err := c.client.eth.CodeAt(ctx, addr, nil)
	if err != nil {
		return false
	}
	is := len(code) > 0
	c.cache.Store(addr, is)
	return is
}
