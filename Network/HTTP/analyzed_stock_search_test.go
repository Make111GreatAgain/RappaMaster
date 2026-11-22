package HTTP

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"runtime"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAnalyzedStocksWindowSearch(t *testing.T) {
	e := &HttpEngine{}
	snapshot, err := newAnalyzedStockSnapshot([]analyzedStockSearchItem{
		{stockCode: "1600001", stockName: "浦发银行样本", taskID: "tsk-10010",
			data: map[string]interface{}{"taskId": "TSK-10010"}},
		{stockCode: "600001", stockName: "浦发银行", taskID: "tsk-1001",
			data: map[string]interface{}{"taskId": "TSK-1001"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := e.analyzedStockSearch.publish(snapshot); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(e.analyzedStockSearch.close)
	service, err := e.GetHttpService(ANALYZED_STOCKS)
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.GET(service.Url, service.Handler)

	for _, test := range []struct {
		name, searchType, keyword, mode, paging string
		want                                    []string
		total                                   int
	}{
		{"legacy initial full list", "stockCode", "", "", "", []string{"TSK-10010", "TSK-1001"}, 0},
		{"exact ranked first", "all", "600001", "", "&pageNo=1&pageSize=1", []string{"TSK-1001"}, 2},
		{"partial on next page", "all", "TSK-1001", "", "&pageNo=2&pageSize=1", []string{"TSK-10010"}, 2},
		{"Chinese substring", "all", "浦发", "", "&pageNo=1&pageSize=20", []string{"TSK-10010", "TSK-1001"}, 2},
		{"exact only", "all", "TSK-1001", "exact", "&pageNo=1&pageSize=20", []string{"TSK-1001"}, 1},
		{"legacy category order", "stockCode", "600001", "", "", []string{"TSK-10010", "TSK-1001"}, 0},
		{"page past end", "all", "浦发", "", "&pageNo=3&pageSize=1", []string{}, 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			query := url.Values{"searchType": {test.searchType}, "keyword": {test.keyword}}
			if test.mode != "" {
				query.Set("matchMode", test.mode)
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, service.Url+"?"+query.Encode()+test.paging, nil))
			var body struct{ Data json.RawMessage }
			if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &body) != nil {
				t.Fatalf("unexpected response: %d %s", response.Code, response.Body)
			}
			type item struct{ TaskID string }
			var items []item
			if test.paging != "" {
				var page struct {
					Items []item
					Total int
				}
				if err := json.Unmarshal(body.Data, &page); err != nil {
					t.Fatal(err)
				}
				if page.Total != test.total || page.Items == nil {
					t.Fatalf("unexpected page: %s", body.Data)
				}
				items = page.Items
			} else if err := json.Unmarshal(body.Data, &items); err != nil {
				t.Fatal(err)
			}
			got := make([]string, 0, len(items))
			for _, item := range items {
				got = append(got, item.TaskID)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("got %v, want %v", got, test.want)
			}
		})
	}
}

// A refresh may replace the buffer while an older request is still streaming it.
func TestAnalyzedStockBufferCanFinishDuringRefresh(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("requires Linux-style unlink of open files")
	}
	var cache analyzedStockSearchCache
	t.Cleanup(cache.close)
	makeSnapshot := func(id string) *analyzedStockSnapshot {
		t.Helper()
		snapshot, err := newAnalyzedStockSnapshot([]analyzedStockSearchItem{
			{data: map[string]interface{}{"taskId": id}},
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := cache.publish(snapshot); err != nil {
			t.Fatal(err)
		}
		return snapshot
	}
	old := makeSnapshot("previous")
	reader, size, err := cache.openFullResponse()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	fresh := makeSnapshot("current")
	if _, err := os.Stat(old.filePath); !os.IsNotExist(err) {
		t.Fatalf("old path should be removed, got %v", err)
	}
	body, err := io.ReadAll(reader)
	if err != nil || int64(len(body)) != size || !json.Valid(body) || !bytes.Contains(body, []byte("previous")) {
		t.Fatalf("old reader lost its response: %q (%v)", body, err)
	}
	cache.close()
	if _, err := os.Stat(fresh.filePath); !os.IsNotExist(err) {
		t.Fatalf("current path should be removed after cache close, got %v", err)
	}
}
