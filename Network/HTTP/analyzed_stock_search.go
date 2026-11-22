package HTTP

import (
	"BHLayer2Node/Database"
	"BHLayer2Node/Network/HTTP/abm"
	"BHLayer2Node/paradigm"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

const analyzedStockRefreshInterval = 20 * time.Second
const analyzedStockRefreshTimeout = 5 * time.Second

type analyzedStockSearchItem struct {
	data      map[string]interface{}
	stockCode string
	stockName string
	taskID    string
	taskName  string
}

type analyzedStockSnapshot struct {
	items    []analyzedStockSearchItem
	filePath string
	fileSize int64
}

type analyzedStockSearchCache struct {
	mu      sync.RWMutex
	current *analyzedStockSnapshot
	closed  bool
	refresh singleflight.Group
}

func (cache *analyzedStockSearchCache) publish(snapshot *analyzedStockSnapshot) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if cache.closed {
		_ = os.Remove(snapshot.filePath)
		return fmt.Errorf("analyzed stocks cache is closed")
	}
	old := cache.current
	cache.current = snapshot
	if old != nil {
		removeAnalyzedStockFile(old.filePath)
	}
	return nil
}

func (cache *analyzedStockSearchCache) close() {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.closed = true
	if cache.current != nil {
		removeAnalyzedStockFile(cache.current.filePath)
		cache.current = nil
	}
}

func removeAnalyzedStockFile(path string) {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		paradigm.Log("WARN", fmt.Sprintf("Remove analyzed stocks buffer: %v", err))
	}
}

func buildAnalyzedStockItem(task Database.AnalyzedStockTask) (analyzedStockSearchItem, bool) {
	p := task.Params
	stockCode := strings.TrimSpace(stringifyTaskParam(p["stockCode"]))
	if stockCode == "" {
		return analyzedStockSearchItem{}, false
	}
	stockName := abm.ResolveStockDisplayName(stockCode, stringifyTaskParam(p["stockName"]))
	stockID := p["stockId"]
	if stockID == nil {
		stockID = stockCode
	}
	taskID := task.Sign
	if task.PlatformTaskID != nil && strings.TrimSpace(*task.PlatformTaskID) != "" {
		taskID = *task.PlatformTaskID
	}
	taskName := fmt.Sprintf("%v 风险监测", stockName)
	if task.PlatformTaskName != nil {
		taskName = *task.PlatformTaskName
	}

	return analyzedStockSearchItem{
		data: map[string]interface{}{
			"stockId": stockID, "stockCode": stockCode, "stockName": stockName,
			"label":  fmt.Sprintf("%s %s", stockCode, stockName),
			"taskId": taskID, "taskName": taskName, "date": task.StartTime.Format("2006-01-02"),
		},
		stockCode: strings.ToLower(stockCode), stockName: strings.ToLower(stockName),
		taskID: strings.ToLower(taskID), taskName: strings.ToLower(taskName),
	}, true
}

func (e *HttpEngine) refreshAnalyzedStockSearch() <-chan singleflight.Result {
	return e.analyzedStockSearch.refresh.DoChan("snapshot", func() (interface{}, error) {
		ctx, cancel := context.WithTimeout(context.Background(), analyzedStockRefreshTimeout)
		defer cancel()
		var items []analyzedStockSearchItem
		err := e.dbService.VisitAnalyzedStockTasks(ctx, func(task Database.AnalyzedStockTask) {
			if item, ok := buildAnalyzedStockItem(task); ok {
				items = append(items, item)
			}
		})
		if err != nil {
			return nil, err
		}
		snapshot, err := newAnalyzedStockSnapshot(items)
		if err != nil {
			return nil, err
		}
		if err := e.analyzedStockSearch.publish(snapshot); err != nil {
			return nil, err
		}
		return snapshot, nil
	})
}

func (e *HttpEngine) runAnalyzedStockRefresh(ctx context.Context) {
	ticker := time.NewTicker(analyzedStockRefreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case result := <-e.refreshAnalyzedStockSearch():
			if result.Err != nil {
				paradigm.Log("WARN", fmt.Sprintf("Refresh analyzed stocks search failed: %v", result.Err))
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func newAnalyzedStockSnapshot(items []analyzedStockSearchItem) (_ *analyzedStockSnapshot, err error) {
	file, err := os.CreateTemp("", "rappamaster-analyzed-stocks-*.json")
	if err != nil {
		return nil, err
	}
	path := file.Name()
	defer func() {
		if err != nil {
			_ = file.Close()
			_ = os.Remove(path)
		}
	}()

	writer := bufio.NewWriter(file)
	if _, err = writer.WriteString(`{"msg":"操作成功","status":"S000000","data":[`); err != nil {
		return nil, err
	}
	encoder := json.NewEncoder(writer)
	for i, item := range items {
		if i > 0 {
			if err = writer.WriteByte(','); err != nil {
				return nil, err
			}
		}
		if err = encoder.Encode(item.data); err != nil {
			return nil, err
		}
	}
	if _, err = writer.WriteString("]}"); err != nil {
		return nil, err
	}
	if err = writer.Flush(); err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if err = file.Close(); err != nil {
		return nil, err
	}
	return &analyzedStockSnapshot{items: items, filePath: path, fileSize: info.Size()}, nil
}

func (e *HttpEngine) getAnalyzedStockSnapshot(ctx context.Context) (*analyzedStockSnapshot, error) {
	cache := &e.analyzedStockSearch
	cache.mu.RLock()
	snapshot := cache.current
	cache.mu.RUnlock()
	if snapshot != nil {
		return snapshot, nil
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-e.refreshAnalyzedStockSearch():
		if result.Err != nil {
			return nil, result.Err
		}
		return result.Val.(*analyzedStockSnapshot), nil
	}
}

// Open under the same lock used to replace the buffer. On Linux an open reader
// can finish streaming after the previous buffer has been unlinked.
func (cache *analyzedStockSearchCache) openFullResponse() (*os.File, int64, error) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	if cache.current == nil {
		return nil, 0, fmt.Errorf("analyzed stocks buffer is unavailable")
	}
	file, err := os.Open(cache.current.filePath)
	return file, cache.current.fileSize, err
}

func (item analyzedStockSearchItem) match(searchType, keyword string, matches func(string, string) bool) bool {
	switch searchType {
	case "stockCode":
		return matches(item.stockCode, keyword)
	case "stockName":
		return matches(item.stockName, keyword)
	case "task":
		return matches(item.taskID, keyword) || matches(item.taskName, keyword)
	case "all":
		return matches(item.stockCode, keyword) || matches(item.stockName, keyword) ||
			matches(item.taskID, keyword) || matches(item.taskName, keyword)
	default:
		return false
	}
}

func (snapshot *analyzedStockSnapshot) search(searchType, keyword, matchMode string) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	exactMatch := func(value, keyword string) bool { return value == keyword }
	if matchMode == "exact" {
		for _, item := range snapshot.items {
			if item.match(searchType, keyword, exactMatch) {
				result = append(result, item.data)
			}
		}
		return result
	}

	// Preserve date ordering within each group and the legacy order for category searches.
	var partial []map[string]interface{}
	for _, item := range snapshot.items {
		if searchType == "all" && item.match(searchType, keyword, exactMatch) {
			result = append(result, item.data)
		} else if item.match(searchType, keyword, strings.Contains) {
			partial = append(partial, item.data)
		}
	}
	return append(result, partial...)
}
