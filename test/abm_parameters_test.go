package test

import (
	"BHLayer2Node/Network/HTTP"
	abmhttp "BHLayer2Node/Network/HTTP/abm"
	"BHLayer2Node/paradigm"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestABMParameterIndexRemoteModeFiltersParamStocksByUniverseAndDatabase(t *testing.T) {
	dataDir := t.TempDir()
	paramsDir := t.TempDir()
	universeRoot := t.TempDir()
	writeABMParameterParamFile(t, paramsDir, "600000", 123)
	writeABMParameterParamFile(t, paramsDir, "600002", 456)
	writeUniverseSnapshot(t, universeRoot, "hs300", "2026Q1", map[string]string{
		"600000": "浦发银行",
		"600001": "测试银行",
		"600002": "测试股份",
	})

	restore := abmhttp.SetStockDataAvailabilityCheckerForTest(&fakeStockDataChecker{
		results: map[string]abmhttp.StockDataAvailabilityResult{
			"600002": {
				Available: false,
				Source:    "dolphindb",
				Reason:    "no remote rows",
			},
		},
	})
	defer restore()

	index, err := abmhttp.BuildSupportedStockIndex(testABMParametersBase(), &paradigm.BHLayer2NodeConfig{
		ABMStockDataDir:         dataDir,
		ABMStockParamDir:        paramsDir,
		ABMStockDataSource:      "auto",
		ABMUniverseSnapshotRoot: universeRoot,
		ABMParameterUniverse:    "hs300",
	})
	if err != nil {
		t.Fatalf("build index: %v", err)
	}
	if index.Total != 1 || index.TunedCount != 1 {
		t.Fatalf("expected one available tuned stock, got total=%d tuned=%d list=%#v", index.Total, index.TunedCount, index.SupportedStockList)
	}
	meta, ok := index.StockMap["600000"]
	if !ok || !meta.SupportSimulation || !meta.HasTunedParams {
		t.Fatalf("expected 600000 available with tuned params, got %#v", meta)
	}
	if _, exists := index.StockMap["600002"]; exists {
		t.Fatalf("expected unavailable stock filtered out, got %#v", index.StockMap["600002"])
	}
}

func TestABMParameterIndexRemoteModeUsesUniverseWhenParamDirEmpty(t *testing.T) {
	dataDir := t.TempDir()
	paramsDir := t.TempDir()
	universeRoot := t.TempDir()
	writeUniverseSnapshot(t, universeRoot, "hs300", "2026Q1", map[string]string{
		"600000": "浦发银行",
		"600001": "测试银行",
	})

	restore := abmhttp.SetStockDataAvailabilityCheckerForTest(&fakeStockDataChecker{
		results: map[string]abmhttp.StockDataAvailabilityResult{
			"600001": {
				Available: false,
				Source:    "dolphindb",
				Reason:    "no remote rows",
			},
		},
	})
	defer restore()

	index, err := abmhttp.BuildSupportedStockIndex(testABMParametersBase(), &paradigm.BHLayer2NodeConfig{
		ABMStockDataDir:         dataDir,
		ABMStockParamDir:        paramsDir,
		ABMStockDataSource:      "dolphindb",
		ABMUniverseSnapshotRoot: universeRoot,
		ABMParameterUniverse:    "hs300",
	})
	if err != nil {
		t.Fatalf("build index: %v", err)
	}
	if index.Total != 1 || index.TunedCount != 0 {
		t.Fatalf("expected one available default stock, got total=%d tuned=%d list=%#v", index.Total, index.TunedCount, index.SupportedStockList)
	}
	meta, ok := index.StockMap["600000"]
	if !ok || !meta.SupportSimulation || meta.HasTunedParams {
		t.Fatalf("expected 600000 available with default params, got %#v", meta)
	}
}

func TestABMParametersEndpointReturnsSupportedStockListOnly(t *testing.T) {
	dataDir := t.TempDir()
	paramsDir := t.TempDir()
	t.Setenv("ABM_STOCK_DATA_DIR", dataDir)
	t.Setenv("ABM_STOCK_PARAM_DIR", paramsDir)

	writeABMParameterParamFile(t, paramsDir, "600000", 123)
	writeABMParameterRawParamFile(t, paramsDir, "600001", `{}`)

	service := newABMParametersService(t)
	data := callABMParametersObject(t, service, http.MethodGet, "")

	if _, exists := data["parameters"]; exists {
		t.Fatalf("list response must not include parameters: %#v", data)
	}
	if _, exists := data["parameterSchema"]; exists {
		t.Fatalf("list response must not include parameterSchema: %#v", data)
	}
	if !sameJSONValue(data["total"], 2) {
		t.Fatalf("expected total=2, got %#v", data["total"])
	}

	items := data["items"].([]interface{})
	if len(items) != 2 {
		t.Fatalf("expected 2 supported stocks, got %d: %#v", len(items), items)
	}

	stock600000 := findStockItem(t, items, "600000")
	if stock600000["stockName"] != "浦发银行" {
		t.Fatalf("expected mapped stockName, got %#v", stock600000["stockName"])
	}
	if stock600000["hasTunedParams"] != true {
		t.Fatalf("expected 600000 has tuned params, got %#v", stock600000["hasTunedParams"])
	}
	if _, exists := stock600000["defaults"]; exists {
		t.Fatalf("list item must not include defaults: %#v", stock600000)
	}

	stock600001 := findStockItem(t, items, "600001")
	if stock600001["hasTunedParams"] != false {
		t.Fatalf("expected 600001 no tuned params, got %#v", stock600001["hasTunedParams"])
	}
}

func TestABMParametersEndpointSingleStockWithTunedParams(t *testing.T) {
	dataDir := t.TempDir()
	paramsDir := t.TempDir()
	t.Setenv("ABM_STOCK_DATA_DIR", dataDir)
	t.Setenv("ABM_STOCK_PARAM_DIR", paramsDir)

	writeABMParameterParamFile(t, paramsDir, "600000", 123)

	service := newABMParametersService(t)
	data := callABMParametersObject(t, service, http.MethodGet, "stockCode=600000")

	if data["stockCode"] != "600000" || data["stockName"] != "浦发银行" {
		t.Fatalf("unexpected single stock identity: %#v", data)
	}
	if data["supportSimulation"] != true || data["hasTunedParams"] != true {
		t.Fatalf("expected supported tuned stock, got %#v", data)
	}
	parameters := data["parameters"].(map[string]interface{})
	if len(parameters) != 20 {
		t.Fatalf("expected 20 parameters, got %d: %#v", len(parameters), parameters)
	}
	if !sameJSONValue(findParameter(t, parameters, "N_FT")["default"], 123) {
		t.Fatalf("expected tuned N_FT default 123, got %#v", findParameter(t, parameters, "N_FT"))
	}
	if findParameter(t, parameters, "N_FT")["source"] != "tuned" {
		t.Fatalf("expected N_FT source=tuned, got %#v", findParameter(t, parameters, "N_FT"))
	}
	if !sameJSONValue(findParameter(t, parameters, "K1")["default"], 1.9855) {
		t.Fatalf("expected tuned K1 default 1.9855, got %#v", findParameter(t, parameters, "K1"))
	}
	if findParameter(t, parameters, "K1")["source"] != "tuned" {
		t.Fatalf("expected K1 source=tuned, got %#v", findParameter(t, parameters, "K1"))
	}
	if !sameJSONValue(findParameter(t, parameters, "DELTA_NT")["default"], 1.0325) {
		t.Fatalf("expected default DELTA_NT=1.0325, got %#v", findParameter(t, parameters, "DELTA_NT"))
	}
	if findParameter(t, parameters, "DELTA_NT")["source"] != "default" {
		t.Fatalf("expected DELTA_NT source=default, got %#v", findParameter(t, parameters, "DELTA_NT"))
	}
	if findParameter(t, parameters, "MU_L")["label"] != "限价订单价格距离对数正态分布均值" {
		t.Fatalf("expected MU_L label, got %#v", findParameter(t, parameters, "MU_L"))
	}
}

func TestABMParametersEndpointSingleStockWithoutTunedParamsUsesDefault(t *testing.T) {
	dataDir := t.TempDir()
	paramsDir := t.TempDir()
	t.Setenv("ABM_STOCK_DATA_DIR", dataDir)
	t.Setenv("ABM_STOCK_PARAM_DIR", paramsDir)

	writeABMParameterRawParamFile(t, paramsDir, "600001", `{}`)

	service := newABMParametersService(t)
	data := callABMParametersObject(t, service, http.MethodGet, "stockCode=600001")

	if data["supportSimulation"] != true || data["hasTunedParams"] != false {
		t.Fatalf("expected supported stock without tuned params, got %#v", data)
	}
	parameters := data["parameters"].(map[string]interface{})
	if !sameJSONValue(findParameter(t, parameters, "N_FT")["default"], 300) {
		t.Fatalf("expected default N_FT=300, got %#v", findParameter(t, parameters, "N_FT"))
	}
	if findParameter(t, parameters, "N_FT")["source"] != "default" {
		t.Fatalf("expected N_FT source=default, got %#v", findParameter(t, parameters, "N_FT"))
	}
	if !sameJSONValue(findParameter(t, parameters, "K1")["default"], 0.2855) {
		t.Fatalf("expected default K1=0.2855, got %#v", findParameter(t, parameters, "K1"))
	}
}

func TestABMParametersEndpointStockFromParamDirWithoutInputCsv(t *testing.T) {
	dataDir := t.TempDir()
	paramsDir := t.TempDir()
	t.Setenv("ABM_STOCK_DATA_DIR", dataDir)
	t.Setenv("ABM_STOCK_PARAM_DIR", paramsDir)
	writeABMParameterParamFile(t, paramsDir, "999999", 77)

	service := newABMParametersService(t)
	data := callABMParametersObject(t, service, http.MethodGet, "stockCode=999999")

	if data["stockCode"] != "999999" || data["stockName"] != "999999" {
		t.Fatalf("expected unknown stockCode/name fallback, got %#v", data)
	}
	if data["supportSimulation"] != true || data["hasTunedParams"] != true {
		t.Fatalf("expected stock from params dir to be supported, got %#v", data)
	}
	parameters := data["parameters"].(map[string]interface{})
	if !sameJSONValue(findParameter(t, parameters, "N_FT")["default"], 77) {
		t.Fatalf("expected tuned N_FT=77, got %#v", findParameter(t, parameters, "N_FT"))
	}
}

func TestABMParametersEndpointInvalidTunedParamsFallbackToDefault(t *testing.T) {
	dataDir := t.TempDir()
	paramsDir := t.TempDir()
	t.Setenv("ABM_STOCK_DATA_DIR", dataDir)
	t.Setenv("ABM_STOCK_PARAM_DIR", paramsDir)

	writeABMParameterRawParamFile(t, paramsDir, "600000", `{"structural_params":{"N_FT":999999}}`)

	service := newABMParametersService(t)
	data := callABMParametersObject(t, service, http.MethodGet, "stockCode=600000")

	if data["supportSimulation"] != true || data["hasTunedParams"] != false {
		t.Fatalf("expected invalid tuned params to be ignored, got %#v", data)
	}
	parameters := data["parameters"].(map[string]interface{})
	if !sameJSONValue(findParameter(t, parameters, "N_FT")["default"], 300) {
		t.Fatalf("expected default N_FT=300, got %#v", findParameter(t, parameters, "N_FT"))
	}
}

func TestABMParametersEndpointStartupFailureReturnsEmptyList(t *testing.T) {
	dataDir := t.TempDir()
	paramsDir := t.TempDir()
	t.Setenv("ABM_STOCK_DATA_DIR", dataDir)
	t.Setenv("ABM_STOCK_PARAM_DIR", filepath.Join(paramsDir, "missing"))

	service := newABMParametersService(t)
	data := callABMParametersObject(t, service, http.MethodGet, "")
	if !sameJSONValue(data["total"], 0) {
		t.Fatalf("expected empty list when startup index build fails, got %#v", data)
	}
	if len(data["items"].([]interface{})) != 0 {
		t.Fatalf("expected empty items when startup index build fails, got %#v", data["items"])
	}
}

func TestABMParametersInternalRefreshFailureKeepsOldIndex(t *testing.T) {
	dataDir := t.TempDir()
	paramsDir := t.TempDir()
	t.Setenv("ABM_STOCK_DATA_DIR", dataDir)
	t.Setenv("ABM_STOCK_PARAM_DIR", paramsDir)

	writeABMParameterParamFile(t, paramsDir, "600000", 123)
	service := newABMParametersService(t)
	refreshService := newABMParametersRefreshService(t)

	first := callABMParametersObject(t, service, http.MethodGet, "")
	if !sameJSONValue(first["total"], 1) {
		t.Fatalf("expected initial total=1, got %#v", first)
	}

	t.Setenv("ABM_STOCK_PARAM_DIR", filepath.Join(t.TempDir(), "missing"))
	recorder := callABMParameters(t, refreshService, http.MethodPost, "")
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected refresh failure status 500, got %d: %s", recorder.Code, recorder.Body.String())
	}

	kept := callABMParametersObject(t, service, http.MethodGet, "")
	if !sameJSONValue(kept["total"], 1) {
		t.Fatalf("expected old index kept after refresh failure, got %#v", kept)
	}
}

