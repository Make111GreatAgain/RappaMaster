//go:build ignore

package ChainUpper

import (
	"BHLayer2Node/Query"
	"BHLayer2Node/paradigm"
	"context"
	"fmt"
	"time"

	"github.com/FISCO-BCOS/go-sdk/v3/types"
	"github.com/ethereum/go-ethereum/common"
)

func (c *ChainUpper) handleQuery() {
	for query := range c.channel.BlockchainQueryChannel {
		c.handle(query)
	}
}

func (c *ChainUpper) handle(query paradigm.Query) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	switch query.(type) {
	// 交易和最新信息的query都在oracle处理了,TODO @XQ 判断一下在oracle里的信息是否完整/正确
	case *Query.BlockchainBlockHashQuery:
		// 通过client获取到block
		item := query.(*Query.BlockchainBlockHashQuery)
		blockHash := item.BlockHash
		if blockHash == "" {
			item.SendInfo(paradigm.NewInvalidBlockInfo("blockHash parameter missing or invalid"))
			return
			//return *paradigm.NewErrorResponse(-1, "blockHash parameter missing or invalid", int(req.QueryType))
		}
		hash := common.HexToHash(blockHash)
		block, err := c.client.GetBlockByHash(ctx, hash, false, false)
		if err != nil {
			item.SendInfo(paradigm.NewInvalidBlockInfo(fmt.Sprintf("Failed to get block: %v", err)))
			//return *paradigm.NewErrorResponse(-1, fmt.Sprintf("Failed to get block: %v", err), int(req.QueryType))
			return
		}
		blockInfo := c.getBlockInfo(*block)
		item.SendInfo(blockInfo)

	case *Query.BlockchainBlockNumberQuery:
		// 通过client获取到block
		item := query.(*Query.BlockchainBlockNumberQuery)
		blockNumber := item.BlockNumber
		block, err := c.client.GetBlockByNumber(ctx, int64(blockNumber), false, false)
		if err != nil {
			item.SendInfo(paradigm.NewInvalidBlockInfo(fmt.Sprintf("Failed to get block: %v", err)))
			return
		}
		blockInfo := c.getBlockInfo(*block)
		item.SendInfo(blockInfo)

	case *Query.BlockchainTransactionQuery:
		item := query.(*Query.BlockchainTransactionQuery)
		txHash := common.HexToHash(item.TxHash)
		receipt, err := c.client.GetTransactionReceipt(ctx, txHash, false)
		if err != nil {
			item.SendInfo(paradigm.NewInvalidTransactionInfo(fmt.Sprintf("Failed to get transaction: %v", err)))
			return
		}
		// blockHash, err := c.client.GetBlockHashByNumber(ctx, int64(receipt.BlockNumber))
		// if err != nil {
		// 	item.SendInfo(paradigm.NewInvalidTransactionInfo(fmt.Sprintf("Failed to get transaction blockHash: %v", err)))
		// 	return
		// }
		// txInfo := c.getTransactionInfo(receipt, blockHash.Hex())

		// block 提供 blockHash 和 upchainTime
		block, err := c.client.GetBlockByNumber(ctx, int64(receipt.BlockNumber), false, false)
		txInfo := c.getTransactionInfo(receipt, block.Hash, block.Timestamp)
		item.SendInfo(txInfo)

	default:
		paradigm.Error(paradigm.RuntimeError, "Unsupported Query Type In ChainUpper")
	}
}

func (c *ChainUpper) getBlockInfo(block types.Block) paradigm.BlockInfo {
	var parentHash string
	if len(block.ParentInfo) > 0 {
		// 取第一个父区块的哈希（通常只有一个父区块）
		parentHash = block.ParentInfo[0].GetBlockHash()
	} else {
		parentHash = "genesis" // 创世区块没有父哈希
	}

	//var txDetails []TxDetail
	txs := make([]paradigm.TransactionInfo, 0) // 只要txHash，剩余在Oracle里获取
	// LogWriter.Log("DEBUG", fmt.Sprintf("Block.GetTransactions() return %s", block.GetTransactions()...))
	for _, tx := range block.Transactions {
		// 首先尝试将 tx 转换为 *types.TransactionDetail
		//var txHash string
		if txObj, ok := tx.(*types.TransactionDetail); ok {
			txs = append(txs, paradigm.TransactionInfo{
				TxHash:       txObj.GetHash(),
				Contract:     txObj.GetTo(),
				Abi:          txObj.GetAbi(),
				BlockHash:    block.Hash,
				UpchainTime:  paradigm.TimestampConvert(block.Timestamp),
				Invalid:      false,
				ErrorMessage: "",
			})
		} else {
			//		// 如果转换失败，尝试转换为 map[string]interface{}
			if txMap, ok := tx.(map[string]interface{}); ok {
				// TODO @XQ 这里修正一下，包括exist的判断
				txs = append(txs, paradigm.TransactionInfo{
					TxHash:       txMap["hash"].(string),
					Contract:     txMap["to"].(string),
					Abi:          txMap["abi"].(string),
					BlockHash:    block.Hash,
					UpchainTime:  paradigm.TimestampConvert(block.Timestamp),
					Invalid:      false,
					ErrorMessage: "",
				})
			}
		}
	}
	return paradigm.BlockInfo{
		BlockHash:       block.Hash,
		ParentHash:      parentHash,
		BlockHeight:     int32(block.Number),
		TransactionRoot: block.TxsRoot,
		Txs:             txs,
		Invalid:         false,
		ErrorMessage:    "",
	}
}

func (c *ChainUpper) getTransactionInfo(receipt *types.Receipt, blockHash string, timestamp uint64) paradigm.TransactionInfo {
	return paradigm.TransactionInfo{
		TxHash:   receipt.TransactionHash,
		Contract: receipt.To,
		Abi:      "processTask", // todo
		// Abi:       abi, // nil?
		BlockHash:    blockHash,
		UpchainTime:  paradigm.TimestampConvert(timestamp),
		Invalid:      false,
		ErrorMessage: "",
	}
}
