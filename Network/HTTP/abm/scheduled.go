package abm

import (
	"BHLayer2Node/paradigm"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

const ScheduledV2Horizon = "1天 (T+1)"

func IsScheduledCreateTask(c *gin.Context) bool {
	return strings.EqualFold(c.Query("isScheduled"), "true")
}

// 构造定时任务的股票列表
// “既有调参文件、又有真实输入 CSV”的股票才会被加入全市场定时任务
func BuildScheduledV2RawTasks(config *paradigm.BHLayer2NodeConfig) ([]map[string]interface{}, error) {
	stockCodes, err := listSupportedStockCodes(config)
	if err != nil {
		return nil, err
	}
	if len(stockCodes) == 0 {
		return nil, fmt.Errorf("no supported ABM stocks found")
	}

	tasks := make([]map[string]interface{}, 0, len(stockCodes))
	for _, stockCode := range stockCodes {
		// 定时任务由外部调度方按周期触发；任务内容固定为所有已调参股票的 T+1 推演，
		// 不接受请求体里的额外覆盖参数，避免不同周期任务口径不一致。
		tasks = append(tasks, map[string]interface{}{
			"stockCode": stockCode,
			"stockName": scheduledV2StockName(stockCode),
			"horizon":   ScheduledV2Horizon,
		})
	}
	return tasks, nil
}

func listSupportedStockCodes(config *paradigm.BHLayer2NodeConfig) ([]string, error) {
	paramsDir := StockParamDir(config)
	dataDir := StockDataDir(config)

	entries, err := os.ReadDir(paramsDir)
	if err != nil {
		return nil, err
	}

	stockCodes := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		stockCode := NormalizeStockCode(entry.Name())
		if stockCode == "" {
			continue
		}
		if _, err := os.Stat(filepath.Join(paramsDir, stockCode, "model_params.json")); err != nil {
			continue
		}
		if _, err := os.Stat(filepath.Join(dataDir, stockCode+".csv")); err != nil {
			continue
		}
		stockCodes = append(stockCodes, stockCode)
	}
	sort.Strings(stockCodes)
	return stockCodes, nil
}

func scheduledV2StockName(stockCode string) string {
	return ResolveStockDisplayName(stockCode, stockCode)
}
