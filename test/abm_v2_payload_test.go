package test

import (
	abmhttp "BHLayer2Node/Network/HTTP/abm"
	"BHLayer2Node/paradigm"
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/types/known/structpb"
)

func TestBuildABMV2TaskParamsProducesProtoStructCompatibleParams(t *testing.T) {
	dataDir := t.TempDir()
	paramsDir := t.TempDir()
	writeABMParameterDataFile(t, dataDir, "600000")
	writeABMV2PayloadParamFile(t, paramsDir, "600000")
	t.Setenv("ABM_STOCK_DATA_DIR", dataDir)
	t.Setenv("ABM_STOCK_PARAM_DIR", paramsDir)
	abmhttp.InitABMParameterIndex(testABMParametersBase(), nil)

	params, err := abmhttp.BuildV2TaskParams(map[string]interface{}{
		"stockCode": "600000",
		"stockName": "浦发银行",
		"horizon":   "1天 (T+1)",
	}, 0)
	if err != nil {
		t.Fatalf("build ABM_V2 params: %v", err)
	}

	predict, ok := params["predict"].(map[string]interface{})
	if !ok {
		t.Fatalf("predict params should be a map, got %#v", params["predict"])
	}
	if predict["method"] != "kalman_rw" {
		t.Fatalf("expected default predict method kalman_rw, got %#v", predict["method"])
	}

	levels, ok := predict["risk_drop_levels"].([]interface{})
	if !ok {
		t.Fatalf("risk_drop_levels must be []interface{} for structpb, got %T", predict["risk_drop_levels"])
	}
	if len(levels) != 1 || levels[0] != 0.05 {
		t.Fatalf("unexpected risk_drop_levels: %#v", levels)
	}

	abmCfg, ok := params["abm"].(map[string]interface{})
	if !ok {
		t.Fatalf("abm params should be a map, got %#v", params["abm"])
	}
	if abmCfg["model_params_root"] != abmhttp.StockParamDir(nil) {
		t.Fatalf("unexpected model_params_root: %#v", abmCfg["model_params_root"])
	}
	if abmCfg["mode"] != "auto" {
		t.Fatalf("expected default abm mode auto, got %#v", abmCfg["mode"])
	}
	if params["hasTunedParams"] != true {
		t.Fatalf("expected hasTunedParams=true, got %#v", params["hasTunedParams"])
	}
	evaluation, ok := params["evaluation"].(map[string]interface{})
	if !ok {
		t.Fatalf("evaluation params should be a map, got %#v", params["evaluation"])
	}
	if evaluation["generate_models"] != true {
		t.Fatalf("expected default generate_models=true, got %#v", evaluation["generate_models"])
	}

	if _, err := structpb.NewStruct(params); err != nil {
		t.Fatalf("params should be protobuf Struct compatible: %v", err)
	}
}

func TestBuildABMV2TaskParamsMarksMissingTunedParams(t *testing.T) {
	dataDir := t.TempDir()
	paramsDir := t.TempDir()
	writeABMParameterDataFile(t, dataDir, "600000")
	t.Setenv("ABM_STOCK_DATA_DIR", dataDir)
	t.Setenv("ABM_STOCK_PARAM_DIR", paramsDir)
	abmhttp.InitABMParameterIndex(testABMParametersBase(), nil)

	params, err := abmhttp.BuildV2TaskParams(map[string]interface{}{
		"stockCode": "600000",
		"stockName": "浦发银行",
		"horizon":   "1天 (T+1)",
	}, -1)
	if err != nil {
		t.Fatalf("build ABM_V2 params: %v", err)
	}
	if params["hasTunedParams"] != false {
		t.Fatalf("expected hasTunedParams=false, got %#v", params["hasTunedParams"])
	}
}

func TestHasStockInputFile(t *testing.T) {
	dataDir := t.TempDir()
	paramsDir := t.TempDir()
	t.Setenv("ABM_STOCK_DATA_DIR", dataDir)
	t.Setenv("ABM_STOCK_PARAM_DIR", paramsDir)
	abmhttp.InitABMParameterIndex(testABMParametersBase(), nil)
	if abmhttp.HasStockInputFile("600000", nil) {
		t.Fatalf("stock without offline params should be unsupported")
	}
	if err := os.WriteFile(filepath.Join(dataDir, "600000.csv"), []byte("date,close\n2022-01-04,1.0\n"), 0o644); err != nil {
		t.Fatalf("write stock csv: %v", err)
	}
	abmhttp.InitABMParameterIndex(testABMParametersBase(), nil)
	if abmhttp.HasStockInputFile("600000", nil) {
		t.Fatalf("stock with only input csv should be unsupported")
	}
	writeABMV2PayloadParamFile(t, paramsDir, "600000")
	abmhttp.InitABMParameterIndex(testABMParametersBase(), nil)
	if !abmhttp.HasStockInputFile("600000", nil) {
		t.Fatalf("stock with offline params should be supported")
	}
}

func testABMParametersBase() map[string]interface{} {
	return paradigm.BHLayer2NodeConfig{
		AbmParameters: map[string]interface{}{
			"N_FT": map[string]interface{}{
				"label":   "基本面交易者数量",
				"type":    "int",
				"default": 300,
				"min":     10,
				"max":     8000,
			},
		},
	}.AbmParameters
}

func writeABMV2PayloadParamFile(t *testing.T, root string, stockCode string) {
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

func TestBuildABMV2TaskParamsCanOmitAssignedNode(t *testing.T) {
	params, err := abmhttp.BuildV2TaskParams(map[string]interface{}{
		"stockCode": "600000",
		"stockName": "浦发银行",
		"horizon":   "1天 (T+1)",
	}, -1)
	if err != nil {
		t.Fatalf("build ABM_V2 params: %v", err)
	}
	if _, exists := params["assigned_node_id"]; exists {
		t.Fatalf("assigned_node_id should be omitted for dynamic scheduling, got %#v", params["assigned_node_id"])
	}
}

func TestBuildABMV2TaskParamsCarriesDataWindow(t *testing.T) {
	params, err := abmhttp.BuildV2TaskParams(map[string]interface{}{
		"stockCode":     "600000",
		"stockName":     "浦发银行",
		"horizon":       "1天 (T+1)",
		"dataStartDate": "2026.01.05",
	}, -1)
	if err != nil {
		t.Fatalf("build ABM_V2 params: %v", err)
	}
	if params["dataStartDate"] != "2026-01-05" || params["dataEndDate"] != "2026-01-05" {
		t.Fatalf("unexpected data window: start=%#v end=%#v", params["dataStartDate"], params["dataEndDate"])
	}
}
