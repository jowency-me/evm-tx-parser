package chain

import (
	"math/big"
	"testing"
)

func TestFormat(t *testing.T) {
	cases := []struct {
		raw      string
		decimals uint8
		maxDp    int
		want     string
	}{
		{"1000000", 6, 4, "1"},
		{"1234500", 6, 4, "1.2345"},
		{"123456789", 6, 2, "123.45"},
		{"1000000000000000000", 18, 4, "1"},
		{"1234567890000000000", 18, 4, "1.2345"},
		{"0", 18, 4, "0"},
	}
	for _, c := range cases {
		raw := new(big.Int)
		raw.SetString(c.raw, 10)
		got := Format(raw, c.decimals, c.maxDp)
		if got != c.want {
			t.Errorf("Format(%s, %d, %d) = %q, want %q", c.raw, c.decimals, c.maxDp, got, c.want)
		}
	}
}
