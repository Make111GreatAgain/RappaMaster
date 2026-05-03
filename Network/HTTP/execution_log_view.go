package HTTP

import (
	"BHLayer2Node/paradigm"
	"time"
)

type ExecutionLogTaskView struct {
	ID             string                `json:"id"`
	TaskName       string                `json:"taskName"`
	Parameters     string                `json:"parameters"`
	ExecutionType  string                `json:"executionType"`
	Status         string                `json:"status"`
	CompletionTime string                `json:"completionTime"`
	IsScheduled    bool                  `json:"isScheduled"`
	CreatedAt      string                `json:"createdAt"`
	UpdatedAt      string                `json:"updatedAt"`
	SubTasks       []ExecutionLogSubTask `json:"subTasks"`
}

type ExecutionLogSubTask struct {
	Sign           string                 `json:"sign"`
	Name           string                 `json:"name"`
	Slot           int32                  `json:"slot"`
	Model          string                 `json:"model"`
	Params         map[string]interface{} `json:"params"`
	StartTime      string                 `json:"startTime"`
	EndTime        string                 `json:"endTime"`
	Status         string                 `json:"status"`
	PlatformTaskID string                 `json:"platformTaskId,omitempty"`
}

func buildExecutionLogViews(tasks []*paradigm.PlatformTask) []ExecutionLogTaskView {
	result := make([]ExecutionLogTaskView, 0, len(tasks))
	for _, task := range tasks {
		if task == nil {
			continue
		}
		result = append(result, buildExecutionLogView(task))
	}
	return result
}

func buildExecutionLogView(task *paradigm.PlatformTask) ExecutionLogTaskView {
	subTasks := make([]ExecutionLogSubTask, 0, len(task.SubTasks))
	for _, subTask := range task.SubTasks {
		subTasks = append(subTasks, buildExecutionLogSubTask(subTask))
	}

	return ExecutionLogTaskView{
		ID:             task.ID,
		TaskName:       task.TaskName,
		Parameters:     task.Parameters,
		ExecutionType:  task.ExecutionType,
		Status:         task.Status,
		CompletionTime: task.CompletionTime,
		IsScheduled:    task.IsScheduled,
		CreatedAt:      formatExecutionLogTime(task.CreatedAt),
		UpdatedAt:      formatExecutionLogTime(task.UpdatedAt),
		SubTasks:       subTasks,
	}
}

func BuildExecutionLogView(task *paradigm.PlatformTask) ExecutionLogTaskView {
	return buildExecutionLogView(task)
}

func buildExecutionLogSubTask(task paradigm.Task) ExecutionLogSubTask {
	platformTaskID := ""
	if task.PlatformTaskID != nil {
		platformTaskID = *task.PlatformTaskID
	}
	return ExecutionLogSubTask{
		Sign:           task.Sign,
		Name:           task.Name,
		Slot:           task.Slot,
		Model:          safeModelTypeName(task.Model),
		Params:         sanitizeExecutionLogParams(task.Params),
		StartTime:      formatExecutionLogTime(task.StartTime),
		EndTime:        formatExecutionLogTime(task.EndTime),
		Status:         slotStatusName(task.Status),
		PlatformTaskID: platformTaskID,
	}
}

func sanitizeExecutionLogParams(params map[string]interface{}) map[string]interface{} {
	if params == nil {
		return nil
	}
	sanitized := make(map[string]interface{}, len(params))
	for key, value := range params {
		if nested, ok := value.(map[string]interface{}); ok {
			sanitized[key] = sanitizeExecutionLogParams(nested)
			continue
		}
		sanitized[key] = value
	}

	if abmParams, ok := sanitized["abm"].(map[string]interface{}); ok {
		if _, exists := abmParams["model_params_root"]; exists {
			delete(abmParams, "model_params_root")
			abmParams["modelParamsSource"] = "offline"
		}
	}
	return sanitized
}

func formatExecutionLogTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339)
}

func safeModelTypeName(model paradigm.SupportModelType) (name string) {
	defer func() {
		if recover() != nil {
			name = "UNKNOWN"
		}
	}()
	return paradigm.ModelTypeToString(model)
}

func slotStatusName(status paradigm.SlotStatus) string {
	switch status {
	case paradigm.Finished:
		return "finished"
	case paradigm.Processing:
		return "processing"
	case paradigm.Failed:
		return "failed"
	default:
		return "unknown"
	}
}
