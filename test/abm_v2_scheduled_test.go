package test

import (
	"BHLayer2Node/Network/HTTP/abm"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBuildScheduledABMV2RawTasksUsesOfflineParamStocks(t *testing.T) {
	paramsDir := t.TempDir()
	dataDir := t.TempDir()
	writeScheduledABMParamFile(t, paramsDir, "600000")
	writeScheduledABMParamFile(t, paramsDir, "000001")
	writeScheduledABMParamFile(t, paramsDir, "600016")
	t.Setenv("ABM_STOCK_PARAM_DIR", paramsDir)
	t.Setenv("ABM_STOCK_DATA_DIR", dataDir)
	abm.InitABMParameterIndex(testABMParametersBase(), nil)
	restore := abm.SetStockDataAvailabilityCheckerForTest(&fakeStockDataChecker{})
	defer restore()

	tasks, err := abm.BuildScheduledV2RawTasks(nil)
	if err != nil {
		t.Fatalf("build scheduled raw tasks: %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("expected 3 scheduled tasks, got %d", len(tasks))
	}

	if tasks[0]["stockCode"] != "000001" || tasks[1]["stockCode"] != "600000" || tasks[2]["stockCode"] != "600016" {
		t.Fatalf("scheduled stock codes should be sorted and normalized, got %#v", tasks)
	}
	for _, task := range tasks {
		if task["horizon"] != abm.ScheduledV2Horizon {
			t.Fatalf("scheduled horizon should be fixed to %s, got %#v", abm.ScheduledV2Horizon, task["horizon"])
		}
		if _, exists := task["N_FT"]; exists {
			t.Fatalf("scheduled task should not carry request override params: %#v", task)
		}
		if task["dataStartDate"] == "" || task["dataEndDate"] == "" {
			t.Fatalf("scheduled task should carry validated data window: %#v", task)
		}
	}
}

func TestBuildScheduledABMV2RawTasksSkipsUnavailableStockData(t *testing.T) {
	paramsDir := t.TempDir()
	dataDir := t.TempDir()
	writeScheduledABMParamFile(t, paramsDir, "600000")
	writeScheduledABMParamFile(t, paramsDir, "600001")
	t.Setenv("ABM_STOCK_PARAM_DIR", paramsDir)
	t.Setenv("ABM_STOCK_DATA_DIR", dataDir)
	abm.InitABMParameterIndex(testABMParametersBase(), nil)
	restore := abm.SetStockDataAvailabilityCheckerForTest(&fakeStockDataChecker{
		results: map[string]abm.StockDataAvailabilityResult{
			"600001": {
				Available: false,
				Source:    "dolphindb",
				Reason:    "no remote rows",
			},
		},
	})
	defer restore()

	tasks, err := abm.BuildScheduledV2RawTasks(nil)
	if err != nil {
		t.Fatalf("build scheduled raw tasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0]["stockCode"] != "600000" {
		t.Fatalf("expected only available stock to be scheduled, got %#v", tasks)
	}
}

func TestIsScheduledCreateTaskUsesIsScheduledOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(nil)
	ctx.Request = httptest.NewRequest("POST", "/simulation/create-task?isScheduled=true", nil)
	if !abm.IsScheduledCreateTask(ctx) {
		t.Fatalf("expected isScheduled=true to be accepted as scheduled")
	}

	ctx, _ = gin.CreateTestContext(nil)
	ctx.Request = httptest.NewRequest("POST", "/simulation/create-task?isSchdule=true", nil)
	if abm.IsScheduledCreateTask(ctx) {
		t.Fatalf("isSchdule should not be accepted")
	}
}

func writeScheduledABMParamFile(t *testing.T, root string, stockCode string) {
	t.Helper()
	dir := filepath.Join(root, stockCode)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create stock param dir: %v", err)
	}
	content := []byte(`{"structural_params":{"N_FT":30},"calibrated_params":{"K1":1.0}}`)
	if err := os.WriteFile(filepath.Join(dir, "model_params.json"), content, 0o644); err != nil {
		t.Fatalf("write model params: %v", err)
	}
}
