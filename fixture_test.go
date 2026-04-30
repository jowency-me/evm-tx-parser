package evmsemantic_test

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/core/types"
)

// fixture wraps a test fixture JSON file.
type fixture struct {
	Description string          `json:"description"`
	Category    string          `json:"category"`
	ChainID     uint64          `json:"chainId"`
	TxHash      string          `json:"txHash"`
	BlockNumber uint64          `json:"blockNumber,omitempty"`
	From        string          `json:"from,omitempty"`
	ExplorerURL string          `json:"explorerURL"`
	Raw         json.RawMessage `json:"raw"` // contains { "tx": ..., "receipt": ... }
}

// rawBundle is the JSON structure inside fixture.Raw.
type rawBundle struct {
	Tx      json.RawMessage `json:"tx"`
	Receipt json.RawMessage `json:"receipt"`
}

// loadedFixture holds a fully-parsed fixture with Go types.
type loadedFixture struct {
	Fixture *fixture
	Tx      *types.Transaction
	Receipt *types.Receipt
}

// loadFixture reads a fixture JSON file and returns parsed components.
func loadFixture(path string) (*loadedFixture, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	var f fixture
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("unmarshal wrapper: %w", err)
	}

	var raw rawBundle
	if err := json.Unmarshal(f.Raw, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal raw: %w", err)
	}

	tx := &types.Transaction{}
	if err := tx.UnmarshalJSON(raw.Tx); err != nil {
		return nil, fmt.Errorf("unmarshal tx: %w", err)
	}

	receipt := &types.Receipt{}
	if err := json.Unmarshal(raw.Receipt, receipt); err != nil {
		return nil, fmt.Errorf("unmarshal receipt: %w", err)
	}

	return &loadedFixture{
		Fixture: &f,
		Tx:      tx,
		Receipt: receipt,
	}, nil
}
