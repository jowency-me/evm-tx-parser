// Package semtypes contains the public data types returned by the classifier.
// Kept in a separate package to break the import cycle between the root engine
// package and the classifier package.
package semtypes

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Category is the high-level semantic class of a transaction.
type Category string

const (
	CategoryTransfer     Category = "Transfer"     // ETH or ERC-20/721/1155 transfer
	CategorySwap         Category = "Swap"         // DEX swap (any DEX)
	CategoryLiquidity    Category = "Liquidity"    // Add/remove LP
	CategoryLend         Category = "Lend"         // Aave/Compound/Morpho/etc.
	CategoryStake        Category = "Stake"        // Lido/Rocket Pool/EigenLayer/native staking
	CategoryWrap         Category = "Wrap"         // WETH/WMATIC deposit/withdraw
	CategoryBridge       Category = "Bridge"       // Cross-chain bridge
	CategoryNFT          Category = "NFT"          // Mint/sale/transfer of ERC-721/1155
	CategoryApprove      Category = "Approve"      // ERC-20/721 approval (no other action)
	CategoryContractCall Category = "ContractCall" // Generic contract call
	CategoryUnknown      Category = "Unknown"
)

// AssetKind distinguishes asset types in TokenFlow.
type AssetKind string

const (
	AssetNative  AssetKind = "native" // ETH, MATIC, BNB
	AssetERC20   AssetKind = "erc20"
	AssetERC721  AssetKind = "erc721"
	AssetERC1155 AssetKind = "erc1155"
)

// TokenFlow represents one token movement extracted from logs (Transfer/Swap/Deposit/Withdrawal/etc.).
type TokenFlow struct {
	Kind     AssetKind
	Token    common.Address // zero address for native ETH
	Symbol   string         // resolved on-chain (cached) when Resolver provided; empty otherwise
	Decimals uint8
	From     common.Address
	To       common.Address
	Amount   *big.Int // raw amount (wei or token base units); nil for ERC-721 if TokenID is used
	TokenID  *big.Int // for ERC-721/1155
}

// Action is the result of classification — a structured human-readable description.
type Action struct {
	Category Category
	Protocol string // e.g. "Uniswap V3", "Aave V3", "Lido", "WETH"
	Method   string // top-level method signature (e.g. "swap(...)") — best effort
	Summary  string // one-line natural-language description, e.g. "Swap 100 USDC -> 0.03 WETH on Uniswap V3"
	Flows    []TokenFlow
	// Diagnostic counters and raw signals — useful for debugging or building UIs.
	Signals map[string]int
	Notes   []string
}

// ChainID is a typed wrapper for clarity.
type ChainID uint64

const (
	ChainEthereum  ChainID = 1
	ChainOptimism  ChainID = 10
	ChainBSC       ChainID = 56
	ChainPolygon   ChainID = 137
	ChainBase      ChainID = 8453
	ChainArbitrum  ChainID = 42161
	ChainAvalanche ChainID = 43114
)
