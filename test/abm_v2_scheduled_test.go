package test

import (
	"BHLayer2Node/Network/HTTP/abm"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBuildScheduledABMV2RawTasksUsesSupportedStockIntersection(t *testing.T) {
	paramsDir := t.TempDir()
	dataDir := t.TempDir()
	writeScheduledABMParamFile(t, paramsDir, "600000")
	writeScheduledABMParamFile(t, paramsDir, "000001")
	writeScheduledABMParamFile(t, paramsDir, "600016")
	writeScheduledABMDataFile(t, dataDir, "600000")
	writeScheduledABMDataFile(t, dataDir, "000001")
	t.Setenv("ABM_STOCK_PARAM_DIR", paramsDir)
	t.Setenv("ABM_STOCK_DATA_DIR", dataDir)

	tasks, err := abm.BuildScheduledV2RawTasks(nil)
	if err != nil {
		t.Fatalf("build scheduled raw tasks: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 scheduled tasks, got %d", len(tasks))
	}

	if tasks[0]["stockCode"] != "000001" || tasks[1]["stockCode"] != "600000" {
		t.Fatalf("scheduled stock codes should be sorted and normalized, got %#v", tasks)
	}
	for _, task := range tasks {
		if task["horizon"] != abm.ScheduledV2Horizon {
			t.Fatalf("scheduled horizon should be fixed to %s, got %#v", abm.ScheduledV2Horizon, task["horizon"])
		}
		if _, exists := task["N_FT"]; exists {
			t.Fatalf("scheduled task should not carry request override params: %#v", task)
		}
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

func writeScheduledABMDataFile(t *testing.T, root string, stockCode string) {
	t.Helper()
	content := []byte("date,close\n2022-01-04 09:30:00,1.0\n")
	if err := os.WriteFile(filepath.Join(root, stockCode+".csv"), content, 0o644); err != nil {
		t.Fatalf("write stock data csv: %v", err)
	}
}
