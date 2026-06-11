package test

import (
	"BHLayer2Node/Network/HTTP"
	"BHLayer2Node/paradigm"
	"testing"
	"time"
)

func TestBuildLatestMarketTaskResponsePrefersDefaultStock(t *testing.T) {
	platformTaskID := "TSK-NEWEST"
	view := HTTP.BuildLatestMarketTaskResponse(&paradigm.PlatformTask{
		ID:       platformTaskID,
		TaskName: "平台任务申报",
		SubTasks: []paradigm.Task{
			newLatestMarketSubTask(platformTaskID, "600016", "民生银行", paradigm.Finished, paradigm.ABM_V2),
			newLatestMarketSubTask(platformTaskID, "600000", "浦发银行", paradigm.Finished, paradigm.ABM_V2),
		},
	})

	if view == nil {
		t.Fatal("expected latest market task response")
	}
	if view["taskId"] != platformTaskID {
		t.Fatalf("expected taskId %s, got %#v", platformTaskID, view["taskId"])
	}
	if view["stockCode"] != "600000" || view["stockId"] != "600000" {
		t.Fatalf("expected default stock 600000, got %#v", view)
	}
	if view["taskName"] != "平台任务申报" {
		t.Fatalf("expected platform task name, got %#v", view["taskName"])
	}
	if view["date"] != "2026-05-14" {
		t.Fatalf("expected formatted date, got %#v", view["date"])
	}
}

func TestBuildLatestMarketTaskResponseFallsBackToStockCodeOrder(t *testing.T) {
	platformTaskID := "TSK-NO-DEFAULT"
	view := HTTP.BuildLatestMarketTaskResponse(&paradigm.PlatformTask{
		ID:       platformTaskID,
		TaskName: "平台任务申报",
		SubTasks: []paradigm.Task{
			newLatestMarketSubTask(platformTaskID, "600019", "宝钢股份", paradigm.Finished, paradigm.ABM_V2),
			newLatestMarketSubTask(platformTaskID, "600016", "民生银行", paradigm.Finished, paradigm.ABM_V2),
		},
	})

	if view == nil {
		t.Fatal("expected latest market task response")
	}
	if view["stockCode"] != "600016" {
		t.Fatalf("expected first stock by code order, got %#v", view)
	}
}

func TestBuildLatestMarketTaskResponseIgnoresInvalidSubTasks(t *testing.T) {
	platformTaskID := "TSK-MIXED"
	view := HTTP.BuildLatestMarketTaskResponse(&paradigm.PlatformTask{
		ID: platformTaskID,
		SubTasks: []paradigm.Task{
			newLatestMarketSubTask(platformTaskID, "600000", "浦发银行", paradigm.Processing, paradigm.ABM_V2),
			newLatestMarketSubTask(platformTaskID, "000001", "平安银行", paradigm.Finished, paradigm.CTGAN),
			newLatestMarketSubTask(platformTaskID, "999999", "测试股票", paradigm.Finished, paradigm.ABM_V2),
		},
	})

	if view == nil {
		t.Fatal("expected latest market task response")
	}
	if view["stockCode"] != "999999" {
		t.Fatalf("expected only valid completed ABM_V2 stock, got %#v", view)
	}
	if view["taskName"] != "测试股票 风险监测" {
		t.Fatalf("expected fallback task name, got %#v", view["taskName"])
	}
}

func TestBuildLatestMarketTaskResponseReturnsNilWithoutValidStock(t *testing.T) {
	platformTaskID := "TSK-EMPTY"
	view := HTTP.BuildLatestMarketTaskResponse(&paradigm.PlatformTask{
		ID: platformTaskID,
		SubTasks: []paradigm.Task{
			newLatestMarketSubTask(platformTaskID, "600000", "浦发银行", paradigm.Processing, paradigm.ABM_V2),
		},
	})

	if view != nil {
		t.Fatalf("expected nil response, got %#v", view)
	}
}

func newLatestMarketSubTask(platformTaskID, stockCode, stockName string, status paradigm.SlotStatus, model paradigm.SupportModelType) paradigm.Task {
	return paradigm.Task{
		Sign:  "SubTask-" + platformTaskID + "-" + stockCode,
		Name:  stockName + "推演",
		Model: model,
		Params: map[string]interface{}{
			"stockCode": stockCode,
			"stockName": stockName,
		},
		Status:         status,
		StartTime:      time.Date(2026, 5, 14, 9, 30, 0, 0, time.UTC),
		PlatformTaskID: &platformTaskID,
	}
}
