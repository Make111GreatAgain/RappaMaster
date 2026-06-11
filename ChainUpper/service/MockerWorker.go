//go:build ignore

package service

import (
	"BHLayer2Node/paradigm"
	"github.com/FISCO-BCOS/go-sdk/v3/types"
)

type MockerUpChainWorker struct {
	id                   int
	queue                chan paradigm.Transaction
	devPackedTransaction chan []*paradigm.PackedTransaction // add by zhmye 这里是用来给dev的，所有已经上链的交易都要给
	//instance             *SlotCommit.SlotCommit
	//client               *client.Client
	batchSize int
	//signs                                                  [][32]byte
	//slotsBigInt, processesBigInt, nidsBigInt, epochsBigInt []*big.Int
	params map[paradigm.TransactionType]paradigm.PackedParams // 这里记录各种类型的交易参数 add by zhmye
	count  int                                                // 交易计数
}

func (w *MockerUpChainWorker) Process() {
	for {
		select {
		case tx := <-w.queue: // 尝试从通道中接收数据
			if tx != nil { // 判断是否接收到有效值
				// log.Printf("Worker %d Received result: %v", id, result)
				//paradigm.Log("CHAINUP", fmt.Sprintf("Worker %d Received Transaction: %v", w.id, tx))

				switch tx.(type) {
				case *paradigm.InitTaskTransaction:
					w.params[paradigm.INIT_TASK_TRANSACTION].UpdateFromTransaction(tx)
				case *paradigm.TaskProcessTransaction:
					w.params[paradigm.TASK_PROCESS_TRANSACTION].UpdateFromTransaction(tx)
				case *paradigm.EpochRecordTransaction:
					w.params[paradigm.EPOCH_RECORD_TRANSACTION].UpdateFromTransaction(tx)
				default:
					panic("Invalid Transaction Type!!!")
				}
				w.count++
				if w.count >= w.batchSize {
					// 每当收集到batchSize个transaction的信息时，调用批量上链函数
					w.consumer()
					w.count = 0
				}
			} else {
				//paradigm.Log("ERROR", fmt.Sprintf("Upchain channel closed, received nil value"))
				return
			}
		}
	}
}
func (w *MockerUpChainWorker) consumer() {
	for _, packedParam := range w.params {
		param := packedParam.GetParams()
		if len(param) != packedParam.ParamsLen() && packedParam.ParamsLen() != -1 { // todo
			panic("Param Length Error...Please check the code in paradigm!!!")
		}
		receipt := types.Receipt{}
		ptxs := packedParam.BuildDevTransactions([]*types.Receipt{&receipt}, "")
		w.devPackedTransaction <- ptxs // 传递到dev
	}
	w.params = paradigm.NewParamsMap()
}
func NewMockerUpChainWorker(id int, queue chan paradigm.Transaction, dev chan []*paradigm.PackedTransaction) *MockerUpChainWorker {

	return &MockerUpChainWorker{
		id:                   id,
		queue:                queue,
		devPackedTransaction: dev,
		//instance:             instance,
		//client:               client,
		batchSize: 1,
		params:    paradigm.NewParamsMap(),
		count:     0,
	}
}
