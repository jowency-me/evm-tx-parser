# EVM TX Parser

A production-grade EVM transaction semantic classifier with **single RPC call** design and **multi-signal fusion** classification.

## Supported Transaction Types

| Type | Details |
|------|---------|
| **Swap** | DEX swaps including Uniswap V2/V3/V4/Universal Router, Curve (stableswap/crypto/NG), Balancer V2, 1inch, 0x, CowSwap, PancakeSwap V2/V3, QuickSwap, Aerodrome/Velodrome, SushiSwap, Trader Joe LB, GMX, ParaSwap, Kyber, DODO |
| **Lend** | Supply/borrow/repay/withdraw/liquidate on Aave V2/V3, Compound V2/V3, Morpho Blue, MakerDAO, Spark, Venus |
| **Stake** | Lido (stETH/wstETH), Rocket Pool (rETH), EigenLayer, Coinbase (cbETH), Frax (sfrxETH), Stakewise (sETH2/osETH), Eth2 Beacon Deposit |
| **Liquidity** | Add/remove liquidity on Uniswap V2/V3 LP, Curve LP, Balancer V2 LP |
| **Bridge** | Stargate, Across, Hop, CCTP, Wormhole, Synapse, Celer, LI.FI |
| **NFT** | OpenSea Seaport 1.5/1.6, Blur, ERC-721/1155 transfers |
| **Transfer** | Native ETH, ERC-20, ERC-721, ERC-1155 |
| **Wrap** | WETH / WMATIC / WBNB and other wrapped native tokens |
| **Approve** | ERC-20/721 approval and Permit2 |
| **ContractCall** | Fallback for uncategorized contract interactions |

### What Gets Extracted

Every classified transaction returns:
- Category, Protocol name, Method signature
- Human-readable summary (e.g. "Swap 100 USDC for 0.03 WETH on Uniswap V3")
- Token transfer flows: from/to addresses, token, amount, symbol, decimals
- Reverted status detection
- Signal breakdown (event types, method categories)

## Classification Architecture

The classifier combines three signal sources:

1. **Event logs** — Decode receipt log topics against known event signatures (e.g. Uniswap `Swap`, Aave `Supply`, Lido `Submitted`)
2. **Calldata selector** — Match 4-byte method selectors to known method signatures
3. **Address tags** — Look up `to` address and log-emitting addresses against a known contract address registry

Priority rules: Bridge > NFT > Swap > Lend > Stake > Wrap > Liquidity > Approve > Transfer > Method fallbacks > ContractCall.

## Supported Protocols

### Supported

| Category | Protocol | Chains |
|----------|----------|--------|
| Swap | Uniswap V2 | Ethereum |
| Swap | Uniswap V3 | Ethereum, Base, Arbitrum, Optimism |
| Swap | Uniswap V4 PoolManager | Ethereum |
| Swap | Uniswap Universal Router | Ethereum, Base, Arbitrum |
| Swap | Curve 3pool / stETH | Ethereum |
| Swap | Balancer V2 | Ethereum |
| Swap | CowSwap | Ethereum |
| Swap | PancakeSwap V2 / V3 | BSC |
| Swap | Aerodrome / SlipStream | Base |
| Swap | BaseSwap | Base |
| Swap | GMX | Arbitrum |
| Swap | 1inch | Ethereum |
| Swap | 0x | Ethereum |
| Swap | SushiSwap | Ethereum |
| Swap | Kyber | Ethereum |
| Swap | ParaSwap | Ethereum |
| Lend | Aave V2 | Ethereum |
| Lend | Aave V3 | Ethereum, Base, Arbitrum, Optimism |
| Lend | Compound V2 / V3 | Ethereum |
| Lend | MakerDAO | Ethereum |
| Stake | Lido | Ethereum |
| Stake | Rocket Pool | Ethereum |
| Stake | EigenLayer | Ethereum |
| Stake | Eth2 Beacon Deposit | Ethereum |
| Stake | Coinbase (cbETH) | Ethereum |
| Stake | Frax (sfrxETH) | Ethereum |
| Stake | Stakewise | Ethereum |
| Bridge | Stargate | Arbitrum, BSC |
| Bridge | CCTP | Arbitrum |
| Bridge | Wormhole | Ethereum |
| Bridge | LI.FI | Ethereum |
| NFT | OpenSea Seaport | Ethereum, Base, Arbitrum, Optimism |
| NFT | Blur | Ethereum |

