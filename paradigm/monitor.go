package paradigm

import (
	"BHLayer2Node/pb/service"
	"fmt"
	"strconv"
)

type NodeHeartbeatReport struct {
	NodeID      int32
	CPUUsage    int
	DiskUsage   int32
	DiskStorage int32
	IsError     bool
	ErrMessage  string
}

func NewNodeHeartbeatReportFromHeartbeat(heartbeat *service.HeartbeatResponse) NodeHeartbeatReport {
	status := heartbeat.NodeStatus
	nodeID := heartbeat.NodeId
	if _, exist := status["cpu"]; !exist {
		e := Error(ExecutorError, "Invalid Heartbeat Node Status, Error status key: cpu")
		return NewErrorNodeHeartbeatReport(nodeID, e.Error())
	}
	if _, exist := status["disk"]; !exist {
		e := Error(ExecutorError, "Invalid Heartbeat Node Status, Error status key: disk")
		return NewErrorNodeHeartbeatReport(nodeID, e.Error())
	}

	if _, exist := status["total"]; !exist {
		e := Error(ExecutorError, "Invalid Heartbeat Node Status, Error status key: total")
		return NewErrorNodeHeartbeatReport(nodeID, e.Error())
	}
	c, d, t := status["cpu"], status["disk"], status["total"]
	cpuUsage, ok := strconv.Atoi(c)
	if ok != nil {
		e := Error(ExecutorError, fmt.Sprintf("Invalid Heartbeat Node Status, Error cpu status value: %s", c))
		return NewErrorNodeHeartbeatReport(nodeID, e.Error())
	}
	diskUsage, ok := strconv.Atoi(d)
	if ok != nil {
		e := Error(ExecutorError, fmt.Sprintf("Invalid Heartbeat Node Status, Error disk usage value: %s", d))
		return NewErrorNodeHeartbeatReport(nodeID, e.Error())
	}
	diskStorage, ok := strconv.Atoi(t)
	if ok != nil {
		e := Error(ExecutorError, fmt.Sprintf("Invalid Heartbeat Node Status, Error disk storage value: %s", t))
		return NewErrorNodeHeartbeatReport(nodeID, e.Error())
	}
	Log("INFO", fmt.Sprintf("Monitor Update Node %d Status, CPU Usage: %d %%, Disk Usage: %d, Total Disk Space: %d", heartbeat.NodeId, cpuUsage, diskUsage, diskStorage))
	return NodeHeartbeatReport{
		NodeID:      nodeID,
		CPUUsage:    cpuUsage,
		DiskUsage:   int32(diskUsage),
		DiskStorage: int32(diskStorage),
		IsError:     false,
		ErrMessage:  "",
	}
}
func NewErrorNodeHeartbeatReport(nodeID int32, error string) NodeHeartbeatReport {
	return NodeHeartbeatReport{
		NodeID:      nodeID,
		CPUUsage:    0,
		DiskUsage:   0,
		DiskStorage: 0,
		IsError:     true,
		ErrMessage:  error,
	}
}

type NodeStatus struct {
	NodeID          int32
	Address         BHNodeAddress
	AverageCPUUsage int                        // 平均cpu使用率,%
	DiskUsage       int32                      // 存储状况，B为单位 todo 这里要考虑是否会溢出，以及是否用小数比较好
	FinishedSlots   []SlotHash                 // 已经完成的任务， todo 这里其实要考虑节点虚报，对于真正的完整的monitor来说应该是需要根据oracle来统计的
	PendingSlots    map[SlotHash]bool          // 处理中的任务 todo 同上
	SynthData       map[SupportModelType]int32 // 合成数据总量, 根据modeltype划分 todo 同上
	DiskStorage     int32                      // 磁盘总量，GB为单位
	isError         bool                       // 是否存在错误
	errMessage      string                     // 错误信息
	Rate            float64                    // 评分，用于Monitor Advice，暂时先不写
}

func (s *NodeStatus) UpdateUsage(cpu int, disk int32, total int32) {
	if cpu >= 0 {
		s.AverageCPUUsage = cpu
	}
	if disk >= 0 {
		s.DiskUsage = disk
	}
	if total >= 0 {
		s.DiskStorage = total
	}
	s.DiscardError()
}
func (s *NodeStatus) UpdatePendingSlot(slotHash string) {
	s.PendingSlots[slotHash] = true
	s.DiscardError()
}
func (s *NodeStatus) UpdateFinishSlot(slotHash string, process int32, modelType SupportModelType) {
	delete(s.PendingSlots, slotHash)
	s.FinishedSlots = append(s.FinishedSlots, slotHash)
	if value, ok := s.SynthData[modelType]; ok {
		s.SynthData[modelType] = value + process
	} else {
		s.SynthData[modelType] = process
	}
	s.DiscardError()
}
func (s *NodeStatus) IsError() bool {
	return s.isError
}
func (s *NodeStatus) ErrorMessage() string {
	return s.errMessage
}
func (s *NodeStatus) SetError(errMessage string) {
	s.isError = true
	s.errMessage = errMessage
}
func (s *NodeStatus) DiscardError() {
	s.isError = false
	s.errMessage = ""
}
func NewNodeStatus(nodeID int32, address BHNodeAddress) *NodeStatus {
	return &NodeStatus{
		NodeID:          nodeID,
		Address:         address,
		AverageCPUUsage: 0,
		DiskUsage:       0,
		FinishedSlots:   make([]SlotHash, 0),
		PendingSlots:    make(map[SlotHash]bool),
		SynthData:       make(map[SupportModelType]int32),
		DiskStorage:     0,
		Rate:            0,
		isError:         false,
		errMessage:      "",
	}
}

// AdviceRequest Scheduler向Monitor请求调度方式
// TODO 这里细节有待思考
type AdviceRequest struct {
	Size     int32 // 全量的大小
	SlotSize int32 //指定的大小
	Model    SupportModelType
	response chan AdviceResponse
}

func (r *AdviceRequest) SendResponse(resp AdviceResponse) {
	r.response <- resp
}
func (r *AdviceRequest) ReceiveResponse() AdviceResponse {
	return <-r.response
}

func NewAdviceRequest(size int32, slotSize int32) *AdviceRequest {
	return &AdviceRequest{
		SlotSize: slotSize,
		Size:     size,
		Model:    CTGAN,
		response: make(chan AdviceResponse, 1),
	}
}

func NewModelAdviceRequest(size int32, slotSize int32, model SupportModelType) *AdviceRequest {
	return &AdviceRequest{
		SlotSize: slotSize,
		Size:     size,
		Model:    model,
		response: make(chan AdviceResponse, 1),
	}
}

// AdviceResponse 调度结果，TODO
type AdviceResponse struct {
	NodeIDs      []int32
	ScheduleSize []int32
}

func NewAdviceResponse(nodeIDs []int32, size []int32) *AdviceResponse {
	return &AdviceResponse{
		NodeIDs:      nodeIDs,
		ScheduleSize: size,
	}
}
