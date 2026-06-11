package abm

import (
	"BHLayer2Node/paradigm"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

var ErrAnalyticsNotFound = errors.New("analytics result not found")

var hiddenPerformanceMetrics = map[string]bool{
	"Kurtosis (峰度)": true,
	"Skewness (偏度)": true,
	"Volume Mean":   true,
}

type AnalyticsQueryItem struct {
	Task      *paradigm.Task
	TaskID    string
	TaskName  string
	StockID   string
	StockCode string
	StockName string
	Date      string
	Data      interface{}
}

func HideUnstablePerformanceMetrics(data interface{}) interface{} {
	switch payload := data.(type) {
	case map[string]interface{}:
		if tableData, ok := payload["tableData"]; ok {
			payload["tableData"] = filterPerformanceTableData(tableData)
		}
		if nested, ok := payload["data"]; ok {
			payload["data"] = HideUnstablePerformanceMetrics(nested)
		}
		return payload
	case []map[string]interface{}:
		for index := range payload {
			payload[index] = HideUnstablePerformanceMetrics(payload[index]).(map[string]interface{})
		}
		return payload
	case []interface{}:
		for index := range payload {
			payload[index] = HideUnstablePerformanceMetrics(payload[index])
		}
		return payload
	default:
		return data
	}
}

func filterPerformanceTableData(tableData interface{}) interface{} {
	switch rows := tableData.(type) {
	case []interface{}:
		filtered := make([]interface{}, 0, len(rows))
		for _, raw := range rows {
			row, ok := raw.(map[string]interface{})
			if ok && hiddenPerformanceMetrics[strings.TrimSpace(fmt.Sprintf("%v", row["indicator"]))] {
				continue
			}
			filtered = append(filtered, raw)
		}
		return filtered
	case []map[string]interface{}:
		filtered := make([]map[string]interface{}, 0, len(rows))
		for _, row := range rows {
			if hiddenPerformanceMetrics[strings.TrimSpace(fmt.Sprintf("%v", row["indicator"]))] {
				continue
			}
			filtered = append(filtered, row)
		}
		return filtered
	default:
		return tableData
	}
}

func NormalizeInvestorCompositionResponse(data interface{}, selectedDate, selectedType string) interface{} {
	normalizedType := strings.TrimSpace(strings.ToLower(selectedType))
	if normalizedType == "" {
		normalizedType = "custom"
	}

	switch payload := data.(type) {
	case map[string]interface{}:
		if nested, ok := payload["data"].(map[string]interface{}); ok {
			payload["data"] = normalizeInvestorCompositionPayload(nested, selectedDate, normalizedType)
			return payload
		}
		return normalizeInvestorCompositionPayload(payload, selectedDate, normalizedType)
	case []map[string]interface{}:
		for index := range payload {
			payload[index] = NormalizeInvestorCompositionResponse(payload[index], selectedDate, normalizedType).(map[string]interface{})
		}
		return payload
	case []interface{}:
		for index := range payload {
			payload[index] = NormalizeInvestorCompositionResponse(payload[index], selectedDate, normalizedType)
		}
		return payload
	default:
		return data
	}
}

func normalizeInvestorCompositionPayload(payload map[string]interface{}, selectedDate, selectedType string) map[string]interface{} {
	if payload == nil {
		return map[string]interface{}{}
	}

	meta, _ := payload["meta"].(map[string]interface{})
	if meta == nil {
		meta = map[string]interface{}{}
	}
	meta["selectedDate"] = strings.TrimSpace(selectedDate)
	meta["selectedType"] = selectedType
	payload["meta"] = meta

	if history, ok := payload["historyData"].(map[string]interface{}); ok {
		categoryLabel := "当前配置"
		if strings.TrimSpace(selectedDate) != "" {
			categoryLabel = strings.TrimSpace(selectedDate)
		} else if selectedType == "history" {
			categoryLabel = "历史快照"
		}
		history["categories"] = []string{categoryLabel}
		payload["historyData"] = history
	}
	return payload
}

func AttachCrashRiskTopRiskListToItems(items []AnalyticsQueryItem) []AnalyticsQueryItem {
	topRiskList := BuildCrashRiskTopRiskList(items)
	for i := range items {
		items[i].Data = InjectCrashRiskTopRiskList(items[i].Data, topRiskList)
	}
	return items
}

func InjectCrashRiskTopRiskList(data interface{}, topRiskList []map[string]interface{}) interface{} {
	payload, ok := data.(map[string]interface{})
	if !ok {
		return data
	}
	payload["topRiskList"] = topRiskList
	return payload
}

func BuildCrashRiskTopRiskList(items []AnalyticsQueryItem) []map[string]interface{} {
	type entry struct {
		Rank        int
		Code        string
		Name        string
		Probability float64
	}

	entries := make([]entry, 0, len(items))
	for _, item := range items {
		score, ok := extractCrashRiskScore(item.Data)
		if !ok {
			continue
		}
		code := strings.TrimSpace(item.StockCode)
		if code == "" {
			code = strings.TrimSpace(item.StockID)
		}
		entries = append(entries, entry{
			Code:        code,
			Name:        strings.TrimSpace(item.StockName),
			Probability: score,
		})
	}

	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Probability == entries[j].Probability {
			return entries[i].Code < entries[j].Code
		}
		return entries[i].Probability > entries[j].Probability
	})

	result := make([]map[string]interface{}, 0, len(entries))
	for i := range entries {
		entries[i].Rank = i + 1
		result = append(result, map[string]interface{}{
			"rank":        entries[i].Rank,
			"code":        entries[i].Code,
			"name":        entries[i].Name,
			"probability": roundFloat(entries[i].Probability, 6),
		})
	}
	return result
}

