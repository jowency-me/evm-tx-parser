package classifier

import (
	"fmt"
	"strings"

	"github.com/jowency-me/evm-tx-parser/chain"
	semantic "github.com/jowency-me/evm-tx-parser/semtypes"
)

// summarize produces a concise human-readable description of the action.
func (c *Classifier) summarize(a *semantic.Action) string {
	proto := a.Protocol
	switch a.Category {
	case semantic.CategoryTransfer:
		return summarizeTransfer(a.Flows)
	case semantic.CategorySwap:
		return summarizeSwap(a.Flows, proto)
	case semantic.CategoryWrap:
		return summarizeWrap(a.Flows, proto)
	case semantic.CategoryStake:
		return summarizeWithProto("Stake", a.Flows, proto)
	case semantic.CategoryLend:
		return summarizeWithProto("Lend", a.Flows, proto)
	case semantic.CategoryLiquidity:
		return summarizeWithProto("Liquidity", a.Flows, proto)
	case semantic.CategoryNFT:
		return summarizeNFT(a.Flows, proto)
	case semantic.CategoryBridge:
		return summarizeWithProto("Bridge", a.Flows, proto)
	case semantic.CategoryApprove:
		return "Approve token allowance"
	case semantic.CategoryContractCall:
		s := "Contract call"
		if proto != "" {
			s += " — " + proto
		}
		if a.Method != "" {
			s += " via " + truncSig(a.Method)
		}
		return s
	}
	return "Unknown"
}

func summarizeTransfer(flows []semantic.TokenFlow) string {
	if len(flows) == 0 {
		return "Transfer (no flows decoded)"
	}
	parts := []string{}
	for _, f := range flows {
		parts = append(parts, formatFlow(f))
	}
	return "Transfer " + strings.Join(parts, " + ")
}

func summarizeSwap(flows []semantic.TokenFlow, proto string) string {
	in, out := pickInOut(flows)
	if in != nil && out != nil {
		s := fmt.Sprintf("Swap %s → %s", formatFlow(*in), formatFlow(*out))
		if proto != "" {
			s += " on " + proto
		}
		return s
	}
	if proto != "" {
		return "Swap on " + proto
	}
	return "Swap"
}

func summarizeWrap(flows []semantic.TokenFlow, proto string) string {
	if proto == "" {
		proto = "WETH"
	}
	if len(flows) > 0 {
		return fmt.Sprintf("Wrap/Unwrap %s on %s", formatFlow(flows[0]), proto)
	}
	return "Wrap/Unwrap on " + proto
}

func summarizeWithProto(verb string, flows []semantic.TokenFlow, proto string) string {
	if proto == "" {
		proto = "?"
	}
	if len(flows) > 0 {
		// Show first non-zero amount flow
		for _, f := range flows {
			if f.Amount != nil && f.Amount.Sign() > 0 {
				return fmt.Sprintf("%s %s on %s", verb, formatFlow(f), proto)
			}
		}
	}
	return verb + " on " + proto
}

func summarizeNFT(flows []semantic.TokenFlow, proto string) string {
	cnt := 0
	for _, f := range flows {
		if f.Kind == semantic.AssetERC721 || f.Kind == semantic.AssetERC1155 {
			cnt++
		}
	}
	if proto == "" {
		proto = "NFT"
	}
	if cnt > 0 {
		return fmt.Sprintf("NFT × %d on %s", cnt, proto)
	}
	return "NFT on " + proto
}

// pickInOut heuristic: for a swap, in = the first ERC-20 transfer from sender's perspective
// = whatever flow has the largest "round-trip" delta. Without sender context here we
// just take the first ERC-20 flow as "in" and the last as "out".
func pickInOut(flows []semantic.TokenFlow) (*semantic.TokenFlow, *semantic.TokenFlow) {
	erc20 := []semantic.TokenFlow{}
	for _, f := range flows {
		if f.Kind == semantic.AssetERC20 && f.Amount != nil && f.Amount.Sign() > 0 {
			erc20 = append(erc20, f)
		}
	}
	if len(erc20) >= 2 {
		first := erc20[0]
		last := erc20[len(erc20)-1]
		return &first, &last
	}
	return nil, nil
}

func formatFlow(f semantic.TokenFlow) string {
	sym := f.Symbol
	if sym == "" {
		switch f.Kind {
		case semantic.AssetNative:
			sym = "NATIVE"
		case semantic.AssetERC721:
			if f.TokenID != nil {
				return fmt.Sprintf("NFT#%s", f.TokenID.String())
			}
			return "NFT"
		case semantic.AssetERC1155:
			if f.TokenID != nil {
				return fmt.Sprintf("1155#%s×%s", f.TokenID.String(), f.Amount)
			}
			return "1155"
		default:
			short := f.Token.Hex()
			if len(short) > 10 {
				short = short[:6] + "…" + short[len(short)-4:]
			}
			sym = short
		}
	}
	if f.Amount == nil {
		return sym
	}
	dp := 4
	return chain.Format(f.Amount, f.Decimals, dp) + " " + sym
}

func truncSig(s string) string {
	if i := strings.IndexByte(s, '('); i > 0 {
		return s[:i]
	}
	if len(s) > 10 {
		return s[:10]
	}
	return s
}
