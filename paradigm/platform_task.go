package paradigm

import (
	"time"
)

// PlatformTask 表示一个大的复合任务（平台任务），可能包含多个子任务
type PlatformTask struct {
	ID             string    `gorm:"primaryKey;type:varchar(256)" json:"id"`
	TaskName       string    `gorm:"type:varchar(256)" json:"taskName"`
	Parameters     string    `gorm:"type:text" json:"parameters"`
	ExecutionType  string    `gorm:"type:varchar(128)" json:"executionType"`
	UserName       string    `gorm:"column:user_name;type:varchar(128)" json:"user"`
	Status         string    `gorm:"type:varchar(64)" json:"status"`
	CompletionTime string    `gorm:"type:varchar(128)" json:"completionTime"`
	IsScheduled    bool      `gorm:"type:tinyint;default:0" json:"isScheduled"`
	SubTasks       []Task    `gorm:"foreignKey:PlatformTaskID" json:"subTasks"` // 一对多关系
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
}