func TestABMParametersEndpointPaginationKeywordAndInternalRefresh(t *testing.T) {
	dataDir := t.TempDir()
	paramsDir := t.TempDir()
	t.Setenv("ABM_STOCK_DATA_DIR", dataDir)
	t.Setenv("ABM_STOCK_PARAM_DIR", paramsDir)

	writeABMParameterParamFile(t, paramsDir, "600000", 123)
	service := newABMParametersService(t)
	refreshService := newABMParametersRefreshService(t)

	first := callABMParametersObject(t, service, http.MethodGet, "")
	if !sameJSONValue(first["total"], 1) {
		t.Fatalf("expected first total=1, got %#v", first)
	}

	writeABMParameterRawParamFile(t, paramsDir, "600001", `{}`)
	cached := callABMParametersObject(t, service, http.MethodGet, "")
	if !sameJSONValue(cached["total"], 1) {
		t.Fatalf("expected cached total=1 before internal refresh, got %#v", cached)
	}

	callABMParametersObject(t, refreshService, http.MethodPost, "")
	refreshed := callABMParametersObject(t, service, http.MethodGet, "pageNo=1&pageSize=1")
	if !sameJSONValue(refreshed["total"], 2) {
		t.Fatalf("expected refreshed total=2, got %#v", refreshed)
	}
	if len(refreshed["items"].([]interface{})) != 1 {
		t.Fatalf("expected pageSize=1, got %#v", refreshed["items"])
	}

	byCode := callABMParametersObject(t, service, http.MethodGet, "keyword=600001")
	if !sameJSONValue(byCode["total"], 1) {
		t.Fatalf("expected keyword by stockCode total=1, got %#v", byCode)
	}

	byName := callABMParametersObject(t, service, http.MethodGet, "keyword=浦发")
	if !sameJSONValue(byName["total"], 1) {
		t.Fatalf("expected keyword by stockName total=1, got %#v", byName)
	}
}

