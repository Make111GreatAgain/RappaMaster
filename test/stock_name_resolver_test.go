package test

import (
	"BHLayer2Node/Network/HTTP"
	"BHLayer2Node/Network/HTTP/abm"
	"BHLayer2Node/paradigm"
	"testing"
)

func TestResolveStockDisplayNameFromMap(t *testing.T) {
	cases := []struct {
		name     string
		code     string
		fallback string
		expected string
	}{
		{
			name:     "missing fallback uses stock map",
			code:     "600000",
			expected: "浦发银行",
		},
		{
			name:     "code fallback is replaced by stock map",
			code:     "600000",
			fallback: "600000",
			expected: "浦发银行",
		},
		{
			name:     "unknown code keeps custom fallback",
			code:     "999999",
			fallback: "测试股票",
			expected: "测试股票",
		},
		{
			name:     "market suffix is normalized",
			code:     "600519.SH",
			fallback: "600519",
			expected: "贵州茅台",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := abm.ResolveStockDisplayName(tc.code, tc.fallback)
			if got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestExtractTaskStockNameUsesStockMapFallback(t *testing.T) {
	cases := []struct {
		name     string
		params   map[string]interface{}
		expected string
	}{
		{
			name:     "missing stockName",
			params:   map[string]interface{}{"stockCode": "600000"},
			expected: "浦发银行",
		},
		{
			name:     "stockName equals stockCode",
			params:   map[string]interface{}{"stockCode": "600000", "stockName": "600000"},
			expected: "浦发银行",
		},
		{
			name:     "unknown code keeps raw stockName",
			params:   map[string]interface{}{"stockCode": "999999", "stockName": "测试股票"},
			expected: "测试股票",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			task := &paradigm.Task{
				Sign:   "SubTask-TSK-TEST-600000",
				Params: tc.params,
			}
			got := HTTP.ExtractTaskStockName(task)
			if got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}
