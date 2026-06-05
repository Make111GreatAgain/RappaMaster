package abm

import (
	"BHLayer2Node/paradigm"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var horizonRegexp = regexp.MustCompile(`T\+(\d+)`)

// BuildV2TaskParams 将前端的扁平 ABM_V2 请求补齐成执行链路需要的完整 payload。
// 前端只传股票信息 + 结构参数，predict/abm/evaluation 等运行参数由这里兜底构造。
func BuildV2TaskParams(raw map[string]interface{}, nodeID int32) (map[string]interface{}, error) {
	return BuildV2TaskParamsWithConfig(raw, nodeID, nil)
}

func BuildV2TaskParamsWithConfig(raw map[string]interface{}, nodeID int32, config *paradigm.BHLayer2NodeConfig) (map[string]interface{}, error) {
	stockCode := strings.TrimSpace(stringValue(raw["stockCode"]))
	if stockCode == "" {
		return nil, fmt.Errorf("stockCode is required")
	}
	stockName := strings.TrimSpace(stringValue(raw["stockName"]))
	if stockName == "" {
		return nil, fmt.Errorf("stockName is required")
	}

	horizonDays, err := parseV2HorizonDays(raw)
	if err != nil {
		return nil, err
	}

	params := cloneMap(raw)
	params["stockCode"] = stockCode
	params["stockName"] = stockName
	params["dataset"] = stockCode
	window, err := NormalizeABMDataWindow(raw, config)
	if err != nil {
		return nil, err
	}
	params["dataStartDate"] = window.StartDate
	params["dataEndDate"] = window.EndDate
	if nodeID >= 0 {
		params["assigned_node_id"] = nodeID
	}
	// 不再接受前端传绝对路径，统一只传逻辑文件名，由节点在本地数据目录解析。
	params["input_csv"] = fmt.Sprintf("%s.csv", stockCode)

	market := strings.ToUpper(strings.TrimSpace(stringValue(raw["market"])))
	if market == "" {
		market = inferStockMarket(stockCode)
	}
	params["market"] = market

	predictCfg := mapFromAny(raw["predict"])
	if len(predictCfg) == 0 {
		predictCfg = map[string]interface{}{}
	}
	predictCfg["enabled"] = true
	if strings.TrimSpace(stringValue(predictCfg["method"])) == "" {
		predictCfg["method"] = "kalman_rw"
	}
	predictCfg["horizon"] = horizonDays
	if _, ok := predictCfg["risk_drop_levels"]; !ok {
		predictCfg["risk_drop_levels"] = []interface{}{0.05}
	}
	params["predict"] = predictCfg

	abmCfg := mapFromAny(raw["abm"])
	if len(abmCfg) == 0 {
		abmCfg = map[string]interface{}{}
	}
	if strings.TrimSpace(stringValue(abmCfg["mode"])) == "" {
		abmCfg["mode"] = "auto"
	}
	if strings.TrimSpace(stringValue(abmCfg["model_params_root"])) == "" {
		// 离线参数按股票统一放在共享目录，避免多节点重复复制大规模参数文件。
		abmCfg["model_params_root"] = StockParamDir(config)
	}
	structuralParams := mapFromAny(abmCfg["structural_params"])
	if len(structuralParams) == 0 {
		structuralParams = map[string]interface{}{}
	}
	tunedParams, hasTunedParams := TunedParamsFromIndex(stockCode)
	if hasTunedParams {
		for _, key := range TunableParamKeys {
			if value, exists := tunedParams[key]; exists {
				structuralParams[key] = value
			}
		}
	}
	for _, key := range TunableParamKeys {
		if value, ok := raw[key]; ok && value != nil {
			structuralParams[key] = value
		}
	}
	abmCfg["structural_params"] = structuralParams
	params["abm"] = abmCfg
	params["hasTunedParams"] = hasTunedParams

	evaluationCfg := mapFromAny(raw["evaluation"])
	if len(evaluationCfg) == 0 {
		evaluationCfg = map[string]interface{}{}
	}
	evaluationCfg["code"] = stockCode
	if strings.TrimSpace(stringValue(evaluationCfg["market"])) == "" {
		evaluationCfg["market"] = "SM"
	}
	if _, ok := evaluationCfg["generate_models"]; !ok {
		// 默认生成 ABM/VRNN/TimeGAN 三种模型的评估结果，供 performance_comparison 按 selectedModel 切换。
		evaluationCfg["generate_models"] = true
	}
	if _, ok := evaluationCfg["vrnn_epochs"]; !ok {
		evaluationCfg["vrnn_epochs"] = 1
	}
	if _, ok := evaluationCfg["timegan_epochs"]; !ok {
		evaluationCfg["timegan_epochs"] = 1
	}
	if _, ok := evaluationCfg["min_deep_samples"]; !ok {
		evaluationCfg["min_deep_samples"] = 200
	}
	if _, ok := evaluationCfg["allow_fallback"]; !ok {
		evaluationCfg["allow_fallback"] = true
	}
	params["evaluation"] = evaluationCfg

	return params, nil
}

func parseV2HorizonDays(raw map[string]interface{}) (int, error) {
	switch value := raw["horizon"].(type) {
	case nil:
		return 1, nil
	case float64:
		days := int(value)
		if days <= 0 {
			return 0, fmt.Errorf("horizon must be positive")
		}
		return days, nil
	case int:
		if value <= 0 {
			return 0, fmt.Errorf("horizon must be positive")
		}
		return value, nil
	case string:
		text := strings.TrimSpace(value)
		if text == "" {
			return 1, nil
		}
		if strings.EqualFold(text, "custom") {
			return parseCustomHorizonDays(raw["custom_horizon_date"])
		}
		if days, err := strconv.Atoi(text); err == nil && days > 0 {
			return days, nil
		}
		if match := horizonRegexp.FindStringSubmatch(text); len(match) == 2 {
			days, _ := strconv.Atoi(match[1])
			if days > 0 {
				return days, nil
			}
		}
		return 0, fmt.Errorf("unsupported horizon value: %s", text)
	default:
		return 0, fmt.Errorf("unsupported horizon type: %T", value)
	}
}

func parseCustomHorizonDays(raw interface{}) (int, error) {
	rangeValues, ok := raw.([]interface{})
	if !ok || len(rangeValues) == 0 {
		return 0, fmt.Errorf("custom horizon requires custom_horizon_date")
	}

	parseDate := func(value interface{}) (time.Time, error) {
		text := strings.TrimSpace(stringValue(value))
		if text == "" {
			return time.Time{}, fmt.Errorf("custom horizon date is empty")
		}
		return time.Parse("2006-01-02", text)
	}

	start, err := parseDate(rangeValues[0])
	if err != nil {
		return 0, err
	}
	end := start
	if len(rangeValues) > 1 {
		end, err = parseDate(rangeValues[1])
		if err != nil {
			return 0, err
		}
	}
	if end.Before(start) {
		return 0, fmt.Errorf("custom_horizon_date end before start")
	}

	// 这里把自定义日期区间折算成交易日步数，供 predict.horizon 复用。
	days := 0
	for current := start; !current.After(end); current = current.AddDate(0, 0, 1) {
		if current.Weekday() != time.Saturday && current.Weekday() != time.Sunday {
			days++
		}
	}
	if days <= 0 {
		return 1, nil
	}
	return days, nil
}

func inferStockMarket(stockCode string) string {
	if strings.HasPrefix(stockCode, "6") {
		return "SH"
	}
	return "SZ"
}

func cloneMap(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{}, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func mapFromAny(raw interface{}) map[string]interface{} {
	if raw == nil {
		return nil
	}
	if value, ok := raw.(map[string]interface{}); ok {
		return cloneMap(value)
	}
	return nil
}

func stringValue(raw interface{}) string {
	if raw == nil {
		return ""
	}
	return fmt.Sprintf("%v", raw)
}
