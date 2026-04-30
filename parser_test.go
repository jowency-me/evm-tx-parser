package evmsemantic_test

import (
	"math/big"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/jowency-me/evm-tx-parser/chain"
	"github.com/jowency-me/evm-tx-parser/classifier"
	semantic "github.com/jowency-me/evm-tx-parser/semtypes"
)

// fixturePaths returns all fixture JSON files in testdata/transactions.
func fixturePaths() ([]string, error) {
	return filepath.Glob("testdata/transactions/*.json")
}

// classifyFixture loads and classifies a fixture.
func classifyFixture(t *testing.T, path string) (*fixture, *semantic.Action) {
	t.Helper()
	lf, err := loadFixture(path)
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	cls := classifier.New(lf.Fixture.ChainID, nil)
	bundle := &chain.TxBundle{
		Hash:    common.HexToHash(lf.Fixture.TxHash),
		Tx:      lf.Tx,
		Receipt: lf.Receipt,
		From:    common.HexToAddress(lf.Fixture.From),
		Block:   big.NewInt(int64(lf.Fixture.BlockNumber)),
	}
	action := cls.Classify(t.Context(), bundle)
	return lf.Fixture, action
}

// TestFixtureLoader verifies that every saved fixture can round-trip through
// JSON back into a parseable transaction.
func TestFixtureLoader(t *testing.T) {
	paths, err := fixturePaths()
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("no fixtures found in testdata/transactions")
	}

	for _, p := range paths {
		t.Run(filepath.Base(p), func(t *testing.T) {
			lf, err := loadFixture(p)
			if err != nil {
				t.Fatalf("load: %v", err)
			}
			if lf.Fixture.TxHash == "" {
				t.Fatal("missing txHash")
			}
			if lf.Tx == nil {
				t.Fatal("nil transaction")
			}
			if lf.Receipt == nil {
				t.Fatal("nil receipt")
			}
		})
	}
}

// TestClassifyAllFixtures classifies every saved fixture.
// Any transaction saved to testdata must be fully recognized:
// the category must not be Unknown, and key fields must be populated.
func TestClassifyAllFixtures(t *testing.T) {
	paths, err := fixturePaths()
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(paths) == 0 {
		t.Skip("no fixtures yet — run scan_test.go to collect")
	}

	for _, p := range paths {
		name := strings.TrimSuffix(filepath.Base(p), ".json")
		t.Run(name, func(t *testing.T) {
			f, action := classifyFixture(t, p)

			if action.Category == semantic.CategoryUnknown {
				t.Fatalf("%s: classified as Unknown (category=%s protocol=%s method=%s)",
					f.Description, action.Category, action.Protocol, action.Method)
			}

			// Category-specific field validation
			switch action.Category {
			case semantic.CategorySwap:
				validateSwap(t, f, action)
			case semantic.CategoryTransfer:
				validateTransfer(t, f, action)
			case semantic.CategoryLend:
				validateLend(t, f, action)
			case semantic.CategoryStake:
				validateStake(t, f, action)
			case semantic.CategoryWrap:
				validateWrap(t, f, action)
			case semantic.CategoryLiquidity:
				validateLiquidity(t, f, action)
			case semantic.CategoryBridge:
				validateBridge(t, f, action)
			case semantic.CategoryNFT:
				validateNFT(t, f, action)
			case semantic.CategoryApprove:
				// Approve is valid on its own
			}
		})
	}
}

func validateSwap(t *testing.T, f *fixture, a *semantic.Action) {
	t.Helper()
	erc20Flows := 0
	for _, flow := range a.Flows {
		if flow.Kind == semantic.AssetERC20 && flow.Amount != nil && flow.Amount.Sign() > 0 {
			erc20Flows++
		}
	}
	if erc20Flows == 0 && a.Signals["native_value"] == 0 {
		t.Errorf("swap has no ERC-20 flows and no native value")
	}
	if a.Protocol == "" {
		t.Error("swap missing protocol")
	}
}

func validateTransfer(t *testing.T, f *fixture, a *semantic.Action) {
	t.Helper()
	if len(a.Flows) == 0 && a.Signals["native_value"] == 0 {
		t.Error("transfer has no flows and no native value")
	}
}

func validateLend(t *testing.T, f *fixture, a *semantic.Action) {
	t.Helper()
	if a.Protocol == "" {
		t.Error("lend missing protocol")
	}
}

func validateStake(t *testing.T, f *fixture, a *semantic.Action) {
	t.Helper()
	if a.Protocol == "" {
		t.Error("stake missing protocol")
	}
}

func validateWrap(t *testing.T, f *fixture, a *semantic.Action) {
	t.Helper()
	if len(a.Flows) == 0 {
		t.Error("wrap has no flows")
	}
}

func validateLiquidity(t *testing.T, f *fixture, a *semantic.Action) {
	t.Helper()
	if a.Protocol == "" {
		t.Error("liquidity missing protocol")
	}
}

func validateBridge(t *testing.T, f *fixture, a *semantic.Action) {
	t.Helper()
	if a.Protocol == "" {
		t.Error("bridge missing protocol")
	}
}

func validateNFT(t *testing.T, f *fixture, a *semantic.Action) {
	t.Helper()
	nftFlows := 0
	for _, flow := range a.Flows {
		if flow.Kind == semantic.AssetERC721 || flow.Kind == semantic.AssetERC1155 {
			nftFlows++
		}
	}
	if nftFlows == 0 {
		t.Error("NFT category but no ERC-721/1155 flows")
	}
}
