package abm

import (
	"BHLayer2Node/paradigm"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	stockDataSourceAuto      = "auto"
	stockDataSourceLocal     = "local"
	stockDataSourceDolphinDB = "dolphindb"
)

type ABMDataWindow struct {
	StartDate string
	EndDate   string
}

type StockDataAvailabilityRequest struct {
	StockCode string
	Window    ABMDataWindow
	Config    *paradigm.BHLayer2NodeConfig
}

type StockDataAvailabilityResult struct {
	Available bool
	Source    string
	Reason    string
	Rows      int64
}

type StockDataAvailabilityChecker interface {
	CheckStockDataAvailable(req StockDataAvailabilityRequest) (StockDataAvailabilityResult, error)
}

type BatchStockDataAvailabilityChecker interface {
	CheckStockDataAvailableBatch(reqs []StockDataAvailabilityRequest) (map[string]StockDataAvailabilityResult, error)
}

var (
	stockDataAvailabilityCheckerMu sync.Mutex
	stockDataAvailabilityChecker   StockDataAvailabilityChecker = defaultStockDataAvailabilityChecker{}
)

func SetStockDataAvailabilityCheckerForTest(checker StockDataAvailabilityChecker) func() {
	stockDataAvailabilityCheckerMu.Lock()
	old := stockDataAvailabilityChecker
	stockDataAvailabilityChecker = checker
	stockDataAvailabilityCheckerMu.Unlock()
	return func() {
		stockDataAvailabilityCheckerMu.Lock()
		stockDataAvailabilityChecker = old
		stockDataAvailabilityCheckerMu.Unlock()
	}
}

func NormalizeABMDataWindow(raw map[string]interface{}, config *paradigm.BHLayer2NodeConfig) (ABMDataWindow, error) {
	startText := strings.TrimSpace(stringValue(raw["dataStartDate"]))
	endText := strings.TrimSpace(stringValue(raw["dataEndDate"]))
	if startText == "" {
		startText = strings.TrimSpace(stringValue(raw["startDate"]))
	}
	if endText == "" {
		endText = strings.TrimSpace(stringValue(raw["endDate"]))
	}

	if startText == "" && endText == "" {
		offset := 1
		if config != nil && config.ABMRemoteDefaultEndOffsetDays > 0 {
			offset = config.ABMRemoteDefaultEndOffsetDays
		}
		day := time.Now().AddDate(0, 0, -offset).Format("2006-01-02")
		return ABMDataWindow{StartDate: day, EndDate: day}, nil
	}
	if startText == "" {
		startText = endText
	}
	if endText == "" {
		endText = startText
	}

	start, err := parseABMDate(startText)
	if err != nil {
		return ABMDataWindow{}, fmt.Errorf("invalid dataStartDate: %w", err)
	}
	end, err := parseABMDate(endText)
	if err != nil {
		return ABMDataWindow{}, fmt.Errorf("invalid dataEndDate: %w", err)
	}
	if end.Before(start) {
		return ABMDataWindow{}, fmt.Errorf("dataEndDate before dataStartDate")
	}
	return ABMDataWindow{
		StartDate: start.Format("2006-01-02"),
		EndDate:   end.Format("2006-01-02"),
	}, nil
}

func ValidateStockDataAvailable(stockCode string, raw map[string]interface{}, config *paradigm.BHLayer2NodeConfig) (StockDataAvailabilityResult, ABMDataWindow, error) {
	stockCode = NormalizeStockCode(stockCode)
	if stockCode == "" {
		return StockDataAvailabilityResult{}, ABMDataWindow{}, fmt.Errorf("stockCode is required")
	}
	window, err := NormalizeABMDataWindow(raw, config)
	if err != nil {
		return StockDataAvailabilityResult{}, ABMDataWindow{}, err
	}

	stockDataAvailabilityCheckerMu.Lock()
	checker := stockDataAvailabilityChecker
	stockDataAvailabilityCheckerMu.Unlock()
	if checker == nil {
		checker = defaultStockDataAvailabilityChecker{}
	}
	result, err := checker.CheckStockDataAvailable(StockDataAvailabilityRequest{
		StockCode: stockCode,
		Window:    window,
		Config:    config,
	})
	return result, window, err
}

