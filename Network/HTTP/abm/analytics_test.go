package abm

import "testing"

func TestHideUnstablePerformanceMetricsFiltersDirectPayload(t *testing.T) {
	payload := map[string]interface{}{
		"selectedModel": "ABM",
		"tableData": []interface{}{
			map[string]interface{}{"indicator": "Mean (均值)", "trueData": 1},
			map[string]interface{}{"indicator": "Kurtosis (峰度)", "trueData": 2},
			map[string]interface{}{"indicator": "Skewness (偏度)", "trueData": 3},
			map[string]interface{}{"indicator": "Volume Mean", "trueData": 4},
			map[string]interface{}{"indicator": "Pearson Corr", "trueData": 5},
		},
	}

	got := HideUnstablePerformanceMetrics(payload).(map[string]interface{})
	rows := got["tableData"].([]interface{})
	if len(rows) != 2 {
		t.Fatalf("expected 2 visible rows, got %#v", rows)
	}
	assertIndicator(t, rows[0], "Mean (均值)")
	assertIndicator(t, rows[1], "Pearson Corr")
}

func TestHideUnstablePerformanceMetricsFiltersWrappedPayloads(t *testing.T) {
	payload := []map[string]interface{}{
		{
			"stockCode": "600000",
			"data": map[string]interface{}{
				"tableData": []map[string]interface{}{
					{"indicator": "Kurtosis (峰度)"},
					{"indicator": "Wasserstein"},
					{"indicator": "Volume Mean"},
				},
			},
		},
		{
			"stockCode": "600016",
			"data": map[string]interface{}{
				"tableData": []map[string]interface{}{
					{"indicator": "Skewness (偏度)"},
					{"indicator": "KL Divergence"},
				},
			},
		},
	}

	got := HideUnstablePerformanceMetrics(payload).([]map[string]interface{})
	firstData := got[0]["data"].(map[string]interface{})
	firstRows := firstData["tableData"].([]map[string]interface{})
	if len(firstRows) != 1 || firstRows[0]["indicator"] != "Wasserstein" {
		t.Fatalf("unexpected first filtered rows: %#v", firstRows)
	}

	secondData := got[1]["data"].(map[string]interface{})
	secondRows := secondData["tableData"].([]map[string]interface{})
	if len(secondRows) != 1 || secondRows[0]["indicator"] != "KL Divergence" {
		t.Fatalf("unexpected second filtered rows: %#v", secondRows)
	}
}

func assertIndicator(t *testing.T, row interface{}, expected string) {
	t.Helper()
	typed, ok := row.(map[string]interface{})
	if !ok {
		t.Fatalf("expected row map, got %#v", row)
	}
	if typed["indicator"] != expected {
		t.Fatalf("expected indicator %q, got %#v", expected, typed["indicator"])
	}
}
