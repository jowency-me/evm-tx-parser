package protocols

import (
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// AddressTag attaches semantic info to a known contract address on a specific chain.
// AddressTag attaches semantic info to a known contract address on a specific chain.
type AddressTag struct {
	Chain    uint64
	Address  common.Address
	Protocol string
	Role     string // e.g. "router", "pool", "vault"
}

// raw catalog — case-insensitive lookups via map below.
var rawTags = []AddressTag{
	// === Ethereum mainnet (chain 1) ===
	{1, common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"), "WETH", "wrapped_native"},
	{1, common.HexToAddress("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D"), "Uniswap V2", "router"},
	{1, common.HexToAddress("0xE592427A0AEce92De3Edee1F18E0157C05861564"), "Uniswap V3", "router"},
	{1, common.HexToAddress("0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45"), "Uniswap V3", "router02"},
	{1, common.HexToAddress("0x66a9893cC07D91D95644AEDD05D03f95e1dBA8Af"), "Uniswap Universal Router", "router"},
	{1, common.HexToAddress("0x3fC91A3afd70395Cd496C647d5a6CC9D4B2b7FAD"), "Uniswap Universal Router", "router"},
	{1, common.HexToAddress("0x000000000022D473030F116dDEE9F6B43aC78BA3"), "Permit2", "permit2"},
	{1, common.HexToAddress("0x000000000004444c5dc75cB358380D2e3dE08A90"), "Uniswap V4 PoolManager", "pool_manager"},
	{1, common.HexToAddress("0xBA12222222228d8Ba445958a75a0704d566BF2C8"), "Balancer V2", "vault"},
	{1, common.HexToAddress("0x87870Bca3F3fD6335C3F4ce8392D69350B4fA4E2"), "Aave V3", "pool"},
	{1, common.HexToAddress("0x7d2768dE32b0b80b7a3454c06BdAc94A69DDc7A9"), "Aave V2", "pool"},
	{1, common.HexToAddress("0xae7ab96520DE3A18E5e111B5EaAb095312D7fE84"), "Lido", "stETH"},
	{1, common.HexToAddress("0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0"), "Lido", "wstETH"},
	{1, common.HexToAddress("0xae78736Cd615f374D3085123A210448E74Fc6393"), "Rocket Pool", "rETH"},
	{1, common.HexToAddress("0x39AA39c021dfbaE8faC545936693aC917d5E7563"), "Compound V2", "cUSDC"},
	{1, common.HexToAddress("0xc3d688B66703497DAA19211EEdff47f25384cdc3"), "Compound V3", "cUSDCv3"},
	{1, common.HexToAddress("0x00000000219ab540356cBB839Cbe05303d7705Fa"), "Eth2 Beacon Deposit", "deposit"},
	{1, common.HexToAddress("0x858646372CC42E1A627fcE94aa7A7033e7CF075A"), "EigenLayer", "strategy_manager"},
	{1, common.HexToAddress("0xAAAA4AcE7dB229d166c3a59be1A4909b735E8e61"), "Morpho Blue", "pool"},
	{1, common.HexToAddress("0x1111111254EEB25477B68fb85Ed929f73A960582"), "1inch", "router_v5"},
	{1, common.HexToAddress("0x111111125421cA6dc452d289314280a0f8842A65"), "1inch", "router_v6"},
	{1, common.HexToAddress("0xDef1C0ded9bec7F1a1670819833240f027b25EfF"), "0x", "exchange_proxy"},
	{1, common.HexToAddress("0x00000000000000ADc04C56Bf30aC9d3c0aAF14dC"), "OpenSea Seaport", "seaport_1_5"},
	{1, common.HexToAddress("0x0000000000000068F116a894984e2DB1123eB395"), "OpenSea Seaport", "seaport_1_6"},
	{1, common.HexToAddress("0xb2ecfE4E4D61f8790bbb9DE2D1259B9e2410CEA5"), "Blur", "marketplace"},
	{1, common.HexToAddress("0xa0b86991c6218B36c1d19D4a2e9Eb0cE3606eB48"), "USDC", "stablecoin"},
	{1, common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7"), "USDT", "stablecoin"},
	{1, common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F"), "DAI", "stablecoin"},
	{1, common.HexToAddress("0xbEbc44782C7dB0a1A60Cb6fe97d0b483032FF1C7"), "Curve 3pool", "pool"},
	{1, common.HexToAddress("0xDC24316b9AE028F1497c275EB9192a3Ea0f67022"), "Curve stETH", "pool"},

	// === Base (chain 8453) ===
	{8453, common.HexToAddress("0x4200000000000000000000000000000000000006"), "WETH (Base)", "wrapped_native"},
	{8453, common.HexToAddress("0x2626664c2603336E57B271c5C0b26F421741e481"), "Uniswap V3", "router02"},
	{8453, common.HexToAddress("0x6fF5693b99212Da76ad316178A184AB56D299b43"), "Uniswap Universal Router", "router"},
	{8453, common.HexToAddress("0x3154Cf16ccdb4C6d922629664174b904d80F2C35"), "BaseSwap", "router"},
	{8453, common.HexToAddress("0xcF77a3Ba9A5CA399B7c97c74d54e5b1Beb874E43"), "Aerodrome", "router"},
	{8453, common.HexToAddress("0x6Cb442acF35158D5eDa88fe602221b67B400Be3E"), "Aerodrome SlipStream", "router"},
	{8453, common.HexToAddress("0xA238Dd80C259a72e81d7e4664a9801593F98d1c5"), "Aave V3 (Base)", "pool"},
	{8453, common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"), "USDC (Base)", "stablecoin"},
	{8453, common.HexToAddress("0x2Ae3F1Ec7F1F5012CFEab0185bfc7aa3cf0DEc22"), "cbETH (Base)", "lst"},

	// Seaport is deployed at the same canonical address on every EVM chain.
	{8453, common.HexToAddress("0x0000000000000068F116a894984e2DB1123eB395"), "OpenSea Seaport", "seaport_1_6"},
	{10, common.HexToAddress("0x0000000000000068F116a894984e2DB1123eB395"), "OpenSea Seaport", "seaport_1_6"},
	{137, common.HexToAddress("0x0000000000000068F116a894984e2DB1123eB395"), "OpenSea Seaport", "seaport_1_6"},
	{42161, common.HexToAddress("0x0000000000000068F116a894984e2DB1123eB395"), "OpenSea Seaport", "seaport_1_6"},

	// === Arbitrum (42161) ===
	{42161, common.HexToAddress("0x82aF49447D8a07e3bd95BD0d56f35241523fBab1"), "WETH (Arbitrum)", "wrapped_native"},
	{42161, common.HexToAddress("0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45"), "Uniswap V3", "router02"},
	{42161, common.HexToAddress("0x5E325eDA8064b456f4781070C0738d849c824258"), "Uniswap Universal Router", "router"},
	{42161, common.HexToAddress("0x794a61358D6845594F94dc1DB02A252b5b4814aD"), "Aave V3", "pool"},
	{42161, common.HexToAddress("0xaf88d065e77c8cC2239327C5EDb3A432268e5831"), "USDC (Arbitrum)", "stablecoin"},
	{42161, common.HexToAddress("0xfd086bC7CD5C481DCC9C85ebE478A1C0b69FCbb9"), "USDT (Arbitrum)", "stablecoin"},
	{42161, common.HexToAddress("0xCAFcD85D8ca7Ad1e1C6F82F651fA15E33AEfD07b"), "GMX", "router"},

	// === Optimism (10) ===
	{10, common.HexToAddress("0x4200000000000000000000000000000000000006"), "WETH (Optimism)", "wrapped_native"},
	{10, common.HexToAddress("0x794a61358D6845594F94dc1DB02A252b5b4814aD"), "Aave V3", "pool"},
	{10, common.HexToAddress("0x9c12939390052919aF7155f3E06C4531303D5c14"), "Velodrome", "router"},
	{10, common.HexToAddress("0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45"), "Uniswap V3", "router02"},

	// === Polygon (137) ===
	{137, common.HexToAddress("0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270"), "WMATIC", "wrapped_native"},
	{137, common.HexToAddress("0xa5E0829CaCEd8fFDD4De3c43696c57F7D7A678ff"), "QuickSwap", "router"},
	{137, common.HexToAddress("0xE592427A0AEce92De3Edee1F18E0157C05861564"), "Uniswap V3", "router"},
	{137, common.HexToAddress("0x794a61358D6845594F94dc1DB02A252b5b4814aD"), "Aave V3", "pool"},

	// === BSC (56) ===
	{56, common.HexToAddress("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c"), "WBNB", "wrapped_native"},
	{56, common.HexToAddress("0x10ED43C718714eb63d5aA57B78B54704E256024E"), "PancakeSwap V2", "router"},
	{56, common.HexToAddress("0x13f4EA83D0bd40E75C8222255bc855a974568Dd4"), "PancakeSwap V3", "router"},

	// === Ethereum mainnet additions ===
	{1, common.HexToAddress("0x9aA8B59a45FD85F6d3D2C98B36200A12dC6Fbc02"), "Stargate", "router"},
	{1, common.HexToAddress("0x36c687b439F7b3a1cA7E54f12cA2496b0e715DC6"), "Across V3", "spoke_pool"},
	{1, common.HexToAddress("0xb55D9B8eE3438B2be2271d1E7033D3553e59193E"), "Hop", "bridge"},
	{1, common.HexToAddress("0xF2ec4b1104e205CF86a1786eB91DE1d57503e20e"), "CCTP", "token_messenger"},
	{1, common.HexToAddress("0x9008D19f58AAbD9eD0D60971565AA8510560ab41"), "CowSwap", "settlement"},
	{1, common.HexToAddress("0xAf5B6cE0fA9e3a8BBcce0060766063c79BA36c30"), "SushiSwap", "router"},
	{1, common.HexToAddress("0xd9e1cE17f2641f24aE83637ab66a2cca9C378B9F"), "SushiSwap", "router_old"},

	// === Avalanche (43114) ===
	{43114, common.HexToAddress("0xB31f66AA3C1e785363F0875A1B74E27b85FD66c7"), "WAVAX", "wrapped_native"},
	{43114, common.HexToAddress("0x60aE616a2155Ee3d9A68541Ba4544862310a1fA1"), "Trader Joe", "router"},
	{43114, common.HexToAddress("0xE54Ca2eDb4C7e0bdd2C2e6923e5ddBB435b64e10"), "SushiSwap", "router"},
	{43114, common.HexToAddress("0xbb00E867772A426C82AafeE68694Bc57aC51C1F5"), "Aave V3", "pool"},
	{43114, common.HexToAddress("0x6B78598a1c4a8aCD92e15F7E2165085A8c16f3B1"), "Stargate", "router"},

	// === Base additions ===
	{8453, common.HexToAddress("0x3206695CaE29952f4b0c22a1691c7d7565E5e29e"), "Stargate", "router"},

	// === Arbitrum additions ===
	{42161, common.HexToAddress("0x53Bf833A5d6c4ddA888F69c22C88C9f356a41614"), "Stargate", "router"},

	// === Optimism additions ===
	{10, common.HexToAddress("0xB244b31eCbdECCa14034D5F8eCDEaC784F3eCCd9"), "Stargate", "router"},

	// === Polygon additions ===
	{137, common.HexToAddress("0x9aA8B59a45FD85F6d3D2C98B36200A12dC6Fbc02"), "Stargate", "router"},
	{137, common.HexToAddress("0xaC1468BF81a489971882a6e01f1aF68735a1348E"), "SushiSwap", "router"},

	// === BSC additions ===
	{56, common.HexToAddress("0x4a364f8c717cAAD9A442737Eb7b8A55cc6cf18D8"), "Stargate", "router"},
	{56, common.HexToAddress("0x1b02dA8Cb0d097eB8D57A175b88c7D8b47997506"), "SushiSwap", "router"},
	{56, common.HexToAddress("0xcF6BB5389c92Bdda8a3747Ddb454cB7a64626C63"), "Venus", "pool"},

	// === DEX addresses ===
	{1, common.HexToAddress("0xD51a44d3FaE010294C616388b506cdaa1c89a1bd"), "Curve", "tricrypto_pool"},
	{1, common.HexToAddress("0x866A2BF7E120Fe3e88F7c4cc4DC7687Bd7248670"), "Curve", "ng_pool"},
	{1, common.HexToAddress("0x6175a8bdecCb296985AcFb04d8C10cE1FbaF27D6"), "Kyber", "dmm_router"},
	{1, common.HexToAddress("0x6131B5fae19EA4f9D964eAc0408E4408b66337b5"), "Kyber", "elastic_router"},
	{1, common.HexToAddress("0x3B6067D4CAa8A14eF37a8C6f52AF354AaF7E7902"), "DODO", "proxy"},
	{1, common.HexToAddress("0xDEF171Fe48CF0115B1D80b88dc8eAB59176FEe57"), "ParaSwap", "router_v5"},

	// === Lending addresses ===
	{1, common.HexToAddress("0x35D1b3F3D7966A1DFe207aa4514C12a259A0492B"), "MakerDAO", "vat"},
	{1, common.HexToAddress("0x9759A6Ac90977b93B58547b4A71c78317f391A28"), "MakerDAO", "dai_join"},
	{1, common.HexToAddress("0xC13e21B648A5EbbC9cB5A5fD31944639767772c3"), "Spark", "pool"},
	{56, common.HexToAddress("0xfD36E2c2a6789Db23113685031d7F163d1A3B44f"), "Venus", "comptroller"},

	// === Staking addresses ===
	{1, common.HexToAddress("0xBe9895146f7AF43049ca1c1AE358B0541Ea49704"), "Coinbase", "cbETH"},
	{1, common.HexToAddress("0xac3E018457B222d93114458476f3E3416Abbe38F"), "Frax", "sfrxETH"},
	{1, common.HexToAddress("0xFe2e63720205603020314F81D46Aa66c0c89c06C"), "Stakewise", "sETH2"},
	{1, common.HexToAddress("0xf1C9acDc66974dFB6dEcB12aA385b9cD01190E38"), "Stakewise", "osETH"},

	// === Bridge addresses ===
	{1, common.HexToAddress("0x98f3c9e6E3fAce36bAAd05FE09d375Ef1464288B"), "Wormhole", "core_bridge"},
	{1, common.HexToAddress("0x3ee18B2214AFF97000D974cf647E7C347E8fa585"), "Wormhole", "token_bridge"},
	{1, common.HexToAddress("0x2796D7fA2a2948C2c3a18B10D8c0e5a2a51f8Da"), "Synapse", "bridge"},
	{1, common.HexToAddress("0x5427FEFA711Eff9841244BBED8E50fDD7c87458A"), "Celer", "cbridge"},
	{1, common.HexToAddress("0x1231DEB6f5749EF6cE6943a275A1D3E7486F4EaE"), "LI.FI", "router"},

	// Across SpokePools
	{42161, common.HexToAddress("0xe4deC6490d96267Fa60397a94c7df4e742bE2896"), "Across", "spoke_pool"},
	{10, common.HexToAddress("0x6B36d87e64CA3e8Fbd8Ad0040e46b2c5D25e3F34"), "Across", "spoke_pool"},
	{137, common.HexToAddress("0x9f7C684b37968f7B7ac0E0165D5Bf2d00aB47DE1"), "Across", "spoke_pool"},
	{56, common.HexToAddress("0xE0B0668D091392B3966c97c3b24adE9B87c4F2a5"), "Across", "spoke_pool"},

	// Hop bridges
	{42161, common.HexToAddress("0x36238419Ba8Df3F6F4625F809D2E73b1d1Bda0E8"), "Hop", "bridge"},
	{10, common.HexToAddress("0x36238419Ba8Df3F6F4625F809D2E73b1d1Bda0E8"), "Hop", "bridge"},
	{137, common.HexToAddress("0x1a1F6b3542A68EdC01dC5199eD7F37f6eE9E58e4"), "Hop", "bridge"},

	// CCTP TokenMessengers
	{43114, common.HexToAddress("0x818635159F7F4aeE64cc249C94f53f94B2666DF6"), "CCTP", "token_messenger"},
	{42161, common.HexToAddress("0x19330d10D9Cc8751218eaf51E8885D058642E08A"), "CCTP", "token_messenger"},
	{10, common.HexToAddress("0x2B406f025Dac2ff37BF1E2AD2Fb7a300C33C6676"), "CCTP", "token_messenger"},
	{137, common.HexToAddress("0x9daF8c91AEFAE50b9c0E69629D3A5E2B78CA2A44"), "CCTP", "token_messenger"},
	{8453, common.HexToAddress("0x1682Ae6375C4E4A97e2B5832a0Fc74Da2A1640e6"), "CCTP", "token_messenger"},

	// === Liquidity/NFT addresses ===
	{1, common.HexToAddress("0xC36442b4a4522E871399CD717aBDD847Ab11FE88"), "Uniswap V3", "nonfungible_position_manager"},
	{42161, common.HexToAddress("0xC36442b4a4522E871399CD717aBDD847Ab11FE88"), "Uniswap V3", "nonfungible_position_manager"},
	{137, common.HexToAddress("0xC36442b4a4522E871399CD717aBDD847Ab11FE88"), "Uniswap V3", "nonfungible_position_manager"},
	{10, common.HexToAddress("0xC36442b4a4522E871399CD717aBDD847Ab11FE88"), "Uniswap V3", "nonfungible_position_manager"},
	{8453, common.HexToAddress("0xC36442b4a4522E871399CD717aBDD847Ab11FE88"), "Uniswap V3", "nonfungible_position_manager"},

	// === GMX addresses ===
	{42161, common.HexToAddress("0xA90e8e93E7c6c8CC70c38497E4B0EA0727E0D095"), "GMX", "router_v2"},
	{42161, common.HexToAddress("0x87a4088B272Cc7e4E579eb9927f7e7f2f7f4D7b2"), "GMX", "position_manager"},
}

type addrKey struct {
	Chain uint64
	Addr  common.Address
}

var addrIndex map[addrKey]*AddressTag

func init() {
	addrIndex = make(map[addrKey]*AddressTag, len(rawTags))
	for i := range rawTags {
		t := &rawTags[i]
		addrIndex[addrKey{t.Chain, t.Address}] = t
	}
}

// LookupAddress returns the tag for (chain, addr) or nil.
func LookupAddress(chain uint64, addr common.Address) *AddressTag {
	if t, ok := addrIndex[addrKey{chain, addr}]; ok {
		return t
	}
	return nil
}

// MatchProtocolByRole tries to find any address on chain whose protocol/role contains substr.
// Mostly used for testing/debugging.
func MatchProtocolByRole(chain uint64, role string) []*AddressTag {
	var out []*AddressTag
	for _, t := range rawTags {
		if t.Chain == chain && strings.Contains(strings.ToLower(t.Role), strings.ToLower(role)) {
			out = append(out, &t)
		}
	}
	return out
}

// AllTags returns all registered address tags (for scan tool use).
func AllTags() []AddressTag {
	return rawTags
}
