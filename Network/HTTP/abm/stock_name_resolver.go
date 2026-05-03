package abm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	stockNameMapOnce sync.Once
	stockNameMap     map[string]string
)

func ResolveStockDisplayName(stockCode, fallback string) string {
	code := normalizeDisplayStockCode(stockCode)
	nameMap := loadStockNameMap()
	if code != "" {
		if name := strings.TrimSpace(nameMap[code]); name != "" {
			return name
		}
	}

	fallback = strings.TrimSpace(fallback)
	if fallback != "" {
		return fallback
	}
	return code
}

func loadStockNameMap() map[string]string {
	stockNameMapOnce.Do(func() {
		stockNameMap = map[string]string{}
		for _, path := range stockNameMapCandidatePaths() {
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			parsed := map[string]string{}
			if err := json.Unmarshal(data, &parsed); err != nil {
				continue
			}
			for code, name := range parsed {
				code = normalizeDisplayStockCode(code)
				name = strings.TrimSpace(name)
				if code != "" && name != "" {
					stockNameMap[code] = name
				}
			}
			if len(stockNameMap) > 0 {
				break
			}
		}
	})
	return stockNameMap
}

func stockNameMapCandidatePaths() []string {
	paths := []string{}
	if envPath := strings.TrimSpace(os.Getenv("STOCK_NAME_MAP_PATH")); envPath != "" {
		paths = append(paths, envPath)
	}
	paths = append(paths,
		filepath.Join("Config", "stock_name_map.json"),
		"/root/rappa/RappaMaster/Config/stock_name_map.json",
	)
	return paths
}

func normalizeDisplayStockCode(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) == 6 && isDigits(value) {
		return value
	}
	for i := 0; i+6 <= len(value); i++ {
		part := value[i : i+6]
		if isDigits(part) {
			return part
		}
	}
	return value
}

func isDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}
