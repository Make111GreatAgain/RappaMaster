package test

import (
	"BHLayer2Node/Network/HTTP/abm"
	"BHLayer2Node/paradigm"
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

	tasks, err := abm.BuildScheduledV2RawTasks(nil, abm.ScheduledV2Options{})
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

	tasks, err := abm.BuildScheduledV2RawTasks(nil, abm.ScheduledV2Options{})
	if err != nil {
		t.Fatalf("build scheduled raw tasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0]["stockCode"] != "600000" {
		t.Fatalf("expected only available stock to be scheduled, got %#v", tasks)
	}
}

func TestBuildScheduledABMV2RawTasksUsesQueryDataWindow(t *testing.T) {
	paramsDir := t.TempDir()
	dataDir := t.TempDir()
	writeScheduledABMParamFile(t, paramsDir, "600000")
	t.Setenv("ABM_STOCK_PARAM_DIR", paramsDir)
	t.Setenv("ABM_STOCK_DATA_DIR", dataDir)
	abm.InitABMParameterIndex(testABMParametersBase(), nil)
	checker := &fakeStockDataChecker{}
	restore := abm.SetStockDataAvailabilityCheckerForTest(checker)
	defer restore()

	tasks, err := abm.BuildScheduledV2RawTasks(nil, abm.ScheduledV2Options{
		DataStartDate: "2026.01.05",
		DataEndDate:   "2026-01-09",
	})
	if err != nil {
		t.Fatalf("build scheduled raw tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected one task, got %#v", tasks)
	}
	if tasks[0]["dataStartDate"] != "2026-01-05" || tasks[0]["dataEndDate"] != "2026-01-09" {
		t.Fatalf("unexpected scheduled data window: %#v", tasks[0])
	}
	if len(checker.calls) != 1 || checker.calls[0].Window.StartDate != "2026-01-05" || checker.calls[0].Window.EndDate != "2026-01-09" {
		t.Fatalf("availability checker should receive query data window, got %#v", checker.calls)
	}
}

func TestBuildScheduledABMV2RawTasksFiltersHS300Universe(t *testing.T) {
	paramsDir := t.TempDir()
	dataDir := t.TempDir()
	universeRoot := t.TempDir()
	writeScheduledABMParamFile(t, paramsDir, "600000")
	writeScheduledABMParamFile(t, paramsDir, "600001")
	writeScheduledABMParamFile(t, paramsDir, "000001")
	writeUniverseSnapshot(t, universeRoot, "hs300", "2022Q1", map[string]string{
		"600000": "浦发银行",
		"000001": "平安银行",
	})
	t.Setenv("ABM_STOCK_PARAM_DIR", paramsDir)
	t.Setenv("ABM_STOCK_DATA_DIR", dataDir)
	abm.InitABMParameterIndex(testABMParametersBase(), nil)
	restore := abm.SetStockDataAvailabilityCheckerForTest(&fakeStockDataChecker{})
	defer restore()

	tasks, err := abm.BuildScheduledV2RawTasks(&paradigm.BHLayer2NodeConfig{
		ABMUniverseSnapshotRoot: universeRoot,
	}, abm.ScheduledV2Options{
		Universe:    "hs300",
		DataEndDate: "2021-12-31",
	})
	if err != nil {
		t.Fatalf("build scheduled raw tasks: %v", err)
	}
	if len(tasks) != 2 || tasks[0]["stockCode"] != "000001" || tasks[1]["stockCode"] != "600000" {
		t.Fatalf("expected hs300 snapshot intersection, got %#v", tasks)
	}
}

func TestBuildScheduledABMV2RawTasksFiltersCSI1000Universe(t *testing.T) {
	paramsDir := t.TempDir()
	dataDir := t.TempDir()
	universeRoot := t.TempDir()
	writeScheduledABMParamFile(t, paramsDir, "600000")
	writeScheduledABMParamFile(t, paramsDir, "600001")
	writeUniverseSnapshot(t, universeRoot, "csi1000", "2022Q1", map[string]string{
		"600001": "邯郸钢铁",
	})
	t.Setenv("ABM_STOCK_PARAM_DIR", paramsDir)
	t.Setenv("ABM_STOCK_DATA_DIR", dataDir)
	abm.InitABMParameterIndex(testABMParametersBase(), nil)
	restore := abm.SetStockDataAvailabilityCheckerForTest(&fakeStockDataChecker{})
	defer restore()

	tasks, err := abm.BuildScheduledV2RawTasks(&paradigm.BHLayer2NodeConfig{
		ABMUniverseSnapshotRoot: universeRoot,
	}, abm.ScheduledV2Options{
		Universe:    "csi1000",
		DataEndDate: "2022-02-01",
	})
	if err != nil {
		t.Fatalf("build scheduled raw tasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0]["stockCode"] != "600001" {
		t.Fatalf("expected csi1000 snapshot intersection, got %#v", tasks)
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

func TestBuildScheduledV2OptionsFromQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(nil)
	ctx.Request = httptest.NewRequest("POST", "/simulation/create-task?isScheduled=true&universe=hs300&dataStartDate=2026-01-05&dataEndDate=2026-01-09", nil)
	options, err := abm.BuildScheduledV2OptionsFromQuery(ctx)
	if err != nil {
		t.Fatalf("build scheduled options: %v", err)
	}
	if options.Universe != abm.ScheduledUniverseHS300 || options.DataStartDate != "2026-01-05" || options.DataEndDate != "2026-01-09" {
		t.Fatalf("unexpected scheduled options: %#v", options)
	}

	ctx, _ = gin.CreateTestContext(nil)
	ctx.Request = httptest.NewRequest("POST", "/simulation/create-task?isScheduled=true&universe=bad", nil)
	if _, err := abm.BuildScheduledV2OptionsFromQuery(ctx); err == nil {
		t.Fatalf("expected invalid universe error")
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

func writeUniverseSnapshot(t *testing.T, root string, universe string, quarter string, stocks map[string]string) {
	t.Helper()
	dir := filepath.Join(root, universe)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create universe dir: %v", err)
	}
	content := "stockCode,stockName\n"
	for code, name := range stocks {
		content += code + "," + name + "\n"
	}
	if err := os.WriteFile(filepath.Join(dir, quarter+".csv"), []byte(content), 0o644); err != nil {
		t.Fatalf("write universe snapshot: %v", err)
	}
}
