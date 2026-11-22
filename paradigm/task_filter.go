package paradigm

import "time"

// SynthTaskFilters 的非空条件以 AND 组合；结束日期使用次日零点的开区间。
type SynthTaskFilters struct {
	TaskID    string
	TaskName  string
	TxHash    string
	Status    *SlotStatus
	Model     *SupportModelType
	StartDate *time.Time
	EndDate   *time.Time
}