func newABMParametersService(t *testing.T) *HTTP.HttpService {
	t.Helper()
	return newABMHTTPService(t, HTTP.ABM_PARAMETERS)
}

func newABMParametersRefreshService(t *testing.T) *HTTP.HttpService {
	t.Helper()
	return newABMHTTPService(t, HTTP.ABM_PARAMETERS_REFRESH)
}

func newABMHTTPService(t *testing.T, serviceType HTTP.HttpServiceEnum) *HTTP.HttpService {
	t.Helper()
	gin.SetMode(gin.TestMode)

	engine := &HTTP.HttpEngine{}
	engine.Setup(paradigm.BHLayer2NodeConfig{
		AbmParameters: map[string]interface{}{
			"N_FT": map[string]interface{}{
				"label":   "基本面交易者数量",
				"type":    "int",
				"default": 300,
				"min":     10,
				"max":     8000,
			},
			"S_FT": map[string]interface{}{
				"label":   "基本面交易者交易间隔（step 维度）",
				"type":    "int",
				"default": 1,
				"min":     1,
				"max":     20,
			},
			"N_LMT": map[string]interface{}{
				"label":   "长期动量交易者数量",
				"type":    "int",
				"default": 300,
				"min":     10,
				"max":     8000,
			},
			"ALPHA_L": map[string]interface{}{
				"label":   "长期动量趋势信号衰减/更新权重（EMA 系数）",
				"type":    "float",
				"default": 0.001,
				"min":     0.001,
				"max":     1,
			},
			"N_SMT": map[string]interface{}{
				"label":   "短期动量交易者数量",
				"type":    "int",
				"default": 300,
				"min":     10,
				"max":     8000,
			},
			"ALPHA_S": map[string]interface{}{
				"label":   "短期动量趋势信号衰减/更新权重（EMA 系数）",
				"type":    "float",
				"default": 0.9,
				"min":     0.001,
				"max":     1,
			},
			"N_NT": map[string]interface{}{
				"label":   "噪音交易者数量",
				"type":    "int",
				"default": 300,
				"min":     10,
				"max":     8000,
			},
		},
	})

	service, err := engine.GetHttpService(serviceType)
	if err != nil {
		t.Fatalf("get service %v: %v", serviceType, err)
	}
	return service
}

