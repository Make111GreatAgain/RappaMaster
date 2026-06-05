package Oracle

import (
	"BHLayer2Node/Query"
	"BHLayer2Node/paradigm"
	"fmt"
)

// TODO 这里会有很多并发读写的问题，后面要用TryLock,如果发现被锁住了，那么直接返回还未更新
func (o *PersistedOracle) processDBQuery() {
	// 来自Http的query,需要发回去，因此Query里需要有一个通道
	// TODO 这里的txMap应该会有并发读写的问题，而且不能将这个协程合并，会影响性能
	for query := range o.channel.QueryChannel {
		paradigm.Print("ORACLE", fmt.Sprintf("Receive a Query: %v", query.ToHttpJson()))
		switch query.(type) {
		case *Query.EvidencePreserveTaskIDQuery:
			item := query.(*Query.EvidencePreserveTaskIDQuery)
			task, err := o.dbService.GetTaskByID(item.TaskID)
			if err != nil {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.ValueError, err.Error()))
				item.SendResponse(errorResponse)
				paradigm.Error(paradigm.ValueError, err.Error())
				continue
			}
			item.SendResponse(item.GenerateResponse(task))
		case *Query.EvidencePreserveTaskTxQuery:
			// 根据txHash查询Task
			item := query.(*Query.EvidencePreserveTaskTxQuery)
			task, err := o.dbService.GetTaskByTxHash(item.TxHash)
			if err != nil {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.ValueError, err.Error()))
				item.SendResponse(errorResponse)
				paradigm.Error(paradigm.ValueError, err.Error())
				continue
			}
			item.SendResponse(item.GenerateResponse(task))
		case *Query.EvidencePreserveEpochIDQuery:
			item := query.(*Query.EvidencePreserveEpochIDQuery)
			epoch, err := o.dbService.GetEpochByID(item.EpochID)
			if err != nil {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.ValueError, err.Error()))
				item.SendResponse(errorResponse)
				paradigm.Error(paradigm.ValueError, err.Error())
				continue
			}
			item.SendResponse(item.GenerateResponse(epoch))
		case *Query.EvidencePreserveEpochTxQuery:
			item := query.(*Query.EvidencePreserveEpochTxQuery)
			epoch, err := o.dbService.GetEpochByTxHash(item.TxHash)
			if err != nil {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.ValueError, err.Error()))
				item.SendResponse(errorResponse)
				paradigm.Error(paradigm.ValueError, err.Error())
				continue
			}
			item.SendResponse(item.GenerateResponse(epoch))
		case *Query.BlockchainLatestInfoQuery:
			item := query.(*Query.BlockchainLatestInfoQuery)
			// 获取最新交易
			latestTxs, err := o.dbService.GetLatestTransactions(20)
			if err != nil {
				paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Failed to query latest transactions: %v", err))
				continue
			}
			// 获取最新epochs
			latestEpochs, err := o.dbService.GetLatestEpochs(20)
			if err != nil {
				paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Failed to query latest epochs: %v", err))
				continue
			}
			// 获取已完成的slot数量
			nbFinalized, err := o.dbService.GetFinalizedSlotsCount()
			if err != nil {
				paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Failed to count finalized slots: %v", err))
				continue
			}
			// 获取合成数据统计
			synthData, err := o.dbService.GetSynthDataByModel()
			if err != nil {
				paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Failed to get synth data: %v", err))
				continue
			}

			// 获取交易总数
			nbTransaction, err := o.dbService.GetTransactionCount()
			if err != nil {
				paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Failed to count transactions: %v", err))
				continue
			}
			// 构建响应信息
			info := paradigm.LatestBlockchainInfo{
				LatestTxs:     latestTxs,
				LatestEpoch:   latestEpochs,
				NbFinalized:   int32(nbFinalized),
				SynthData:     synthData,
				NbEpoch:       int32(len(latestEpochs)),
				NbTransaction: int32(nbTransaction),
			}

			if len(latestTxs) > 0 {
				info.NbBlock = int32(latestTxs[0].TxReceipt.BlockNumber)
			}

			// 发送响应
			item.SendResponse(item.GenerateResponse(info))
		case *Query.BlockchainBlockNumberQuery: // 链上查询
			item := query.(*Query.BlockchainBlockNumberQuery)
			o.channel.BlockchainQueryChannel <- item
			block := item.ReceiveInfo()
			item.SendResponse(item.GenerateResponse(block))
		case *Query.BlockchainBlockHashQuery: // 链上查询
			item := query.(*Query.BlockchainBlockHashQuery)
			o.channel.BlockchainQueryChannel <- item
			block := item.ReceiveInfo()
			item.SendResponse(item.GenerateResponse(block))
		case *Query.BlockchainTransactionQuery:
			item := query.(*Query.BlockchainTransactionQuery)
			// 从区块链获取交易信息
			o.channel.BlockchainQueryChannel <- item
			tx := item.ReceiveInfo()
			txInfo := tx.(paradigm.TransactionInfo)
			// 从数据库获得交易信息
			ref, err := o.dbService.GetTransactionByHash(txInfo.TxHash)
			if err != nil {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.ValueError, err.Error()))
				item.SendResponse(errorResponse)
				paradigm.Error(paradigm.ValueError, err.Error())
				continue
			}
			switch ref.Rf {
			case paradigm.InitTaskTx:
				txInfo.Abi = "InitTask"
			case paradigm.SlotTX:
				txInfo.Abi = "TaskProcess"
			case paradigm.EpochTx:
				txInfo.Abi = "EpochRecord"
			default:
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.RuntimeError, "Unknown transaction reference type"))
				item.SendResponse(errorResponse)
				paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Unknown transaction reference type: %v", ref.Rf))
			}

			item.SendResponse(item.GenerateResponse(txInfo))
		case *Query.NodesStatusQuery:
			item := query.(*Query.NodesStatusQuery)
			o.channel.MonitorQueryChannel <- item
			status := item.ReceiveInfo()
			item.SendResponse(item.GenerateResponse(status))
		case *Query.DateSynthDataQuery:
			item := query.(*Query.DateSynthDataQuery)
			records, err := o.dbService.GetDateRecords()
			proccessTasks := o.dbService.CountTaskByStatus(paradigm.Processing)
			finishTasks := o.dbService.CountTaskByStatus(paradigm.Finished)
			data := map[string]interface{}{
				"records":         records,
				"proccessTaskNum": proccessTasks,
				"finishTaskNum":   finishTasks,
			}
			if err != nil {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.RuntimeError, fmt.Sprintf("Failed to query date records: %v", err)))
				item.SendResponse(errorResponse)
				paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Date synth data query failed: %v", err))
				continue
			}
			item.SendResponse(item.GenerateResponse(data))
		case *Query.DateTransactionQuery:
			item := query.(*Query.DateTransactionQuery)
			records, err := o.dbService.GetDateRecords()
			if err != nil {
				errorResponse := paradigm.NewErrorResponse(paradigm.NewRappaError(paradigm.RuntimeError, fmt.Sprintf("Failed to query date records: %v", err)))
				item.SendResponse(errorResponse)
				paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Date synth data query failed: %v", err))
				continue
			}
			item.SendResponse(item.GenerateResponse(records))
		case *Query.SynthTaskQuery:
			item := query.(*Query.SynthTaskQuery)
			// 获取所有任务
			tasksMap, err := o.dbService.GetAllTasks()
			if err != nil {
				errorResponse := paradigm.NewErrorResponse(
					paradigm.NewRappaError(paradigm.RuntimeError,
						fmt.Sprintf("Failed to query tasks: %v", err)))
				item.SendResponse(errorResponse)
				paradigm.Error(paradigm.RuntimeError,
					fmt.Sprintf("Synth task query failed: %v", err))
				continue
			}
			item.SendResponse(item.GenerateResponse(tasksMap))
		case *Query.TaskOnNodesQuery:
			item := query.(*Query.TaskOnNodesQuery)
			slots := o.dbService.QueryFinishedSlotsByTask(item.Sign)
			item.SendResponse(item.GenerateResponse(slots))
		case *Query.CollectTaskQuery:
			item := query.(*Query.CollectTaskQuery)
			task, err := o.dbService.GetTaskByID(item.TaskID())
			// task.SetCollector(o.collectors[task.Sign])
			if err != nil {
				errorResponse := paradigm.NewErrorResponse(
					paradigm.NewRappaError(paradigm.RuntimeError,
						fmt.Sprintf("Failed to query tasks: %v", err)))
				item.SendResponse(errorResponse)
				paradigm.Error(paradigm.RuntimeError,
					fmt.Sprintf("Synth task query failed: %v", err))
				continue
			}
			// 从数据库中恢复Collector
			err = o.dbService.RecoverCollector(task, o.manager)
			if err != nil {
				errorResponse := paradigm.NewErrorResponse(
					paradigm.NewRappaError(paradigm.RuntimeError,
						fmt.Sprintf("Failed to query tasks: %v", err)))
				item.SendResponse(errorResponse)
				paradigm.Error(paradigm.RuntimeError,
					fmt.Sprintf("Synth task query failed: %v", err))
				continue
			}
			go item.SendResponse(item.GenerateResponse(task.GetCollector()))
		case *Query.SlotIntegrityVerification: // 完整性验证
			item := query.(*Query.SlotIntegrityVerification)
			slots := o.dbService.QueryFinishedSlotsBySlot(item.SlotHash)
			item.SendResponse(item.GenerateResponse(slots))
		case *Query.UploadTaskQuery:
			item := query.(*Query.UploadTaskQuery)
			task, err := o.dbService.GetTaskByID(item.TaskID())
			if err != nil {
				errorResponse := paradigm.NewErrorResponse(
					paradigm.NewRappaError(paradigm.RuntimeError,
						fmt.Sprintf("Failed to query task %s: %v", item.TaskID(), err)))
				item.SendResponse(errorResponse)
				paradigm.Error(paradigm.RuntimeError,
					fmt.Sprintf("Upload task query failed: %v", err))
				continue
			}
			// task.SetCollector(o.collectors[task.Sign])
			// 从数据库中恢复Collector
			err = o.dbService.RecoverCollector(task, o.manager)
			if err != nil {
				errorResponse := paradigm.NewErrorResponse(
					paradigm.NewRappaError(paradigm.RuntimeError,
						fmt.Sprintf("Failed to query tasks: %v", err)))
				item.SendResponse(errorResponse)
				paradigm.Error(paradigm.RuntimeError,
					fmt.Sprintf("Synth task query failed: %v", err))
				continue
			}
			item.SendResponse(item.GenerateResponse(task))
		default:
			panic("Unsupported Query Type!!!")
		}
	}
}
