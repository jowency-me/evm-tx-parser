package classifier

import (
	"math/big"

	semantic "github.com/jowency-me/evm-tx-parser/semtypes"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// decodeERC20Transfer decodes the standard ERC-20 Transfer(address,address,uint256)
// from a log with 3 topics + 32 bytes of data.
func decodeERC20Transfer(lg *types.Log) semantic.TokenFlow {
	from := common.BytesToAddress(lg.Topics[1].Bytes())
	to := common.BytesToAddress(lg.Topics[2].Bytes())
	amount := new(big.Int).SetBytes(lg.Data)
	return semantic.TokenFlow{
		Kind:   semantic.AssetERC20,
		Token:  lg.Address,
		From:   from,
		To:     to,
		Amount: amount,
	}
}

// decodeERC721Transfer: Transfer(address from, address to, uint256 tokenId) — all indexed.
func decodeERC721Transfer(lg *types.Log) semantic.TokenFlow {
	from := common.BytesToAddress(lg.Topics[1].Bytes())
	to := common.BytesToAddress(lg.Topics[2].Bytes())
	tokenID := new(big.Int).SetBytes(lg.Topics[3].Bytes())
	return semantic.TokenFlow{
		Kind:    semantic.AssetERC721,
		Token:   lg.Address,
		From:    from,
		To:      to,
		TokenID: tokenID,
	}
}

// decodeERC1155Single: TransferSingle(operator, from, to, id, value)
func decodeERC1155Single(lg *types.Log) semantic.TokenFlow {
	if len(lg.Topics) < 4 || len(lg.Data) < 64 {
		return semantic.TokenFlow{Kind: semantic.AssetERC1155, Token: lg.Address}
	}
	from := common.BytesToAddress(lg.Topics[2].Bytes())
	to := common.BytesToAddress(lg.Topics[3].Bytes())
	id := new(big.Int).SetBytes(lg.Data[:32])
	val := new(big.Int).SetBytes(lg.Data[32:64])
	return semantic.TokenFlow{
		Kind:    semantic.AssetERC1155,
		Token:   lg.Address,
		From:    from,
		To:      to,
		TokenID: id,
		Amount:  val,
	}
}

// decodeWETHEvent decodes WETH-style Deposit(address,uint256) or Withdrawal(address,uint256).
func decodeWETHEvent(lg *types.Log, cat any) semantic.TokenFlow {
	user := common.BytesToAddress(lg.Topics[1].Bytes())
	amount := new(big.Int)
	if len(lg.Data) >= 32 {
		amount.SetBytes(lg.Data[:32])
	}
	// We treat the WETH contract as the To (deposit) or From (withdrawal).
	// Without knowing direction, we just put user in From and zero in To; the engine
	// knows the protocol name.
	return semantic.TokenFlow{
		Kind:   semantic.AssetERC20,
		Token:  lg.Address,
		From:   user,
		Amount: amount,
	}
}
