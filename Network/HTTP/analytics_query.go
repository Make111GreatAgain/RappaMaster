package HTTP

import (
	"BHLayer2Node/Network/HTTP/abm"
	"BHLayer2Node/paradigm"
	"BHLayer2Node/pb/service"
	"BHLayer2Node/utils"
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errAnalyticsNotFound = errors.New("analytics result not found")

type analyticsQueryItem struct {
	Task      *paradigm.Task
	TaskID    string
	TaskName  string
	StockID   string
	StockCode string
	StockName string
	Date      string
	Data      interface{}
}

type AnalyticsQueryItem = analyticsQueryItem

// handleAnalyticsQuery 统一处理 4 种查询口径：
// 1. taskId + stockId -> 精确返回单股票分析结果
// 2. only taskId      -> 返回该平台任务下所有股票的分析结果列表
// 3. only stockId     -> 返回该股票最新一条可读取的分析结果
// 4. empty            -> 返回所有股票各自最新一条可读取的分析结果列表
func (e *HttpEngine) handleAnalyticsQuery(c *gin.Context, analType paradigm.AnalysisType, notFoundMsg, internalMsg, code string) {
	taskID := strings.TrimSpace(c.Query("taskId"))
	stockID := strings.TrimSpace(c.Query("stockId"))
	options := buildAnalyticsQueryOptions(c, analType)

	data, err := e.QueryAnalytics(taskID, stockID, analType, options)
	if err != nil {
		if errors.Is(err, errAnalyticsNotFound) {
			c.JSON(404, paradigm.HttpResponse{
				Message: notFoundMsg + ": " + err.Error(),
				Code:    code,
				Data:    nil,
			})
			return
		}
		paradigm.Log("ERROR", fmt.Sprintf("Failed to fetch %s: %v", analType.String(), err))
		c.JSON(500, paradigm.HttpResponse{
			Message: internalMsg + ": " + err.Error(),
			Code:    code,
			Data:    nil,
		})
		return
	}

	c.JSON(200, paradigm.HttpResponse{Message: "操作成功", Data: data, Code: "S000000"})
}

func normalizeInvestorCompositionResponse(data interface{}, selectedDate, selectedType string) interface{} {
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
			payload[index] = normalizeInvestorCompositionResponse(payload[index], selectedDate, normalizedType).(map[string]interface{})
		}
		return payload
	case []interface{}:
		for index := range payload {
			payload[index] = normalizeInvestorCompositionResponse(payload[index], selectedDate, normalizedType)
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

func (e *HttpEngine) QueryAnalytics(taskID, stockID string, analType paradigm.AnalysisType, options map[string]string) (interface{}, error) {
	switch {
	case taskID != "" && stockID != "":
		task, err := e.resolveTaskByPlatformTaskAndStock(taskID, stockID)
		if err != nil {
			return nil, err
		}
		payload, err := e.fetchNodeAnalyticsByTask(task, analType, options)
		if err != nil {
			return nil, err
		}
		if analType == paradigm.CrashRisk {
			payload = e.attachCrashRiskTopRiskListToPayload(taskID, task, payload)
		}
		return payload, nil
	case taskID != "":
		items, err := e.queryAnalyticsByTaskID(taskID, analType, options)
		if err != nil {
			return nil, err
		}
		if analType == paradigm.CrashRisk {
			items = attachCrashRiskTopRiskListToItems(items)
		}
		return e.wrapAnalyticsItems(items), nil
	case stockID != "":
		item, err := e.queryLatestAnalyticsByStockID(stockID, analType, options)
		if err != nil {
			return nil, err
		}
		if analType == paradigm.CrashRisk {
			if items, err := e.queryLatestAnalyticsForAllStocks(analType, options); err == nil {
				topRiskList := buildCrashRiskTopRiskList(items)
				item.Data = injectCrashRiskTopRiskList(item.Data, topRiskList)
			}
		}
		return e.wrapAnalyticsItem(item), nil
	default:
		items, err := e.queryLatestAnalyticsForAllStocks(analType, options)
		if err != nil {
			return nil, err
		}
		if analType == paradigm.CrashRisk {
			items = attachCrashRiskTopRiskListToItems(items)
		}
		return e.wrapAnalyticsItems(items), nil
	}
}

func (e *HttpEngine) attachCrashRiskTopRiskListToPayload(taskID string, task *paradigm.Task, payload interface{}) interface{} {
	var items []analyticsQueryItem
	switch {
	case taskID != "":
		if resolved, err := e.queryAnalyticsByTaskID(taskID, paradigm.CrashRisk, nil); err == nil {
			items = resolved
		}
	case task != nil && task.PlatformTaskID != nil && strings.TrimSpace(*task.PlatformTaskID) != "":
		if resolved, err := e.queryAnalyticsByTaskID(strings.TrimSpace(*task.PlatformTaskID), paradigm.CrashRisk, nil); err == nil {
			items = resolved
		}
	}

	if len(items) == 0 && task != nil {
		items = []analyticsQueryItem{e.buildAnalyticsItem(task, payload)}
	}
	return injectCrashRiskTopRiskList(payload, buildCrashRiskTopRiskList(items))
}

func attachCrashRiskTopRiskListToItems(items []analyticsQueryItem) []analyticsQueryItem {
	topRiskList := buildCrashRiskTopRiskList(items)
	for i := range items {
		items[i].Data = injectCrashRiskTopRiskList(items[i].Data, topRiskList)
	}
	return items
}

func AttachCrashRiskTopRiskListToItems(items []AnalyticsQueryItem) []AnalyticsQueryItem {
	return attachCrashRiskTopRiskListToItems(items)
}

func injectCrashRiskTopRiskList(data interface{}, topRiskList []map[string]interface{}) interface{} {
	payload, ok := data.(map[string]interface{})
	if !ok {
		return data
	}
	payload["topRiskList"] = topRiskList
	return payload
}

func buildCrashRiskTopRiskList(items []analyticsQueryItem) []map[string]interface{} {
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

func BuildCrashRiskTopRiskList(items []AnalyticsQueryItem) []map[string]interface{} {
	return buildCrashRiskTopRiskList(items)
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

func (e *HttpEngine) queryAnalyticsByTaskID(taskID string, analType paradigm.AnalysisType, options map[string]string) ([]analyticsQueryItem, error) {
	platformTask, err := e.dbService.GetPlatformTaskByID(taskID)
	if err != nil {
		return nil, err
	}
	if platformTask == nil {
		return nil, fmt.Errorf("%w: platform task %s not found", errAnalyticsNotFound, taskID)
	}

	subTasks := make([]*paradigm.Task, 0, len(platformTask.SubTasks))
	for i := range platformTask.SubTasks {
		task := platformTask.SubTasks[i]
		if task.Model != paradigm.ABM_V2 {
			continue
		}
		subTasks = append(subTasks, &task)
	}
	sort.SliceStable(subTasks, func(i, j int) bool {
		return extractTaskStockCode(subTasks[i]) < extractTaskStockCode(subTasks[j])
	})

	items := make([]analyticsQueryItem, 0, len(subTasks))
	for _, task := range subTasks {
		payload, err := e.fetchNodeAnalyticsByTask(task, analType, options)
		if err != nil {
			paradigm.Log("WARN", fmt.Sprintf("Skip unreadable analytics task %s: %v", task.Sign, err))
			continue
		}
		items = append(items, e.buildAnalyticsItem(task, payload))
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("%w: no readable analytics found under task %s", errAnalyticsNotFound, taskID)
	}
	return items, nil
}

func (e *HttpEngine) queryLatestAnalyticsByStockID(stockID string, analType paradigm.AnalysisType, options map[string]string) (analyticsQueryItem, error) {
	tasks, err := e.dbService.GetFinishedTasks()
	if err != nil {
		return analyticsQueryItem{}, err
	}

	candidates := make([]*paradigm.Task, 0)
	for _, task := range tasks {
		if matchTaskStock(task, stockID) {
			candidates = append(candidates, task)
		}
	}
	sortTasksByStartTimeDesc(candidates)

	for _, task := range candidates {
		payload, err := e.fetchNodeAnalyticsByTask(task, analType, options)
		if err != nil {
			paradigm.Log("WARN", fmt.Sprintf("Skip unreadable latest analytics task %s: %v", task.Sign, err))
			continue
		}
		return e.buildAnalyticsItem(task, payload), nil
	}
	return analyticsQueryItem{}, fmt.Errorf("%w: no readable analytics found for stock %s", errAnalyticsNotFound, stockID)
}

func (e *HttpEngine) queryLatestAnalyticsForAllStocks(analType paradigm.AnalysisType, options map[string]string) ([]analyticsQueryItem, error) {
	tasks, err := e.dbService.GetFinishedTasks()
	if err != nil {
		return nil, err
	}

	grouped := make(map[string][]*paradigm.Task)
	for _, task := range tasks {
		stockCode := extractTaskStockCode(task)
		if stockCode == "" {
			continue
		}
		grouped[stockCode] = append(grouped[stockCode], task)
	}

	stockCodes := make([]string, 0, len(grouped))
	for stockCode := range grouped {
		stockCodes = append(stockCodes, stockCode)
	}
	sort.Strings(stockCodes)

	items := make([]analyticsQueryItem, 0, len(stockCodes))
	for _, stockCode := range stockCodes {
		group := grouped[stockCode]
		sortTasksByStartTimeDesc(group)
		for _, task := range group {
			payload, err := e.fetchNodeAnalyticsByTask(task, analType, options)
			if err != nil {
				paradigm.Log("WARN", fmt.Sprintf("Skip unreadable analytics task %s for stock %s: %v", task.Sign, stockCode, err))
				continue
			}
			items = append(items, e.buildAnalyticsItem(task, payload))
			break
		}
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("%w: no readable analytics found", errAnalyticsNotFound)
	}
	return items, nil
}

func (e *HttpEngine) resolveTaskByPlatformTaskAndStock(taskID, stockID string) (*paradigm.Task, error) {
	platformTask, err := e.dbService.GetPlatformTaskByID(taskID)
	if err != nil {
		return nil, err
	}
	if platformTask == nil {
		// 兼容旧调用：允许直接传子任务 sign，或者继续按旧格式拼接一次。
		if task, err := e.dbService.GetTaskByID(taskID); err == nil && matchTaskStock(task, stockID) {
			return task, nil
		}
		task, err := e.dbService.GetTaskByID(fmt.Sprintf("SubTask-%s-%s", taskID, stockID))
		if err == nil {
			return task, nil
		}
		return nil, fmt.Errorf("%w: platform task %s not found", errAnalyticsNotFound, taskID)
	}

	for i := range platformTask.SubTasks {
		task := platformTask.SubTasks[i]
		if task.Model != paradigm.ABM_V2 {
			continue
		}
		if matchTaskStock(&task, stockID) {
			return &task, nil
		}
	}
	return nil, fmt.Errorf("%w: stock %s not found under task %s", errAnalyticsNotFound, stockID, taskID)
}

func (e *HttpEngine) buildAnalyticsItem(task *paradigm.Task, payload interface{}) analyticsQueryItem {
	item := analyticsQueryItem{
		Task:      task,
		TaskID:    displayTaskID(task),
		TaskName:  fmt.Sprintf("%s 风险监测", extractTaskStockName(task)),
		StockID:   extractTaskStockID(task),
		StockCode: extractTaskStockCode(task),
		StockName: extractTaskStockName(task),
		Date:      task.StartTime.Format("2006-01-02"),
		Data:      payload,
	}

	if task.PlatformTaskID != nil {
		pt, err := e.dbService.GetPlatformTaskByID(*task.PlatformTaskID)
		if err == nil && pt != nil && strings.TrimSpace(pt.TaskName) != "" {
			item.TaskName = pt.TaskName
		}
	}
	if item.StockID == "" {
		item.StockID = item.StockCode
	}
	return item
}

func (e *HttpEngine) wrapAnalyticsItem(item analyticsQueryItem) map[string]interface{} {
	return map[string]interface{}{
		"taskId":    item.TaskID,
		"taskName":  item.TaskName,
		"stockId":   item.StockID,
		"stockCode": item.StockCode,
		"stockName": item.StockName,
		"date":      item.Date,
		"data":      item.Data,
	}
}

func (e *HttpEngine) wrapAnalyticsItems(items []analyticsQueryItem) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		result = append(result, e.wrapAnalyticsItem(item))
	}
	return result
}

func (e *HttpEngine) fetchNodeAnalyticsByTask(task *paradigm.Task, analType paradigm.AnalysisType, options map[string]string) (interface{}, error) {
	if task == nil {
		return nil, fmt.Errorf("%w: empty task", errAnalyticsNotFound)
	}

	nodeID, err := e.resolveAnalyticsNodeID(task)
	if err != nil {
		return nil, err
	}

	conn, err := e.grpcManager.GetConn(nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node %d connection: %v", nodeID, err)
	}

	client := service.NewRappaExecutorClient(conn)
	resp, err := client.GetAnalytics(context.Background(), &service.AnalyticalRequest{
		Sign:         task.Sign,
		AnalysisType: encodeAnalysisTypeRequest(analType, options),
	})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			return nil, fmt.Errorf("%w: %s", errAnalyticsNotFound, st.Message())
		}
		return nil, fmt.Errorf("grpc GetAnalytics error: %v", err)
	}

	if resp.Data != nil {
		return resp.Data.AsMap(), nil
	}
	return nil, fmt.Errorf("%w: empty analytics response for %s", errAnalyticsNotFound, task.Sign)
}

