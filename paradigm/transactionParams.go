package paradigm

import (
	"BHLayer2Node/utils"
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

type PackedParams interface {
	UpdateFromTransaction(tx Transaction) // 根据交易得到参数
	GetParams() []interface{}
	ParamsLen() int
	IsEmpty() bool
	ParamsToKVPairs() ([][32]byte, [][32]byte) // batch setItems
	BuildDevTransactions(receipts []*Receipt, blockHash string) []*PackedTransaction
}

// TODO 这里每多一个交易类型就要加上对应的参数

type InitTaskTransactionParams struct {
	txs []*InitTaskTransaction
	// TODO @XQ
	// signs [][32]byte
	// sizes  []*big.Int
	// models [][32]byte
	// isReliables [][32]byte
	// paramsList  [][32]byte
	callDatas []map[string]interface{}
}

func (p *InitTaskTransactionParams) UpdateFromTransaction(tx Transaction) {
	switch tx.(type) {
	case *InitTaskTransaction:
		p.txs = append(p.txs, tx.(*InitTaskTransaction))
		p.callDatas = append(p.callDatas, tx.CallData())
		// calldata := tx.CallData()
		// p.signs = append(p.signs, utils.StringToBytes32(calldata["Sign"].(string)))
		// p.sizes = append(p.sizes, new(big.Int).SetUint64(uint64(calldata["Size"].(int32))))
		// p.models = append(p.models, utils.StringToBytes32(ModelTypeToString(calldata["Model"].(SupportModelType))))
		// isRel := calldata["IsReliable"].(bool)
		// reliableStr := "false"
		// if isRel {
		// 	reliableStr = "true"
		// }
		// p.isReliables = append(p.isReliables, utils.StringToBytes32(reliableStr))
		// p.paramsList = append(p.paramsList, utils.SerializeParams(calldata["Params"].(map[string]interface{})))

	default:
		e := Error(RuntimeError, "TaskProcessTransactionParams should be updated from TaskProcessTransaction!!!")
		panic(e.Error())

	}
}

//	func (p *InitTaskTransactionParams) getSigns() [][32]byte {
//		return [][32]byte{}
//	}
func (p *InitTaskTransactionParams) GetParams() []interface{} {
	// return []interface{}{}
	return []interface{}{p.txs, p.callDatas}
}
func (p *InitTaskTransactionParams) ParamsLen() int {
	return 2
}
func (p *InitTaskTransactionParams) IsEmpty() bool {
	return len(p.txs) == 0
}

func (p *InitTaskTransactionParams) ParamsToKVPairs() ([][32]byte, [][32]byte) {
	n := len(p.txs)
	keys := make([][32]byte, n)
	values := make([][32]byte, n)
	for i := 0; i < n; i++ {
		callData := p.callDatas[i]
		keys[i] = utils.StringToBytes32(callData["Sign"].(string))
		jsonData, _ := json.Marshal(callData)
		values[i] = sha256.Sum256(jsonData)
		// keys[i] = p.signs[i] // 使用任务sign作为key
		// var composite []byte
		// composite = append(composite, p.signs[i][:]...)
		// composite = append(composite, utils.BigIntToBytes32(p.sizes[i])...)
		// composite = append(composite, p.models[i][:]...)
		// composite = append(composite, p.isReliables[i][:]...)
		// composite = append(composite, p.paramsList[i][:]...)
		// values[i] = composite
	}
	return keys, values
}
func (p *InitTaskTransactionParams) BuildDevTransactions(receipts []*Receipt, blockHash string) []*PackedTransaction {
	receipt := receipts[0]
	result := make([]*PackedTransaction, 0)
	for _, tx := range p.txs {
		result = append(result, NewPackedTransaction(tx, receipt, blockHash))
	}
	return result
}
func NewInitTaskTransactionParams() *InitTaskTransactionParams {
	return &InitTaskTransactionParams{
		txs:       make([]*InitTaskTransaction, 0),
		callDatas: make([]map[string]interface{}, 0),
		// signs: make([][32]byte, 0),
		// sizes:       make([]*big.Int, 0),
		// models:      make([][32]byte, 0),
		// isReliables: make([][32]byte, 0),
		// paramsList:  make([][32]byte, 0),
	}
}

type TaskProcessTransactionParams struct {
	// TODO @XQ
	// signs, hashs                                           [][32]byte
	// slotsBigInt, processesBigInt, nidsBigInt, epochsBigInt []*big.Int
	txs       []*TaskProcessTransaction
	callDatas []map[string]interface{}
}

func (p *TaskProcessTransactionParams) UpdateFromTransaction(tx Transaction) {
	switch tx.(type) {
	case *TaskProcessTransaction:
		p.txs = append(p.txs, tx.(*TaskProcessTransaction))
		p.callDatas = append(p.callDatas, tx.CallData())
		// calldata := tx.CallData()
		// p.signs = append(p.signs, utils.StringToBytes32(calldata["Sign"].(string)))
		// p.hashs = append(p.hashs, utils.StringToBytes32(calldata["Hash"].(string)))
		// p.slotsBigInt = append(p.slotsBigInt, new(big.Int).SetUint64(uint64(calldata["Slot"].(int32))))
		// p.processesBigInt = append(p.processesBigInt, new(big.Int).SetUint64(uint64(calldata["Process"].(int32))))
		// // p.nidsBigInt = append(p.nidsBigInt, new(big.Int).SetUint64(uint64(calldata["ID"].(int32))))
		// var idValue int32 = 0
		// if calldata["ID"] != nil {
		// 	idValue = calldata["ID"].(int32)
		// }
		// p.nidsBigInt = append(p.nidsBigInt, new(big.Int).SetUint64(uint64(idValue)))
		// p.epochsBigInt = append(p.epochsBigInt, new(big.Int).SetUint64(uint64(calldata["Epoch"].(int32))))
	default:
		e := Error(RuntimeError, "TaskProcessTransactionParams should be updated from TaskProcessTransaction!!!")
		panic(e.Error())

	}
}
func (p *TaskProcessTransactionParams) GetParams() []interface{} {
	return []interface{}{p.callDatas}
}
func (p *TaskProcessTransactionParams) ParamsLen() int {
	return 1
}
func (p *TaskProcessTransactionParams) IsEmpty() bool {
	return len(p.txs) == 0
}

func (p *TaskProcessTransactionParams) ParamsToKVPairs() ([][32]byte, [][32]byte) {
	n := len(p.txs)
	keys := make([][32]byte, n)
	values := make([][32]byte, n)
	for i := 0; i < n; i++ {
		callData := p.callDatas[i]
		keys[i] = utils.StringToBytes32(callData["Hash"].(SlotHash))
		jsonData, _ := json.Marshal(callData)
		values[i] = sha256.Sum256(jsonData)
		// keys[i] = p.hashs[i] // 使用hash作为key
		// var composite []byte
		// // 直接拼接各个[32]byte字段
		// composite = append(composite, p.signs[i][:]...)
		// composite = append(composite, p.hashs[i][:]...)
		// composite = append(composite, utils.BigIntToBytes32(p.slotsBigInt[i])...)
		// composite = append(composite, utils.BigIntToBytes32(p.processesBigInt[i])...)
		// composite = append(composite, utils.BigIntToBytes32(p.nidsBigInt[i])...)
		// composite = append(composite, utils.BigIntToBytes32(p.epochsBigInt[i])...)
		// values[i] = composite
	}
	return keys, values
}

func (p *TaskProcessTransactionParams) BuildDevTransactions(receipts []*Receipt, blockHash string) []*PackedTransaction {
	// todo 这里暂时先为多个receipt做好准备
	// 要判断receipt和transaction长度 todo

	// 现在只有一个receipt
	receipt := receipts[0]
	result := make([]*PackedTransaction, 0)
	for _, tx := range p.txs {
		result = append(result, NewPackedTransaction(tx, receipt, blockHash))
	}
	return result
}
func NewTaskProcessTransactionParams() *TaskProcessTransactionParams {
	return &TaskProcessTransactionParams{
		// signs:           make([][32]byte, 0),
		// hashs:           make([][32]byte, 0),
		// slotsBigInt:     make([]*big.Int, 0),
		// processesBigInt: make([]*big.Int, 0),
		// nidsBigInt:      make([]*big.Int, 0),
		// epochsBigInt:    make([]*big.Int, 0),
		txs:       make([]*TaskProcessTransaction, 0),
		callDatas: make([]map[string]interface{}, 0),
	}
}

type EpochRecordTransactionParams struct {
	// TODO @XQ
	txs       []*EpochRecordTransaction
	callDatas []map[string]interface{}
	// ids        []*big.Int
	// justifieds [][][32]byte
	// commits    [][][32]byte
	// invalids   []map[string]interface{}
}

func (p *EpochRecordTransactionParams) UpdateFromTransaction(tx Transaction) {
	switch tx.(type) {
	case *EpochRecordTransaction:
		p.txs = append(p.txs, tx.(*EpochRecordTransaction))
		p.callDatas = append(p.callDatas, tx.CallData())
		// calldata := tx.CallData()
		// p.ids = append(p.ids, new(big.Int).SetUint64(uint64(calldata["id"].(int32))))
		// // 处理Justified (带类型安全检查和默认值)
		// var justifiedsArr [][32]byte
		// if justified, ok := calldata["Justified"].([]SlotHash); ok {
		// 	for _, s := range justified {
		// 		justifiedsArr = append(justifiedsArr, utils.StringToBytes32(s))
		// 	}
		// }
		// p.justifieds = append(p.justifieds, justifiedsArr)
		// // 处理Commits (同上)
		// var commitArr [][32]byte
		// if commits, ok := calldata["Commits"].([]SlotHash); ok {
		// 	for _, s := range commits {
		// 		commitArr = append(commitArr, utils.StringToBytes32(s))
		// 	}
		// }
		// p.commits = append(p.commits, commitArr)
		// // 异常slot处理
		// var invalidMap = make(map[string]interface{})
		// for k, v := range calldata["invalids"].(map[SlotHash]InvalidCommitType) {
		// 	invalidMap[k] = v
		// }
		// p.invalids = append(p.invalids, invalidMap)

	default:
		e := Error(RuntimeError, "TaskProcessTransactionParams should be updated from TaskProcessTransaction!!!")
		panic(e.Error())

	}
}
func (p *EpochRecordTransactionParams) GetParams() []interface{} {
	return []interface{}{p.callDatas}
}
func (p *EpochRecordTransactionParams) ParamsLen() int {
	return 1
}
func (p *EpochRecordTransactionParams) IsEmpty() bool {
	return len(p.txs) == 0
}
func (p *EpochRecordTransactionParams) ParamsToKVPairs() ([][32]byte, [][32]byte) {
	n := len(p.txs)
	// LogWriter.Log("DEBUG", fmt.Sprintf("EPOCH PARAMS: %s", p.GetParams()...))
	// 会出现某个epoch没有任何新提交的情况，但是epoch仍需要上链
	keys := make([][32]byte, n)
	values := make([][32]byte, n)
	for i := 0; i < n; i++ {
		callData := p.callDatas[i]
		keys[i] = utils.StringToBytes32(fmt.Sprintf("EPOCH_%d", callData["id"].(int32)))
		jsonData, _ := json.Marshal(callData)
		values[i] = sha256.Sum256(jsonData)
		// keys[i] = utils.StringToBytes32(fmt.Sprintf("EPOCH_%d", p.ids[i])) // 使用epochid作为key
		// var composite []byte
		// // 直接拼接各个[32]byte字段
		// composite = append(composite, utils.FlattenByte32Slice(p.justifieds[i])...)
		// composite = append(composite, utils.FlattenByte32Slice(p.commits[i])...)
		// invalids := utils.SerializeParams(p.invalids[i])
		// composite = append(composite, invalids[:]...)
		// values[i] = composite
	}
	return keys, values
}
func (p *EpochRecordTransactionParams) BuildDevTransactions(receipts []*Receipt, blockHash string) []*PackedTransaction {
	receipt := receipts[0]
	result := make([]*PackedTransaction, 0)
	for _, tx := range p.txs {
		result = append(result, NewPackedTransaction(tx, receipt, blockHash))
	}
	return result
}
func NewEpochRecordTransactionParams() *EpochRecordTransactionParams {
	return &EpochRecordTransactionParams{
		txs:       make([]*EpochRecordTransaction, 0),
		callDatas: make([]map[string]interface{}, 0),
		// ids:        make([]*big.Int, 0),
		// justifieds: make([][][32]byte, 0),
		// commits:    make([][][32]byte, 0),
		// invalids:   make([]map[string]interface{}, 0),
	}
}

func NewParamsMap() map[TransactionType]PackedParams {
	params := make(map[TransactionType]PackedParams)
	params[INIT_TASK_TRANSACTION] = NewInitTaskTransactionParams()
	params[TASK_PROCESS_TRANSACTION] = NewTaskProcessTransactionParams()
	params[EPOCH_RECORD_TRANSACTION] = NewEpochRecordTransactionParams()
	return params
}
