package abm

import (
	"BHLayer2Node/paradigm"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const defaultParametersIndexRefreshInterval = time.Hour

var TunableParamKeys = []string{
	"N_FT",
	"S_FT",
	"N_LMT",
	"ALPHA_L",
	"N_SMT",
	"ALPHA_S",
	"N_NT",
	"MU_L",
	"SIGMA_L",
	"K1",
	"K2",
	"BETA_L",
	"BETA_S",
	"DELTA_NT",
	"THETA",
	"MU",
	"DELTA",
	"RHO",
	"VOLUME",
	"GAMMA",
}

var intParamKeys = map[string]bool{
	"N_FT":   true,
	"S_FT":   true,
	"N_LMT":  true,
	"N_SMT":  true,
	"N_NT":   true,
	"VOLUME": true,
}

var paramLabels = map[string]string{
	"N_FT":     "基本面交易者数量",
	"S_FT":     "基本面交易者交易间隔（step 维度）",
	"N_LMT":    "长期动量交易者数量",
	"ALPHA_L":  "长期动量趋势信号衰减/更新权重（EMA 系数）",
	"N_SMT":    "短期动量交易者数量",
	"ALPHA_S":  "短期动量趋势信号衰减/更新权重（EMA 系数）",
	"N_NT":     "噪音交易者数量",
	"MU_L":     "限价订单价格距离对数正态分布均值",
	"SIGMA_L":  "限价订单价格距离对数正态分布标准差",
	"K1":       "基本面交易者线性需求系数",
	"K2":       "基本面交易者非线性需求系数",
	"BETA_L":   "长期动量交易者需求计算系数",
	"BETA_S":   "短期动量交易者需求计算系数",
	"DELTA_NT": "噪音交易者总需求水平参数",
	"THETA":    "提交限价订单概率",
	"MU":       "提交市价订单概率",
	"DELTA":    "限价订单取消概率",
	"RHO":      "市价单与限价单提交概率之比",
	"VOLUME":   "单笔订单体积",
	"GAMMA":    "动量交易者需求计算系数",
}

var runtimeParamDefaults = map[string]interface{}{
	"MU_L":     1.1,
	"SIGMA_L":  0.3,
	"K1":       0.2855,
	"K2":       0.4058,
	"BETA_L":   0.6905,
	"BETA_S":   0.0554,
	"DELTA_NT": 1.0325,
	"THETA":    0,
	"MU":       0,
	"DELTA":    0.005,
	"RHO":      0.2,
	"VOLUME":   100,
	"GAMMA":    10,
}

var (
	stockCodeRegexp          = regexp.MustCompile(`\d{6}`)
	currentABMParameterIndex atomic.Value
)

type modelParamsFile struct {
	StructuralParams map[string]interface{} `json:"structural_params"`
	CalibratedParams map[string]interface{} `json:"calibrated_params"`
}

type ParameterType string

const (
	ParameterTypeInt   ParameterType = "int"
	ParameterTypeFloat ParameterType = "float"
)

type ParameterSpec struct {
	Key     string
	Label   string
	Type    ParameterType
	Default float64
	Min     *float64
	Max     *float64
}

type ParameterResponseItem struct {
	Key     string      `json:"key"`
	Label   string      `json:"label"`
	Type    string      `json:"type"`
	Default interface{} `json:"default"`
	Min     interface{} `json:"min,omitempty"`
	Max     interface{} `json:"max,omitempty"`
	Source  string      `json:"source,omitempty"`
}

type StockSimulationListItem struct {
	StockCode      string `json:"stockCode"`
	StockName      string `json:"stockName"`
	HasTunedParams bool   `json:"hasTunedParams"`
}

type StockSimulationMeta struct {
	StockCode              string
	StockName              string
	SupportSimulation      bool
	HasInputCsv            bool
	HasTunedParams         bool
	TunedParams            map[string]float64
	InputCsvPath           string
	TunedParamPath         string
	InputCsvLastModified   int64
	TunedParamLastModified int64
}

type SupportedStockIndex struct {
	Version            string
	LastUpdatedAt      string
	StockDataDir       string
	StockParamDir      string
	ParameterSpecs     map[string]ParameterSpec
	StockMap           map[string]StockSimulationMeta
	SupportedStockList []StockSimulationListItem
	Total              int
	TunedCount         int
}

type RefreshABMParameterIndexResult struct {
	OldVersion string `json:"oldVersion"`
	NewVersion string `json:"newVersion"`
	Total      int    `json:"total"`
	TunedCount int    `json:"tunedCount"`
}

func InitABMParameterIndex(base map[string]interface{}, config *paradigm.BHLayer2NodeConfig) {
	start := time.Now()
	index, err := BuildSupportedStockIndex(base, config)
	if err != nil {
		paradigm.Log("ERROR", fmt.Sprintf("AbmParameterIndex initialize failed, use empty index, error=%v", err))
		index = BuildEmptySupportedStockIndex(base, config)
	}
	currentABMParameterIndex.Store(index)
	paradigm.Log("INFO", fmt.Sprintf("AbmParameterIndex initialized, version=%s, total=%d, tunedCount=%d, costMs=%d",
		index.Version, index.Total, index.TunedCount, time.Since(start).Milliseconds()))
}

func StartABMParameterIndexRefresher(base map[string]interface{}, config *paradigm.BHLayer2NodeConfig) {
	ticker := time.NewTicker(defaultParametersIndexRefreshInterval)
	go func() {
		for range ticker.C {
			if _, err := RefreshABMParameterIndex(base, config); err != nil {
				// RefreshABMParameterIndex 已记录保留旧版本的日志。
			}
		}
	}()
}

func CurrentABMParameterIndex() *SupportedStockIndex {
	if value := currentABMParameterIndex.Load(); value != nil {
		if index, ok := value.(*SupportedStockIndex); ok && index != nil {
			return index
		}
	}
	return BuildEmptySupportedStockIndex(nil, nil)
}

func RefreshABMParameterIndex(base map[string]interface{}, config *paradigm.BHLayer2NodeConfig) (RefreshABMParameterIndexResult, error) {
	start := time.Now()
	oldIndex := CurrentABMParameterIndex()
	newIndex, err := BuildSupportedStockIndex(base, config)
	if err != nil {
		paradigm.Log("ERROR", fmt.Sprintf("AbmParameterIndex refresh failed, keep old version=%s, error=%v", oldIndex.Version, err))
		return RefreshABMParameterIndexResult{OldVersion: oldIndex.Version}, err
	}
	currentABMParameterIndex.Store(newIndex)
	paradigm.Log("INFO", fmt.Sprintf("AbmParameterIndex refreshed, oldVersion=%s, newVersion=%s, total=%d, tunedCount=%d, costMs=%d",
		oldIndex.Version, newIndex.Version, newIndex.Total, newIndex.TunedCount, time.Since(start).Milliseconds()))
	return RefreshABMParameterIndexResult{
		OldVersion: oldIndex.Version,
		NewVersion: newIndex.Version,
		Total:      newIndex.Total,
		TunedCount: newIndex.TunedCount,
	}, nil
}

func BuildEmptySupportedStockIndex(base map[string]interface{}, config *paradigm.BHLayer2NodeConfig) *SupportedStockIndex {
	now := time.Now()
	return &SupportedStockIndex{
		Version:            now.Format("20060102_150405"),
		LastUpdatedAt:      now.Format("2006-01-02 15:04:05"),
		StockDataDir:       StockDataDir(config),
		StockParamDir:      StockParamDir(config),
		ParameterSpecs:     BuildParameterSpecMap(base),
		StockMap:           map[string]StockSimulationMeta{},
		SupportedStockList: []StockSimulationListItem{},
		Total:              0,
		TunedCount:         0,
	}
}

func BuildSupportedStockIndex(base map[string]interface{}, config *paradigm.BHLayer2NodeConfig) (*SupportedStockIndex, error) {
	now := time.Now()
	dataDir := StockDataDir(config)
	paramDir := StockParamDir(config)
	specs := BuildParameterSpecMap(base)

	entries, err := os.ReadDir(paramDir)
	if err != nil {
		return nil, err
	}

	stockMap := map[string]StockSimulationMeta{}
	supportedList := make([]StockSimulationListItem, 0, len(entries))
	tunedCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		stockCode := NormalizeStockCode(entry.Name())
		if stockCode == "" {
			continue
		}
		if _, exists := stockMap[stockCode]; exists {
			continue
		}

		paramPath := filepath.Join(paramDir, entry.Name(), "model_params.json")
		if _, err := os.Stat(paramPath); err != nil {
			continue
		}
		tunedParams, tunedMTime, hasTuned := loadStockTunedParamsFromPath(stockCode, paramPath, specs)
		if hasTuned {
			tunedCount++
		}

		inputPath := filepath.Join(dataDir, stockCode+".csv")
		inputMTime := int64(0)
		hasInputCsv := false
		if info, err := os.Stat(inputPath); err == nil && !info.IsDir() {
			hasInputCsv = true
			inputMTime = info.ModTime().Unix()
		}

		stockName := ResolveStockDisplayName(stockCode, stockCode)
		meta := StockSimulationMeta{
			StockCode:              stockCode,
			StockName:              stockName,
			SupportSimulation:      true,
			HasInputCsv:            hasInputCsv,
			HasTunedParams:         hasTuned,
			TunedParams:            tunedParams,
			InputCsvPath:           inputPath,
			TunedParamPath:         paramPath,
			InputCsvLastModified:   inputMTime,
			TunedParamLastModified: tunedMTime,
		}
		stockMap[stockCode] = meta
		supportedList = append(supportedList, StockSimulationListItem{
			StockCode:      stockCode,
			StockName:      stockName,
			HasTunedParams: hasTuned,
		})
	}

	sort.Slice(supportedList, func(i, j int) bool {
		return supportedList[i].StockCode < supportedList[j].StockCode
	})

	return &SupportedStockIndex{
		Version:            now.Format("20060102_150405"),
		LastUpdatedAt:      now.Format("2006-01-02 15:04:05"),
		StockDataDir:       dataDir,
		StockParamDir:      paramDir,
		ParameterSpecs:     specs,
		StockMap:           stockMap,
		SupportedStockList: supportedList,
		Total:              len(supportedList),
		TunedCount:         tunedCount,
	}, nil
}