func buildAnalyticsQueryOptions(c *gin.Context, analType paradigm.AnalysisType) map[string]string {
	options := map[string]string{}
	switch analType {
	case paradigm.PerformanceComparison:
		if model := strings.ToUpper(strings.TrimSpace(c.Query("selectedModel"))); model != "" {
			options["selectedModel"] = model
		}
	case paradigm.OrderDynamics:
		if date := strings.TrimSpace(c.Query("date")); date != "" {
			options["date"] = date
		}
	}
	return options
}

func encodeAnalysisTypeRequest(analType paradigm.AnalysisType, options map[string]string) string {
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

func EncodeAnalysisTypeRequest(analType paradigm.AnalysisType, options map[string]string) string {
	return encodeAnalysisTypeRequest(analType, options)
}

func sortTasksByStartTimeDesc(tasks []*paradigm.Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		return tasks[i].StartTime.After(tasks[j].StartTime)
	})
}

func displayTaskID(task *paradigm.Task) string {
	if task == nil {
		return ""
	}
	if task.PlatformTaskID != nil && strings.TrimSpace(*task.PlatformTaskID) != "" {
		return strings.TrimSpace(*task.PlatformTaskID)
	}
	return task.Sign
}

func extractTaskStockCode(task *paradigm.Task) string {
	if task == nil {
		return ""
	}
	if code := strings.TrimSpace(stringifyTaskParam(task.Params["stockCode"])); code != "" {
		return code
	}
	if code := strings.TrimSpace(stringifyTaskParam(task.Params["stockId"])); code != "" {
		return code
	}
	parts := strings.Split(task.Sign, "-")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func extractTaskStockID(task *paradigm.Task) string {
	if task == nil {
		return ""
	}
	if stockID := strings.TrimSpace(stringifyTaskParam(task.Params["stockId"])); stockID != "" {
		return stockID
	}
	return extractTaskStockCode(task)
}

func extractTaskStockName(task *paradigm.Task) string {
	if task == nil {
		return ""
	}
	return abm.ResolveStockDisplayName(
		extractTaskStockCode(task),
		stringifyTaskParam(task.Params["stockName"]),
	)
}

func ExtractTaskStockName(task *paradigm.Task) string {
	return extractTaskStockName(task)
}

func stringifyTaskParam(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

func matchTaskStock(task *paradigm.Task, stockID string) bool {
	target := strings.TrimSpace(strings.ToLower(stockID))
	if target == "" {
		return false
	}
	return strings.EqualFold(extractTaskStockID(task), target) || strings.EqualFold(extractTaskStockCode(task), target)
}

func (e *HttpEngine) resolveAnalyticsNodeID(task *paradigm.Task) (int, error) {
	if task != nil {
		slots := e.dbService.QueryFinishedSlotsByTask(task.Sign)
		if len(slots) > 0 {
			sort.Slice(slots, func(i, j int) bool {
				if slots[i].ScheduleID == slots[j].ScheduleID {
					return slots[i].SlotID > slots[j].SlotID
				}
				return slots[i].ScheduleID > slots[j].ScheduleID
			})
			return int(slots[0].NodeID), nil
		}
	}

	if task == nil {
		return 0, fmt.Errorf("finished slot node_id not found for empty task")
	}
	if nodeID, ok := utils.ExtractAssignedNodeID(task.Params); ok {
		return int(nodeID), nil
	}

	return 0, fmt.Errorf("finished slot node_id not found for task %s", task.Sign)
}
