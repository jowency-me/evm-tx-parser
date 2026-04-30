package protocols

import (
	"encoding/hex"
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
)

// MethodCategory hints what action the top-level calldata represents.
type MethodCategory string

const (
	MtdERC20Transfer      MethodCategory = "erc20_transfer"
	MtdERC20TransferFrom  MethodCategory = "erc20_transfer_from"
	MtdERC20Approve       MethodCategory = "erc20_approve"
	MtdERC721Mint         MethodCategory = "erc721_mint"
	MtdERC721SafeTransfer MethodCategory = "erc721_safe_transfer"
	MtdSwap               MethodCategory = "swap"
	MtdLPAdd              MethodCategory = "lp_add"
	MtdLPRemove           MethodCategory = "lp_remove"
	MtdAaveSupply         MethodCategory = "aave_supply"
	MtdAaveBorrow         MethodCategory = "aave_borrow"
	MtdAaveRepay          MethodCategory = "aave_repay"
	MtdAaveWithdraw       MethodCategory = "aave_withdraw"
	MtdLidoSubmit         MethodCategory = "lido_submit"
	MtdWETHDeposit        MethodCategory = "weth_deposit"
	MtdWETHWithdraw       MethodCategory = "weth_withdraw"
	MtdMulticall          MethodCategory = "multicall"
	MtdUniversalRouter    MethodCategory = "universal_router_execute"
	MtdSafeExec           MethodCategory = "safe_exec"
	MtdPermit2            MethodCategory = "permit2"
	MtdBridge             MethodCategory = "bridge"
	MtdCompoundV2Supply   MethodCategory = "compoundv2_supply"
	MtdCompoundV2Redeem   MethodCategory = "compoundv2_redeem"
	MtdCompoundV2Borrow   MethodCategory = "compoundv2_borrow"
	MtdCompoundV3Supply   MethodCategory = "compoundv3_supply"
	MtdCompoundV3Borrow   MethodCategory = "compoundv3_borrow"
	MtdLPAddV3            MethodCategory = "lp_add_v3"
	MtdLPRemoveV3         MethodCategory = "lp_remove_v3"
	MtdLend               MethodCategory = "lend"
	MtdSeaportFulfill     MethodCategory = "seaport_fulfill"
)

// MethodDef maps a 4-byte selector to a semantic category and hint.
type MethodDef struct {
	Signature string
	Selector  [4]byte
	Category  MethodCategory
	Hint      string // human-readable hint
}

// sel computes the 4-byte ABI selector for a method signature.
func sel(sig string) [4]byte {
	var s [4]byte
	copy(s[:], crypto.Keccak256([]byte(sig))[:4])
	return s
}

