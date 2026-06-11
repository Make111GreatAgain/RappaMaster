package paradigm

import (
	"fmt"
	"strings"
	"time"
)

// Task 描述一个合成任务
//type Task struct {
//	Sign        string
//	Slot        int32
//	Model       SupportModelType
//	Params      map[string]interface{}
//	Size        int32                // 总的数据量
//	Process     int32                // 已经完成的数据量
//	isReliable  bool                 // 是否可信 TODO: @YZM 这里需要加入任务是否可信的部分，这个需要在http前端得到
//	OutputType  ModelOutputType      // 模型类型，这里假定一个任务一个模型
//	Schedules   []*SynthTaskSchedule // 该任务的所有的调度
//	ScheduleMap map[ScheduleHash]int // 为了防止乱序
//	TxID        int
//	TxReceipt   *Receipt
//	// TxBlock     *types.Block
//	TxBlockHash string
//	// 以下是测试字段
//	HasbeenCollect bool
//	StartTime      time.Time
//	EndTime        time.Time
//	Collector      RappaCollector
//	//records    []paradigm.SlotRecord // 记录每个slot的调度和完成情况
//}

type Task struct {
	Sign string `gorm:"primaryKey;type:varchar(256);not null;comment:唯一任务标识符（必填）"`
	Name string `gorm:"type:varchar(256)"`
	// TODO 这里我改成指定slot的大小
	Slot           int32                  `gorm:"type:int;not null"`
	Model          SupportModelType       `gorm:"type:tinyint;not null;comment:支持的模型类型（必填，0=CTGAN, 1=BAED, 2=FINKAN, 3=ABM, 4=ABM_V2）"`
	Params         map[string]interface{} `gorm:"type:json;serializer:json"`
	Size           int32                  `gorm:"type:int;not null;default:0;comment:总数据处理量（必填，默认0）"`
	Process        int32                  `gorm:"type:int;not null;default:0;comment:已完成数据量（必填，默认0）"`
	isReliable     bool                   `gorm:"type:tinyint;not null;default:false;comment:任务可信标记（必填，默认false）"`
	OutputType     ModelOutputType        `gorm:"type:tinyint;not null;comment:模型输出类型（必填，0=DATAFRAME, 1=NETWORK）"`
	Schedules      []*SynthTaskSchedule   `gorm:"type:json;serializer:json"`
	ScheduleMap    map[ScheduleHash]int   `gorm:"type:json;serializer:json"`
	TID            int64                  `gorm:"not null"`
	TxHash         string                 `gorm:"-"`
	TxReceipt      *Receipt               `gorm:"-"`
	TxBlockHash    string                 `gorm:"-"`
	HasbeenCollect bool                   `gorm:"-"`
	// 这里本来就有时间
	StartTime      time.Time      `gorm:"type:datetime;not null;comment:任务启动时间戳"`
	EndTime        time.Time      `gorm:"type:datetime;comment:任务结束时间戳"`
	Status         SlotStatus     `gorm:"type:tinyint;not null;comment:完成状态"`
	PlatformTaskID *string        `gorm:"type:varchar(256);index;comment:所属平台任务ID"`
	Collector      RappaCollector `gorm:"-"`
}

func (t *Task) Speed() float64 {
	//if !t.IsFinish() {
	//	return -1 // 还没完成
	//}
	processTime := t.EndTime.Sub(t.StartTime).Seconds()
	if processTime == 0 {
		return -1 // 极短的时间，这里前端要适配，比如--MB/s
	}
	return float64(t.Process) / processTime // 这里用的process
}

func (t *Task) Print() {
	var sb strings.Builder
	sb.WriteString("Task Details:\n")
	sb.WriteString(fmt.Sprintf("Sign: %s\n", t.Sign))
	sb.WriteString(fmt.Sprintf("Name: %s\n", t.Name))
	sb.WriteString(fmt.Sprintf("Slot: %d\n", t.Slot))
	sb.WriteString(fmt.Sprintf("Model: %v\n", t.Model))
	sb.WriteString("Params:\n")
	for key, value := range t.Params {
		sb.WriteString(fmt.Sprintf("  - %s: %v\n", key, value))
	}
	sb.WriteString(fmt.Sprintf("Size: %d\n", t.Size))
	sb.WriteString(fmt.Sprintf("Process: %d\n", t.Process))
	sb.WriteString(fmt.Sprintf("Is Reliable: %t\n", t.isReliable))
	sb.WriteString(fmt.Sprintf("Output Type: %v\n", t.OutputType))
	sb.WriteString("Schedules:\n")
	for _, schedule := range t.Schedules {
		sb.WriteString(fmt.Sprintf("  - Schedule %d: [Size: %d]\n", schedule.ScheduleID, schedule.Size))
		for nodeID, slot := range schedule.Slots {
			sb.WriteString(fmt.Sprintf("    - Slots %s [NodeID: %d], ScheduleSize: %d, Status: %d, Err: %s\n", slot.SlotID, nodeID, slot.ScheduleSize, slot.Status, slot.Err))
		}

		//sb.WriteString(fmt.Sprintf("  - %v\n", schedule))
	}
	sb.WriteString(fmt.Sprintf("TxHash: %s\n", t.TxHash))
	sb.WriteString(fmt.Sprintf("TxReceipt: %v\n", t.TxReceipt))
	sb.WriteString(fmt.Sprintf("TxBlockHash: %v\n", t.TxBlockHash))
	//sb.WriteString(fmt.Sprintf("Has Been Collected: %t\n", t.HasbeenCollect))

	fmt.Println(sb.String())
}
func (t *Task) UpdateTxInfo(ptx *PackedTransaction) {
	switch ptx.Tx.(type) {
	case *InitTaskTransaction:
		t.TxReceipt = ptx.Receipt
		//t.TxID = ptx.Id
		t.TxBlockHash = ptx.BlockHash
		//return &DevTask{
		//	Task:      ptx.Tx.Blob().(*Task),
		//	Slots:     make([]*SlotRecord, 0),
		//	TxID:      ptx.Id,
		//	TxReceipt: ptx.Receipt,
		//}
	default:
		panic("A DevTask should be init from InitTaskTransaction!!!")
	}
}
func (t *Task) InitTrack() *SynthTaskTrackItem {
	unprocessedTask := &UnprocessedTask{
		TaskID:   t.Sign,
		SlotSize: t.Slot,
		Size:     t.Size,
		Model:    t.Model,
		Params:   t.Params,
	}
	return &SynthTaskTrackItem{
		UnprocessedTask: unprocessedTask,
		Total:           t.Size,
		History:         0,
		IsReliable:      t.isReliable,
	}
}

