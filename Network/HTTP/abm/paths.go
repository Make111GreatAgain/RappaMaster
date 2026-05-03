package abm

import (
	"BHLayer2Node/paradigm"
	"os"
	"path/filepath"
	"strings"
)

const defaultStockDataDir = "/root/rappa/stockdata"

func StockDataDir(config *paradigm.BHLayer2NodeConfig) string {
	if value := strings.TrimSpace(os.Getenv("ABM_STOCK_DATA_DIR")); value != "" {
		return value
	}
	if config != nil {
		if value := strings.TrimSpace(config.ABMStockDataDir); value != "" {
			return value
		}
	}
	return defaultStockDataDir
}

func StockParamDir(config *paradigm.BHLayer2NodeConfig) string {
	if value := strings.TrimSpace(os.Getenv("ABM_STOCK_PARAM_DIR")); value != "" {
		return value
	}
	if config != nil {
		if value := strings.TrimSpace(config.ABMStockParamDir); value != "" {
			return value
		}
	}
	return filepath.Join(StockDataDir(config), "params")
}