func DataSourceMode(config *paradigm.BHLayer2NodeConfig) string {
	mode := stockDataSourceAuto
	if config != nil && strings.TrimSpace(config.ABMStockDataSource) != "" {
		mode = strings.ToLower(strings.TrimSpace(config.ABMStockDataSource))
	}
	switch mode {
	case stockDataSourceLocal, stockDataSourceDolphinDB, stockDataSourceAuto:
		return mode
	default:
		return stockDataSourceAuto
	}
}

type defaultStockDataAvailabilityChecker struct{}

func (defaultStockDataAvailabilityChecker) CheckStockDataAvailable(req StockDataAvailabilityRequest) (StockDataAvailabilityResult, error) {
	mode := DataSourceMode(req.Config)
	localPath := filepath.Join(StockDataDir(req.Config), req.StockCode+".csv")
	if mode == stockDataSourceLocal || mode == stockDataSourceAuto {
		if info, err := os.Stat(localPath); err == nil && !info.IsDir() {
			return StockDataAvailabilityResult{Available: true, Source: stockDataSourceLocal, Reason: "local csv exists"}, nil
		}
		if mode == stockDataSourceLocal {
			return StockDataAvailabilityResult{Available: false, Source: stockDataSourceLocal, Reason: "local csv missing"}, nil
		}
	}
	return checkDolphinDBStockData(req)
}

func (defaultStockDataAvailabilityChecker) CheckStockDataAvailableBatch(reqs []StockDataAvailabilityRequest) (map[string]StockDataAvailabilityResult, error) {
	results := make(map[string]StockDataAvailabilityResult, len(reqs))
	remoteGroups := map[string][]StockDataAvailabilityRequest{}

	for _, req := range reqs {
		stockCode := NormalizeStockCode(req.StockCode)
		if stockCode == "" {
			continue
		}
		req.StockCode = stockCode

		mode := DataSourceMode(req.Config)
		localPath := filepath.Join(StockDataDir(req.Config), req.StockCode+".csv")
		if mode == stockDataSourceLocal || mode == stockDataSourceAuto {
			if info, err := os.Stat(localPath); err == nil && !info.IsDir() {
				results[req.StockCode] = StockDataAvailabilityResult{Available: true, Source: stockDataSourceLocal, Reason: "local csv exists"}
				continue
			}
			if mode == stockDataSourceLocal {
				results[req.StockCode] = StockDataAvailabilityResult{Available: false, Source: stockDataSourceLocal, Reason: "local csv missing"}
				continue
			}
		}

		groupKey := strings.Join([]string{
			remoteDBForStock(req.Config, req.StockCode),
			remoteTableForStock(req.Config, req.StockCode),
			req.Window.StartDate,
			req.Window.EndDate,
		}, "\x00")
		remoteGroups[groupKey] = append(remoteGroups[groupKey], req)
	}

	for _, group := range remoteGroups {
		groupResults, err := checkDolphinDBStockDataBatch(group)
		if err != nil {
			return results, err
		}
		for stockCode, result := range groupResults {
			results[stockCode] = result
		}
	}

	return results, nil
}

func checkDolphinDBStockData(req StockDataAvailabilityRequest) (StockDataAvailabilityResult, error) {
	results, err := checkDolphinDBStockDataBatch([]StockDataAvailabilityRequest{req})
	if err != nil {
		return StockDataAvailabilityResult{Available: false, Source: stockDataSourceDolphinDB, Reason: err.Error()}, err
	}
	if result, ok := results[NormalizeStockCode(req.StockCode)]; ok {
		return result, nil
	}
	return StockDataAvailabilityResult{Available: false, Source: stockDataSourceDolphinDB, Reason: "remote rows missing"}, nil
}

