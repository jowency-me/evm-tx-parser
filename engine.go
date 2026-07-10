package evmsemantic

import (
	"context"
	"fmt"

	"github.com/jowency-me/evm-tx-parser/chain"
	"github.com/jowency-me/evm-tx-parser/classifier"
	"github.com/jowency-me/evm-tx-parser/resolver"
	"github.com/jowency-me/evm-tx-parser/semtypes"

	"github.com/ethereum/go-ethereum/common"
)

// ensure import used (in case the build context drops it elsewhere)
var _ = semtypes.CategoryUnknown

// Engine is the top-level facade for the library.
//
//	e, err := evmsemantic.New(ctx, evmsemantic.ChainEthereum, "https://ethereum-rpc.publicnode.com")
//	if err != nil { ... }
//	defer e.Close()
//	a, err := e.ClassifyTx(ctx, common.HexToHash("0x..."))
//	fmt.Println(a.Summary)
type Engine struct {
	chain      *chain.Client
	classifier *classifier.Classifier
	tokens     *chain.TokenCache
	sigRes     *resolver.SignatureResolver // optional, for unknown selectors
	abiRes     *resolver.ABIResolver       // optional, for richer decoding
}

// New constructs an Engine bound to a single chain + RPC endpoint.
// Multiple engines may be created in parallel for multi-chain analysis.
func New(ctx context.Context, chainID ChainID, rpcURL string) (*Engine, error) {
	c, err := chain.New(ctx, uint64(chainID), rpcURL)
	if err != nil {
		return nil, fmt.Errorf("chain client: %w", err)
	}
	tokens := chain.NewTokenCache(c)
	code := chain.NewCodeCache(c)
	cls := classifier.NewWithCode(uint64(chainID), tokens, code)
	return &Engine{
		chain:      c,
		classifier: cls,
		tokens:     tokens,
		sigRes:     resolver.NewSignatureResolver(),
		abiRes:     resolver.NewABIResolver(),
	}, nil
}

// Close releases resources.
func (e *Engine) Close() {
	if e.chain != nil {
		e.chain.Close()
	}
}

// ClassifyTx classifies a single transaction by hash.
func (e *Engine) ClassifyTx(ctx context.Context, txHash common.Hash) (*semtypes.Action, error) {
	b, err := e.chain.FetchTx(ctx, txHash)
	if err != nil {
		return nil, err
	}
	return e.classifier.Classify(ctx, b), nil
}

// ClassifyBundle classifies a pre-fetched transaction bundle (tx + receipt + sender),
// avoiding an extra RPC round-trip when the caller already has the data — e.g. a
// webhook pipeline that parsed the block payload and has the calldata + logs in hand.
// This is the same classification as ClassifyTx but without the chain fetch.
func (e *Engine) ClassifyBundle(ctx context.Context, b *chain.TxBundle) (*semtypes.Action, error) {
	if b == nil {
		return nil, fmt.Errorf("classify bundle: nil bundle")
	}
	return e.classifier.Classify(ctx, b), nil
}

// SignatureResolver returns the signature resolver (callers may prime it or look up sigs directly).
func (e *Engine) SignatureResolver() *resolver.SignatureResolver { return e.sigRes }

// ABIResolver returns the ABI resolver.
func (e *Engine) ABIResolver() *resolver.ABIResolver { return e.abiRes }

// TokenCache returns the ERC-20 metadata cache.
func (e *Engine) TokenCache() *chain.TokenCache { return e.tokens }
