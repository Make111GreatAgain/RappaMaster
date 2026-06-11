//go:build ignore

package service

import (
	"context"
	"fmt"
	"math/big"
	"time"

	Store "BHLayer2Node/ChainUpper/contract/storeData"
	"BHLayer2Node/paradigm"

	"github.com/FISCO-BCOS/go-sdk/v3/client"
	"github.com/FISCO-BCOS/go-sdk/v3/types"
	"github.com/ethereum/go-ethereum/common"
)

// 单笔上链，不做批量 链上存储具体的交易数据
type UpChainWorker struct {
	id                   int
	queue                chan paradigm.Transaction
	devPackedTransaction chan []*paradigm.PackedTransaction
	instance             *Store.StoreData
	client               *client.Client
}

func NewUpchainWorker(
	id int,
	queue chan paradigm.Transaction,
	dev chan []*paradigm.PackedTransaction,
	instance *Store.StoreData,
	client *client.Client,
) *UpChainWorker {
	return &UpChainWorker{
		id:                   id,
		queue:                queue,
		devPackedTransaction: dev,
		instance:             instance,
		client:               client,
	}
}

func (w *UpChainWorker) Process() {
	for tx := range w.queue {
		if tx == nil {
			return
		}
		params, err := tx.StoreParams()
		paradigm.Log("CHAINUP", fmt.Sprintf("Transaction %s:\n %+v\n upchain params: %+v", tx.Call(), tx.CallData(), params))
		if err != nil {
			paradigm.Error(paradigm.UpchainError, "转换上链参数失败: "+err.Error())
			continue
		}
		session := &Store.StoreDataSession{Contract: w.instance, TransactOpts: *w.client.GetTransactOpts()}
		// 收到单笔交易，直接上链
		var receipt *types.Receipt
		switch tx.(type) {
		case *paradigm.InitTaskTransaction:
			_, receipt, err = session.StoreInitTask(
				params[0].([32]byte),
				params[1].(string),
				params[2].(uint32),
				params[3].([32]byte),
				params[4].(bool),
				params[5].([]byte),
			)
			Tx := tx.Blob().(*paradigm.Task)
			Txjson, err := w.GetInitTaskJSON(Tx.Sign)
			if err != nil {
				paradigm.Log("ERROR", fmt.Sprintf("query InitTaskTransaction from blockchain By task %s failed %s", Tx.Sign, err))
			} else {
				paradigm.Log("CHAINUP", fmt.Sprintf("query InitTaskTransaction from blockchain By task %s: %s", Tx.Sign, Txjson))
			}
		case *paradigm.TaskProcessTransaction:
			_, receipt, err = session.StoreTaskProcess(
				params[0].([32]byte),
				params[1].([32]byte),
				params[2].(uint32),
				params[3].(uint32),
				params[4].(uint32),
				params[5].(uint32),
				params[6].(common.Hash),
				params[7].([]byte),
				params[8].([][]byte),
			)
			Tx := tx.Blob().(*paradigm.CommitSlotItem)
			Txjson, err := w.GetTaskProcessJSON(Tx.SlotHash())
			if err != nil {
				paradigm.Log("ERROR", fmt.Sprintf("query TaskProcessTransaction from blockchain By Slot %s failed %s", Tx.SlotHash(), err))
			} else {
				paradigm.Log("CHAINUP", fmt.Sprintf("query TaskProcessTransaction from blockchain By Slot %s: %s", Tx.SlotHash(), Txjson))
			}
		case *paradigm.EpochRecordTransaction:
			_, receipt, err = session.StoreEpochRecord(
				params[0].(*big.Int),
				params[1].([][32]byte),
				params[2].([][32]byte),
				params[3].([][32]byte),
				params[4].([]uint8),
			)
			Tx := tx.Blob().(*paradigm.EpochRecord)
			Txjson, err := w.GetEpochRecordJSON(uint64(Tx.Id))
			if err != nil {
				paradigm.Log("ERROR", fmt.Sprintf("query EpochRecordTransaction from blockchain By ID %d failed %s", Tx.Id, err))
			} else {
				paradigm.Log("CHAINUP", fmt.Sprintf("query EpochRecordTransaction from blockchain By ID %d: %s", Tx.Id, Txjson))
			}
		default:
			panic("Unknown transaction type")
		}
		if err != nil {
			paradigm.Error(paradigm.UpchainError, "上链失败: "+err.Error())
			continue
		}
		w.waitAndDevPush(receipt, tx)
	}
}

// 交易上链后，将回执和交易打包发送给Oracle
func (w *UpChainWorker) waitAndDevPush(receipt *types.Receipt, tx paradigm.Transaction) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rcpt, err := w.client.GetTransactionReceipt(ctx, common.HexToHash(receipt.TransactionHash), true)
	if err != nil {
		paradigm.Error(paradigm.RuntimeError, "GetReceipt error: "+err.Error())
		return
	}
	blockHash, _ := w.client.GetBlockHashByNumber(ctx, int64(rcpt.BlockNumber))
	devTxs := []*paradigm.PackedTransaction{
		paradigm.NewPackedTransaction(tx, rcpt, blockHash.Hex()),
	}
	w.devPackedTransaction <- devTxs
}
