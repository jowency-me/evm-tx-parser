// Package classifier implements multi-signal fusion to determine the semantic
// Category of an EVM transaction. It does NOT make any RPC calls itself —
// callers pass a fully-fetched TxBundle plus a TokenCache for ERC-20 metadata.
//
// Design: we walk the receipt's logs and decode any well-known event topics
// into TokenFlows + protocol counters. We also peek at the top-level calldata
// selector and value transfer. We then map signal counts to a Category using
// a small set of rules:
//
//   - any *Swap event present  → Swap (protocol = best-known swap fingerprint)
//   - any Aave/Compound event → Lend
//   - any Lido/RocketPool/EigenLayer event → Stake
//   - any WETH Deposit/Withdrawal (and to-address is a wrapped token) → Wrap
//   - any Seaport/Blur or ERC-721/1155 transfer → NFT
//   - bridge messengers (Stargate/Across/CCTP) → Bridge
//   - else if the only logs are ERC-20 Transfer(s) → Transfer
//   - else → ContractCall
package classifier

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/jowency-me/evm-tx-parser/chain"
	"github.com/jowency-me/evm-tx-parser/protocols"
	semantic "github.com/jowency-me/evm-tx-parser/semtypes"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Classifier converts a TxBundle into a semantic Action.
type Classifier struct {
	chainID uint64
	tokens  *chain.TokenCache // optional — when present, flow Symbol/Decimals are filled
	code    *chain.CodeCache  // optional — when present, used to detect EOA destinations
}

// New creates a classifier for a given chain id. tokens may be nil.
func New(chainID uint64, tokens *chain.TokenCache) *Classifier {
	return &Classifier{chainID: chainID, tokens: tokens}
}

// NewWithCode creates a classifier with both token-metadata and code caches.
// The code cache lets us detect transfers to EOAs (where the calldata, if any,
// has no execution effect — think "ethscription" / inscription transactions).
func NewWithCode(chainID uint64, tokens *chain.TokenCache, code *chain.CodeCache) *Classifier {
	return &Classifier{chainID: chainID, tokens: tokens, code: code}
}

// Classify produces an Action from a TxBundle.
func (c *Classifier) Classify(ctx context.Context, b *chain.TxBundle) *semantic.Action {
	a := &semantic.Action{
		Signals: map[string]int{},
		Notes:   []string{},
	}
	if b.Receipt != nil && b.Receipt.Status == 0 {
		a.Notes = append(a.Notes, "transaction reverted (status=0)")
		a.Signals["reverted"] = 1
	}

	// 1) Top-level method + to-address tags
	c.fillTopLevel(b, a)

	// 2) Walk logs
	flows, sigs, protocolHints := c.scanLogs(ctx, b.Receipt)
	a.Flows = flows
	for k, v := range sigs {
		a.Signals[k] = v
	}

	// 3) Native value transfer
	if b.Tx.Value() != nil && b.Tx.Value().Sign() > 0 {
		a.Signals["native_value"] = 1
		// Add a synthetic native flow
		flow := semantic.TokenFlow{
			Kind:   semantic.AssetNative,
			Token:  common.Address{},
			From:   b.From,
			Amount: new(big.Int).Set(b.Tx.Value()),
		}
		if b.Tx.To() != nil {
			flow.To = *b.Tx.To()
		}
		// chain.SymbolForChain not implemented yet; use heuristic.
		flow.Symbol = nativeSymbol(c.chainID)
		flow.Decimals = 18
		a.Flows = append(a.Flows, flow)
	}

	// 4) Determine category via priority rules
	c.decide(a, b, protocolHints)

	// 5) Build summary string
	a.Summary = c.summarize(a)
	return a
}

// fillTopLevel reads top-level "to" address, calldata selector, and looks up tags.
func (c *Classifier) fillTopLevel(b *chain.TxBundle, a *semantic.Action) {
	if b.Tx.To() != nil {
		if tag := protocols.LookupAddress(c.chainID, *b.Tx.To()); tag != nil {
			a.Protocol = tag.Protocol
			a.Notes = append(a.Notes, fmt.Sprintf("to=%s (%s/%s)", b.Tx.To().Hex(), tag.Protocol, tag.Role))
			a.Signals["to_known_"+tag.Role] = 1
		}
	}
	data := b.Tx.Data()
	if len(data) >= 4 {
		var sel [4]byte
		copy(sel[:], data[:4])
		if md := protocols.LookupMethod(sel); md != nil {
			a.Method = md.Signature
			a.Signals["method_"+string(md.Category)] = 1
			a.Notes = append(a.Notes, fmt.Sprintf("method=%s (%s)", md.Signature, md.Hint))
		} else {
			a.Method = protocols.SelectorHex(sel)
		}
	}
}

