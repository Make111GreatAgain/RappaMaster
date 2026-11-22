package Query

import (
	"BHLayer2Node/paradigm"
	"strconv"

	"github.com/gin-gonic/gin"
)

type HttpInitTaskRequest struct {
	Sign string // task sign
	//Slot       int32                  // slot index
	Name       string                 // 任务名称
	Size       int32                  // data size
	SlotSize   int32                  // 可选参数，自定义slot大小
	Model      string                 // 模型名称
	Params     map[string]interface{} // 不确定的模型参数
	IsReliable bool                   // 是否需要可信证明
}

type HttpOracleQueryRequest struct {
	Query string                      // 查询类型
	Data  map[interface{}]interface{} // 查询内容rawData
}

type HttpUploadTaskRequest struct {
	TaskID      string
	DataType    string
	Purpose     string
	Description string
	CreateBy    string
}

func (r *HttpOracleQueryRequest) BuildQueryFromGETRequest(c *gin.Context) (bool, paradigm.Query) {
	query := c.DefaultQuery("query", "")
	if query == "" {
		//fmt.Println(222)
		return false, nil
	}
	//fmt.Println(111, query)
	r.Query = query
	switch query {
	case "EvidencePreserveTaskTxQuery":
		txHash := c.DefaultQuery("txHash", "")
		if txHash == "" {
			return false, nil
		}
		return true, NewEvidencePreserveTaskTxQuery(map[interface{}]interface{}{
			"txHash": txHash,
		})
	case "EvidencePreserveTaskIDQuery":
		taskID := c.DefaultQuery("taskID", "")
		if taskID == "" {
			return false, nil
		}
		//t, err := strconv.Atoi(taskID)
		//if err != nil {
		//	return false, nil
		//}
		return true, NewEvidencePreserveTaskIDQuery(map[interface{}]interface{}{
			"taskID": taskID,
		})
	case "EvidencePreserveEpochTxQuery":
		txHash := c.DefaultQuery("txHash", "")
		if txHash == "" {
			return false, nil
		}
		return true, NewEvidencePreserveEpochTxQuery(map[interface{}]interface{}{
			"txHash": txHash,
		})
	case "EvidencePreserveEpochIDQuery":
		epochID := c.DefaultQuery("epochID", "")
		//fmt.Println(epochID)
		if epochID == "" {
			return false, nil
		}
		e, err := strconv.Atoi(epochID)
		if err != nil {
			return false, nil
		}
		return true, NewEvidencePreserveEpochIDQuery(map[interface{}]interface{}{
			"epochID": e,
		})
	case "BlockchainLatestInfoQuery":
		return true, NewBlockchainLatestInfoQuery()
	case "BlockchainBlockHashQuery":
		blockHash := c.DefaultQuery("blockHash", "")
		if blockHash == "" {
			return false, nil
		}
		return true, NewBlockchainBlockHashQuery(map[interface{}]interface{}{
			"blockHash": blockHash,
		})
	case "BlockchainBlockNumberQuery":
		blockNumber := c.DefaultQuery("blockNumber", "")
		if blockNumber == "" {
			return false, nil
		}
		b, err := strconv.Atoi(blockNumber)
		if err != nil {
			return false, nil
		}
		return true, NewBlockchainBlockNumberQuery(map[interface{}]interface{}{
			"blockNumber": b,
		})
	case "BlockchainTransactionQuery":
		txHash := c.DefaultQuery("txHash", "")
		if txHash == "" {
			return false, nil
		}
		return true, NewBlockchainTransactionQuery(map[interface{}]interface{}{
			"txHash": txHash,
		})
	case "NodesStatusQuery":
		return true, NewDataSynthMonitorQuery()
	case "DateSynthDataQuery":
		return true, NewDateSynthDataQuery()
	case "DateTransactionQuery":
		return true, NewDateTransactionQuery()
	case "SynthTaskQuery":
		item := NewSynthTaskQuery()
		filters := make(map[interface{}]interface{})
		for _, key := range []string{"taskID", "taskName", "txHash", "status", "model", "startDate", "endDate"} {
			filters[key] = c.Query(key)
		}
		if !item.ParseRawDataFromHttpEngine(filters) {
			return false, nil
		}
		return true, item
	case "TaskOnNodesQuery":
		taskID := c.DefaultQuery("taskID", "")
		if taskID == "" {
			return false, nil
		}
		return true, NewTaskOnNodesQuery(map[interface{}]interface{}{
			"taskID": taskID,
		})
	case "CollectTaskQuery":
		taskID := c.DefaultQuery("taskID", "")
		if taskID == "" {
			return false, nil
		}
		size := c.DefaultQuery("size", "")
		if size == "" {
			return false, nil
		}
		s, err := strconv.Atoi(size)
		if err != nil {
			return false, nil
		}
		return true, NewCollectTaskQuery(map[interface{}]interface{}{
			"taskID": taskID,
			"size":   s,
		})
	case "UploadTaskQuery":
		taskID := c.DefaultQuery("taskID", "")
		dataType := c.DefaultQuery("dataType", "marketRisk")
		purpose := c.DefaultQuery("purpose", "")
		description := c.DefaultQuery("description", "")
		createBy := c.DefaultQuery("createBy", "")
		if taskID == "" {
			return false, nil
		}
		// 构造 UploadTaskQuery 对象
		return true, NewUploadTaskQuery(map[interface{}]interface{}{
			"taskID":      taskID,
			"dataType":    dataType,
			"purpose":     purpose,
			"description": description,
			"createBy":    createBy,
		})
	case "SlotIntegrityVerification":
		slotHash := c.DefaultQuery("slotHash", "")
		structType := c.DefaultQuery("type", "merkle")
		if slotHash == "" {
			return false, nil
		}
		return true, NewSlotIntegrityVerification(map[interface{}]interface{}{
			"slotHash": slotHash,
			"type":     structType,
		})
	default:
		return false, nil
	}
}