// UpdateSchedule 更新调度情况
func (t *Task) UpdateSchedule(schedule *SynthTaskSchedule) {
	// todo 这里有代码内部问题是没有调试的
	t.ScheduleMap[schedule.ScheduleID] = len(t.Schedules)
	t.Schedules = append(t.Schedules, schedule) // 这里假设的是依次不错不重复
	//return nil
}

func (t *Task) Commit(slot *CommitRecord) error {
	if slot.State() != FINALIZE {
		return fmt.Errorf("the commit Slot is not finalized") // 只能提交finalized的，因为已经通过投票了所以不需要check
	}
	//for len(t.records) <= int(slot.Slot) {
	//	t.records = append(t.records, paradigm.NewSlotRecord(len(t.records)))
	//}
	//slotRecord := t.records[slot.Slot]
	//slotRecord.Process = append(slotRecord.Process, slot.Record())
	t.Process += slot.Process
	t.Collector.ProcessSlotUpdate(CollectSlotItem{
		Sign: slot.Sign,
		Hash: slot.SlotHash(),
		Size: slot.Process,
		//OutputType:  t.OutputType,
		PaddingSize: slot.Padding,
		StoreMethod: slot.Store,
	})
	//Print("INFO", fmt.Sprintf("Task %s Process %d, Total: %d, Process: %d", slot.Sign, slot.Process, t.Size, t.Process))
	//LogWriter.Log("DEBUG", fmt.Sprintf("Epoch %s process %d by node %d", slot.Sign, slot.Process, slot.Nid))
	//t.records[slot.Slot] = slotRecord
	return nil
}
func (t *Task) GetCollector() RappaCollector {
	return t.Collector
}
func (t *Task) IsReliable() bool {
	return t.isReliable
}

//func (t *Task) Next() (UnprocessedTask, error) {
//	if t.IsFinish() {
//		return UnprocessedTask{}, fmt.Errorf("task %s has been finished", t.Sign)
//	}
//	t.Slot++
//	slot := UnprocessedTask{
//		Sign:   t.Sign,
//		Slot:   t.Slot,
//		Size:   t.Remain(),
//		Model:  t.Model,
//		Params: t.Params,
//	}
//	return slot, nil
//}

func (t *Task) Remain() int32 {
	if t.IsFinish() {
		return 0
	}
	return t.Size - t.Process
}
func (t *Task) IsFinish() bool {
	return t.Process >= t.Size
}
func (t *Task) SetCollected() {
	t.HasbeenCollect = true
}
func (t *Task) SetEndTime() {
	t.EndTime = time.Now()
	t.Status = Finished
}
func (t *Task) SetCollector(c RappaCollector) {
	t.Collector = c
}
func (t *Task) GetDataset() string {
	if dataset, exist := t.Params["dataset"]; exist {
		return dataset.(string)
	} else {
		Error(ValueError, "Dataset is not given in params")
		return ""
	}
}
func NewTask(sign string, name string, model SupportModelType, slotSize int32, params map[string]interface{}, total int32, isReliable bool) *Task {
	outputType := DATAFRAME
	switch model {
	case CTGAN:
		outputType = DATAFRAME
	case BAED:
		outputType = NETWORK
	case FINKAN:
		outputType = DATAFRAME
	case ABM:
		outputType = DATAFRAME
	case ABM_V2:
		outputType = DATAFRAME
	default:
		e := Error(RuntimeError, "Unsupported Model Type!!!")
		panic(e.Error())
	}
	return &Task{
		Sign:        sign,
		Name:        name,
		Slot:        slotSize,
		Model:       model,
		OutputType:  outputType,
		ScheduleMap: make(map[ScheduleHash]int),
		Schedules:   make([]*SynthTaskSchedule, 0),
		Params:      params,
		Size:        total,
		Process:     0,
		Status:      Processing,
		//records:    make([]paradigm.SlotRecord, 0),
		isReliable:     isReliable,
		HasbeenCollect: false,
		StartTime:      time.Now(), // 包括上链时间
	}
}