func checkDolphinDBStockDataBatch(reqs []StockDataAvailabilityRequest) (map[string]StockDataAvailabilityResult, error) {
	results := make(map[string]StockDataAvailabilityResult, len(reqs))
	if len(reqs) == 0 {
		return results, nil
	}
	req := reqs[0]
	config := req.Config
	python := "python3"
	script := "tools/abm_remote_data_check.py"
	if config != nil {
		if strings.TrimSpace(config.ABMRemoteCheckPython) != "" {
			python = strings.TrimSpace(config.ABMRemoteCheckPython)
		}
		if strings.TrimSpace(config.ABMRemoteCheckScript) != "" {
			script = strings.TrimSpace(config.ABMRemoteCheckScript)
		}
	}

	stockCodes := make([]string, 0, len(reqs))
	for _, item := range reqs {
		if stockCode := NormalizeStockCode(item.StockCode); stockCode != "" {
			stockCodes = append(stockCodes, stockCode)
		}
	}

	args := []string{
		script,
		"--host", remoteConfigString(config, "host"),
		"--port", remoteConfigString(config, "port"),
		"--user", remoteConfigString(config, "user"),
		"--password", remoteConfigString(config, "password"),
		"--db", remoteDBForStock(config, req.StockCode),
		"--table", remoteTableForStock(config, req.StockCode),
		"--stock-codes", strings.Join(stockCodes, ","),
		"--start-date", req.Window.StartDate,
		"--end-date", req.Window.EndDate,
	}

	timeout := remoteCheckTimeout(config)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, python, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		reason := strings.TrimSpace(string(output))
		if ctx.Err() == context.DeadlineExceeded {
			reason = fmt.Sprintf("remote data check timeout after %s", timeout)
		}
		for _, stockCode := range stockCodes {
			results[stockCode] = StockDataAvailabilityResult{Available: false, Source: stockDataSourceDolphinDB, Reason: reason}
		}
		return results, fmt.Errorf("%s", reason)
	}

	var parsed struct {
		Exists bool `json:"exists"`
		Rows   int64
		Reason string `json:"reason"`
		Stocks map[string]struct {
			Exists bool   `json:"exists"`
			Rows   int64  `json:"rows"`
			Reason string `json:"reason"`
		} `json:"stocks"`
	}
	if err := json.Unmarshal(output, &parsed); err != nil {
		reason := strings.TrimSpace(string(output))
		for _, stockCode := range stockCodes {
			results[stockCode] = StockDataAvailabilityResult{Available: false, Source: stockDataSourceDolphinDB, Reason: reason}
		}
		return results, err
	}

	if len(parsed.Stocks) == 0 && len(stockCodes) == 1 {
		parsed.Stocks = map[string]struct {
			Exists bool   `json:"exists"`
			Rows   int64  `json:"rows"`
			Reason string `json:"reason"`
		}{
			stockCodes[0]: {Exists: parsed.Exists, Rows: parsed.Rows, Reason: parsed.Reason},
		}
	}

	for _, stockCode := range stockCodes {
		stock := parsed.Stocks[stockCode]
		reason := stock.Reason
		if reason == "" && stock.Exists {
			reason = "remote rows found"
		}
		if reason == "" {
			reason = "remote rows missing"
		}
		results[stockCode] = StockDataAvailabilityResult{
			Available: stock.Exists,
			Source:    stockDataSourceDolphinDB,
			Reason:    reason,
			Rows:      stock.Rows,
		}
	}
	return results, nil
}

func remoteCheckTimeout(config *paradigm.BHLayer2NodeConfig) time.Duration {
	raw := ""
	if config != nil {
		raw = strings.TrimSpace(config.ABMRemoteCheckTimeout)
	}
	if raw == "" {
		raw = strings.TrimSpace(paradigm.DefaultBHLayer2NodeConfig.ABMRemoteCheckTimeout)
	}
	if raw == "" {
		return 5 * time.Second
	}
	timeout, err := time.ParseDuration(raw)
	if err != nil || timeout <= 0 {
		return 5 * time.Second
	}
	return timeout
}