func callABMParametersObject(t *testing.T, service *HTTP.HttpService, method string, rawQuery string) map[string]interface{} {
	t.Helper()

	recorder := callABMParameters(t, service, method, rawQuery)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response paradigm.HttpResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected response data object, got %#v", response.Data)
	}
	return data
}

func callABMParameters(t *testing.T, service *HTTP.HttpService, method string, rawQuery string) *httptest.ResponseRecorder {
	t.Helper()

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	request := httptest.NewRequest(method, service.Url+"?"+rawQuery, nil)
	context.Request = request

	service.Handler(context)
	return recorder
}

func findStockItem(t *testing.T, stocks []interface{}, stockCode string) map[string]interface{} {
	t.Helper()
	for _, raw := range stocks {
		stock := raw.(map[string]interface{})
		if stock["stockCode"] == stockCode {
			return stock
		}
	}
	t.Fatalf("missing stock %s in %#v", stockCode, stocks)
	return nil
}

func findParameter(t *testing.T, parameters map[string]interface{}, key string) map[string]interface{} {
	t.Helper()
	raw, ok := parameters[key]
	if !ok {
		t.Fatalf("missing parameter %s in %#v", key, parameters)
	}
	return raw.(map[string]interface{})
}

func writeABMParameterDataFile(t *testing.T, root string, stockCode string) {
	t.Helper()
	content := []byte("date,close\n2022-01-04 09:30:00,1.0\n")
	if err := os.WriteFile(filepath.Join(root, stockCode+".csv"), content, 0o644); err != nil {
		t.Fatalf("write stock data csv: %v", err)
	}
}

func writeABMParameterParamFile(t *testing.T, root string, stockCode string, nFT int) {
	t.Helper()
	writeABMParameterRawParamFile(t, root, stockCode, fmt.Sprintf(`{
  "structural_params": {
    "N_FT": %d,
    "IGNORED": 999
  },
  "calibrated_params": {
    "K1": 1.9855
  }
}`, nFT))
}

func writeABMParameterRawParamFile(t *testing.T, root string, stockCode string, content string) {
	t.Helper()
	dir := filepath.Join(root, stockCode)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create tuned param dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "model_params.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write tuned params: %v", err)
	}
}

func sameJSONValue(actual interface{}, expected interface{}) bool {
	actualFloat, actualNumeric := numericValue(actual)
	expectedFloat, expectedNumeric := numericValue(expected)
	if actualNumeric && expectedNumeric {
		return actualFloat == expectedFloat
	}
	return fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", expected)
}

func numericValue(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}