func BuildABMParameterListResponse(index *SupportedStockIndex, pageNo int, pageSize int, keyword string) map[string]interface{} {
	if index == nil {
		index = CurrentABMParameterIndex()
	}
	pageNo, pageSize = normalizePage(pageNo, pageSize)
	keyword = strings.ToLower(strings.TrimSpace(keyword))

	filtered := make([]StockSimulationListItem, 0, len(index.SupportedStockList))
	for _, item := range index.SupportedStockList {
		if keyword == "" ||
			strings.Contains(strings.ToLower(item.StockCode), keyword) ||
			strings.Contains(strings.ToLower(item.StockName), keyword) {
			filtered = append(filtered, item)
		}
	}

	total := len(filtered)
	start := (pageNo - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	return map[string]interface{}{
		"version":       index.Version,
		"lastUpdatedAt": index.LastUpdatedAt,
		"total":         total,
		"pageNo":        pageNo,
		"pageSize":      pageSize,
		"items":         filtered[start:end],
	}
}

func BuildABMSingleStockDetail(base map[string]interface{}, index *SupportedStockIndex, stockCode string) map[string]interface{} {
	rawStockCode := strings.TrimSpace(stockCode)
	stockCode = NormalizeStockCode(stockCode)
	if stockCode == "" {
		stockCode = rawStockCode
	}
	if index == nil {
		index = CurrentABMParameterIndex()
	}
	specs := index.ParameterSpecs
	if len(specs) == 0 {
		specs = BuildParameterSpecMap(base)
	}

	meta, exists := index.StockMap[stockCode]
	if !exists {
		meta = StockSimulationMeta{
			StockCode:         stockCode,
			StockName:         ResolveStockDisplayName(stockCode, stockCode),
			SupportSimulation: false,
			HasInputCsv:       false,
			HasTunedParams:    false,
			TunedParams:       map[string]float64{},
		}
	}
	if strings.TrimSpace(meta.StockName) == "" {
		meta.StockName = meta.StockCode
	}

	parameters := make(map[string]ParameterResponseItem, len(TunableParamKeys))
	for _, key := range TunableParamKeys {
		spec := specs[key]
		value := spec.Default
		source := "default"
		if meta.HasTunedParams {
			if tuned, ok := meta.TunedParams[key]; ok {
				value = tuned
				source = "tuned"
			}
		}
		parameters[key] = buildParameterResponseItem(spec, value, source)
	}

	return map[string]interface{}{
		"stockCode":         meta.StockCode,
		"stockName":         meta.StockName,
		"supportSimulation": meta.SupportSimulation,
		"hasTunedParams":    meta.HasTunedParams,
		"parameters":        parameters,
	}
}

func IsStockSupportedByIndex(stockCode string) bool {
	stockCode = NormalizeStockCode(stockCode)
	if stockCode == "" {
		return false
	}
	index := CurrentABMParameterIndex()
	meta, ok := index.StockMap[stockCode]
	return ok && meta.SupportSimulation
}

func StockMetaFromIndex(stockCode string) (StockSimulationMeta, bool) {
	stockCode = NormalizeStockCode(stockCode)
	if stockCode == "" {
		return StockSimulationMeta{}, false
	}
	meta, ok := CurrentABMParameterIndex().StockMap[stockCode]
	return meta, ok
}

func TunedParamsFromIndex(stockCode string) (map[string]interface{}, bool) {
	meta, ok := StockMetaFromIndex(stockCode)
	if !ok || !meta.HasTunedParams {
		return map[string]interface{}{}, false
	}
	result := make(map[string]interface{}, len(meta.TunedParams))
	specs := CurrentABMParameterIndex().ParameterSpecs
	for key, value := range meta.TunedParams {
		spec := specs[key]
		if spec.Key == "" {
			spec = parameterSpecFromDefault(nil, key)
		}
		result[key] = formatParameterValue(spec, value)
	}
	return result, len(result) > 0
}

// BuildParametersResponse is kept for existing callers. It no longer reads tuned
// parameter files directly; tuned values are loaded from the in-memory index.
func BuildParametersResponse(base map[string]interface{}, stockCode string, config *paradigm.BHLayer2NodeConfig) map[string]interface{} {
	response := cloneParameters(base)
	tunedParams, hasTunedParams := TunedParamsFromIndex(stockCode)

	for _, key := range TunableParamKeys {
		spec := ensureParamSpec(response, key)
		if value, ok := tunedParams[key]; hasTunedParams && ok {
			spec["default"] = value
			spec["source"] = "tuned"
			continue
		}
		if _, ok := spec["source"]; !ok {
			spec["source"] = "default"
		}
	}

	return response
}

// LoadStockTunedParams is retained for non-request utilities and tests. Request
// handlers should use the in-memory index instead.
func LoadStockTunedParams(stockCode string, config *paradigm.BHLayer2NodeConfig) (map[string]interface{}, bool) {
	stockCode = NormalizeStockCode(stockCode)
	if stockCode == "" {
		return map[string]interface{}{}, false
	}
	path := filepath.Join(StockParamDir(config), stockCode, "model_params.json")
	values, _, ok := loadStockTunedParamsFromPath(stockCode, path, BuildParameterSpecMap(nil))
	if !ok {
		return map[string]interface{}{}, false
	}
	result := make(map[string]interface{}, len(values))
	specs := BuildParameterSpecMap(nil)
	for key, value := range values {
		result[key] = formatParameterValue(specs[key], value)
	}
	return result, true
}

// HasStockInputFile is kept for compatibility. New request paths should call
// IsStockSupportedByIndex so they do not touch the filesystem.
func HasStockInputFile(stockCode string, config *paradigm.BHLayer2NodeConfig) bool {
	return IsStockSupportedByIndex(stockCode)
}

func BuildParameterSpecMap(base map[string]interface{}) map[string]ParameterSpec {
	specs := make(map[string]ParameterSpec, len(TunableParamKeys))
	for _, key := range TunableParamKeys {
		specs[key] = parameterSpecFromDefault(base, key)
	}
	return specs
}

func parameterSpecFromDefault(base map[string]interface{}, key string) ParameterSpec {
	rawSpec := map[string]interface{}{}
	if base != nil {
		if nested, ok := base[key].(map[string]interface{}); ok {
			rawSpec = nested
		}
	}

	paramType := ParameterTypeFloat
	if intParamKeys[key] || strings.EqualFold(stringValue(rawSpec["type"]), string(ParameterTypeInt)) {
		paramType = ParameterTypeInt
	}

	label := strings.TrimSpace(stringValue(rawSpec["label"]))
	if label == "" {
		label = paramLabels[key]
	}

	defaultValue := 0.0
	if value, ok := numericValue(rawSpec["default"]); ok {
		defaultValue = value
	} else if value, exists := runtimeParamDefaults[key]; exists {
		if parsed, ok := numericValue(value); ok {
			defaultValue = parsed
		}
	}

	var minValue *float64
	if value, ok := numericValue(rawSpec["min"]); ok {
		minValue = &value
	}
	var maxValue *float64
	if value, ok := numericValue(rawSpec["max"]); ok {
		maxValue = &value
	}

	return ParameterSpec{
		Key:     key,
		Label:   label,
		Type:    paramType,
		Default: defaultValue,
		Min:     minValue,
		Max:     maxValue,
	}
}

func loadStockTunedParamsFromPath(stockCode string, path string, specs map[string]ParameterSpec) (map[string]float64, int64, bool) {
	info, statErr := os.Stat(path)
	if statErr != nil {
		return map[string]float64{}, 0, false
	}

	file, err := os.Open(path)
	if err != nil {
		paradigm.Log("WARN", fmt.Sprintf("abm tuned parameter parse failed, stockCode=%s, path=%s, error=%v", stockCode, path, err))
		return map[string]float64{}, info.ModTime().Unix(), false
	}
	defer file.Close()

	var payload modelParamsFile
	decoder := json.NewDecoder(file)
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		paradigm.Log("WARN", fmt.Sprintf("abm tuned parameter parse failed, stockCode=%s, path=%s, error=%v", stockCode, path, err))
		return map[string]float64{}, info.ModTime().Unix(), false
	}

	allowed := make(map[string]bool, len(TunableParamKeys))
	for _, key := range TunableParamKeys {
		allowed[key] = true
	}

	result := map[string]float64{}
	mergeValidatedParamValues(result, payload.StructuralParams, allowed, specs, stockCode)
	mergeValidatedParamValues(result, payload.CalibratedParams, allowed, specs, stockCode)
	return result, info.ModTime().Unix(), len(result) > 0
}

