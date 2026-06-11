package paradigm

import (
	"fmt"

	"time"
)

// =============== 以下是Reference部分==============

type RFType int

const (
	InitTaskTx = iota
	EpochTx
	SlotTX
)

// DevReference 指代一个txMap得到的结果
//type DevReference struct {
//	TxHash      string
//	TxReceipt   Receipt // 以上是交易信息
//	TxBlockHash string
//	Rf          RFType // 类型
//	// 如果是InitTask，那么就是一个交易->TaskID，没有额外信息
//	// 如果是EpochTx，那么就是一个交易->EpochID，没有额外信息
//	// 如是果SlotTx, 那么需要包含两类信息
//	// 1. Slot所在epoch; 2. Slot所在Task
//	TaskID      TaskHash
//	EpochID     int32
//	UpchainTime time.Time
//	//ScheduleID ScheduleHash
//}

//	type DevReference struct {
//		TID         int64         `gorm:"primaryKey;autoIncrement"`
//		TxHash      string        `gorm:"type:char(66)"`
//		TxReceipt   Receipt `gorm:"type:json;serializer:json"` // JSON 类型需要数据库支持
//		TxBlockHash string        `gorm:"not null;type:char(66)"`
//		Rf          RFType        `gorm:"not null;type:tinyint"` // 枚举存储为整数类型
//		TaskID      TaskHash      `gorm:"type:varchar(255)"`     // 假设 TaskHash 是字符串类型
//		EpochID     int32         `gorm:"type:int"`
//		UpchainTime time.Time     `gorm:"not null;type:datetime"`
//	}
type DevReference struct {
	TID         int64     `gorm:"primaryKey;autoIncrement"`
	TxHash      string    `gorm:"type:char(66);Index:idx_tx_hash"` // 添加普通索引
	TxReceipt   Receipt   `gorm:"type:json;serializer:json"`
	TxBlockHash string    `gorm:"not null;type:char(66);index:idx_block_hash"`   // 添加普通索引
	Rf          RFType    `gorm:"not null;type:tinyint;index:idx_rf"`            // 添加普通索引
	TaskID      TaskHash  `gorm:"type:varchar(255);index:idx_task_id"`           // 添加普通索引
	EpochID     int32     `gorm:"type:int;index:idx_epoch_id"`                   // 添加普通索引
	UpchainTime time.Time `gorm:"not null;type:datetime;index:idx_upchain_time"` // 添加普通索引
}

// CommitRecord 每个commitRecord对应一个完成finalize的commitSlotItem，对应一笔TaskProcessTransaction
type CommitRecord struct {
	*CommitSlotItem
	TxReceipt *Receipt // 交易回执
	TxID      int64    // 这个交易的id
}

func (r *CommitRecord) Print() {
	fmt.Printf("		CommitSlot Hash: %s\n", r.hash)
	fmt.Printf("		Commit Size: %d\n", r.Process)
	fmt.Printf("		Commit Commitment: %s\n", r.Commitment)
	fmt.Printf("		Commit Epoch: %d\n", r.Epoch)
	fmt.Printf("		Commit Sign: %s\n", r.Sign)
	fmt.Printf("		Commit Slot: %d\n", r.Slot)
	fmt.Printf("		Commit State: %d\n", r.State())
}
func NewCommitRecord(ptx *PackedTransaction) *CommitRecord {
	switch ptx.Tx.(type) {
	case *TaskProcessTransaction:
		return &CommitRecord{
			CommitSlotItem: ptx.Tx.Blob().(*CommitSlotItem),
			TxReceipt:      ptx.Receipt,
		}
	default:
		e := Error(RuntimeError, "A CommitRecord should be init from an TaskProcessTransaction!!!")
		panic(e.Error())
	}
}

// =================== 以下是epoch部分=========================

/*
TODO: EpochID在每次系统启动时自动获取上次运行的最后一个epoch编号
数据一致性：

	1.Epoch编号必须通过数据库生成/修改，确保slot等其他数据里的epochid也一致
	2.崩溃恢复后数据一致
*/
type DevEpoch struct {
	// TODO 这里需要根据根据任务类型去分类，要不然前端这边就没办法判断出来了
	EpochID     int32                           `gorm:"primaryKey;autoIncrement:false"` // 明确指定为主键且禁用自增
	Process     map[SupportModelType]int32      `gorm:"type:json;serializer:json"`
	Commits     map[SupportModelType][]SlotHash `gorm:"type:json;serializer:json"` // Epoch里只存储SlotHash，溯源时从数据库里查Slot By SlotHash
	Justifieds  map[SupportModelType][]SlotHash `gorm:"type:json;serializer:json"`
	Finalizes   map[SupportModelType][]SlotHash `gorm:"type:json;serializer:json"`
	Invalids    []*Slot                         `gorm:"type:json;serializer:json"`
	InitTasks   []*Task                         `gorm:"type:json;serializer:json"`
	TxReceipt   *Receipt                        `gorm:"-"`
	TID         int64                           `gorm:"not null"`
	TxHash      string                          `gorm:"-"`
	TxBlockHash string                          `gorm:"-"`
	CreatedAt   time.Time                       `gorm:"type:timestamp"` // 创建时间
	SlotMap     map[SlotHash]*Slot              `gorm:"-"`              // 临时字段，用于查询时存储需要的Slots完整信息
}

