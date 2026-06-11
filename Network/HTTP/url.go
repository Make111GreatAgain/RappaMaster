package HTTP

import (
	"BHLayer2Node/Collector"
	"BHLayer2Node/Network/HTTP/abm"
	"BHLayer2Node/Query"
	"BHLayer2Node/paradigm"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type HttpService struct {
	Url     string
	Method  string
	Handler func(c *gin.Context)
}

//func (s *HttpService) HandleRequest(request paradigm.HttpRequestInterface) paradigm.HttpResponse {
//	return s.Handler(request)
//}

type HttpServiceEnum int

const (
	INIT_TASK = iota
	ORACLE_QUERY
	COLLECT_TASK
	UPLOAD_TASK
	BLOCKCHAIN_QUERY
	DATASYNTH_QUERY
	EXECUTION_LOG
	EXECUTION_LOG_TASK
	CREATE_SIM_TASK
	ANALYZED_STOCKS
	ABM_PARAMETERS
	ORDER_DYNAMICS
	PRICE_SYNTH_DOWNLOAD
	PRICE_SYNTH
	CRASH_RISK
	INVESTOR_COMP
	PERF_COMPARISON
	PLATFORM_TASK_DOWNLOAD
	ABM_PARAMETERS_REFRESH
	LATEST_MARKET_TASK
	DEFAULT_MARKET_TASK
)

func (e *HttpEngine) SupportUrl() []HttpServiceEnum {
	return []HttpServiceEnum{INIT_TASK, ORACLE_QUERY, BLOCKCHAIN_QUERY, DATASYNTH_QUERY, COLLECT_TASK, UPLOAD_TASK, EXECUTION_LOG, EXECUTION_LOG_TASK, CREATE_SIM_TASK, ANALYZED_STOCKS, LATEST_MARKET_TASK, DEFAULT_MARKET_TASK, ABM_PARAMETERS, ORDER_DYNAMICS, PRICE_SYNTH_DOWNLOAD, PRICE_SYNTH, CRASH_RISK, INVESTOR_COMP, PERF_COMPARISON, PLATFORM_TASK_DOWNLOAD, ABM_PARAMETERS_REFRESH}
}
func (e *HttpEngine) HandleGET(c *gin.Context) {
	var requestBody Query.HttpOracleQueryRequest
	// 解析请求体中的 JSON 数据

	if success, query := requestBody.BuildQueryFromGETRequest(c); success {
		//fmt.Println(query.ToHttpJson())
		e.channel.QueryChannel <- query
		r := query.ReceiveResponse() // 这里会阻塞

		if r.Error() != "" {
			c.JSON(http.StatusInternalServerError, paradigm.HttpResponse{
				Message: r.Error(),
				Code:    "ERROR",
				Data:    r.ToHttpJson(),
			})
			return
		}

		response := paradigm.HttpResponse{
			Message: fmt.Sprintf("Query Data Successfully, query type: %s, query: %v", requestBody.Query, requestBody.Data),
			Code:    "OK",
			Data:    r.ToHttpJson(),
		}
		c.JSON(http.StatusOK, response)
		//fmt.Println(r.ToHttpJson(), r.Error())
	} else {
		c.JSON(http.StatusBadRequest, paradigm.HttpResponse{
			Message: "Invalid Request data",
			Code:    "ERROR",
			Data:    nil,
		})
	}
}

// HandleSimulationDownload 复用已有下载功能的逻辑
func (e *HttpEngine) HandleSimulationDownload(c *gin.Context) {
	taskId := c.Query("taskId")
	stockId := c.Query("stockId")
	if taskId == "" || stockId == "" {
		c.JSON(http.StatusBadRequest, paradigm.HttpResponse{Message: "taskId and stockId required", Code: "E100005", Data: nil})
		return
	}

	// 映射到系统的 Sign (SubTask-taskId-stockId)
	sign := fmt.Sprintf("SubTask-%s-%s", taskId, stockId)

	// 获取任务信息来确定Size
	task, err := e.dbService.GetTaskByID(sign)
	if err != nil || task == nil {
		c.JSON(http.StatusNotFound, paradigm.HttpResponse{Message: "任务未找到", Code: "E100006", Data: nil})
		return
	}

	// 构造 CollectTaskQuery 并直接调用 HandleDownload 的核心逻辑
	query := Query.NewCollectTaskQuery(map[interface{}]interface{}{
		"taskID": sign,
		"size":   int(task.Size),
	})

	e.channel.QueryChannel <- query
	r := query.ReceiveResponse() // 会阻塞
	fileJson := r.ToHttpJson()
	interfaceReader, ok := fileJson["fileReader"]
	if !ok || interfaceReader == nil {
		c.JSON(http.StatusInternalServerError, paradigm.HttpResponse{Message: "结果文件尚未准备好", Code: "E100007", Data: nil})
		return
	}

	reader := interfaceReader.(io.Reader)
	filename := fileJson["filename"].(string)

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.DataFromReader(http.StatusOK, -1, "application/octet-stream", reader, nil)
}

func (e *HttpEngine) HandleDownload(c *gin.Context) {
	var requestBody Query.HttpOracleQueryRequest
	// 解析请求体中的 JSON 数据
	if success, query := requestBody.BuildQueryFromGETRequest(c); success {
		//fmt.Println(query.ToHttpJson())
		e.channel.QueryChannel <- query
		r := query.ReceiveResponse() // 这里会阻塞
		fileJson := r.ToHttpJson()
		reader := fileJson["fileReader"].(*io.PipeReader) // 这里改成流式reader
		filename := fileJson["filename"].(string)
		// 将 byte 数组转换为 Reader
		//reader := bytes.NewReader(data)
		// 设置响应头
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		// 流式传输数据
		c.DataFromReader(
			http.StatusOK,
			-1, // 内容长度未知，使用 -1
			"application/octet-stream",
			reader, // 流传输
			nil,    // 可选的额外 headers
		)
	} else {
		c.JSON(http.StatusBadRequest, paradigm.HttpResponse{
			Message: "Invalid Request data",
			Code:    "ERROR",
			Data:    nil,
		})
	}
}

func matchAnalyzedStockFilter(searchType, keyword string, stock map[string]interface{}) bool {
	if searchType == "" || keyword == "" {
		return true
	}

	lowerKeyword := strings.ToLower(strings.TrimSpace(keyword))
	match := func(value interface{}) bool {
		return strings.Contains(strings.ToLower(fmt.Sprintf("%v", value)), lowerKeyword)
	}

	switch searchType {
	case "stockCode":
		return match(stock["stockCode"])
	case "stockName":
		return match(stock["stockName"])
	case "task":
		return match(stock["taskId"]) || match(stock["taskName"])
	default:
		return true
	}
}

func BuildLatestMarketTaskResponse(platformTask *paradigm.PlatformTask) map[string]interface{} {
	if platformTask == nil {
		return nil
	}

	selected := selectDefaultLatestMarketSubTask(platformTask.SubTasks)
	if selected == nil {
		return nil
	}

	stockCode := strings.TrimSpace(abm.ExtractTaskStockCode(selected))
	if stockCode == "" {
		return nil
	}
	stockID := strings.TrimSpace(abm.ExtractTaskStockID(selected))
	if stockID == "" {
		stockID = stockCode
	}
	stockName := abm.ExtractTaskStockName(selected)
	taskName := strings.TrimSpace(platformTask.TaskName)
	if taskName == "" {
		taskName = fmt.Sprintf("%v 风险监测", stockName)
	}

	return map[string]interface{}{
		"stockId":   stockID,
		"stockCode": stockCode,
		"stockName": stockName,
		"label":     fmt.Sprintf("%v %v", stockCode, stockName),
		"taskId":    platformTask.ID,
		"taskName":  taskName,
		"date":      selected.StartTime.Format("2006-01-02"),
	}
}

func BuildDefaultMarketTaskResponse() map[string]interface{} {
	stockCode := "600028"
	stockName := abm.ResolveStockDisplayName(stockCode, stockCode)
	return map[string]interface{}{
		"stockId":   stockCode,
		"stockCode": stockCode,
		"stockName": stockName,
		"label":     fmt.Sprintf("%v %v", stockCode, stockName),
		"taskId":    "TSK-1001",
		"taskName":  "平台任务申报",
		"date":      "2026-05-14",
	}
}

func selectDefaultLatestMarketSubTask(tasks []paradigm.Task) *paradigm.Task {
	candidates := make([]*paradigm.Task, 0, len(tasks))
	for i := range tasks {
		task := &tasks[i]
		if task.Status != paradigm.Finished || task.Model != paradigm.ABM_V2 {
			continue
		}
		if strings.TrimSpace(abm.ExtractTaskStockCode(task)) == "" {
			continue
		}
		candidates = append(candidates, task)
	}
	if len(candidates) == 0 {
		return nil
	}

	for _, task := range candidates {
		if abm.ExtractTaskStockCode(task) == "600000" {
			return task
		}
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return abm.ExtractTaskStockCode(candidates[i]) < abm.ExtractTaskStockCode(candidates[j])
	})
	return candidates[0]
}

