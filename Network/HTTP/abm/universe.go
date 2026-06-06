package abm

import (
	"BHLayer2Node/paradigm"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	ScheduledUniverseAll     = "all"
	ScheduledUniverseHS300   = "hs300"
	ScheduledUniverseCSI1000 = "csi1000"
	firstUniverseQuarter     = "2022Q1"
)

type UniverseStock struct {
	StockCode string
	StockName string
}

func NormalizeScheduledUniverse(raw string) (string, error) {
	universe := strings.ToLower(strings.TrimSpace(raw))
	if universe == "" {
		return ScheduledUniverseAll, nil
	}
	switch universe {
	case ScheduledUniverseAll, ScheduledUniverseHS300, ScheduledUniverseCSI1000:
		return universe, nil
	default:
		return "", fmt.Errorf("unsupported universe: %s", raw)
	}
}

func LoadUniverseStocks(universe string, targetDate string, config *paradigm.BHLayer2NodeConfig) (map[string]UniverseStock, error) {
	universe, err := NormalizeScheduledUniverse(universe)
	if err != nil {
		return nil, err
	}
	if universe == ScheduledUniverseAll {
		return nil, nil
	}
	target, err := parseABMDate(targetDate)
	if err != nil {
		return nil, err
	}
	quarter, err := resolveUniverseQuarter(filepath.Join(UniverseSnapshotRoot(config), universe), target)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(UniverseSnapshotRoot(config), universe, quarter+".csv")
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open universe snapshot %s: %w", path, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read universe snapshot %s: %w", path, err)
	}
	stocks := make(map[string]UniverseStock, len(rows))
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 1 {
			continue
		}
		code := NormalizeStockCode(row[0])
		if code == "" {
			continue
		}
		name := code
		if len(row) > 1 && strings.TrimSpace(row[1]) != "" {
			name = strings.TrimSpace(row[1])
		}
		stocks[code] = UniverseStock{StockCode: code, StockName: name}
	}
	if len(stocks) == 0 {
		return nil, fmt.Errorf("universe snapshot %s has no stocks", path)
	}
	return stocks, nil
}

func UniverseSnapshotRoot(config *paradigm.BHLayer2NodeConfig) string {
	if config != nil && strings.TrimSpace(config.ABMUniverseSnapshotRoot) != "" {
		return strings.TrimSpace(config.ABMUniverseSnapshotRoot)
	}
	return paradigm.DefaultBHLayer2NodeConfig.ABMUniverseSnapshotRoot
}

func resolveUniverseQuarter(root string, target time.Time) (string, error) {
	quarters, err := availableUniverseQuarters(root)
	if err != nil {
		return "", err
	}
	if len(quarters) == 0 {
		return "", fmt.Errorf("no universe snapshots under %s", root)
	}
	targetQuarter := quarterLabel(target)
	if compareQuarter(targetQuarter, firstUniverseQuarter) < 0 {
		targetQuarter = firstUniverseQuarter
	}
	selected := quarters[0]
	for _, quarter := range quarters {
		if compareQuarter(quarter, targetQuarter) <= 0 {
			selected = quarter
			continue
		}
		break
	}
	return selected, nil
}

func availableUniverseQuarters(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read universe snapshot dir %s: %w", root, err)
	}
	quarters := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".csv" {
			continue
		}
		quarter := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		if _, _, ok := parseQuarter(quarter); ok {
			quarters = append(quarters, quarter)
		}
	}
	sort.Slice(quarters, func(i, j int) bool {
		return compareQuarter(quarters[i], quarters[j]) < 0
	})
	return quarters, nil
}

func quarterLabel(day time.Time) string {
	quarter := (int(day.Month())-1)/3 + 1
	return fmt.Sprintf("%dQ%d", day.Year(), quarter)
}

func compareQuarter(left string, right string) int {
	leftYear, leftQuarter, leftOK := parseQuarter(left)
	rightYear, rightQuarter, rightOK := parseQuarter(right)
	if !leftOK || !rightOK {
		return strings.Compare(left, right)
	}
	if leftYear != rightYear {
		return leftYear - rightYear
	}
	return leftQuarter - rightQuarter
}

func parseQuarter(raw string) (int, int, bool) {
	parts := strings.Split(strings.ToUpper(strings.TrimSpace(raw)), "Q")
	if len(parts) != 2 {
		return 0, 0, false
	}
	var year int
	var quarter int
	if _, err := fmt.Sscanf(parts[0], "%d", &year); err != nil {
		return 0, 0, false
	}
	if _, err := fmt.Sscanf(parts[1], "%d", &quarter); err != nil {
		return 0, 0, false
	}
	return year, quarter, year > 0 && quarter >= 1 && quarter <= 4
}
