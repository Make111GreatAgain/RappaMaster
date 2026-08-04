package HTTP

import (
	"BHLayer2Node/Network/HTTP/abm"
	"BHLayer2Node/paradigm"
	"BHLayer2Node/pb/service"
	"BHLayer2Node/utils"
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errAnalyticsNotFound = abm.ErrAnalyticsNotFound

type AnalyticsQueryItem = abm.AnalyticsQueryItem

// handleAnalyticsQuery 统一处理 4 种查询口径：
// 1. taskId + stockId -> 精确返回单股票分析结果
// 2. only taskId      -> 返回该平台任务下所有股票的分析结果列表
// 3. only stockId     -> 返回该股票最新一条可读取的分析结果
// 4. empty            -> 返回所有股票各自最新一条可读取的分析结果列表
func (e *HttpEngine) handleAnalyticsQuery(c *gin.Context, analType paradigm.AnalysisType, notFoundMsg, internalMsg, code string) {
	taskID := strings.TrimSpace(c.Query("taskId"))
	stockID := strings.TrimSpace(c.Query("stockId"))
	options := abm.BuildAnalyticsQueryOptions(c.Query, analType)

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

	if analType == paradigm.PerformanceComparison {
		data = abm.HideUnstablePerformanceMetrics(data)
	}

	c.JSON(200, paradigm.HttpResponse{Message: "操作成功", Data: data, Code: "S000000"})
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
			items = abm.AttachCrashRiskTopRiskListToItems(items)
		}
		return e.wrapAnalyticsItems(items), nil
	case stockID != "":
		item, err := e.queryLatestAnalyticsByStockID(stockID, analType, options)
		if err != nil {
			return nil, err
		}
		if analType == paradigm.CrashRisk {
			if items, err := e.queryLatestAnalyticsForAllStocks(analType, options); err == nil {
				topRiskList := abm.BuildCrashRiskTopRiskList(items)
				item.Data = abm.InjectCrashRiskTopRiskList(item.Data, topRiskList)
			}
		}
		return e.wrapAnalyticsItem(item), nil
	default:
		items, err := e.queryLatestAnalyticsForAllStocks(analType, options)
		if err != nil {
			return nil, err
		}
		if analType == paradigm.CrashRisk {
			items = abm.AttachCrashRiskTopRiskListToItems(items)
		}
		return e.wrapAnalyticsItems(items), nil
	}
}

func (e *HttpEngine) attachCrashRiskTopRiskListToPayload(taskID string, task *paradigm.Task, payload interface{}) interface{} {
	var items []abm.AnalyticsQueryItem
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
		items = []abm.AnalyticsQueryItem{e.buildAnalyticsItem(task, payload)}
	}
	return abm.InjectCrashRiskTopRiskList(payload, abm.BuildCrashRiskTopRiskList(items))
}

func (e *HttpEngine) queryAnalyticsByTaskID(taskID string, analType paradigm.AnalysisType, options map[string]string) ([]abm.AnalyticsQueryItem, error) {
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
		return abm.ExtractTaskStockCode(subTasks[i]) < abm.ExtractTaskStockCode(subTasks[j])
	})

	items := make([]abm.AnalyticsQueryItem, 0, len(subTasks))
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

func (e *HttpEngine) queryLatestAnalyticsByStockID(stockID string, analType paradigm.AnalysisType, options map[string]string) (abm.AnalyticsQueryItem, error) {
	tasks, err := e.dbService.GetFinishedTasks()
	if err != nil {
		return abm.AnalyticsQueryItem{}, err
	}

	candidates := make([]*paradigm.Task, 0)
	for _, task := range tasks {
		if abm.MatchTaskStock(task, stockID) {
			candidates = append(candidates, task)
		}
	}
	abm.SortTasksByStartTimeDesc(candidates)

	for _, task := range candidates {
		payload, err := e.fetchNodeAnalyticsByTask(task, analType, options)
		if err != nil {
			paradigm.Log("WARN", fmt.Sprintf("Skip unreadable latest analytics task %s: %v", task.Sign, err))
			continue
		}
		return e.buildAnalyticsItem(task, payload), nil
	}
	return abm.AnalyticsQueryItem{}, fmt.Errorf("%w: no readable analytics found for stock %s", errAnalyticsNotFound, stockID)
}

func (e *HttpEngine) queryLatestAnalyticsForAllStocks(analType paradigm.AnalysisType, options map[string]string) ([]abm.AnalyticsQueryItem, error) {
	tasks, err := e.dbService.GetFinishedTasks()
	if err != nil {
		return nil, err
	}

	grouped := make(map[string][]*paradigm.Task)
	for _, task := range tasks {
		stockCode := abm.ExtractTaskStockCode(task)
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

	items := make([]abm.AnalyticsQueryItem, 0, len(stockCodes))
	for _, stockCode := range stockCodes {
		group := grouped[stockCode]
		abm.SortTasksByStartTimeDesc(group)
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
		if task, err := e.dbService.GetTaskByID(taskID); err == nil && abm.MatchTaskStock(task, stockID) {
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
		if abm.MatchTaskStock(&task, stockID) {
			return &task, nil
		}
	}
	return nil, fmt.Errorf("%w: stock %s not found under task %s", errAnalyticsNotFound, stockID, taskID)
}

func (e *HttpEngine) buildAnalyticsItem(task *paradigm.Task, payload interface{}) abm.AnalyticsQueryItem {
	item := abm.AnalyticsQueryItem{
		Task:      task,
		TaskID:    abm.DisplayTaskID(task),
		TaskName:  fmt.Sprintf("%s 风险监测", abm.ExtractTaskStockName(task)),
		StockID:   abm.ExtractTaskStockID(task),
		StockCode: abm.ExtractTaskStockCode(task),
		StockName: abm.ExtractTaskStockName(task),
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

func (e *HttpEngine) wrapAnalyticsItem(item abm.AnalyticsQueryItem) map[string]interface{} {
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

func (e *HttpEngine) wrapAnalyticsItems(items []abm.AnalyticsQueryItem) []map[string]interface{} {
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
		AnalysisType: abm.EncodeAnalysisTypeRequest(analType, options),
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

func AttachCrashRiskTopRiskListToItems(items []AnalyticsQueryItem) []AnalyticsQueryItem {
	return abm.AttachCrashRiskTopRiskListToItems(items)
}

func BuildCrashRiskTopRiskList(items []AnalyticsQueryItem) []map[string]interface{} {
	return abm.BuildCrashRiskTopRiskList(items)
}

func EncodeAnalysisTypeRequest(analType paradigm.AnalysisType, options map[string]string) string {
	return abm.EncodeAnalysisTypeRequest(analType, options)
}

func ExtractTaskStockName(task *paradigm.Task) string {
	return abm.ExtractTaskStockName(task)
}

func stringifyTaskParam(value interface{}) string {
	return abm.StringifyTaskParam(value)
}
