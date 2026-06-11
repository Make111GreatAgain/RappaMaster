package paradigm

import (
	"BHLayer2Node/utils"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"time"
)

/*** Transaction相关内容 ***/

type TransactionType int

const (
	INIT_TASK_TRANSACTION = iota
	TASK_PROCESS_TRANSACTION
	EPOCH_RECORD_TRANSACTION
)

// TODO @YZM 这里的CallData可以改成遍历结构体中的所有成员变量，然后结合reflect变成map里的内容,这样新定义的transaction无需写CallData()

type Transaction interface {
	Call() string                        // 调用合约哪个函数
	CallData() map[string]interface{}    // 调用合约函数时候的参数
	Blob() interface{}                   // 提供额外的附属内容
	StoreParams() ([]interface{}, error) // 返回上链函数需要的参数列表
}

// InitTaskTransaction 表示RappaMaster新收到了一个合成任务，将任务的一些相关参数传上去，同时在合约里其他一些和任务相关的地方准备新建这个相关的字段
/* 交易类型一: 新建任务交易 */
type InitTaskTransaction struct {
	//Sign       string                 // 合成任务标识
	//Size       int32                  // 合成数据量
	//Model      string                 // 模型,这里可以用某个特殊的函数映射一下也行
	//IsReliable bool                   // 是否需要可信证明，这个一定要上链
	//Params     map[string]interface{} // 合成任务相关参数
	*Task // 这里有很多字段
	// 大概是这样
}

func (t *InitTaskTransaction) Call() string {
	return "InitTask"
}
func (t *InitTaskTransaction) CallData() map[string]interface{} {
	result := make(map[string]interface{})
	result["Sign"] = t.Sign
	result["Name"] = t.Name
	result["Size"] = t.Size
	result["Model"] = t.Model
	result["IsReliable"] = t.IsReliable()
	result["Params"] = t.Params
	return result
}
func (t *InitTaskTransaction) Blob() interface{} {
	return t.Task
}
func (t *InitTaskTransaction) StoreParams() ([]interface{}, error) {
	// 1) sign -> [32]byte
	sign := utils.StringToBytes32(t.Sign)
	// 2) name
	name := t.Name
	// 3) size -> uint32
	size := uint32(t.Size)
	// 4) model -> bytes32
	model := utils.IntToBytes32(int(t.Model))
	// 5) isReliable
	reliable := t.IsReliable()
	// 6) params -> JSON bytes 允许 nil 或 empty map
	var paramsJSON []byte
	if t.Params == nil {
		paramsJSON = []byte("{}")
	} else {
		b, err := json.Marshal(t.Params)
		if err != nil {
			// 这里不直接返回错误，而是降级成 empty JSON
			paramsJSON = []byte("{}")
		} else {
			paramsJSON = b
		}
	}

	return []interface{}{
		sign,
		name,
		size,
		model,
		reliable,
		paramsJSON,
	}, nil
}
func NewInitTaskTransaction(task *Task) *InitTaskTransaction {
	return &InitTaskTransaction{task}
}

// TaskProcessTransaction 表示在RappaMaster中一个commitSlot通过了验证(1. 存储, 2. 可信证明(可选))
/* 交易类型二: 任务进度更新交易 */
// 这里上链以后要维护的是若干个task结构体，这些结构体的状态会不断更新，其中会有一个slots数组，slots根据下面提交的slot和nid更新字段，可能会比较大，后续用于展示
// 我现在想的我们未来的前端里，溯源界面，会有一个EpochChain，也就是每个epoch完成了哪些任务的slot；会有一个Task的搜索，给出TaskSlotChain，也就是每个task的每个slot在哪些epoch里被完成了多少
type TaskProcessTransaction struct {
	*CommitSlotItem                  // 这里有很多的字段
	Proof           Proof            // 这个是可信证明，暂时先不用放到链上，先在链上准备一下相关的字段和函数
	Signatures      [][]byte         // 这里后面我们要加入节点的公私钥，每次投票都会附上自己的签名，我打算在这里签名上链验证的，和上面一样，先留下字段
	Model           SupportModelType // TODO 这里的目的是暂时拿到任务类型给monitor，先这样写
}