// Common method selectors. We only register signatures that strongly imply a category.
// For everything else, we fall back to openchain/4byte online lookup.
var MethodCatalog = []MethodDef{
	// ERC-20 / 721
	{"transfer(address,uint256)", sel("transfer(address,uint256)"), MtdERC20Transfer, "ERC-20 transfer"},
	{"transferFrom(address,address,uint256)", sel("transferFrom(address,address,uint256)"), MtdERC20TransferFrom, "ERC-20 transferFrom (also ERC-721)"},
	{"approve(address,uint256)", sel("approve(address,uint256)"), MtdERC20Approve, "ERC-20 approve"},
	{"safeTransferFrom(address,address,uint256)", sel("safeTransferFrom(address,address,uint256)"), MtdERC721SafeTransfer, "ERC-721 safeTransferFrom"},
	{"safeTransferFrom(address,address,uint256,bytes)", sel("safeTransferFrom(address,address,uint256,bytes)"), MtdERC721SafeTransfer, "ERC-721 safeTransferFrom"},

	// Uniswap V2 router family
	{"swapExactTokensForTokens(uint256,uint256,address[],address,uint256)", sel("swapExactTokensForTokens(uint256,uint256,address[],address,uint256)"), MtdSwap, "Uniswap V2: swapExactTokensForTokens"},
	{"swapTokensForExactTokens(uint256,uint256,address[],address,uint256)", sel("swapTokensForExactTokens(uint256,uint256,address[],address,uint256)"), MtdSwap, "Uniswap V2: swapTokensForExactTokens"},
	{"swapExactETHForTokens(uint256,address[],address,uint256)", sel("swapExactETHForTokens(uint256,address[],address,uint256)"), MtdSwap, "Uniswap V2: swapExactETHForTokens"},
	{"swapExactTokensForETH(uint256,uint256,address[],address,uint256)", sel("swapExactTokensForETH(uint256,uint256,address[],address,uint256)"), MtdSwap, "Uniswap V2: swapExactTokensForETH"},
	{"swapExactTokensForTokensSupportingFeeOnTransferTokens(uint256,uint256,address[],address,uint256)", sel("swapExactTokensForTokensSupportingFeeOnTransferTokens(uint256,uint256,address[],address,uint256)"), MtdSwap, "V2 swap (FoT)"},
	{"swapExactETHForTokensSupportingFeeOnTransferTokens(uint256,address[],address,uint256)", sel("swapExactETHForTokensSupportingFeeOnTransferTokens(uint256,address[],address,uint256)"), MtdSwap, "V2 swap ETH (FoT)"},
	{"swapExactTokensForETHSupportingFeeOnTransferTokens(uint256,uint256,address[],address,uint256)", sel("swapExactTokensForETHSupportingFeeOnTransferTokens(uint256,uint256,address[],address,uint256)"), MtdSwap, "V2 swap to ETH (FoT)"},

	// Uniswap V3 SwapRouter / SwapRouter02
	{"exactInputSingle((address,address,uint24,address,uint256,uint256,uint256,uint160))", sel("exactInputSingle((address,address,uint24,address,uint256,uint256,uint256,uint160))"), MtdSwap, "Uniswap V3: exactInputSingle"},
	{"exactInput((bytes,address,uint256,uint256,uint256))", sel("exactInput((bytes,address,uint256,uint256,uint256))"), MtdSwap, "Uniswap V3: exactInput"},
	{"exactOutputSingle((address,address,uint24,address,uint256,uint256,uint256,uint160))", sel("exactOutputSingle((address,address,uint24,address,uint256,uint256,uint256,uint160))"), MtdSwap, "Uniswap V3: exactOutputSingle"},

	// Uniswap Universal Router
	{"execute(bytes,bytes[],uint256)", sel("execute(bytes,bytes[],uint256)"), MtdUniversalRouter, "Uniswap Universal Router"},
	{"execute(bytes,bytes[])", sel("execute(bytes,bytes[])"), MtdUniversalRouter, "Uniswap Universal Router"},

	// Solidly / Aerodrome
	{"swapExactTokensForTokens(uint256,uint256,(address,address,bool,address)[],address,uint256)", sel("swapExactTokensForTokens(uint256,uint256,(address,address,bool,address)[],address,uint256)"), MtdSwap, "Solidly/Aerodrome swap"},
	{"swapExactETHForTokens(uint256,(address,address,bool,address)[],address,uint256)", sel("swapExactETHForTokens(uint256,(address,address,bool,address)[],address,uint256)"), MtdSwap, "Solidly/Aerodrome swap ETH"},
	{"swapExactTokensForETH(uint256,uint256,(address,address,bool,address)[],address,uint256)", sel("swapExactTokensForETH(uint256,uint256,(address,address,bool,address)[],address,uint256)"), MtdSwap, "Solidly/Aerodrome swap to ETH"},
	{"swapExactETHForTokensSupportingFeeOnTransferTokens(uint256,(address,address,bool,address)[],address,uint256)", sel("swapExactETHForTokensSupportingFeeOnTransferTokens(uint256,(address,address,bool,address)[],address,uint256)"), MtdSwap, "Solidly swap ETH (FoT)"},
	{"swapExactTokensForETHSupportingFeeOnTransferTokens(uint256,uint256,(address,address,bool,address)[],address,uint256)", sel("swapExactTokensForETHSupportingFeeOnTransferTokens(uint256,uint256,(address,address,bool,address)[],address,uint256)"), MtdSwap, "Solidly swap to ETH (FoT)"},

	// Uniswap V2 LP add/remove
	{"addLiquidity(address,address,uint256,uint256,uint256,uint256,address,uint256)", sel("addLiquidity(address,address,uint256,uint256,uint256,uint256,address,uint256)"), MtdLPAdd, "V2 addLiquidity"},
	{"removeLiquidity(address,address,uint256,uint256,uint256,address,uint256)", sel("removeLiquidity(address,address,uint256,uint256,uint256,address,uint256)"), MtdLPRemove, "V2 removeLiquidity"},

	// Aave V3 Pool
	{"supply(address,uint256,address,uint16)", sel("supply(address,uint256,address,uint16)"), MtdAaveSupply, "Aave V3 supply"},
	{"withdraw(address,uint256,address)", sel("withdraw(address,uint256,address)"), MtdAaveWithdraw, "Aave V3 withdraw"},
	{"borrow(address,uint256,uint256,uint16,address)", sel("borrow(address,uint256,uint256,uint16,address)"), MtdAaveBorrow, "Aave V3 borrow"},
	{"repay(address,uint256,uint256,address)", sel("repay(address,uint256,uint256,address)"), MtdAaveRepay, "Aave V3 repay"},

	// Lido
	{"submit(address)", sel("submit(address)"), MtdLidoSubmit, "Lido submit (stake ETH)"},

	// WETH
	{"deposit()", sel("deposit()"), MtdWETHDeposit, "WETH deposit"},
	{"withdraw(uint256)", sel("withdraw(uint256)"), MtdWETHWithdraw, "WETH withdraw"},

	// Multicall
	{"multicall(bytes[])", sel("multicall(bytes[])"), MtdMulticall, "Multicall"},
	{"multicall(uint256,bytes[])", sel("multicall(uint256,bytes[])"), MtdMulticall, "Multicall (with deadline)"},
	{"aggregate((address,bytes)[])", sel("aggregate((address,bytes)[])"), MtdMulticall, "Multicall.aggregate"},

	// Safe (Gnosis) execTransaction
	{"execTransaction(address,uint256,bytes,uint8,uint256,uint256,uint256,address,address,bytes)", sel("execTransaction(address,uint256,bytes,uint8,uint256,uint256,uint256,address,address,bytes)"), MtdSafeExec, "Safe execTransaction"},

	// Permit2
	{"approve(address,address,uint160,uint48)", sel("approve(address,address,uint160,uint48)"), MtdPermit2, "Permit2 approve"},
	{"transferFrom(address,address,uint160,address)", sel("transferFrom(address,address,uint160,address)"), MtdPermit2, "Permit2 transferFrom"},

	// Compound V2
	{"mint(uint256)", sel("mint(uint256)"), MtdCompoundV2Supply, "Compound V2: supply (mint cToken)"},
	{"redeem(uint256)", sel("redeem(uint256)"), MtdCompoundV2Redeem, "Compound V2: redeem"},
	{"borrow(uint256)", sel("borrow(uint256)"), MtdCompoundV2Borrow, "Compound V2: borrow"},

	// Compound V3
	{"supply(address,uint256)", sel("supply(address,uint256)"), MtdCompoundV3Supply, "Compound V3: supply"},
	{"borrow(uint256)", sel("borrow(uint256)"), MtdCompoundV3Borrow, "Compound V3: borrow"},

	// Uniswap V4 PoolManager
	{"unlock(bytes)", sel("unlock(bytes)"), MtdSwap, "Uniswap V4: unlock"},
	{"swap((address,address,int24,int24,int256,uint256,bytes),bytes)", sel("swap((address,address,int24,int24,int256,uint256,bytes),bytes)"), MtdSwap, "Uniswap V4: swap"},
	{"modifyLiquidity((address,address,int24,int24,int256,bytes32),bytes)", sel("modifyLiquidity((address,address,int24,int24,int256,bytes32),bytes)"), MtdLPAddV3, "Uniswap V4: modifyLiquidity"},

	// Uniswap V3 LP
	{"mint(address,address,uint24,int24,int24,uint256,uint256,uint256,uint256,address,uint256)", sel("mint(address,address,uint24,int24,int24,uint256,uint256,uint256,uint256,address,uint256)"), MtdLPAddV3, "Uniswap V3: mint LP"},
	{"burn(int24,int24,int128,uint256)", sel("burn(int24,int24,int128,uint256)"), MtdLPRemoveV3, "Uniswap V3: burn LP"},

	// Stargate bridge
	{"swap(uint8,uint8,uint256,uint256,uint256,address,bytes,uint256)", sel("swap(uint8,uint8,uint256,uint256,uint256,address,bytes,uint256)"), MtdBridge, "Stargate: swap"},

	// Aave V2
	{"deposit(address,uint256,address,uint16)", sel("deposit(address,uint256,address,uint16)"), MtdAaveSupply, "Aave V2: deposit"},

	// Curve stableswap
	{"exchange(int128,int128,uint256,uint256)", sel("exchange(int128,int128,uint256,uint256)"), MtdSwap, "Curve: exchange"},
	{"exchange_underlying(int128,int128,uint256,uint256)", sel("exchange_underlying(int128,int128,uint256,uint256)"), MtdSwap, "Curve: exchange_underlying"},
	{"get_dy(int128,int128,uint256)", sel("get_dy(int128,int128,uint256)"), MtdSwap, "Curve: get_dy"},

	// Curve crypto pool
	{"exchange(uint256,uint256,uint256,uint256)", sel("exchange(uint256,uint256,uint256,uint256)"), MtdSwap, "Curve crypto: exchange"},
	{"get_dy(uint256,uint256,uint256)", sel("get_dy(uint256,uint256,uint256)"), MtdSwap, "Curve crypto: get_dy"},

	// Curve NG
	{"exchange(uint256,uint256,uint256,uint256,uint256)", sel("exchange(uint256,uint256,uint256,uint256,uint256)"), MtdSwap, "Curve NG: exchange"},

	// Curve LP
	{"add_liquidity(uint256[],uint256)", sel("add_liquidity(uint256[],uint256)"), MtdLPAdd, "Curve: add_liquidity"},
	{"remove_liquidity(uint256,uint256[])", sel("remove_liquidity(uint256,uint256[])"), MtdLPRemove, "Curve: remove_liquidity"},

	// Balancer V2 Vault
	{"swap((bytes32,address,address,uint256,bytes),(address,bool,address,uint256))", sel("swap((bytes32,address,address,uint256,bytes),(address,bool,address,uint256))"), MtdSwap, "Balancer V2: swap"},
	{"batchSwap(uint8,(bytes32,uint256,uint256,uint256,bytes)[],(address,bool,address,uint256))", sel("batchSwap(uint8,(bytes32,uint256,uint256,uint256,bytes)[],(address,bool,address,uint256))"), MtdSwap, "Balancer V2: batchSwap"},

	// 1inch v5 AggregationRouter
	{"swap(address,uint256,uint256,uint256[],bytes)", sel("swap(address,uint256,uint256,uint256[],bytes)"), MtdSwap, "1inch v5: swap"},
	{"unoswap(address,uint256,uint256,uint256[])", sel("unoswap(address,uint256,uint256,uint256[])"), MtdSwap, "1inch v5: unoswap"},
	{"uniswapV3Swap(uint256,uint256,uint256[])", sel("uniswapV3Swap(uint256,uint256,uint256[])"), MtdSwap, "1inch v5: uniswapV3Swap"},
	{"clipperSwap(address,address,uint256,uint256,bytes)", sel("clipperSwap(address,address,uint256,uint256,bytes)"), MtdSwap, "1inch v5: clipperSwap"},

	// 0x / Matcha
	{"fillRfqOrder(((address,address,address,address,uint128,uint64),((address,address,address,address,uint128,uint128,bytes32),bytes32)),(uint128,uint128,bytes32))", sel("fillRfqOrder(((address,address,address,address,uint128,uint64),((address,address,address,address,uint128,uint128,bytes32),bytes32)),(uint128,uint128,bytes32))"), MtdSwap, "0x: fillRfqOrder"},
	{"fillOtcOrder((address,address,address,address,uint256,uint256),uint256,(uint256,uint256),bytes32)", sel("fillOtcOrder((address,address,address,address,uint256,uint256),uint256,(uint256,uint256),bytes32)"), MtdSwap, "0x: fillOtcOrder"},

	// PancakeSwap V3 (different tuple from Uni V3)
	{"exactInputSingle((address,address,uint256,address,uint256,uint256,uint256))", sel("exactInputSingle((address,address,uint256,address,uint256,uint256,uint256))"), MtdSwap, "PancakeSwap V3: exactInputSingle"},

	// SushiSwap RouteProcessor
	{"processRoute(address,uint256,uint256,address,uint256,bytes)", sel("processRoute(address,uint256,uint256,address,uint256,bytes)"), MtdSwap, "SushiSwap: processRoute"},

	// CowSwap
	{"settle(address,address,uint256,uint256,bytes)", sel("settle(address,address,uint256,uint256,bytes)"), MtdSwap, "CowSwap: settle"},
	{"invalidateOrder(bytes)", sel("invalidateOrder(bytes)"), MtdSwap, "CowSwap: invalidateOrder"},

	// DODO
	{"sellBase(address,uint256)", sel("sellBase(address,uint256)"), MtdSwap, "DODO: sellBase"},
	{"sellQuote(address,uint256)", sel("sellQuote(address,uint256)"), MtdSwap, "DODO: sellQuote"},

	// Trader Joe LB
	{"swap((uint256,uint256,uint256,bytes),bool)", sel("swap((uint256,uint256,uint256,bytes),bool)"), MtdSwap, "Trader Joe LB: swap"},

	// GMX
	{"increasePosition(address,address,uint256,uint256)", sel("increasePosition(address,address,uint256,uint256)"), MtdSwap, "GMX: increasePosition"},
	{"decreasePosition(address,address,uint256,uint256,uint256)", sel("decreasePosition(address,address,uint256,uint256,uint256)"), MtdSwap, "GMX: decreasePosition"},

	// ParaSwap
	{"swapOnUniswap(address,uint256,uint256)", sel("swapOnUniswap(address,uint256,uint256)"), MtdSwap, "ParaSwap: swapOnUniswap"},
	{"buyOnUniswap(uint256,uint256,address)", sel("buyOnUniswap(uint256,uint256,address)"), MtdSwap, "ParaSwap: buyOnUniswap"},

	// Morpho Blue (supply omitted: same selector as Compound V3)
	{"borrow(address,uint256)", sel("borrow(address,uint256)"), MtdAaveBorrow, "Morpho Blue: borrow"},
	{"repay(address,uint256)", sel("repay(address,uint256)"), MtdAaveRepay, "Morpho Blue: repay"},
	{"withdraw(address,uint256)", sel("withdraw(address,uint256)"), MtdAaveWithdraw, "Morpho Blue: withdraw"},

	// MakerDAO
	{"frob(bytes32,int256,int256)", sel("frob(bytes32,int256,int256)"), MtdLend, "MakerDAO: frob"},
	{"heal(uint256)", sel("heal(uint256)"), MtdLend, "MakerDAO: heal"},
	{"wipe(bytes32,uint256)", sel("wipe(bytes32,uint256)"), MtdLend, "MakerDAO: wipe"},

	// Rocket Pool (deposit omitted: same selector as WETH)
	{"burn(uint256)", sel("burn(uint256)"), MtdLidoSubmit, "Rocket Pool: burn"},

	// EigenLayer
	{"depositIntoStrategy(address,address,uint256)", sel("depositIntoStrategy(address,address,uint256)"), MtdLidoSubmit, "EigenLayer: depositIntoStrategy"},
	{"withdrawFromStrategy(address,address,uint256)", sel("withdrawFromStrategy(address,address,uint256)"), MtdLidoSubmit, "EigenLayer: withdrawFromStrategy"},

	// Coinbase cbETH (deposit omitted: same selector as WETH)
	{"requestWithdrawal(uint256)", sel("requestWithdrawal(uint256)"), MtdLidoSubmit, "Coinbase cbETH: requestWithdrawal"},

	// Frax sfrxETH
	{"deposit(uint256,uint256)", sel("deposit(uint256,uint256)"), MtdLidoSubmit, "Frax: deposit"},
	{"withdraw(uint256,uint256)", sel("withdraw(uint256,uint256)"), MtdLidoSubmit, "Frax: withdraw"},

	// Stakewise
	{"stake(address,uint256)", sel("stake(address,uint256)"), MtdLidoSubmit, "Stakewise: stake"},

	// Synapse bridge
	{"swap(uint8,uint8,uint256,uint256,uint256)", sel("swap(uint8,uint8,uint256,uint256,uint256)"), MtdBridge, "Synapse: swap"},

	// Across
	{"deposit(uint32,address,uint256,address,uint256,uint256,int64)", sel("deposit(uint32,address,uint256,address,uint256,uint256,int64)"), MtdBridge, "Across: deposit"},
	{"fillRelay(address,address,address,uint256,uint256,uint256,uint256,uint256,uint256,bytes32)", sel("fillRelay(address,address,address,uint256,uint256,uint256,uint256,uint256,uint256,bytes32)"), MtdBridge, "Across: fillRelay"},

	// Hop
	{"send(uint256,uint256,uint32,address,uint256,uint256,uint256)", sel("send(uint256,uint256,uint32,address,uint256,uint256,uint256)"), MtdBridge, "Hop: send"},

	// CCTP
	{"depositForBurn(uint256,uint32,bytes32,address)", sel("depositForBurn(uint256,uint32,bytes32,address)"), MtdBridge, "CCTP: depositForBurn"},
	{"receiveMessage(bytes,(bytes32,uint32,address,bytes32))", sel("receiveMessage(bytes,(bytes32,uint32,address,bytes32))"), MtdBridge, "CCTP: receiveMessage"},

	// Seaport
	{"fulfillBasicOrder((address,uint256,uint256,address,address,address,uint256,uint256,uint256,bytes32,uint256,bytes32,uint256))", sel("fulfillBasicOrder((address,uint256,uint256,address,address,address,uint256,uint256,uint256,bytes32,uint256,bytes32,uint256))"), MtdSeaportFulfill, "Seaport: fulfillBasicOrder"},
	{"fulfillOrder(((address,address,address,uint256,uint256,uint256,uint256,uint256,uint256,uint8,bytes32,uint256,bytes32,uint256),(address,address,uint256,uint256,uint256,uint256)),bytes32)", sel("fulfillOrder(((address,address,address,uint256,uint256,uint256,uint256,uint256,uint256,uint8,bytes32,uint256,bytes32,uint256),(address,address,uint256,uint256,uint256,uint256)),bytes32)"), MtdSeaportFulfill, "Seaport: fulfillOrder"},
	{"fulfillAvailableOrders(((address,address,address,uint256,uint256,uint256,uint256,uint256,uint256,uint8,bytes32,uint256,bytes32,uint256)[],(address,address,uint256,uint256,uint256,uint256)[],bytes32))", sel("fulfillAvailableOrders(((address,address,address,uint256,uint256,uint256,uint256,uint256,uint256,uint8,bytes32,uint256,bytes32,uint256)[],(address,address,uint256,uint256,uint256,uint256)[],bytes32))"), MtdSeaportFulfill, "Seaport: fulfillAvailableOrders"},
}

var (
	methodBySelector     map[[4]byte]*MethodDef
	methodBySelectorOnce sync.Once
)

func LookupMethod(s [4]byte) *MethodDef {
	methodBySelectorOnce.Do(func() {
		methodBySelector = make(map[[4]byte]*MethodDef, len(MethodCatalog))
		for i := range MethodCatalog {
			methodBySelector[MethodCatalog[i].Selector] = &MethodCatalog[i]
		}
	})
	if d, ok := methodBySelector[s]; ok {
		return d
	}
	return nil
}

// SelectorHex returns "0xXXXXXXXX".
func SelectorHex(s [4]byte) string {
	return "0x" + hex.EncodeToString(s[:])
}
