package abm

import (
	"BHLayer2Node/paradigm"
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

func checkDolphinDBStockData(req StockDataAvailabilityRequest) (StockDataAvailabilityResult, error) {
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

	args := []string{
		script,
		"--host", remoteConfigString(config, "host"),
		"--port", remoteConfigString(config, "port"),
		"--user", remoteConfigString(config, "user"),
		"--password", remoteConfigString(config, "password"),
		"--db", remoteDBForStock(config, req.StockCode),
		"--table", remoteTableForStock(config, req.StockCode),
		"--stock-code", req.StockCode,
		"--start-date", req.Window.StartDate,
		"--end-date", req.Window.EndDate,
	}
	cmd := exec.Command(python, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return StockDataAvailabilityResult{Available: false, Source: stockDataSourceDolphinDB, Reason: strings.TrimSpace(string(output))}, err
	}

	var parsed struct {
		Exists bool   `json:"exists"`
		Rows   int64  `json:"rows"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(output, &parsed); err != nil {
		return StockDataAvailabilityResult{Available: false, Source: stockDataSourceDolphinDB, Reason: strings.TrimSpace(string(output))}, err
	}
	reason := parsed.Reason
	if reason == "" && parsed.Exists {
		reason = "remote rows found"
	}
	if reason == "" {
		reason = "remote rows missing"
	}
	return StockDataAvailabilityResult{
		Available: parsed.Exists,
		Source:    stockDataSourceDolphinDB,
		Reason:    reason,
		Rows:      parsed.Rows,
	}, nil
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