func (t *TaskProcessTransaction) Call() string {
	return "Commit" // 简单先写一下，后面具体和合约对齐
}
func (t *TaskProcessTransaction) CallData() map[string]interface{} {
	result := make(map[string]interface{})
	result["Sign"] = t.Sign
	result["Slot"] = t.Slot
	result["Process"] = t.Process
	result["ID"] = t.Nid
	//result["Vote"] = t.Votes
	result["Epoch"] = t.Epoch
	result["Hash"] = t.hash
	result["Commitment"] = t.Commitment
	result["Proof"] = t.Proof
	result["Signatures"] = t.Signatures
	return result
}
func (t *TaskProcessTransaction) Blob() interface{} {
	return t.CommitSlotItem // 将整个commitSlot返回
}
func (t *TaskProcessTransaction) StoreParams() ([]interface{}, error) {
	sign := utils.StringToBytes32(t.Sign)
	var hash [32]byte
	if bs, err := hex.DecodeString(t.hash); err == nil {
		copy(hash[:], bs)
	} else {
		hash = utils.StringToBytes32(t.hash)
	}
	slot, process, id, epoch :=
		uint32(t.Slot), uint32(t.Process), uint32(t.Nid), uint32(t.Epoch)
	var commitment [32]byte
	copy(commitment[:], t.Commitment)
	var proof []byte
	switch p := t.Proof.(type) {
	case nil:
		proof = []byte{}
	case []byte:
		proof = p
	case string:
		if bs, err := hex.DecodeString(p); err == nil {
			proof = bs
		} else {
			proof = []byte(p)
		}
	default:
		// 任意可 JSON 序列化的类型
		if b, err := json.Marshal(p); err == nil {
			proof = b
		} else {
			proof = []byte{}
		}
	}
	signatures := t.Signatures
	if signatures == nil {
		signatures = make([][]byte, 0)
	}

	return []interface{}{
		sign,
		hash,
		slot,
		process,
		id,
		epoch,
		commitment,
		proof,
		signatures,
	}, nil
}

// EpochRecordTransaction
/* 交易类型三：epoch记录交易 */
// 这里每个epoch会给出三个内容:
// 1. 节点提交了哪些slot(commit);
// 2. 哪些slot通过了投票(justified);
// 3. 检测出了哪些不合法的slot
// 后面可能还会加上收到了哪些slot的proof等，先按上面的写
type EpochRecordTransaction struct {
	// 这里所有的slot都是用一个独一无二的hash表示的(目前就是个string也就是sign_slot_nid)，因为我们只需要在链上记录真正落盘的slot就行
	// 其它那些可能在未来检测出不合法的，没必要存，那些落盘的最后会由TaskProcessTransaction给出
	*EpochRecord        // 这里有很多字段
	Id            int32 // epoch id
	CommitsHash   []SlotHash
	JustifiedHash []SlotHash
	Invalids      map[SlotHash]InvalidCommitType //这里不同的异常slot有不同的理由
}

func (t *EpochRecordTransaction) Call() string {
	return "EpochUpdate"
}
func (t *EpochRecordTransaction) CallData() map[string]interface{} {
	result := make(map[string]interface{})
	result["id"] = t.Id
	result["Justified"] = t.JustifiedHash
	result["Commits"] = t.CommitsHash
	result["invalids"] = t.Invalids
	return result
}
func (t *EpochRecordTransaction) Blob() interface{} {
	return t.EpochRecord // 返回epochRecord
}
func (t *EpochRecordTransaction) StoreParams() ([]interface{}, error) {
	// 1) id -> *big.Int
	idBig := big.NewInt(int64(t.Id))

	// 2) justifiedHash -> [][32]byte
	justs := make([][32]byte, len(t.JustifiedHash))
	for i, h := range t.JustifiedHash {
		if bs, err := hex.DecodeString(h); err == nil {
			copy(justs[i][:], bs)
		} else {
			justs[i] = utils.StringToBytes32(h)
		}
	}

	// 3) commitsHash -> [][32]byte
	comms := make([][32]byte, len(t.CommitsHash))
	for i, h := range t.CommitsHash {
		if bs, err := hex.DecodeString(h); err == nil {
			copy(comms[i][:], bs)
		} else {
			comms[i] = utils.StringToBytes32(h)
		}
	}

	// 4) invalids -> 2个切片
	invHashes := make([][32]byte, 0, len(t.Invalids))
	invReasons := make([]uint8, 0, len(t.Invalids))
	for h, r := range t.Invalids {
		var b [32]byte
		if bs, err := hex.DecodeString(h); err == nil {
			copy(b[:], bs)
		} else {
			b = utils.StringToBytes32(h)
		}
		invHashes = append(invHashes, b)
		invReasons = append(invReasons, uint8(r))
	}

	return []interface{}{
		idBig,
		justs,
		comms,
		invHashes,
		invReasons,
	}, nil
}

type PackedTransaction struct {
	Tx          Transaction
	Id          int
	Receipt     *Receipt
	BlockHash   string
	UpchainTime time.Time
}

func (t *PackedTransaction) SetID(id int) {
	t.Id = id
}

func (t *PackedTransaction) SetUpchainTime(time time.Time) {
	t.UpchainTime = time
}

func (t *PackedTransaction) SetBlockInfo(block *Block) {
	t.BlockHash = block.Hash
	t.UpchainTime = TimestampConvert(block.Timestamp)
}

func NewPackedTransaction(tx Transaction, receipt *Receipt, blockHash string) *PackedTransaction {
	return &PackedTransaction{
		Tx:        tx,
		Id:        -1,
		Receipt:   receipt,
		BlockHash: blockHash,
		// MerkleRoot: merkleRoot,
	}
}
