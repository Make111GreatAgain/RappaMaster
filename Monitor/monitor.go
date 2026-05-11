package Monitor

import (
	"BHLayer2Node/Query"
	"BHLayer2Node/paradigm"
	"sync"
)

// Monitor 监视节点状态
// 存储节点的状态信息，并维护一些统计值
type Monitor struct {
	channel           *paradigm.RappaChannel
	nodeStatus        []*paradigm.NodeStatus
	reservedLoads     map[int32]int
	reservedABMV2     int
	pendingSlotModels map[paradigm.SlotHash]paradigm.SupportModelType
	mu                sync.Mutex
}

// processHeartbeatResponse 处理节点的心跳回复，其中包含节点最新的磁盘、cpu等信息用于展示
func (m *Monitor) processHeartbeatResponse() {
	for report := range m.channel.MonitorHeartbeatChannel {
		if report.IsError {
			// 发现是错误的，那么对NodeStatus更新错误
			m.nodeStatus[int(report.NodeID)].SetError(report.ErrMessage)
		} else {
			m.nodeStatus[int(report.NodeID)].UpdateUsage(report.CPUUsage, report.DiskUsage, report.DiskStorage)
		}
	}
}

// processOracleInfo 处理来自Oracle的更新信息，用于更新nodeStatus中的任务和调度相关的内容
func (m *Monitor) processOracleInfo() {
	for info := range m.channel.MonitorOracleChannel {
		switch info.(type) {
		// 传过来的内容要么是schedule，要么是taskprocessTransaction
		case *paradigm.SynthTaskSchedule:
			schedule := info.(*paradigm.SynthTaskSchedule)
			m.mu.Lock()
			for nodeID, index := range schedule.NodeIDMap {
				slot := schedule.Slots[index]
				// 调度时就失败的就不用更新了
				if slot.Status == paradigm.Failed {
					continue
				}
				m.nodeStatus[nodeID].UpdatePendingSlot(slot.SlotID)
				m.pendingSlotModels[slot.SlotID] = schedule.Model
				if m.reservedLoads[int32(nodeID)] > 0 {
					m.reservedLoads[int32(nodeID)]--
				}
				if schedule.Model == paradigm.ABM_V2 && m.reservedABMV2 > 0 {
					m.reservedABMV2--
				}
				//paradigm.Log("INFO", fmt.Sprintf("Monitor Update Node %d Status, New Pending Slot: %s", nodeID, slot.SlotID))

			}
			m.mu.Unlock()
		case *paradigm.TaskProcessTransaction:
			tx := info.(*paradigm.TaskProcessTransaction)
			nodeID := tx.Nid
			m.mu.Lock()
			m.nodeStatus[int(nodeID)].UpdateFinishSlot(tx.SlotHash(), tx.Process, tx.Model)
			delete(m.pendingSlotModels, tx.SlotHash())
			m.mu.Unlock()
			//paradigm.Log("INFO", fmt.Sprintf("Monitor Update Node %d Status, New Finish Slot: %s, process: %d", nodeID, tx.SlotHash(), tx.Process))

		default:
			paradigm.Error(paradigm.RuntimeError, "Error type in oracle channel")
		}
	}
}

// processAdviceRequest 处理来自Scheduler的请求
func (m *Monitor) processAdviceRequest() {
	for request := range m.channel.MonitorAdviceChannel {
		m.advice(request)
	}

}

// processQuery 处理查询，这里针对monitor的查询就是所有的节点
func (m *Monitor) processQuery() {
	for query := range m.channel.MonitorQueryChannel {
		// 就是全部返回
		switch query.(type) {
		case *Query.NodesStatusQuery:
			item := query.(*Query.NodesStatusQuery)
			item.SendInfo(m.nodeStatus)
		default:
			paradigm.Error(paradigm.RuntimeError, "Unsupported Query Type In Monitor")
		}
	}
}