func mergeValidatedParamValues(dst map[string]float64, src map[string]interface{}, allowed map[string]bool, specs map[string]ParameterSpec, stockCode string) {
	for key, raw := range src {
		if !allowed[key] {
			continue
		}
		spec := specs[key]
		if spec.Key == "" {
			spec = parameterSpecFromDefault(nil, key)
		}
		value, ok, reason := validateParamValue(spec, raw)
		if !ok {
			paradigm.Log("WARN", fmt.Sprintf("abm tuned parameter invalid, stockCode=%s, key=%s, value=%v, reason=%s, use default", stockCode, key, raw, reason))
			continue
		}
		dst[key] = value
	}
}

func validateParamValue(spec ParameterSpec, raw interface{}) (float64, bool, string) {
	value, ok := numericValue(raw)
	if !ok {
		return 0, false, "not numeric"
	}
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, false, "not finite"
	}
	if spec.Type == ParameterTypeInt && math.Trunc(value) != value {
		return 0, false, "not integer"
	}
	if spec.Min != nil && value < *spec.Min {
		return 0, false, "below min"
	}
	if spec.Max != nil && value > *spec.Max {
		return 0, false, "above max"
	}
	return value, true, ""
}

func buildParameterResponseItem(spec ParameterSpec, value float64, source string) ParameterResponseItem {
	return ParameterResponseItem{
		Key:     spec.Key,
		Label:   spec.Label,
		Type:    string(spec.Type),
		Default: formatParameterValue(spec, value),
		Min:     formatOptionalParameterValue(spec, spec.Min),
		Max:     formatOptionalParameterValue(spec, spec.Max),
		Source:  source,
	}
}