func extractCrashRiskScore(data interface{}) (float64, bool) {
	payload, ok := data.(map[string]interface{})
	if !ok {
		return 0, false
	}

	best := math.Inf(-1)
	if series, ok := payload["forecastSeries"].([]interface{}); ok {
		// 风险榜单按 5% 跌幅出现概率排序，对应 predict_fv.csv 中的 Prob_Drop_5pct。
		for _, raw := range series {
			row, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}
			if score, ok := parseFloat64(row["probDrop5pct"]); ok && !math.IsNaN(score) {
				if score > best {
					best = score
				}
			}
		}
		if !math.IsInf(best, -1) {
			return clampProbability(best), true
		}
	}
	return 0, false
}

func parseFloat64(value interface{}) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case uint:
		return float64(typed), true
	case uint32:
		return float64(typed), true
	case uint64:
		return float64(typed), true
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		if err == nil {
			return parsed, true
		}
	}
	return 0, false
}

func clampProbability(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func roundFloat(value float64, digits int) float64 {
	factor := math.Pow10(digits)
	return math.Round(value*factor) / factor
}

func BuildAnalyticsQueryOptions(query func(string) string, analType paradigm.AnalysisType) map[string]string {
	options := map[string]string{}
	switch analType {
	case paradigm.PerformanceComparison:
		if model := strings.ToUpper(strings.TrimSpace(query("selectedModel"))); model != "" {
			options["selectedModel"] = model
		}
	case paradigm.OrderDynamics:
		if date := strings.TrimSpace(query("date")); date != "" {
			options["date"] = date
		}
	}
	return options
}

func EncodeAnalysisTypeRequest(analType paradigm.AnalysisType, options map[string]string) string {
	if len(options) == 0 {
		return analType.String()
	}

	queryParts := make([]string, 0, len(options))
	switch analType {
	case paradigm.PerformanceComparison:
		if model := strings.TrimSpace(options["selectedModel"]); model != "" {
			queryParts = append(queryParts, fmt.Sprintf("selectedModel=%s", model))
		}
	case paradigm.OrderDynamics:
		if date := strings.TrimSpace(options["date"]); date != "" {
			queryParts = append(queryParts, fmt.Sprintf("date=%s", date))
		}
	}

	if len(queryParts) == 0 {
		return analType.String()
	}
	return fmt.Sprintf("%s?%s", analType.String(), strings.Join(queryParts, "&"))
}

func SortTasksByStartTimeDesc(tasks []*paradigm.Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		return tasks[i].StartTime.After(tasks[j].StartTime)
	})
}

func DisplayTaskID(task *paradigm.Task) string {
	if task == nil {
		return ""
	}
	if task.PlatformTaskID != nil && strings.TrimSpace(*task.PlatformTaskID) != "" {
		return strings.TrimSpace(*task.PlatformTaskID)
	}
	return task.Sign
}

func ExtractTaskStockCode(task *paradigm.Task) string {
	if task == nil {
		return ""
	}
	if code := strings.TrimSpace(StringifyTaskParam(task.Params["stockCode"])); code != "" {
		return code
	}
	if code := strings.TrimSpace(StringifyTaskParam(task.Params["stockId"])); code != "" {
		return code
	}
	parts := strings.Split(task.Sign, "-")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func ExtractTaskStockID(task *paradigm.Task) string {
	if task == nil {
		return ""
	}
	if stockID := strings.TrimSpace(StringifyTaskParam(task.Params["stockId"])); stockID != "" {
		return stockID
	}
	return ExtractTaskStockCode(task)
}

func ExtractTaskStockName(task *paradigm.Task) string {
	if task == nil {
		return ""
	}
	return ResolveStockDisplayName(
		ExtractTaskStockCode(task),
		StringifyTaskParam(task.Params["stockName"]),
	)
}

func StringifyTaskParam(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

func MatchTaskStock(task *paradigm.Task, stockID string) bool {
	target := strings.TrimSpace(strings.ToLower(stockID))
	if target == "" {
		return false
	}
	return strings.EqualFold(ExtractTaskStockID(task), target) || strings.EqualFold(ExtractTaskStockCode(task), target)
}
