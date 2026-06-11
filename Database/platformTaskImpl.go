package Database

import (
	"BHLayer2Node/paradigm"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// SetPlatformTask 创建或更新大任务
func (o DatabaseService) SetPlatformTask(task *paradigm.PlatformTask) error {
	return o.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("SubTasks").Save(task).Error; err != nil {
			return err
		}

		for i := range task.SubTasks {
			task.SubTasks[i].PlatformTaskID = &task.ID
			if err := tx.Omit("EndTime").Save(&task.SubTasks[i]).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetAllPlatformTasks 获取所有大任务, 按创建时间倒序
func (o DatabaseService) GetAllPlatformTasks() ([]*paradigm.PlatformTask, error) {
	var tasks []*paradigm.PlatformTask
	err := o.db.Preload("SubTasks").Order("created_at DESC").Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetPlatformTaskByID 通过ID获取大任务
func (o DatabaseService) GetPlatformTaskByID(id string) (*paradigm.PlatformTask, error) {
	var task paradigm.PlatformTask
	err := o.db.Preload("SubTasks").Where("id = ?", id).First(&task).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &task, nil
}

// GetLatestFinishedABMPlatformTask returns the newest completed platform task
// that has at least one completed ABM_V2 stock subtask.
func (o DatabaseService) GetLatestFinishedABMPlatformTask() (*paradigm.PlatformTask, error) {
	var task paradigm.PlatformTask
	err := o.db.Model(&paradigm.PlatformTask{}).
		Joins("JOIN tasks ON tasks.platform_task_id = platform_tasks.id").
		Where("platform_tasks.status = ? AND tasks.status = ? AND tasks.model = ?", "finished", paradigm.Finished, paradigm.ABM_V2).
		Order("platform_tasks.completion_time DESC, platform_tasks.updated_at DESC, platform_tasks.created_at DESC").
		Preload("SubTasks", "status = ? AND model = ?", paradigm.Finished, paradigm.ABM_V2).
		First(&task).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &task, nil
}

// UpdatePlatformTask 更新大任务
func (o DatabaseService) UpdatePlatformTask(task *paradigm.PlatformTask) error {
	return o.db.Model(task).Updates(task).Error
}

// DownUnFinishedPlatformTasks 将 Master 重启前仍在运行的平台任务标记为失败。
// 子任务会由 DownUnFinishedTasks 按原逻辑置 Failed，这里同步维护 platform_tasks 总状态，
// 避免 execution_log 中平台任务长期停留在 running。
func (o DatabaseService) DownUnFinishedPlatformTasks() error {
	now := time.Now().Format(time.RFC3339)
	result := o.db.Model(&paradigm.PlatformTask{}).
		Where("status = ?", "running").
		Updates(map[string]interface{}{
			"status":          "failed",
			"completion_time": now,
		})
	if result.Error != nil {
		return fmt.Errorf("failed to reset running platform tasks: %w", result.Error)
	}

	paradigm.Log("INFO", fmt.Sprintf("Reset %d running platform tasks to failed status", result.RowsAffected))
	return nil
}

// RefreshPlatformTaskStatus 根据子任务状态刷新平台任务总状态，便于 execution_log 直接展示。
func (o DatabaseService) RefreshPlatformTaskStatus(platformTaskID string) error {
	task, err := o.GetPlatformTaskByID(platformTaskID)
	if err != nil {
		return err
	}
	if task == nil {
		return nil
	}

	allFinished := len(task.SubTasks) > 0
	hasFailed := false
	for _, subTask := range task.SubTasks {
		switch subTask.Status {
		case paradigm.Finished:
			continue
		case paradigm.Failed:
			hasFailed = true
			allFinished = false
		default:
			allFinished = false
		}
	}

	switch {
	case hasFailed:
		task.Status = "failed"
		task.CompletionTime = time.Now().Format(time.RFC3339)
	case allFinished:
		task.Status = "finished"
		task.CompletionTime = time.Now().Format(time.RFC3339)
	default:
		task.Status = "running"
		task.CompletionTime = ""
	}

	return o.db.Model(task).Select("status", "completion_time").Updates(task).Error
}

// GetNextPlatformTaskID 获取下一个可用的 TSK ID
func (o DatabaseService) GetNextPlatformTaskID() (string, error) {
	var count int64
	err := o.db.Model(&paradigm.PlatformTask{}).Count(&count).Error
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("TSK-%04d", 1001+count), nil
}

// GetFinishedTasks 获取所有已完成的任务，用于分析股票列表
func (o DatabaseService) GetFinishedTasks() ([]*paradigm.Task, error) {
	var tasks []*paradigm.Task
	// 这里只返回平台分析任务。
	// 老的普通合成任务虽然也可能是 finished，但它们没有平台任务ID，也没有 stockCode/stockName，
	// 不能混入 analyzed-stocks 的返回结果里。
	err := o.db.Where("status = ? AND platform_task_id IS NOT NULL AND model = ?", paradigm.Finished, paradigm.ABM_V2).Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}