// scanLogs walks receipt logs, decodes well-known events into TokenFlows,
// and counts signals + collects protocol hints from address tags.
func (c *Classifier) scanLogs(ctx context.Context, receipt *types.Receipt) ([]semantic.TokenFlow, map[string]int, map[string]int) {
	flows := []semantic.TokenFlow{}
	sigs := map[string]int{}
	protoHints := map[string]int{}

	if receipt == nil {
		return flows, sigs, protoHints
	}

	for _, lg := range receipt.Logs {
		if len(lg.Topics) == 0 {
			continue
		}
		topic0 := lg.Topics[0]
		ev := protocols.LookupEvent(topic0)
		if ev == nil {
			sigs["log_unknown"]++
			continue
		}
		sigs["evt_"+string(ev.Category)]++
		// Address-derived protocol hint
		if tag := protocols.LookupAddress(c.chainID, lg.Address); tag != nil {
			protoHints[tag.Protocol]++
		} else if ev.Protocol != "" {
			protoHints[ev.Protocol]++
		}

		// Decode flow if it's a token transfer
		switch ev.Category {
		case protocols.EvtTransferERC20:
			if protocols.IsERC20Transfer(len(lg.Topics), len(lg.Data)) {
				flow := decodeERC20Transfer(lg)
				c.enrichFlow(ctx, &flow)
				flows = append(flows, flow)
			} else if protocols.IsERC721Transfer(len(lg.Topics)) {
				flows = append(flows, decodeERC721Transfer(lg))
				sigs["evt_transfer_erc721"]++
			}
		case protocols.EvtTransferSingle:
			flows = append(flows, decodeERC1155Single(lg))
		case protocols.EvtWETHDeposit, protocols.EvtWETHWithdraw:
			// wrap/unwrap — treat as native↔wrapped flow
			flow := decodeWETHEvent(lg, ev.Category)
			c.enrichFlow(ctx, &flow)
			flows = append(flows, flow)
		}
	}
	return flows, sigs, protoHints
}

func (c *Classifier) enrichFlow(ctx context.Context, f *semantic.TokenFlow) {
	if c.tokens == nil || (f.Kind != semantic.AssetERC20) || f.Token == (common.Address{}) {
		return
	}
	m := c.tokens.Get(ctx, f.Token)
	if m.Symbol != "" {
		f.Symbol = m.Symbol
	}
	if m.Decimals > 0 {
		f.Decimals = m.Decimals
	}
}