func formatOptionalParameterValue(spec ParameterSpec, value *float64) interface{} {
	if value == nil {
		return nil
	}
	return formatParameterValue(spec, *value)
}

func formatParameterValue(spec ParameterSpec, value float64) interface{} {
	if spec.Type == ParameterTypeInt {
		return int(value)
	}
	return value
}

func normalizePage(pageNo int, pageSize int) (int, int) {
	if pageNo < 1 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 500 {
		pageSize = 500
	}
	return pageNo, pageSize
}

func cloneParameters(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{}, len(src))
	for key, value := range src {
		if nested, ok := value.(map[string]interface{}); ok {
			copied := make(map[string]interface{}, len(nested))
			for nestedKey, nestedValue := range nested {
				copied[nestedKey] = nestedValue
			}
			dst[key] = copied
			continue
		}
		dst[key] = value
	}
	return dst
}

func ensureParamSpec(parameters map[string]interface{}, key string) map[string]interface{} {
	if existing, ok := parameters[key].(map[string]interface{}); ok {
		return existing
	}

	paramType := "float"
	if intParamKeys[key] {
		paramType = "int"
	}
	spec := map[string]interface{}{
		"label":   paramLabels[key],
		"type":    paramType,
		"default": nil,
	}
	if value, ok := runtimeParamDefaults[key]; ok {
		spec["default"] = value
	}
	parameters[key] = spec
	return spec
}

func numericValue(raw interface{}) (float64, bool) {
	switch value := raw.(type) {
	case json.Number:
		parsed, err := value.Float64()
		return parsed, err == nil
	case float64:
		return value, true
	case float32:
		return float64(value), true
	case int:
		return float64(value), true
	case int64:
		return float64(value), true
	case int32:
		return float64(value), true
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func NormalizeStockCode(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if _, err := strconv.Atoi(raw); err == nil && len(raw) < 6 {
		return strings.Repeat("0", 6-len(raw)) + raw
	}
	match := stockCodeRegexp.FindString(raw)
	return match
}