// TODO 这里调度还得改下，不能给标记为失败的Node一直发送调度
func (m *Monitor) advice(request *paradigm.AdviceRequest) {
	if request.Size <= request.SlotSize {
		m.mu.Lock()
		if request.Model == paradigm.ABM_V2 && m.abmV2ActiveLocked() >= m.abmV2MaxConcurrency() {
			m.mu.Unlock()
			response := paradigm.NewAdviceResponse([]int32{}, []int32{})
			request.SendResponse(*response)
			return
		}
		nodeID, load := m.selectLeastLoadedNodeAndLoadWithReservationsLocked(m.reservedLoads)
		if nodeID >= 0 && load == 0 {
			m.reservedLoads[nodeID]++
			if request.Model == paradigm.ABM_V2 {
				m.reservedABMV2++
			}
		}
		m.mu.Unlock()
		if nodeID < 0 || load > 0 {
			response := paradigm.NewAdviceResponse([]int32{}, []int32{})
			request.SendResponse(*response)
			return
		}
		response := paradigm.NewAdviceResponse([]int32{nodeID}, []int32{request.Size})
		request.SendResponse(*response)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	nodeIDs := make([]int32, len(m.nodeStatus))
	scheduleSize := make([]int32, len(m.nodeStatus))
	for i := 0; i < len(m.nodeStatus); i++ {
		nodeIDs[i] = int32(i)
		adviceSize := request.Size / int32(len(nodeIDs))

		// 这里设置一个指定的均值
		if adviceSize > request.SlotSize {
			adviceSize = request.SlotSize
		}
		if adviceSize == 0 {
			adviceSize = 1
		}

		scheduleSize[i] = adviceSize
	}

	scheduleSize[0] += request.Size % int32(len(nodeIDs))
	response := paradigm.NewAdviceResponse(nodeIDs, scheduleSize)
	request.SendResponse(*response)
}

func (m *Monitor) abmV2MaxConcurrency() int {
	if m.channel == nil || m.channel.Config == nil || m.channel.Config.ABMV2MaxConcurrency <= 0 {
		return 4
	}
	return m.channel.Config.ABMV2MaxConcurrency
}

func (m *Monitor) abmV2ActiveLocked() int {
	active := m.reservedABMV2
	for _, model := range m.pendingSlotModels {
		if model == paradigm.ABM_V2 {
			active++
		}
	}
	return active
}

// SelectLeastLoadedNodeExcluding 选择负载最低的节点，并允许排除当前批次中已经预留过的节点。
// 这样平台一次创建多个“单股票单节点”子任务时，不会因为监控状态尚未刷新而都落到同一个节点。
func (m *Monitor) SelectLeastLoadedNodeExcluding(excluded map[int32]struct{}) int32 {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.nodeStatus) == 0 {
		return -1
	}

	bestNodeID := int32(-1)
	minLoad := -1

	for i := 0; i < len(m.nodeStatus); i++ {
		nodeID := int32(i)
		if _, skip := excluded[nodeID]; skip {
			continue
		}
		load := len(m.nodeStatus[i].PendingSlots)
		if bestNodeID == -1 || load < minLoad {
			minLoad = load
			bestNodeID = nodeID
		}
	}

	// 如果可选节点都被排除了，就退化回原来的“全局最小负载”策略。
	if bestNodeID == -1 {
		bestNodeID = int32(0)
		minLoad = len(m.nodeStatus[0].PendingSlots)
		for i := 1; i < len(m.nodeStatus); i++ {
			load := len(m.nodeStatus[i].PendingSlots)
			if load < minLoad {
				minLoad = load
				bestNodeID = int32(i)
			}
		}
	}
	return bestNodeID
}

// SelectLeastLoadedNodeWithReservations 按“当前待处理槽位 + 本批次已预留数量”选择负载最低节点。
// 平台 ABM_V2 会一次创建多只股票子任务，Monitor 的 PendingSlots 需要等调度落链后才更新；
// 如果创建阶段只看 PendingSlots，批量任务会在状态刷新前反复选中同一个节点。
func (m *Monitor) SelectLeastLoadedNodeWithReservations(reserved map[int32]int) int32 {
	m.mu.Lock()
	defer m.mu.Unlock()
	nodeID, _ := m.selectLeastLoadedNodeAndLoadWithReservationsLocked(reserved)
	return nodeID
}

func (m *Monitor) selectLeastLoadedNodeWithReservationsLocked(reserved map[int32]int) int32 {
	nodeID, _ := m.selectLeastLoadedNodeAndLoadWithReservationsLocked(reserved)
	return nodeID
}

func (m *Monitor) selectLeastLoadedNodeAndLoadWithReservationsLocked(reserved map[int32]int) (int32, int) {
	if len(m.nodeStatus) == 0 {
		return -1, 0
	}

	bestNodeID := int32(0)
	minLoad := len(m.nodeStatus[0].PendingSlots) + reserved[0]
	for i := 1; i < len(m.nodeStatus); i++ {
		nodeID := int32(i)
		load := len(m.nodeStatus[i].PendingSlots) + reserved[nodeID]
		if load < minLoad {
			minLoad = load
			bestNodeID = nodeID
		}
	}
	return bestNodeID, minLoad
}

// SelectLeastLoadedNode 保留旧接口，供现有调用方继续复用。
func (m *Monitor) SelectLeastLoadedNode() int32 {
	return m.SelectLeastLoadedNodeExcluding(map[int32]struct{}{})
}

func (m *Monitor) Start() {
	go m.processHeartbeatResponse()
	go m.processOracleInfo()
	go m.processAdviceRequest()
	go m.processQuery()
}

func NewMonitor(channel *paradigm.RappaChannel) *Monitor {
	nodeStatus := make([]*paradigm.NodeStatus, len(channel.Config.BHNodeAddressMap)) // 这里假设key是对应的[n]
	for nodeID, address := range channel.Config.BHNodeAddressMap {
		nodeStatus[nodeID] = paradigm.NewNodeStatus(int32(nodeID), *address)
	}
	return &Monitor{
		channel:           channel,
		nodeStatus:        nodeStatus,
		reservedLoads:     make(map[int32]int),
		pendingSlotModels: make(map[paradigm.SlotHash]paradigm.SupportModelType),
	}

}
