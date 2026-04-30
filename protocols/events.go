// Package protocols holds a curated fingerprint database of well-known DeFi/NFT protocols.
//
// We register event topics and method selectors with friendly names and category hints,
// so the Classifier can match logs/calldata against this catalog without needing the
// contract ABI for every counterparty.
package protocols

import (
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// EventCategory is what we infer when we see a given event topic in receipt.logs.
type EventCategory string

const (
	EvtTransferERC20    EventCategory = "transfer_erc20"
	EvtTransferERC721   EventCategory = "transfer_erc721" // same topic, distinguished by indexed args count
	EvtTransferSingle   EventCategory = "transfer_single_erc1155"
	EvtTransferBatch    EventCategory = "transfer_batch_erc1155"
	EvtApprovalERC20    EventCategory = "approval_erc20"
	EvtApprovalForAll   EventCategory = "approval_for_all"
	EvtSwapUniV2        EventCategory = "swap_univ2"
	EvtSwapUniV3        EventCategory = "swap_univ3"
	EvtSwapUniV4        EventCategory = "swap_univ4"
	EvtSwapAerodrome    EventCategory = "swap_aerodrome"
	EvtSwapBalancer     EventCategory = "swap_balancer_v2"
	EvtSwapCurve        EventCategory = "swap_curve"
	EvtSwapCurveUnder   EventCategory = "swap_curve_underlying"
	EvtSwap1inchAggReg  EventCategory = "swap_1inch_aggregation"
	EvtSwap0xRfq        EventCategory = "swap_0x_rfq"
	EvtSwap0xLimit      EventCategory = "swap_0x_limit"
	EvtSwap0xOtc        EventCategory = "swap_0x_otc"
	EvtSwapKyber        EventCategory = "swap_kyber"
	EvtSwapDodo         EventCategory = "swap_dodo"
	EvtSwapPancakeV3    EventCategory = "swap_pancakeswap_v3"
	EvtMintLP           EventCategory = "mint_lp"
	EvtBurnLP           EventCategory = "burn_lp"
	EvtWETHDeposit      EventCategory = "weth_deposit"
	EvtWETHWithdraw     EventCategory = "weth_withdraw"
	EvtAaveSupply       EventCategory = "aave_supply"
	EvtAaveWithdraw     EventCategory = "aave_withdraw"
	EvtAaveBorrow       EventCategory = "aave_borrow"
	EvtAaveRepay        EventCategory = "aave_repay"
	EvtAaveLiquidate    EventCategory = "aave_liquidate"
	EvtCompoundV3Supply EventCategory = "compoundv3_supply"
	EvtMorphoSupply     EventCategory = "morpho_supply"
	EvtLidoSubmit       EventCategory = "lido_submit"
	EvtRocketDeposit    EventCategory = "rocket_pool_deposit"
	EvtEigenDeposit     EventCategory = "eigenlayer_deposit"
	EvtBridgeMessage    EventCategory = "bridge_message"
	EvtSeaportOrder     EventCategory = "nft_seaport_order"
	EvtBlurExecution    EventCategory = "nft_blur_execution"

	// Aave V2
	EvtAaveV2Deposit  EventCategory = "aave_v2_deposit"
	EvtAaveV2Withdraw EventCategory = "aave_v2_withdraw"
	EvtAaveV2Borrow   EventCategory = "aave_v2_borrow"
	EvtAaveV2Repay    EventCategory = "aave_v2_repay"

	// Compound V2
	EvtCompoundV2Redeem    EventCategory = "compoundv2_redeem"
	EvtCompoundV2Borrow    EventCategory = "compoundv2_borrow"
	EvtCompoundV2Repay     EventCategory = "compoundv2_repay"
	EvtCompoundV2Liquidate EventCategory = "compoundv2_liquidate"

	// Compound V3 (extended)
	EvtCompoundV3Borrow EventCategory = "compoundv3_borrow"
	EvtCompoundV3Repay  EventCategory = "compoundv3_repay"

	// Uniswap V3 LP (different from V2)
	EvtMintLPV3 EventCategory = "mint_lp_v3"
	EvtBurnLPV3 EventCategory = "burn_lp_v3"

	// Curve crypto pool
	EvtCurveCryptoSwap EventCategory = "swap_curve_crypto"

	// Bridges
	EvtStargateSwap        EventCategory = "bridge_stargate_swap"
	EvtAcrossDeposit       EventCategory = "bridge_across_deposit"
	EvtAcrossFill          EventCategory = "bridge_across_fill"
	EvtHopTransferSent     EventCategory = "bridge_hop_sent"
	EvtHopTransferReceived EventCategory = "bridge_hop_received"
	EvtCCTPMessageSent     EventCategory = "bridge_cctp_sent"
	EvtCCTPMessageReceived EventCategory = "bridge_cctp_received"

	// Rocket Pool extended
	EvtRocketWithdraw EventCategory = "rocket_pool_withdraw"

	// CowSwap
	EvtCowTrade EventCategory = "swap_cow_trade"

	// Balancer V2 LP
	EvtBalancerPoolJoined EventCategory = "balancer_joined"

	// Curve NG
	EvtCurveNGSwap EventCategory = "swap_curve_ng"

	// Curve LP
	EvtCurveAddLiquidity    EventCategory = "curve_add_liquidity"
	EvtCurveRemoveLiquidity EventCategory = "curve_remove_liquidity"

	// Trader Joe LB
	EvtSwapTraderJoeLB EventCategory = "swap_traderjoe_lb"

	// GMX perps
	EvtGMXIncreasePosition EventCategory = "gmx_increase_position"
	EvtGMXDecreasePosition EventCategory = "gmx_decrease_position"
	EvtGMXClosePosition    EventCategory = "gmx_close_position"

	// ParaSwap
	EvtSwapParaSwap EventCategory = "swap_paraswap"

	// Morpho Blue — Borrow, Withdraw, Liquidate (Repay omitted: same hash as Compound V3)
	EvtMorphoBorrow    EventCategory = "morpho_borrow"
	EvtMorphoWithdraw  EventCategory = "morpho_withdraw"
	EvtMorphoLiquidate EventCategory = "morpho_liquidate"

	// EigenLayer extended
	EvtEigenWithdrawQueued    EventCategory = "eigenlayer_withdraw_queued"
	EvtEigenWithdrawCompleted EventCategory = "eigenlayer_withdraw_completed"

	// Frax sfrxETH
	EvtFraxSubmit   EventCategory = "frax_submit"
	EvtFraxWithdraw EventCategory = "frax_withdraw"

	// Stakewise
	EvtStakewiseStaked   EventCategory = "stakewise_staked"
	EvtStakewiseUnstaked EventCategory = "stakewise_unstaked"

	// Bridges
	EvtWormholeMessage     EventCategory = "bridge_wormhole_message"
	EvtSynapseSwap         EventCategory = "bridge_synapse_swap"
	EvtCelerRelaySent      EventCategory = "bridge_celer_relay_sent"
	EvtCelerRelayConfirmed EventCategory = "bridge_celer_relay_confirmed"

	// Blur batch
	EvtBlurExecutionBatch EventCategory = "nft_blur_execution_batch"
)

// EventDef maps an event topic hash to a semantic category and protocol.
type EventDef struct {
	Signature string
	Topic     common.Hash
	Category  EventCategory
	Protocol  string
}

// keccak256 helper
// k256 computes the keccak256 hash of an event signature.
func k256(s string) common.Hash {
	return crypto.Keccak256Hash([]byte(s))
}

// Catalog of event topics → semantic class.
// Fork-shared topics: protocol attribution is resolved via known-address lookup.
var EventCatalog = []EventDef{
	// ERC-20 / 721 / 1155
	{"Transfer(address,address,uint256)", k256("Transfer(address,address,uint256)"), EvtTransferERC20, "ERC20/721"},
	{"TransferSingle(address,address,address,uint256,uint256)", k256("TransferSingle(address,address,address,uint256,uint256)"), EvtTransferSingle, "ERC1155"},
	{"TransferBatch(address,address,address,uint256[],uint256[])", k256("TransferBatch(address,address,address,uint256[],uint256[])"), EvtTransferBatch, "ERC1155"},
	{"Approval(address,address,uint256)", k256("Approval(address,address,uint256)"), EvtApprovalERC20, "ERC20/721"},
	{"ApprovalForAll(address,address,bool)", k256("ApprovalForAll(address,address,bool)"), EvtApprovalForAll, "ERC721/1155"},

	// DEX swaps
	{"Swap(address,uint256,uint256,uint256,uint256,address)", k256("Swap(address,uint256,uint256,uint256,uint256,address)"), EvtSwapUniV2, "Uniswap V2 (family)"},
	{"Swap(address,address,int256,int256,uint160,uint128,int24)", k256("Swap(address,address,int256,int256,uint160,uint128,int24)"), EvtSwapUniV3, "Uniswap V3 (family)"},
	{"Swap(bytes32,address,int128,int128,uint160,uint128,int24,uint24)", k256("Swap(bytes32,address,int128,int128,uint160,uint128,int24,uint24)"), EvtSwapUniV4, "Uniswap V4"},
	// Aerodrome / Velodrome (Solidly fork) Swap event:
	{"Swap(address,address,uint256,uint256,uint256,uint256)", k256("Swap(address,address,uint256,uint256,uint256,uint256)"), EvtSwapAerodrome, "Aerodrome/Velodrome (Solidly)"},
	// Balancer V2 Vault.Swap
	{"Swap(bytes32,address,address,uint256,uint256)", k256("Swap(bytes32,address,address,uint256,uint256)"), EvtSwapBalancer, "Balancer V2"},
	// Curve TokenExchange
	{"TokenExchange(address,int128,uint256,int128,uint256)", k256("TokenExchange(address,int128,uint256,int128,uint256)"), EvtSwapCurve, "Curve"},
	{"TokenExchangeUnderlying(address,int128,uint256,int128,uint256)", k256("TokenExchangeUnderlying(address,int128,uint256,int128,uint256)"), EvtSwapCurveUnder, "Curve"},
	// PancakeSwap V3 — distinct from Uni V3 (extra protocolFees fields).
	{"Swap(address,address,int256,int256,uint160,uint128,int24,uint128,uint128)", k256("Swap(address,address,int256,int256,uint160,uint128,int24,uint128,uint128)"), EvtSwapPancakeV3, "PancakeSwap V3"},
	// (Uni V3 family entry above also handled.) — distinguished by chain (BSC) + factory addr.
	// 1inch AggregationRouter v5 OrderFilledRFQ
	{"OrderFilledRFQ(bytes32,uint256)", k256("OrderFilledRFQ(bytes32,uint256)"), EvtSwap1inchAggReg, "1inch Aggregation"},
	// 0x: RfqOrderFilled / LimitOrderFilled / OtcOrderFilled
	{"RfqOrderFilled(bytes32,address,address,address,address,uint128,uint128,bytes32)", k256("RfqOrderFilled(bytes32,address,address,address,address,uint128,uint128,bytes32)"), EvtSwap0xRfq, "0x"},
	{"LimitOrderFilled(bytes32,address,address,address,address,address,uint128,uint128,uint128,uint256,bytes32)", k256("LimitOrderFilled(bytes32,address,address,address,address,address,uint128,uint128,uint128,uint256,bytes32)"), EvtSwap0xLimit, "0x"},
	{"OtcOrderFilled(bytes32,address,address,address,address,uint128,uint128)", k256("OtcOrderFilled(bytes32,address,address,address,address,uint128,uint128)"), EvtSwap0xOtc, "0x"},
	// Kyber Network (DMM) Swap
	{"Swap(address,address,address,uint256,uint256,address,uint256)", k256("Swap(address,address,address,uint256,uint256,address,uint256)"), EvtSwapKyber, "Kyber"},
	// DODO trade
	{"DODOSwap(address,address,uint256,uint256,address,address)", k256("DODOSwap(address,address,uint256,uint256,address,address)"), EvtSwapDodo, "DODO"},

	// LP mint/burn (Uniswap V2/V3)
	{"Mint(address,uint256,uint256)", k256("Mint(address,uint256,uint256)"), EvtMintLP, "Uniswap V2 LP"},
	{"Burn(address,uint256,uint256,address)", k256("Burn(address,uint256,uint256,address)"), EvtBurnLP, "Uniswap V2 LP"},

	// Wrapped native
	{"Deposit(address,uint256)", k256("Deposit(address,uint256)"), EvtWETHDeposit, "WETH/WMATIC/etc."},
	{"Withdrawal(address,uint256)", k256("Withdrawal(address,uint256)"), EvtWETHWithdraw, "WETH/WMATIC/etc."},

	// Aave V3
	{"Supply(address,address,address,uint256,uint16)", k256("Supply(address,address,address,uint256,uint16)"), EvtAaveSupply, "Aave V3"},
	{"Withdraw(address,address,address,uint256)", k256("Withdraw(address,address,address,uint256)"), EvtAaveWithdraw, "Aave V3"},
	{"Borrow(address,address,address,uint256,uint8,uint256,uint16)", k256("Borrow(address,address,address,uint256,uint8,uint256,uint16)"), EvtAaveBorrow, "Aave V3"},
	{"Repay(address,address,address,uint256,bool)", k256("Repay(address,address,address,uint256,bool)"), EvtAaveRepay, "Aave V3"},
	{"LiquidationCall(address,address,address,uint256,uint256,address,bool)", k256("LiquidationCall(address,address,address,uint256,uint256,address,bool)"), EvtAaveLiquidate, "Aave V3"},
	// Compound III Supply
	{"Supply(address,address,uint256)", k256("Supply(address,address,uint256)"), EvtCompoundV3Supply, "Compound V3"},

	// Liquid Staking
	{"Submitted(address,uint256,address)", k256("Submitted(address,uint256,address)"), EvtLidoSubmit, "Lido"},
	{"DepositReceived(address,uint256)", k256("DepositReceived(address,uint256)"), EvtRocketDeposit, "Rocket Pool"},

	// EigenLayer StrategyManager.Deposit
	{"Deposit(address,address,address,uint256)", k256("Deposit(address,address,address,uint256)"), EvtEigenDeposit, "EigenLayer"},

	// Seaport OrderFulfilled / Blur Execution
	{"OrderFulfilled(bytes32,address,address,address,(uint8,address,uint256,uint256)[],(uint8,address,uint256,uint256,address)[])", k256("OrderFulfilled(bytes32,address,address,address,(uint8,address,uint256,uint256)[],(uint8,address,uint256,uint256,address)[])"), EvtSeaportOrder, "OpenSea Seaport"},
	{"Execution721Packed(bytes32,uint256,uint256)", k256("Execution721Packed(bytes32,uint256,uint256)"), EvtBlurExecution, "Blur"},

	// Aave V2 events (unique signatures not shared with V3)
	{"Deposit(address,address,uint256,uint16)", k256("Deposit(address,address,uint256,uint16)"), EvtAaveV2Deposit, "Aave V2"},
	{"Withdraw(address,address,uint256,address)", k256("Withdraw(address,address,uint256,address)"), EvtAaveV2Withdraw, "Aave V2"},
	{"Borrow(address,address,uint256,uint256,uint16,address)", k256("Borrow(address,address,uint256,uint256,uint16,address)"), EvtAaveV2Borrow, "Aave V2"},
	{"Repay(address,address,uint256,address,bool)", k256("Repay(address,address,uint256,address,bool)"), EvtAaveV2Repay, "Aave V2"},
	// Shared with Aave V3. Address tags disambiguate.

	// Compound V2 events (unique signatures)
	// Shared with Uniswap V2 LP. Address tags disambiguate.

	{"Redeem(address,uint256,uint256)", k256("Redeem(address,uint256,uint256)"), EvtCompoundV2Redeem, "Compound V2"},
	{"Borrow(address,uint256,uint256,uint256)", k256("Borrow(address,uint256,uint256,uint256)"), EvtCompoundV2Borrow, "Compound V2"},
	{"RepayBorrow(address,address,uint256,uint256,uint256)", k256("RepayBorrow(address,address,uint256,uint256,uint256)"), EvtCompoundV2Repay, "Compound V2"},
	{"LiquidateBorrow(address,address,uint256,address,uint256)", k256("LiquidateBorrow(address,address,uint256,address,uint256)"), EvtCompoundV2Liquidate, "Compound V2"},

	// Compound V3 extended (unique signatures)
	// Shared with Aave V3. Address tags disambiguate.

	{"Borrow(address,uint256)", k256("Borrow(address,uint256)"), EvtCompoundV3Borrow, "Compound V3"},
	{"Repay(address,address,uint256)", k256("Repay(address,address,uint256)"), EvtCompoundV3Repay, "Compound V3"},

	// Uniswap V3 LP (different signatures from V2)
	{"Mint(address,address,int24,int24,int128,uint256,uint256,uint256,uint256,address,uint256)", k256("Mint(address,address,int24,int24,int128,uint256,uint256,uint256,uint256,address,uint256)"), EvtMintLPV3, "Uniswap V3 LP"},
	{"Burn(address,address,int24,int24,int128,uint256,uint256,uint256,uint256,address,uint256)", k256("Burn(address,address,int24,int24,int128,uint256,uint256,uint256,uint256,address,uint256)"), EvtBurnLPV3, "Uniswap V3 LP"},

	// Curve crypto pool
	{"TokenExchange(address,uint256,uint256,uint256)", k256("TokenExchange(address,uint256,uint256,uint256)"), EvtCurveCryptoSwap, "Curve Crypto"},

	// Stargate bridge
	{"Swap(address,uint256,uint256,uint256,uint256,uint256,uint256)", k256("Swap(address,uint256,uint256,uint256,uint256,uint256,uint256)"), EvtStargateSwap, "Stargate"},

	// Across bridge
	{"FundsDeposited(uint256,uint256,uint256,address,uint256,address,uint256,uint256)", k256("FundsDeposited(uint256,uint256,uint256,address,uint256,address,uint256,uint256)"), EvtAcrossDeposit, "Across"},
	{"FilledRelay(address,address,address,address,uint256,uint256,uint256,uint256,uint256,uint256,bytes32)", k256("FilledRelay(address,address,address,address,uint256,uint256,uint256,uint256,uint256,uint256,bytes32)"), EvtAcrossFill, "Across"},

	// Hop bridge
	{"TransferSent(uint256,uint256,address,uint256)", k256("TransferSent(uint256,uint256,address,uint256)"), EvtHopTransferSent, "Hop"},
	{"TransferReceived(uint256,uint256,address,uint256)", k256("TransferReceived(uint256,uint256,address,uint256)"), EvtHopTransferReceived, "Hop"},

	// CCTP (Circle)
	{"MessageSent(bytes32,uint64)", k256("MessageSent(bytes32,uint64)"), EvtCCTPMessageSent, "CCTP"},
	{"MessageReceived(bytes32,uint64)", k256("MessageReceived(bytes32,uint64)"), EvtCCTPMessageReceived, "CCTP"},

	// Rocket Pool extended
	{"WithdrawalProcessed(address,uint256,uint256)", k256("WithdrawalProcessed(address,uint256,uint256)"), EvtRocketWithdraw, "Rocket Pool"},

	// CowSwap (GPv2)
	{"Trade(address,address,address,uint256,uint256,uint256,bytes)", k256("Trade(address,address,address,uint256,uint256,uint256,bytes)"), EvtCowTrade, "CowSwap"},

	// Balancer V2 LP
	{"PoolBalanceChanged(bytes32,address,uint256[],uint256[])", k256("PoolBalanceChanged(bytes32,address,uint256[],uint256[])"), EvtBalancerPoolJoined, "Balancer V2"},

	// Curve NG TokenExchange
	{"TokenExchange(uint256,uint256,uint256)", k256("TokenExchange(uint256,uint256,uint256)"), EvtCurveNGSwap, "Curve NG"},

	// Curve LP add/remove
	{"AddLiquidity(address,uint256[],uint256[],uint256)", k256("AddLiquidity(address,uint256[],uint256[],uint256)"), EvtCurveAddLiquidity, "Curve"},
	{"RemoveLiquidity(address,uint256[],uint256[],uint256)", k256("RemoveLiquidity(address,uint256[],uint256[],uint256)"), EvtCurveRemoveLiquidity, "Curve"},

	// Trader Joe LB pair swap
	{"Swap(address,uint256,uint256,uint256)", k256("Swap(address,uint256,uint256,uint256)"), EvtSwapTraderJoeLB, "Trader Joe LB"},

	// GMX perp positions
	{"IncreasePosition(address,address,address,uint256,uint256,bool,uint256,uint256)", k256("IncreasePosition(address,address,address,uint256,uint256,bool,uint256,uint256)"), EvtGMXIncreasePosition, "GMX"},
	{"DecreasePosition(address,address,address,uint256,uint256,bool,uint256,uint256)", k256("DecreasePosition(address,address,address,uint256,uint256,bool,uint256,uint256)"), EvtGMXDecreasePosition, "GMX"},
	{"ClosePosition(address,uint256,uint256,uint256,uint256)", k256("ClosePosition(address,uint256,uint256,uint256,uint256)"), EvtGMXClosePosition, "GMX"},

	// ParaSwap
	{"Swapped(address,address,uint256,uint256)", k256("Swapped(address,address,uint256,uint256)"), EvtSwapParaSwap, "ParaSwap"},

	// Morpho Blue — Borrow, Withdraw, Liquidate (Repay omitted: same hash as Compound V3)
	{"Borrow(address,address,uint256)", k256("Borrow(address,address,uint256)"), EvtMorphoBorrow, "Morpho Blue"},
	{"Withdraw(address,address,uint256)", k256("Withdraw(address,address,uint256)"), EvtMorphoWithdraw, "Morpho Blue"},
	{"Liquidate(address,address,uint256)", k256("Liquidate(address,address,uint256)"), EvtMorphoLiquidate, "Morpho Blue"},

	// EigenLayer withdrawals
	{"WithdrawalQueued(uint256,address,address,address,uint256)", k256("WithdrawalQueued(uint256,address,address,address,uint256)"), EvtEigenWithdrawQueued, "EigenLayer"},
	{"WithdrawalCompleted(uint256,address,address,uint256)", k256("WithdrawalCompleted(uint256,address,address,uint256)"), EvtEigenWithdrawCompleted, "EigenLayer"},

	// Frax sfrxETH
	{"Submit(address,uint256)", k256("Submit(address,uint256)"), EvtFraxSubmit, "Frax"},
	{"Withdraw(address,uint256)", k256("Withdraw(address,uint256)"), EvtFraxWithdraw, "Frax"},

	// Stakewise
	{"Staked(address,uint256)", k256("Staked(address,uint256)"), EvtStakewiseStaked, "Stakewise"},
	{"Unstaked(address,uint256)", k256("Unstaked(address,uint256)"), EvtStakewiseUnstaked, "Stakewise"},

	// Wormhole
	{"LogMessagePublished(address,uint64,uint32,uint32,bytes,uint8)", k256("LogMessagePublished(address,uint64,uint32,uint32,bytes,uint8)"), EvtWormholeMessage, "Wormhole"},

	// Synapse
	{"TokenSwap(address,uint256,uint256,uint256,uint256)", k256("TokenSwap(address,uint256,uint256,uint256,uint256)"), EvtSynapseSwap, "Synapse"},

	// Celer / cBridge
	{"RelaySent(bytes32,address,uint64)", k256("RelaySent(bytes32,address,uint64)"), EvtCelerRelaySent, "Celer"},
	{"RelayConfirmed(bytes32,address,uint64)", k256("RelayConfirmed(bytes32,address,uint64)"), EvtCelerRelayConfirmed, "Celer"},

	// Blur batch execution
	{"ExecutionBatch721(bytes32,uint256[],uint256[])", k256("ExecutionBatch721(bytes32,uint256[],uint256[])"), EvtBlurExecutionBatch, "Blur"},
}

// LookupEvent returns the matching EventDef for a given topic0, or nil if unknown.
var (
	eventByTopic     map[common.Hash]*EventDef
	eventByTopicOnce sync.Once
)

func LookupEvent(topic0 common.Hash) *EventDef {
	eventByTopicOnce.Do(func() {
		eventByTopic = make(map[common.Hash]*EventDef, len(EventCatalog))
		for i := range EventCatalog {
			if _, dup := eventByTopic[EventCatalog[i].Topic]; dup {
				panic("duplicate event topic: " + EventCatalog[i].Signature)
			}
			eventByTopic[EventCatalog[i].Topic] = &EventCatalog[i]
		}
	})
	if d, ok := eventByTopic[topic0]; ok {
		return d
	}
	return nil
}

// IsERC20Transfer returns true when this Transfer log is ERC-20 (3 topics + 32-byte data),
// false when it's ERC-721 (4 topics, all indexed including tokenId).
func IsERC20Transfer(numTopics, dataLen int) bool {
	return numTopics == 3 && dataLen >= 32
}

func IsERC721Transfer(numTopics int) bool {
	return numTopics == 4
}

// NormalizeSig normalizes a signature for comparison ("transfer ( address , uint256 )" -> "transfer(address,uint256)").
func NormalizeSig(s string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(s, " ", "")), "")
}
