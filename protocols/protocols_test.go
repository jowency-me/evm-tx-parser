package protocols

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestEventCatalogTopicsValid(t *testing.T) {
	for _, ev := range EventCatalog {
		want := crypto.Keccak256Hash([]byte(ev.Signature))
		if ev.Topic != want {
			t.Errorf("topic mismatch for %s: got %s want %s", ev.Signature, ev.Topic.Hex(), want.Hex())
		}
	}
}

func TestLookupEvent(t *testing.T) {
	// ERC-20 Transfer
	topic := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
	ev := LookupEvent(topic)
	if ev == nil {
		t.Fatal("expected to find Transfer event")
	}
	if ev.Category != EvtTransferERC20 {
		t.Errorf("got category %s", ev.Category)
	}
}

func TestLookupMethod_ERC20Transfer(t *testing.T) {
	want := sel("transfer(address,uint256)")
	md := LookupMethod(want)
	if md == nil {
		t.Fatal("nil method def")
	}
	if md.Category != MtdERC20Transfer {
		t.Errorf("category=%s", md.Category)
	}
}

func TestLookupAddress_Multichain(t *testing.T) {
	cases := []struct {
		chain    uint64
		addr     string
		protocol string
	}{
		{1, "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", "WETH"},
		{8453, "0x4200000000000000000000000000000000000006", "WETH (Base)"},
		{42161, "0x794a61358D6845594F94dc1DB02A252b5b4814aD", "Aave V3"},
	}
	for _, c := range cases {
		tag := LookupAddress(c.chain, common.HexToAddress(c.addr))
		if tag == nil {
			t.Errorf("missing tag for chain %d %s", c.chain, c.addr)
			continue
		}
		if !strings.Contains(tag.Protocol, c.protocol) {
			t.Errorf("got protocol %q want contains %q", tag.Protocol, c.protocol)
		}
	}
}

func TestNormalizeSig(t *testing.T) {
	got := NormalizeSig("transfer ( address , uint256 )")
	if got != "transfer(address,uint256)" {
		t.Errorf("got %q", got)
	}
}
