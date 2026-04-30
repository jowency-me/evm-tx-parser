// Package chain wraps go-ethereum's ethclient with helpers needed by the classifier:
// fetching transaction + receipt + sender, and resolving ERC-20 token metadata
// (symbol, decimals) with an in-memory cache.
package chain

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// Client wraps an ethclient with chain-id awareness and convenience methods.
type Client struct {
	rpcURL  string
	chainID uint64
	rpc     *rpc.Client
	eth     *ethclient.Client
}

// New dials the given RPC and verifies its chain ID.
func New(ctx context.Context, chainID uint64, rpcURL string) (*Client, error) {
	rc, err := rpc.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("dial rpc: %w", err)
	}
	ec := ethclient.NewClient(rc)
	got, err := ec.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("chain id: %w", err)
	}
	if chainID != 0 && got.Uint64() != chainID {
		return nil, fmt.Errorf("chain id mismatch: configured=%d, rpc=%d", chainID, got.Uint64())
	}
	return &Client{rpcURL: rpcURL, chainID: got.Uint64(), rpc: rc, eth: ec}, nil
}

// Close releases underlying RPC resources.
func (c *Client) Close() {
	if c.rpc != nil {
		c.rpc.Close()
	}
}

// ChainID returns the chain id.
func (c *Client) ChainID() uint64 { return c.chainID }

// Eth returns the underlying ethclient.
func (c *Client) Eth() *ethclient.Client { return c.eth }

// Raw returns the raw RPC client (useful for trace_* and debug_* calls).
func (c *Client) Raw() *rpc.Client { return c.rpc }

// TxBundle is the data we need to classify a transaction.
type TxBundle struct {
	Hash    common.Hash
	Tx      *types.Transaction
	Receipt *types.Receipt
	From    common.Address
	Block   *big.Int
}

// FetchTx retrieves transaction + receipt + recovered sender.
func (c *Client) FetchTx(ctx context.Context, txHash common.Hash) (*TxBundle, error) {
	tx, isPending, err := c.eth.TransactionByHash(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("tx by hash: %w", err)
	}
	if isPending {
		return nil, errors.New("transaction is pending")
	}
	receipt, err := c.eth.TransactionReceipt(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("receipt: %w", err)
	}
	signer := types.LatestSignerForChainID(new(big.Int).SetUint64(c.chainID))
	from, err := types.Sender(signer, tx)
	if err != nil {
		return nil, fmt.Errorf("recover sender: %w", err)
	}
	return &TxBundle{
		Hash:    txHash,
		Tx:      tx,
		Receipt: receipt,
		From:    from,
		Block:   receipt.BlockNumber,
	}, nil
}
