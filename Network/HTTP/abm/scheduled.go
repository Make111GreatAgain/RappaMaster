package abm

import (
	"BHLayer2Node/paradigm"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

const ScheduledV2Horizon = "1天 (T+1)"

type ScheduledV2Options struct {
	Universe      string
	DataStartDate string
	DataEndDate   string
}

func IsScheduledCreateTask(c *gin.Context) bool {
	return strings.EqualFold(c.Query("isScheduled"), "true")
}

func BuildScheduledV2OptionsFromQuery(c *gin.Context) (ScheduledV2Options, error) {
	universe, err := NormalizeScheduledUniverse(c.Query("universe"))
	if err != nil {
		return ScheduledV2Options{}, err
	}
	return ScheduledV2Options{
		Universe:      universe,
		DataStartDate: strings.TrimSpace(c.Query("dataStartDate")),
		DataEndDate:   strings.TrimSpace(c.Query("dataEndDate")),
	}, nil
}

// 构造定时任务的股票列表。
// 股票列表由离线参数目录驱动，真实行情数据由执行节点的数据源在运行时解析。
func BuildScheduledV2RawTasks(config *paradigm.BHLayer2NodeConfig, options ScheduledV2Options) ([]map[string]interface{}, error) {
	universe, err := NormalizeScheduledUniverse(options.Universe)
	if err != nil {
		return nil, err
	}
	window, err := scheduledV2DataWindow(options)
	if err != nil {
		return nil, err
	}
	stockCodes, universeStocks, err := listScheduledStockCodes(config, universe, window.EndDate)
	if err != nil {
		return nil, err
	}
	if len(stockCodes) == 0 {
		return nil, fmt.Errorf("no supported ABM stocks found")
	}

	tasks := make([]map[string]interface{}, 0, len(stockCodes))
	for _, stockCode := range stockCodes {
		stockName := scheduledV2StockName(stockCode)
		if universeStock, ok := universeStocks[stockCode]; ok && strings.TrimSpace(universeStock.StockName) != "" {
			stockName = universeStock.StockName
		}
		raw := map[string]interface{}{
			"stockCode": stockCode,
			"stockName": stockName,
			"horizon":   ScheduledV2Horizon,
			// 定时任务默认 T-1 单日，可通过接口 query dataStartDate/dataEndDate 覆盖。
			"dataStartDate": window.StartDate,
			"dataEndDate":   window.EndDate,
		}
		availability, window, err := ValidateStockDataAvailable(stockCode, raw, config)
		if err != nil {
			paradigm.Log("WARN", fmt.Sprintf("ABM_V2 scheduled stock skipped, stockCode=%s, reason=%v", stockCode, err))
			continue
		}
		if !availability.Available {
			paradigm.Log("WARN", fmt.Sprintf("ABM_V2 scheduled stock skipped, stockCode=%s, startDate=%s, endDate=%s, source=%s, reason=%s",
				stockCode, window.StartDate, window.EndDate, availability.Source, availability.Reason))
			continue
		}
		tasks = append(tasks, raw)
	}
	if len(tasks) == 0 {
		return nil, fmt.Errorf("no supported ABM stocks found with available stock data")
	}
	return tasks, nil
}

func scheduledV2DataWindow(options ScheduledV2Options) (ABMDataWindow, error) {
	raw := map[string]interface{}{}
	if strings.TrimSpace(options.DataStartDate) != "" {
		raw["dataStartDate"] = strings.TrimSpace(options.DataStartDate)
	}
	if strings.TrimSpace(options.DataEndDate) != "" {
		raw["dataEndDate"] = strings.TrimSpace(options.DataEndDate)
	}
	// 定时任务窗口默认固定 T-1，不使用配置项覆盖。
	return NormalizeABMDataWindow(raw, nil)
}

func listScheduledStockCodes(config *paradigm.BHLayer2NodeConfig, universe string, targetDate string) ([]string, map[string]UniverseStock, error) {
	index := CurrentABMParameterIndex()
	stockCodes := make([]string, 0, index.TunedCount)
	universeStocks := map[string]UniverseStock{}
	if universe != ScheduledUniverseAll {
		stocks, err := LoadUniverseStocks(universe, targetDate, config)
		if err != nil {
			return nil, nil, err
		}
		universeStocks = stocks
	}
	for _, item := range index.SupportedStockList {
		meta, ok := index.StockMap[item.StockCode]
		if ok && meta.SupportSimulation && meta.HasTunedParams {
			if universe != ScheduledUniverseAll {
				if _, exists := universeStocks[item.StockCode]; !exists {
					continue
				}
			}
			stockCodes = append(stockCodes, item.StockCode)
		}
	}
	return stockCodes, universeStocks, nil
}

func scheduledV2StockName(stockCode string) string {
	return ResolveStockDisplayName(stockCode, stockCode)
}
