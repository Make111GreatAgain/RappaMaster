package test

import (
	abmhttp "BHLayer2Node/Network/HTTP/abm"
	"BHLayer2Node/paradigm"
	"testing"
	"time"
)

type fakeStockDataChecker struct {
	results map[string]abmhttp.StockDataAvailabilityResult
	calls   []abmhttp.StockDataAvailabilityRequest
}

func (f *fakeStockDataChecker) CheckStockDataAvailable(req abmhttp.StockDataAvailabilityRequest) (abmhttp.StockDataAvailabilityResult, error) {
	f.calls = append(f.calls, req)
	if result, ok := f.results[req.StockCode]; ok {
		return result, nil
	}
	return abmhttp.StockDataAvailabilityResult{
		Available: true,
		Source:    "mock",
		Reason:    "mock available",
		Rows:      240,
	}, nil
}

func TestNormalizeABMDataWindowDefaultsToTMinusOneSingleDay(t *testing.T) {
	window, err := abmhttp.NormalizeABMDataWindow(map[string]interface{}{}, &paradigm.BHLayer2NodeConfig{
		ABMRemoteDefaultEndOffsetDays: 1,
	})
	if err != nil {
		t.Fatalf("normalize data window: %v", err)
	}
	expected := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	if window.StartDate != expected || window.EndDate != expected {
		t.Fatalf("expected T-1 single day %s, got %#v", expected, window)
	}
}

func TestValidateStockDataAvailableUsesInjectedCheckerAndWindow(t *testing.T) {
	checker := &fakeStockDataChecker{}
	restore := abmhttp.SetStockDataAvailabilityCheckerForTest(checker)
	defer restore()

	result, window, err := abmhttp.ValidateStockDataAvailable("1", map[string]interface{}{
		"dataStartDate": "2026.01.05",
	}, nil)
	if err != nil {
		t.Fatalf("validate stock data: %v", err)
	}
	if !result.Available {
		t.Fatalf("expected stock data available, got %#v", result)
	}
	if window.StartDate != "2026-01-05" || window.EndDate != "2026-01-05" {
		t.Fatalf("unexpected normalized window: %#v", window)
	}
	if len(checker.calls) != 1 || checker.calls[0].StockCode != "000001" {
		t.Fatalf("expected normalized stock code call, got %#v", checker.calls)
	}
}