// decide applies priority-based rules to set Action.Category.
func (c *Classifier) decide(a *semantic.Action, b *chain.TxBundle, protoHints map[string]int) {
	// Highest-priority rule: if the destination is an EOA (or no destination at all
	// for contract creation), the calldata cannot execute — the tx is at most a
	// native ETH transfer carrying inscription/ethscription bytes.
	if b.Tx.To() != nil && c.code != nil {
		if !c.code.IsContract(context.Background(), *b.Tx.To()) {
			if len(b.Tx.Data()) > 0 {
				a.Notes = append(a.Notes, "to is an EOA — calldata is inert (likely inscription)")
				a.Signals["to_is_eoa"] = 1
			}
			a.Category = semantic.CategoryTransfer
			return
		}
	}

	// Contract creation (to == nil)
	if b.Tx.To() == nil {
		a.Category = semantic.CategoryContractCall
		a.Notes = append(a.Notes, "contract creation")
		a.Signals["contract_creation"] = 1
		return
	}

	has := func(key string) bool { _, ok := a.Signals[key]; return ok }
	hasAny := func(prefixes ...string) bool {
		for k := range a.Signals {
			for _, p := range prefixes {
				if strings.HasPrefix(k, p) {
					return true
				}
			}
		}
		return false
	}

	// Helper: pick a protocol name from hints (the most-frequent one wins,
	// excluding the generic ERC20/721 family bucket which appears whenever
	// the tx involves token transfers — it's almost never the actual protocol).
	pickProtocol := func() string {
		best := ""
		bestN := 0
		for p, n := range protoHints {
			if p == "ERC20/721" || p == "ERC1155" || p == "ERC721/1155" {
				continue
			}
			if n > bestN {
				best = p
				bestN = n
			}
		}
		return best
	}
	// Prefer the top-level address tag (already in a.Protocol) when it is more specific.
	preferTop := func() string {
		if a.Protocol != "" {
			return a.Protocol
		}
		return pickProtocol()
	}

	// Bridge
	if hasAny("evt_bridge_") || isKnownBridgeProtocol(a.Protocol) {
		a.Category = semantic.CategoryBridge
		return
	}

	// NFT
	if hasAny("evt_nft_", "evt_transfer_erc721", "evt_transfer_single_erc1155", "evt_transfer_batch_erc1155") {
		a.Category = semantic.CategoryNFT
		a.Protocol = preferTop()
		return
	}

	// Swap (any DEX swap event)
	if hasAny("evt_swap_", "evt_gmx_") {
		a.Category = semantic.CategorySwap
		a.Protocol = preferTop()
		return
	}

	// Lend
	if hasAny("evt_aave_", "evt_compoundv2_", "evt_compoundv3_", "evt_morpho_", "evt_venus_") {
		a.Category = semantic.CategoryLend
		a.Protocol = preferTop()
		return
	}

	// Stake
	if hasAny("evt_lido_", "evt_rocket_", "evt_eigenlayer_", "evt_stakewise_", "evt_frax_") {
		a.Category = semantic.CategoryStake
		a.Protocol = preferTop()
		return
	}

	// Wrap
	if has("evt_weth_deposit") || has("evt_weth_withdraw") {
		a.Category = semantic.CategoryWrap
		a.Protocol = preferTop()
		return
	}

	// LP add/remove
	if has("evt_mint_lp") || has("evt_burn_lp") || has("evt_mint_lp_v3") || has("evt_burn_lp_v3") || has("evt_balancer_joined") || hasAny("evt_curve_add_liquidity", "evt_curve_remove_liquidity") || has("method_lp_add") || has("method_lp_remove") || has("method_lp_add_v3") || has("method_lp_remove_v3") {
		a.Category = semantic.CategoryLiquidity
		a.Protocol = preferTop()
		return
	}

	// Approve only
	if (has("evt_approval_erc20") || has("evt_approval_for_all") || has("method_erc20_approve") || has("method_permit2")) &&
		!has("evt_transfer_erc20") {
		a.Category = semantic.CategoryApprove
		return
	}

	// Pure transfer (only Transfer events, or simple ETH send)
	if has("evt_transfer_erc20") {
		// If the to-address is a known router/protocol, more likely a contract interaction wrapping a transfer
		if !hasAny("to_known_router", "to_known_pool", "to_known_vault") {
			a.Category = semantic.CategoryTransfer
			return
		}
	}
	if a.Signals["native_value"] == 1 && len(b.Receipt.Logs) == 0 {
		a.Category = semantic.CategoryTransfer
		return
	}
	// ETH pushed to a contract with no calldata (calldata_len==0) — receive()/fallback path.
	// Even if logs are emitted, from the user's perspective this is a transfer.
	if a.Signals["native_value"] == 1 && len(b.Tx.Data()) == 0 {
		a.Category = semantic.CategoryTransfer
		a.Notes = append(a.Notes, "native ETH transfer to contract via receive()/fallback")
		return
	}

	// Method-based fallbacks
	if has("method_erc20_transfer") || has("method_erc20_transfer_from") {
		a.Category = semantic.CategoryTransfer
		return
	}
	if has("method_universal_router_execute") || has("method_swap") {
		a.Category = semantic.CategorySwap
		a.Protocol = preferTop()
		return
	}

	a.Category = semantic.CategoryContractCall
	if a.Protocol == "" {
		a.Protocol = pickProtocol()
	}
}

func isKnownBridgeProtocol(name string) bool {
	if name == "" {
		return false
	}
	low := strings.ToLower(name)
	for _, k := range []string{"stargate", "across", "lifi", "hop", "synapse", "celer", "wormhole", "cctp"} {
		if strings.Contains(low, k) {
			return true
		}
	}
	return false
}

// nativeSymbol returns ETH/MATIC/BNB/etc. for known chains.
func nativeSymbol(chain uint64) string {
	switch chain {
	case 1, 10, 42161, 8453:
		return "ETH"
	case 137:
		return "MATIC"
	case 56:
		return "BNB"
	case 43114:
		return "AVAX"
	}
	return "NATIVE"
}