### Unsupported

| Category | Protocol | Chains | Notes |
|----------|----------|--------|-------|
| Swap | DODO | Ethereum | |
| Swap | QuickSwap | Polygon | |
| Swap | Velodrome | Optimism | |
| Swap | Trader Joe LB | Avalanche | |
| Lend | Morpho Blue | Ethereum | |
| Lend | Spark | Ethereum | |
| Lend | Venus | BSC | |
| Bridge | Across | Multi-chain | |
| Bridge | Hop | Multi-chain | |
| Bridge | CCTP | Avalanche | No Avalanche RPC available |
| Bridge | Synapse | Ethereum | |
| Bridge | Celer | Ethereum | |
| NFT | OpenSea Seaport 1.5 | Ethereum | |
| Liquidity | Uniswap V2 LP | Multi-chain | |
| Liquidity | Uniswap V3 LP | Multi-chain | |
| Liquidity | Curve LP | Ethereum | |
| Liquidity | Balancer V2 LP | Ethereum | |

## Multi-Chain Support

| Chain | ChainID |
|-------|---------|
| Ethereum | 1 |
| Optimism | 10 |
| BSC | 56 |
| Polygon | 137 |
| Base | 8453 |
| Arbitrum | 42161 |
| Avalanche | 43114 |

Custom addresses for any EVM chain can be registered in `protocols/addresses.go`.

## Installation

```bash
go get github.com/jowency-me/evm-tx-parser
```

## Usage

### Classify by Transaction Hash (with RPC)

```go
import (
    "context"
    "fmt"
    "log"

    "github.com/ethereum/go-ethereum/common"
    evmtxparser "github.com/jowency-me/evm-tx-parser"
)

func main() {
    ctx := context.Background()
    eng, err := evmtxparser.New(ctx, evmtxparser.ChainEthereum, "https://ethereum-rpc.publicnode.com")
    if err != nil { log.Fatal(err) }
    defer eng.Close()

    a, err := eng.ClassifyTx(ctx, common.HexToHash("0xee80..."))
    if err != nil { log.Fatal(err) }

    fmt.Println(a.Category)  // Lend
    fmt.Println(a.Protocol)  // Aave V3
    fmt.Println(a.Method)    // supply(address,uint256,address,uint16)
    fmt.Println(a.Summary)   // Lend 40000 USDe on Aave V3
}
```

### Classify Pre-fetched Data (no RPC)

```go
a := eng.Classify(ctx, &evmtxparser.TxBundle{Tx: tx, Receipt: receipt, From: fromAddr})
```

## Output Structure

```go
type Action struct {
    Category Category         // Transfer / Swap / Lend / Stake / Wrap / Bridge / Liquidity / NFT / Approve / ContractCall
    Protocol string           // e.g. "Uniswap V3", "Aave V3", "Lido"
    Method   string           // Top-level calldata method signature
    Summary  string           // Human-readable one-liner
    Flows    []TokenFlow      // Decoded token transfer flows
    Signals  map[string]int   // Internal signal counters
    Notes    []string         // Notes (e.g. "transaction reverted")
}

type TokenFlow struct {
    Kind     AssetKind        // Native / ERC20 / ERC721 / ERC1155
    Token    common.Address   // Token contract address (zero for native)
    From     common.Address
    To       common.Address
    Amount   *big.Int
    Symbol   string
    Decimals uint8
}
```

## Design

1. **Single RPC call** — fetch transaction + receipt once; classifier works on pre-fetched data.
2. **Multi-signal fusion** — event logs + calldata selector + address tags are combined to determine category.
3. **Priority-based rules** — unambiguous ordering prevents classification conflicts (e.g. Bridge before Swap).
4. **Collision-safe** — event topic hash duplicates panic at init; method selector collisions use last-writer-wins with explicit comments.
5. **Extensible** — add protocol fingerprints in `protocols/` (events, methods, addresses) with no code changes elsewhere.

## Limitations

- Internal calls via `trace_*` are not parsed; proxy contract coverage relies on event logs
- Bridge send/receive correlation requires querying multiple chains
- Advanced NFT operations (Blur auctions, Sudoswap) receive basic classification

## Testing

```bash
go test ./...                            # Run all tests
go test -run TestClassifyAllFixtures -v  # Run fixture classification tests
```

Tests use local fixture files stored in `testdata/transactions/`.

## License

Apache-2.0
