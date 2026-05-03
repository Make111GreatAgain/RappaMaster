package test

import (
	"BHLayer2Node/Network/HTTP"
	"BHLayer2Node/paradigm"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestExecutionLogViewHidesInternalTaskFields(t *testing.T) {
	platformTaskID := "TSK-TEST"
	view := HTTP.BuildExecutionLogView(&paradigm.PlatformTask{
		ID:            platformTaskID,
		TaskName:      "平台任务申报",
		Parameters:    "Stock: 浦发银行 (600000)",
		ExecutionType: "即时任务",
		Status:        "running",
		IsScheduled:   false,
		CreatedAt:     time.Date(2026, 5, 3, 1, 2, 3, 0, time.UTC),
		SubTasks: []paradigm.Task{
			{
				Sign:  "SubTask-TSK-TEST-600000",
				Name:  "浦发银行推演",
				Slot:  1,
				Model: paradigm.ABM_V2,
				Params: map[string]interface{}{
					"stockCode": "600000",
					"abm": map[string]interface{}{
						"mode":              "auto",
						"model_params_root": "/root/rappa/stockdata/params",
					},
				},
				Size:           1,
				Process:        1,
				OutputType:     paradigm.DATAFRAME,
				TID:            9152,
				TxHash:         "internal-hash",
				TxBlockHash:    "internal-block",
				HasbeenCollect: true,
				Status:         paradigm.Finished,
				StartTime:      time.Date(2026, 5, 3, 1, 2, 3, 0, time.UTC),
				PlatformTaskID: &platformTaskID,
			},
		},
	})

	raw, err := json.Marshal(view)
	if err != nil {
		t.Fatalf("marshal execution log view: %v", err)
	}
	body := string(raw)
	for _, forbidden := range []string{
		"Schedules",
		"ScheduleMap",
		"TID",
		"TxHash",
		"TxReceipt",
		"TxBlockHash",
		"HasbeenCollect",
		"Collector",
		`"model":4`,
		`"modelName"`,
		`"status":0`,
		`"statusName"`,
		`"size"`,
		`"process"`,
		`"outputType"`,
		`"outputTypeName"`,
		`"model_params_root"`,
		"/root/rappa/stockdata/params",
		"internal-hash",
		"internal-block",
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("execution log view leaked internal field/value %q in %s", forbidden, body)
		}
	}

	if !strings.Contains(body, `"model":"ABM_V2"`) {
		t.Fatalf("expected readable model in execution log view, got %s", body)
	}
	if !strings.Contains(body, `"status":"finished"`) {
		t.Fatalf("expected readable status in execution log view, got %s", body)
	}
	if !strings.Contains(body, `"modelParamsSource":"offline"`) {
		t.Fatalf("expected offline model params marker in execution log view, got %s", body)
	}
}
