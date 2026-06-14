package abm

import (
	"BHLayer2Node/paradigm"
	"testing"
)

func TestRemoteMarketRoutingForStock(t *testing.T) {
	config := &paradigm.BHLayer2NodeConfig{
		ABMRemoteDBName:      "legacy-db",
		ABMRemoteSHDBName:    "sh-db",
		ABMRemoteSZDBName:    "sz-db",
		ABMRemoteTableName:   "legacy-table",
		ABMRemoteTableMode:   "market",
		ABMRemoteSHTableName: "sh-table",
		ABMRemoteSZTableName: "sz-table",
	}

	cases := []struct {
		code  string
		db    string
		table string
	}{
		{code: "600028", db: "sh-db", table: "sh-table"},
		{code: "000157", db: "sz-db", table: "sz-table"},
		{code: "300750", db: "sz-db", table: "sz-table"},
	}
	for _, tc := range cases {
		if got := remoteDBForStock(config, tc.code); got != tc.db {
			t.Fatalf("remoteDBForStock(%s) = %s, want %s", tc.code, got, tc.db)
		}
		if got := remoteTableForStock(config, tc.code); got != tc.table {
			t.Fatalf("remoteTableForStock(%s) = %s, want %s", tc.code, got, tc.table)
		}
	}
}

func TestRemoteSingleRoutingKeepsLegacyConfig(t *testing.T) {
	config := &paradigm.BHLayer2NodeConfig{
		ABMRemoteDBName:      "legacy-db",
		ABMRemoteSHDBName:    "sh-db",
		ABMRemoteSZDBName:    "sz-db",
		ABMRemoteTableName:   "legacy-table",
		ABMRemoteTableMode:   "single",
		ABMRemoteSHTableName: "sh-table",
		ABMRemoteSZTableName: "sz-table",
	}

	if got := remoteDBForStock(config, "000157"); got != "legacy-db" {
		t.Fatalf("single mode db = %s, want legacy-db", got)
	}
	if got := remoteTableForStock(config, "000157"); got != "legacy-table" {
		t.Fatalf("single mode table = %s, want legacy-table", got)
	}
}
