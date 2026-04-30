// Package evmsemantic provides production-grade semantic classification of EVM
// transactions across Ethereum, Base, Arbitrum, Optimism, Polygon, BSC and any
// EVM-compatible chain.
//
// The library combines multiple signals — event logs, calldata selector, value
// transfer, and known-contract address tags — to determine whether a tx is a
// Transfer, Swap, Lend, Stake, Wrap, Bridge, NFT trade, or generic Contract Call.
//
// It composes mature upstream libraries (go-ethereum, dbadoy/signature, Sourcify
// REST API) and only adds a thin domain layer of protocol fingerprints +
// multi-signal fusion logic.
//
// Quick start:
//
//	ctx := context.Background()
//	e, err := evmsemantic.New(ctx, evmsemantic.ChainEthereum, "https://ethereum-rpc.publicnode.com")
//	if err != nil { log.Fatal(err) }
//	defer e.Close()
//	a, err := e.ClassifyTx(ctx, common.HexToHash("0xc4b084e8..."))
//	if err != nil { log.Fatal(err) }
//	fmt.Println(a.Category, a.Protocol, a.Summary)
package evmsemantic

import "github.com/jowency-me/evm-tx-parser/semtypes"

// Re-exports of the public data types so callers can write
// `evmsemantic.Action` instead of `semtypes.Action`.

// Action is the result of classification.
type Action = semtypes.Action

// TokenFlow is one decoded token movement.
type TokenFlow = semtypes.TokenFlow

// Category enumerates high-level transaction classes.
type Category = semtypes.Category

// AssetKind enumerates asset types.
type AssetKind = semtypes.AssetKind

// ChainID enumerates supported chains (the value is the standard EVM chain id).
type ChainID = semtypes.ChainID

// Re-exported constants.
const (
	CategoryTransfer     = semtypes.CategoryTransfer
	CategorySwap         = semtypes.CategorySwap
	CategoryLiquidity    = semtypes.CategoryLiquidity
	CategoryLend         = semtypes.CategoryLend
	CategoryStake        = semtypes.CategoryStake
	CategoryWrap         = semtypes.CategoryWrap
	CategoryBridge       = semtypes.CategoryBridge
	CategoryNFT          = semtypes.CategoryNFT
	CategoryApprove      = semtypes.CategoryApprove
	CategoryContractCall = semtypes.CategoryContractCall
	CategoryUnknown      = semtypes.CategoryUnknown

	AssetNative  = semtypes.AssetNative
	AssetERC20   = semtypes.AssetERC20
	AssetERC721  = semtypes.AssetERC721
	AssetERC1155 = semtypes.AssetERC1155

	ChainEthereum  = semtypes.ChainEthereum
	ChainOptimism  = semtypes.ChainOptimism
	ChainBSC       = semtypes.ChainBSC
	ChainPolygon   = semtypes.ChainPolygon
	ChainBase      = semtypes.ChainBase
	ChainArbitrum  = semtypes.ChainArbitrum
	ChainAvalanche = semtypes.ChainAvalanche
)