//func NewDevEpoch(ptx *PackedTransaction) *DevEpoch {
//	switch ptx.Tx.(type) {
//	case *EpochRecordTransaction:
//		return &DevEpoch{
//			EpochRecord: ptx.Tx.Blob().(*EpochRecord),
//			TxReceipt:   ptx.Receipt,
//			TxID:        ptx.Id,
//		}
//	default:
//		panic("A DevEpoch should be init from an EpochRecordTransaction!!!")
//	}
//}

/*** EpochRecord 用于记录一个epoch内情况，由TaskManager更新***/

type EpochRecord struct {
	//commits   []*service.JustifiedSlot
	//finalizes []*service.JustifiedSlot
	//invalids  []*service.JustifiedSlot
	Id         int                         // Epoch id
	Commits    map[SlotHash]SlotCommitment // 在这个epoch里commit的slot，目前状态为undetermined, map的内容为commitment
	Justifieds map[SlotHash]SlotCommitment
	Finalizes  map[SlotHash]SlotCommitment     // 在这个epoch里已经确认finalized的，节点在收到这个后可以确认落盘
	Invalids   map[SlotHash]InvalidCommitError // 在这个epoch里被检测出的问题slot, 节点可以根据这个删、改
	Tasks      map[string]int32                // 新收到的任务sign, 对应的数据量
	Process    int32                           // 一共处理了多少
}

func (r *EpochRecord) UpdateTask(task *Task) {
	if _, exist := r.Tasks[task.Sign]; exist {
		e := Error(RuntimeError, "Repeat Epoch Sign!!!")
		panic(e.Error())
	}
	r.Tasks[task.Sign] = task.Size
}
func (r *EpochRecord) Commit(slot *SignedCommitSlotItem) {
	check := func() bool {
		// 这里判断slot的合法性 todo
		if slot.State() == INVALID {
			return false
		}
		return slot.Check() // 除了这个可能还有别的逻辑
	}
	if check() {
		r.Commits[slot.SlotHash()] = slot.Commitment
	}
}
func (r *EpochRecord) Justified(slot *SignedCommitSlotItem) {
	check := func() bool {
		if slot.State() != JUSTIFIED {
			return false
		}
		return true
	}
	if check() {
		//fmt.Println(slot.State(), len(r.Justifieds))
		r.Justifieds[slot.SlotHash()] = slot.Commitment
	} else {
		slot.SetInvalid(UNKNOWN) // TODO
	}
}
func (r *EpochRecord) Finalize(slot *SignedCommitSlotItem) {
	check := func() bool {
		// 这里判断合法性 todo
		if slot.State() != FINALIZE {
			return false
		}
		return true
	}
	if check() {
		//slot.SetFinalize() // finalize
		//r.finalizes = append(r.finalizes, slot.JustifiedSlot)
		r.Finalizes[slot.SlotHash()] = slot.Commitment
		r.Process += slot.Process
	} else {
		slot.SetInvalid(UNKNOWN) // 这里目前没用，甚至不会进入这里 todo

	}
}
func (r *EpochRecord) Abort(slot *SignedCommitSlotItem, reason InvalidCommitType) {
	check := func() bool {
		return true
	}
	if check() {
		slot.SetInvalid(reason)
		r.Invalids[slot.SlotHash()] = NewInvalidCommitError(reason, InvalidCommitTypeToString(reason))
	} else {
		// TODO
	}

}

// TODO 这里临时写成这样用来适配
func (r *EpochRecord) SampleInvalids() map[SlotHash]InvalidCommitType {
	res := make(map[SlotHash]InvalidCommitType)
	for k, v := range r.Invalids {
		res[k] = v.Error
	}
	return res
}

func (r *EpochRecord) Refresh() {
	r.Id++
	r.Process = 0
	r.Commits = make(map[SlotHash]SlotCommitment)
	r.Justifieds = make(map[SlotHash]SlotCommitment)
	r.Finalizes = make(map[SlotHash]SlotCommitment)
	r.Invalids = make(map[SlotHash]InvalidCommitError)
}
func (r *EpochRecord) Echo() {
	Print("EPOCH", fmt.Sprintf("Epoch %d Record, Commits: %d, Justified: %d, Finalized: %d, Invalid: %d", r.Id, len(r.Commits), len(r.Justifieds), len(r.Finalizes), len(r.Invalids)))
	//Print("EPOCH", fmt.Sprintf("	Commits: %v", r.Commits))
	//Print("EPOCH", fmt.Sprintf("	Justifieds: %v", r.Justifieds))
	//Print("EPOCH", fmt.Sprintf("	Finalizeds: %v", r.Finalizes))
	//Print("EPOCH", fmt.Sprintf("	Invalids: %v", r.Invalids))
}

// EpochRecord的epochID也需要从数据库中初始化
func NewEpochRecord(initEpochID int) *EpochRecord {
	return &EpochRecord{
		Id:         initEpochID,
		Commits:    make(map[SlotHash]SlotCommitment),
		Finalizes:  make(map[SlotHash]SlotCommitment),
		Justifieds: make(map[SlotHash]SlotCommitment),
		Invalids:   make(map[SlotHash]InvalidCommitError),
		Tasks:      make(map[string]int32),
	}
}
