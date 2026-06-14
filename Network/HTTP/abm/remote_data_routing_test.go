package abm

import (
	"BHLayer2Node/paradigm"
	"testing"
)

func TestRemoteStockDataRoutesMarketTables(t *testing.T) {
	config := &paradigm.BHLayer2NodeConfig{
		ABMRemoteDBName:      "dfs://default",
		ABMRemoteSHDBName:    "dfs://sh",
		ABMRemoteSZDBName:    "dfs://sz",
		ABMRemoteTableName:   "default_table",
		ABMRemoteTableMode:   "market",
		ABMRemoteSHTableName: "sh_table",
		ABMRemoteSZTableName: "sz_table",
	}

	cases := []struct {
		name      string
		stockCode string
		wantDB    string
		wantTable string
	}{
		{name: "shanghai", stockCode: "600000", wantDB: "dfs://sh", wantTable: "sh_table"},
		{name: "shenzhen", stockCode: "000001", wantDB: "dfs://sz", wantTable: "sz_table"},
		{name: "chinext", stockCode: "300750", wantDB: "dfs://sz", wantTable: "sz_table"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := remoteDBForStock(config, tc.stockCode); got != tc.wantDB {
				t.Fatalf("remoteDBForStock(%q) = %q, want %q", tc.stockCode, got, tc.wantDB)
			}
			if got := remoteTableForStock(config, tc.stockCode); got != tc.wantTable {
				t.Fatalf("remoteTableForStock(%q) = %q, want %q", tc.stockCode, got, tc.wantTable)
			}
		})
	}
}

func TestRemoteStockDataRoutesSingleTable(t *testing.T) {
	config := &paradigm.BHLayer2NodeConfig{
		ABMRemoteDBName:      "dfs://default",
		ABMRemoteSHDBName:    "dfs://sh",
		ABMRemoteSZDBName:    "dfs://sz",
		ABMRemoteTableName:   "default_table",
		ABMRemoteTableMode:   "single",
		ABMRemoteSHTableName: "sh_table",
		ABMRemoteSZTableName: "sz_table",
	}

	if got := remoteDBForStock(config, "000001"); got != "dfs://default" {
		t.Fatalf("remoteDBForStock(single) = %q, want dfs://default", got)
	}
	if got := remoteTableForStock(config, "600000"); got != "default_table" {
		t.Fatalf("remoteTableForStock(single) = %q, want default_table", got)
	}
}