func (e *HttpEngine) GetHttpService(service HttpServiceEnum) (*HttpService, error) {
	switch service {
	case INIT_TASK:
		// 初始化任务
		httpService := HttpService{
			Url:    "/create",
			Method: "POST",
			Handler: func(c *gin.Context) {
				var requestBody Query.HttpInitTaskRequest

				// 解析请求体中的 JSON 数据
				if err := c.ShouldBindJSON(&requestBody); err != nil {
					// 如果解析失败，返回错误信息
					c.JSON(http.StatusBadRequest, paradigm.HttpResponse{
						Message: "Invalid Request data",
						Code:    "ERROR",
						Data:    nil,
					})
					return
				}

				e.taskIDConsumer <- 1        // 获取ID
				taskID := <-e.taskIDProvider // 这里要阻塞

				//如果没有指定，设置默认值
				if requestBody.SlotSize == 0 {
					requestBody.SlotSize = 3000
				}

				task := paradigm.NewTask(
					taskID,
					requestBody.Name,
					paradigm.NameToModelType(requestBody.Model),
					requestBody.SlotSize,
					requestBody.Params,
					requestBody.Size,
					requestBody.IsReliable,
				)
				task.SetCollector(Collector.NewCollector(task.Sign, task.OutputType, e.channel, e.pkiManager))
				paradigm.Log("HTTP", fmt.Sprintf("Receive Init Task Request: %v, Generate New Task: %s", requestBody, task.Sign))
				// 新建任务用于上链，然后直接返回response
				e.channel.PendingTransactions <- &paradigm.InitTaskTransaction{Task: task}
				// 返回response
				response := paradigm.HttpResponse{
					Message: fmt.Sprintf("Create New SynthTask Successfully, taskID: %s", task.Sign),
					Code:    "OK",
					Data:    nil,
				}
				c.JSON(http.StatusOK, response)

			},
		}
		return &httpService, nil
	case ORACLE_QUERY:
		httpService := HttpService{
			Url:     "/oracle",
			Method:  "GET",
			Handler: e.HandleGET,
		}
		return &httpService, nil
	case COLLECT_TASK:
		httpService := HttpService{
			Url:     "/collect",
			Method:  "GET",
			Handler: e.HandleDownload,
		}
		return &httpService, nil
	case UPLOAD_TASK:
		httpService := HttpService{
			Url:     "/upload",
			Method:  "GET",
			Handler: e.HandleGET,
		}
		return &httpService, nil
	case BLOCKCHAIN_QUERY:
		httpService := HttpService{
			Url:     "/blockchain",
			Method:  "GET",
			Handler: e.HandleGET,
		}
		return &httpService, nil
	case DATASYNTH_QUERY:
		httpService := HttpService{
			Url:     "/dataSynth",
			Method:  "GET",
			Handler: e.HandleGET,
		}
		return &httpService, nil
	case EXECUTION_LOG:
		httpService := HttpService{
			Url:    "/dashboard/execution_log",
			Method: "GET",
			Handler: func(c *gin.Context) {
				tasks, err := e.dbService.GetAllPlatformTasks()
				if err != nil {
					c.JSON(http.StatusInternalServerError, paradigm.HttpResponse{
						Message: "获取执行日志失败",
						Code:    "E100001",
						Data:    nil,
					})
					return
				}

				c.JSON(http.StatusOK, paradigm.HttpResponse{
					Message: "操作成功",
					Data:    buildExecutionLogViews(tasks),
					Code:    "S000000",
				})
			},
		}
		return &httpService, nil
	case EXECUTION_LOG_TASK:
		httpService := HttpService{
			Url:    "/dashboard/execution_log/task",
			Method: "GET",
			Handler: func(c *gin.Context) {
				taskID := strings.TrimSpace(c.Query("taskId"))
				if taskID == "" {
					c.JSON(http.StatusBadRequest, paradigm.HttpResponse{
						Message: "taskId is required",
						Code:    "E100015",
						Data:    nil,
					})
					return
				}

				task, err := e.dbService.GetPlatformTaskByID(taskID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, paradigm.HttpResponse{
						Message: "获取平台任务失败",
						Code:    "E100016",
						Data:    nil,
					})
					return
				}
				if task == nil {
					c.JSON(http.StatusNotFound, paradigm.HttpResponse{
						Message: "平台任务不存在",
						Code:    "E100017",
						Data:    nil,
					})
					return
				}

				c.JSON(http.StatusOK, paradigm.HttpResponse{
					Message: "操作成功",
					Data:    buildExecutionLogView(task),
					Code:    "S000000",
				})
			},
		}
		return &httpService, nil
	case CREATE_SIM_TASK:
		httpService := HttpService{
			Url:    "/simulation/create-task",
			Method: "POST",
			Handler: func(c *gin.Context) {
				isScheduled := abm.IsScheduledCreateTask(c)
				var rawTasks []map[string]interface{}
				if isScheduled {
					var err error
					rawTasks, err = abm.BuildScheduledV2RawTasks(&e.config)
					if err != nil {
						c.JSON(http.StatusInternalServerError, paradigm.HttpResponse{
							Message: "定时任务参数构造失败: " + err.Error(),
							Code:    "E100018",
							Data:    false,
						})
						return
					}
				} else if err := c.ShouldBindJSON(&rawTasks); err != nil {
					c.JSON(http.StatusBadRequest, paradigm.HttpResponse{
						Message: "参数格式错误",
						Code:    "E100002",
						Data:    false,
					})
					return
				}

				// 生成ID
				taskID, _ := e.dbService.GetNextPlatformTaskID()

				var subTasks []paradigm.Task
				for _, raw := range rawTasks {
					taskSize := int32(1)
					stockCode := abm.NormalizeStockCode(stringifyTaskParam(raw["stockCode"]))
					if !isScheduled && !abm.IsStockSupportedByIndex(stockCode) {
						c.JSON(http.StatusBadRequest, paradigm.HttpResponse{
							Message: fmt.Sprintf("ABM_V2 参数错误: stockCode %s is not supported because input csv is missing", stockCode),
							Code:    "E100019",
							Data:    false,
						})
						return
					}
					// ABM_V2 不在创建阶段预分配节点；Scheduler 每轮根据 Monitor 当前负载动态选择节点。
					// 最终产出节点以 slots 表中 Finished slot 的 node_id 为准，供分析接口定位。
					taskParams, err := abm.BuildV2TaskParamsWithConfig(raw, -1, &e.config)
					if err != nil {
						c.JSON(http.StatusBadRequest, paradigm.HttpResponse{
							Message: "ABM_V2 参数错误: " + err.Error(),
							Code:    "E100002",
							Data:    false,
						})
						return
					}

					// 复用 Task 结构
					task := paradigm.Task{
						Sign:        fmt.Sprintf("SubTask-%s-%s", taskID, taskParams["stockCode"]),
						Name:        fmt.Sprintf("%s推演", taskParams["stockName"]),
						Slot:        1,
						Model:       paradigm.ABM_V2,
						Params:      taskParams,
						Size:        taskSize,
						Process:     0,
						OutputType:  paradigm.DATAFRAME,
						Schedules:   make([]*paradigm.SynthTaskSchedule, 0),
						ScheduleMap: make(map[paradigm.ScheduleHash]int),
						Status:      paradigm.Processing,
						StartTime:   time.Now(),
					}
					task.SetCollector(Collector.NewCollector(task.Sign, task.OutputType, e.channel, e.pkiManager))

					subTasks = append(subTasks, task)
				}

				platformTask := &paradigm.PlatformTask{
					ID:          taskID,
					TaskName:    "平台任务申报",
					SubTasks:    subTasks,
					IsScheduled: isScheduled,
					UserName:    executionLogUserName(isScheduled, rawTasks),
					Status:      "running",
					CreatedAt:   time.Now(),
				}

				if isScheduled {
					platformTask.ExecutionType = "定时任务"
				} else {
					platformTask.ExecutionType = "即时任务"
				}

				// 汇总参数描述
				if len(subTasks) > 0 {
					platformTask.Parameters = fmt.Sprintf("Stock: %s (%s); Horizon: %v", subTasks[0].Params["stockName"], subTasks[0].Params["stockCode"], subTasks[0].Params["horizon"])
					if len(subTasks) > 1 {
						platformTask.Parameters += fmt.Sprintf("... (共%d个子任务)", len(subTasks))
					}
				}

				err := e.dbService.SetPlatformTask(platformTask)
				if err != nil {
					c.JSON(http.StatusInternalServerError, paradigm.HttpResponse{
						Message: "任务持久化失败",
						Code:    "E100003",
						Data:    false,
					})
					return
				}

				// 平台子任务也走普通任务的上链初始化流程：
				// 1. 先记录到 platform_tasks/tasks，保证 execution_log 能立即看到任务
				// 2. 再发起 InitTaskTransaction，由 Oracle 在链上确认后统一进入 Tracker/Scheduler
				// 这样后续的进度推进、下载、分析、平台状态刷新都复用原有链路。
				for _, subTask := range subTasks {
					t := subTask
					e.channel.PendingTransactions <- &paradigm.InitTaskTransaction{Task: &t}
				}

				c.JSON(http.StatusOK, paradigm.HttpResponse{
					Message: "操作成功",
					Data: map[string]interface{}{
						"taskId": taskID,
					},
					Code: "S000000",
				})
			},
		}
		return &httpService, nil
	case ANALYZED_STOCKS:
		httpService := HttpService{
			Url:    "/market/analyzed_stocks",
			Method: "GET",
			Handler: func(c *gin.Context) {
				searchType := strings.TrimSpace(c.Query("searchType"))
				keyword := strings.TrimSpace(c.Query("keyword"))
				if searchType != "" && searchType != "stockCode" && searchType != "stockName" && searchType != "task" {
					c.JSON(http.StatusBadRequest, paradigm.HttpResponse{
						Message: "searchType must be one of stockCode, stockName, task",
						Code:    "E100014",
						Data:    nil,
					})
					return
				}

				tasks, err := e.dbService.GetFinishedTasks()
				if err != nil {
					c.JSON(http.StatusInternalServerError, paradigm.HttpResponse{
						Message: "获取分析已完成任务失败",
						Code:    "E100004",
						Data:    nil,
					})
					return
				}

				result := make([]map[string]interface{}, 0)
				for _, t := range tasks {
					p := t.Params
					stockCode := strings.TrimSpace(stringifyTaskParam(p["stockCode"]))
					if stockCode == "" {
						// 历史脏数据或非平台分析任务不应出现在股票分析列表里，直接跳过。
						continue
					}
					stockName := abm.ResolveStockDisplayName(stockCode, stringifyTaskParam(p["stockName"]))
					stockID := p["stockId"]
					if stockID == nil {
						// ABM_V2 当前请求体只稳定传 stockCode，这里回退保证前端有可用主键。
						stockID = stockCode
					}
					taskID := t.Sign
					if t.PlatformTaskID != nil && strings.TrimSpace(*t.PlatformTaskID) != "" {
						taskID = *t.PlatformTaskID
					}
					stock := map[string]interface{}{
						"stockId":   stockID,
						"stockCode": stockCode,
						"stockName": stockName,
						"label":     fmt.Sprintf("%v %v", stockCode, stockName),
						"taskId":    taskID,
						"taskName":  fmt.Sprintf("%v 风险监测", stockName), // 如果没有platform task name就拼一下
						"date":      t.StartTime.Format("2006-01-02"),
					}

					// 如果有关联的平台任务，可以尝试获取它的名字
					if t.PlatformTaskID != nil {
						pt, err := e.dbService.GetPlatformTaskByID(*t.PlatformTaskID)
						if err == nil && pt != nil {
							stock["taskName"] = pt.TaskName
						}
					}

					if !matchAnalyzedStockFilter(searchType, keyword, stock) {
						continue
					}

					result = append(result, stock)
				}

				c.JSON(http.StatusOK, paradigm.HttpResponse{
					Message: "操作成功",
					Data:    result,
					Code:    "S000000",
				})
			},
		}
		return &httpService, nil
	case LATEST_MARKET_TASK:
		httpService := HttpService{
			Url:    "/market/latest_task",
			Method: "GET",
			Handler: func(c *gin.Context) {
				task, err := e.dbService.GetLatestFinishedABMPlatformTask()
				if err != nil {
					c.JSON(http.StatusInternalServerError, paradigm.HttpResponse{
						Message: "获取最新分析任务失败",
						Code:    "E100004",
						Data:    nil,
					})
					return
				}

				c.JSON(http.StatusOK, paradigm.HttpResponse{
					Message: "操作成功",
					Data:    BuildLatestMarketTaskResponse(task),
					Code:    "S000000",
				})
			},
		}
		return &httpService, nil
	case DEFAULT_MARKET_TASK:
		httpService := HttpService{
			Url:    "/market/default_task",
			Method: "GET",
			Handler: func(c *gin.Context) {
				c.JSON(http.StatusOK, paradigm.HttpResponse{
					Message: "操作成功",
					Data:    BuildDefaultMarketTaskResponse(),
					Code:    "S000000",
				})
			},
		}
		return &httpService, nil
	case ABM_PARAMETERS:
		httpService := HttpService{
			Url:    "/simulation/abm_parameters",
			Method: "GET",
			Handler: func(c *gin.Context) {
				stockCode := strings.TrimSpace(c.Query("stockCode"))
				if stockCode == "" {
					stockCode = strings.TrimSpace(c.Query("stockId"))
				}
				index := abm.CurrentABMParameterIndex()
				if stockCode != "" {
					data := abm.BuildABMSingleStockDetail(e.config.AbmParameters, index, stockCode)
					c.JSON(http.StatusOK, paradigm.HttpResponse{
						Message: "操作成功",
						Data:    data,
						Code:    "S000000",
					})
					return
				}
				pageNo, _ := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("pageNo", "1")))
				pageSize, _ := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("pageSize", "50")))
				keyword := strings.TrimSpace(c.Query("keyword"))

				c.JSON(http.StatusOK, paradigm.HttpResponse{
					Message: "操作成功",
					Data:    abm.BuildABMParameterListResponse(index, pageNo, pageSize, keyword),
					Code:    "S000000",
				})
			},
		}
		return &httpService, nil
	case ABM_PARAMETERS_REFRESH:
		httpService := HttpService{
			Url:    "/internal/simulation/abm_parameters/refresh",
			Method: "POST",
			Handler: func(c *gin.Context) {
				result, err := abm.RefreshABMParameterIndex(e.config.AbmParameters, &e.config)
				if err != nil {
					c.JSON(http.StatusInternalServerError, paradigm.HttpResponse{
						Message: "刷新 ABM 参数索引失败: " + err.Error(),
						Code:    "E100021",
						Data:    result,
					})
					return
				}
				c.JSON(http.StatusOK, paradigm.HttpResponse{
					Message: "操作成功",
					Data:    result,
					Code:    "S000000",
				})
			},
		}
		return &httpService, nil
	case ORDER_DYNAMICS:
		httpService := HttpService{
			Url:    "/dashboard/order_dynamics",
			Method: "GET",
			Handler: func(c *gin.Context) {
				e.handleAnalyticsQuery(c, paradigm.OrderDynamics, "未找到订单态势分析结果", "获取实时订单态势失败", "E100009")
			},
		}
		return &httpService, nil
	case PRICE_SYNTH_DOWNLOAD:
		httpService := HttpService{
			Url:     "/dashboard/price_synthesis/download",
			Method:  "GET",
			Handler: e.HandleSimulationDownload,
		}
		return &httpService, nil
	case PRICE_SYNTH:
		httpService := HttpService{
			Url:    "/dashboard/price_synthesis",
			Method: "GET",
			Handler: func(c *gin.Context) {
				e.handleAnalyticsQuery(c, paradigm.PriceSynthesis, "未找到价格合成分析结果", "获取价格合成分析失败", "E100010")
			},
		}
		return &httpService, nil
	case CRASH_RISK:
		httpService := HttpService{
			Url:    "/dashboard/crash_risk_warning",
			Method: "GET",
			Handler: func(c *gin.Context) {
				e.handleAnalyticsQuery(c, paradigm.CrashRisk, "未找到崩盘风险预警结果", "获取崩盘风险预警失败", "E100011")
			},
		}
		return &httpService, nil
	case INVESTOR_COMP:
		httpService := HttpService{
			Url:    "/dashboard/investor_composition",
			Method: "GET",
			Handler: func(c *gin.Context) {
				selectedType := strings.TrimSpace(strings.ToLower(c.Query("type")))
				if selectedType != "" && selectedType != "history" && selectedType != "custom" {
					c.JSON(http.StatusBadRequest, paradigm.HttpResponse{
						Message: "type must be one of history, custom",
						Code:    "E100012",
						Data:    nil,
					})
					return
				}

				taskID := strings.TrimSpace(c.Query("taskId"))
				stockID := strings.TrimSpace(c.Query("stockId"))
				data, err := e.QueryAnalytics(taskID, stockID, paradigm.InvestorComposition, nil)
				if err != nil {
					if errors.Is(err, errAnalyticsNotFound) {
						c.JSON(http.StatusNotFound, paradigm.HttpResponse{
							Message: "未找到投资者构成结果: " + err.Error(),
							Code:    "E100012",
							Data:    nil,
						})
						return
					}
					paradigm.Log("ERROR", fmt.Sprintf("Failed to fetch investor composition: %v", err))
					c.JSON(http.StatusInternalServerError, paradigm.HttpResponse{
						Message: "获取投资者构成失败: " + err.Error(),
						Code:    "E100012",
						Data:    nil,
					})
					return
				}

				data = abm.NormalizeInvestorCompositionResponse(data, c.Query("date"), selectedType)
				c.JSON(http.StatusOK, paradigm.HttpResponse{Message: "操作成功", Data: data, Code: "S000000"})
			},
		}
		return &httpService, nil
	case PERF_COMPARISON:
		httpService := HttpService{
			Url:    "/dashboard/performance_comparison",
			Method: "GET",
			Handler: func(c *gin.Context) {
				e.handleAnalyticsQuery(c, paradigm.PerformanceComparison, "未找到模型评估结果", "获取模型评估失败", "E100013")
			},
		}
		return &httpService, nil
	case PLATFORM_TASK_DOWNLOAD:
		httpService := HttpService{
			Url:     "/dashboard/platform_task/download",
			Method:  "GET",
			Handler: e.HandlePlatformTaskDownload,
		}
		return &httpService, nil
	default:
		paradigm.Error(paradigm.NetworkError, "Unknown HTTP Service")
		//LogWriter.Log("ERROR", fmt.Sprintf("%s: %s", paradigm.ErrorToString(paradigm.NetworkError), "Unknown Http Service"))
		return nil, fmt.Errorf("unknown Http Service")
	}
}