func remoteDBForStock(config *paradigm.BHLayer2NodeConfig, stockCode string) string {
	if config == nil {
		config = &paradigm.DefaultBHLayer2NodeConfig
	}
	defaultDB := strings.TrimSpace(config.ABMRemoteDBName)
	if defaultDB == "" {
		defaultDB = paradigm.DefaultBHLayer2NodeConfig.ABMRemoteDBName
	}
	mode := strings.ToLower(strings.TrimSpace(config.ABMRemoteTableMode))
	if mode == "" {
		mode = strings.ToLower(strings.TrimSpace(paradigm.DefaultBHLayer2NodeConfig.ABMRemoteTableMode))
	}
	if mode == "single" {
		return defaultDB
	}

	code := NormalizeStockCode(stockCode)
	if strings.HasPrefix(code, "0") || strings.HasPrefix(code, "3") {
		if db := strings.TrimSpace(config.ABMRemoteSZDBName); db != "" {
			return db
		}
		if db := strings.TrimSpace(paradigm.DefaultBHLayer2NodeConfig.ABMRemoteSZDBName); db != "" {
			return db
		}
		return defaultDB
	}
	if db := strings.TrimSpace(config.ABMRemoteSHDBName); db != "" {
		return db
	}
	if db := strings.TrimSpace(paradigm.DefaultBHLayer2NodeConfig.ABMRemoteSHDBName); db != "" {
		return db
	}
	return defaultDB
}

func remoteTableForStock(config *paradigm.BHLayer2NodeConfig, stockCode string) string {
	if config == nil {
		config = &paradigm.DefaultBHLayer2NodeConfig
	}
	defaultTable := strings.TrimSpace(config.ABMRemoteTableName)
	if defaultTable == "" {
		defaultTable = paradigm.DefaultBHLayer2NodeConfig.ABMRemoteTableName
	}
	mode := strings.ToLower(strings.TrimSpace(config.ABMRemoteTableMode))
	if mode == "" {
		mode = strings.ToLower(strings.TrimSpace(paradigm.DefaultBHLayer2NodeConfig.ABMRemoteTableMode))
	}
	if mode == "single" {
		return defaultTable
	}

	code := NormalizeStockCode(stockCode)
	if strings.HasPrefix(code, "0") || strings.HasPrefix(code, "3") {
		if table := strings.TrimSpace(config.ABMRemoteSZTableName); table != "" {
			return table
		}
		if table := strings.TrimSpace(paradigm.DefaultBHLayer2NodeConfig.ABMRemoteSZTableName); table != "" {
			return table
		}
		return defaultTable
	}
	if table := strings.TrimSpace(config.ABMRemoteSHTableName); table != "" {
		return table
	}
	if table := strings.TrimSpace(paradigm.DefaultBHLayer2NodeConfig.ABMRemoteSHTableName); table != "" {
		return table
	}
	return defaultTable
}

func remoteConfigString(config *paradigm.BHLayer2NodeConfig, key string) string {
	if config == nil {
		config = &paradigm.DefaultBHLayer2NodeConfig
	}
	switch key {
	case "host":
		return config.ABMRemoteDBHost
	case "port":
		if config.ABMRemoteDBPort > 0 {
			return fmt.Sprintf("%d", config.ABMRemoteDBPort)
		}
		return fmt.Sprintf("%d", paradigm.DefaultBHLayer2NodeConfig.ABMRemoteDBPort)
	case "user":
		return config.ABMRemoteDBUser
	case "password":
		return config.ABMRemoteDBPassword
	default:
		return ""
	}
}

func parseABMDate(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	for _, layout := range []string{"2006-01-02", "2006.01.02"} {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("expected YYYY-MM-DD")
}
