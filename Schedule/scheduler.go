package Schedule

import (
	"BHLayer2Node/paradigm"
	"BHLayer2Node/utils"
	"fmt"
	"sync"
	"time"
)

// Scheduler 用于调度任务，产生Slot
// 接收UnprocessedTask，1. 新建的合成任务(上链后); 2. 未完成的任务(由Tracker发现过期，重新进入调度)
type Scheduler struct {
	channel *paradigm.RappaChannel // channel
	//monitor   *Monitor.Monitor          // 监控节点状态，用于进行调度
	schedules map[paradigm.TaskHash]int // 记录每个Task最新的调度index
	mu        sync.Mutex                // 保护 unprocessedTasks 的读写
}

func (s *Scheduler) Start() error {
	paradigm.Log("INFO", "Scheduler started, waiting for tasks...")
	// 处理待调度的task，这些task有可能是由前端http新发的，也可能是task在前一个slot未完成返工的
	processUnprocessedTasks := func() {
		for request := range s.channel.UnprocessedTasks {
			paradigm.Print("SCHEDULE", fmt.Sprintf("Schedule Unprocessed Task: TaskID=%s, Total Size=%d", request.TaskID, request.Size))
			s.process(request)
		}
	}

	go processUnprocessedTasks()
	//go processPendingSlotRecord()

	return nil
}

// process 分配Slot，并更新任务状态
func (s *Scheduler) process(task paradigm.UnprocessedTask) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// // 获取节点列表
	// // 这里测试分配到所有节点
	// //nIDs := make([]int, 0, len(s.channel.Config.BHNodeAddressMap))
	// //for key := range s.channel.Config.BHNodeAddressMap {
	// //	nIDs = append(nIDs, key)
	// //}
	// //
	// //if len(nIDs) == 0 {
	// //	LogWriter.Log("ERROR", "No connected nodes available for scheduling")
	// //	return
	// //}
	// adviceRequest := paradigm.NewAdviceRequest(task.Size, task.SlotSize)
	// s.channel.MonitorAdviceChannel <- adviceRequest
	// resp := adviceRequest.ReceiveResponse()
	// //allocSizes := s.monitor.Advice(nIDs, task.Size) // todo
	// //allocSizes := resp.ScheduleSize
	// if len(resp.ScheduleSize) != len(resp.NodeIDs) {
	// 	panic("Error in Monitor Advice(), len(nIDs) != len(allocSizes)")
	// }
	// // 构建分配计划
	// schedule := s.generateSynthSchedule(task, resp.NodeIDs, resp.ScheduleSize)

	var nodeIDs []int32
	var scheduleSizes []int32

	if task.Model == paradigm.ABM_V2 {
		if nodeID, ok := utils.ExtractAssignedNodeID(task.Params); ok {
			nodeIDs = []int32{nodeID}
			scheduleSizes = []int32{1}
		} else {
			for {
				adviceRequest := paradigm.NewModelAdviceRequest(1, 1, paradigm.ABM_V2)
				s.channel.MonitorAdviceChannel <- adviceRequest
				resp := adviceRequest.ReceiveResponse()
				if len(resp.NodeIDs) > 0 {
					nodeIDs = []int32{resp.NodeIDs[0]}
					scheduleSizes = []int32{1}
					break
				}
				paradigm.Print("SCHEDULE", fmt.Sprintf("No idle node for ABM_V2 task %s, wait and retry", task.TaskID))
				time.Sleep(5 * time.Second)
			}
		}
	} else {
		adviceRequest := paradigm.NewAdviceRequest(task.Size, task.SlotSize)
		s.channel.MonitorAdviceChannel <- adviceRequest
		resp := adviceRequest.ReceiveResponse()
		if len(resp.ScheduleSize) != len(resp.NodeIDs) {
			panic("Error in Monitor Advice(), len(nIDs) != len(allocSizes)")
		}
		nodeIDs = resp.NodeIDs
		scheduleSizes = resp.ScheduleSize
	}
	// 构建分配计划
	schedule := s.generateSynthSchedule(task, nodeIDs, scheduleSizes)
	paradigm.Print("SCHEDULE", fmt.Sprintf("New Schedule for Task %s, Schedule: %d, Size: %d", task.TaskID, schedule.ScheduleID, schedule.Size))
	s.channel.PendingSchedules <- schedule
}
func (s *Scheduler) generateSynthSchedule(task paradigm.UnprocessedTask, nIDs []int32, size []int32) paradigm.SynthTaskSchedule {
	scheduleIndex := s.schedules[task.TaskID]
	if _, exist := s.schedules[task.TaskID]; !exist {
		s.schedules[task.TaskID] = -1
	}
	s.schedules[task.TaskID]++
	computeSlotHash := func(nID int) paradigm.SlotHash {
		scheduleIndex := s.schedules[task.TaskID]
		return fmt.Sprintf("%s_%d_%d", task.TaskID, paradigm.ScheduleHash(scheduleIndex), nID)
	}
	schedule := paradigm.SynthTaskSchedule{
		TaskID:     task.TaskID,
		ScheduleID: paradigm.ScheduleHash(scheduleIndex),
		SlotSize:   task.SlotSize,
		Size:       task.Size,
		Model:      task.Model,
		Params:     task.Params,
		//Slots: make([]*paradigm.Slot, 0),
	}
	slots := make([]*paradigm.Slot, 0)
	nodeIDMap := make(map[int]int)

	for i := 0; i < len(nIDs); i++ {
		nID, scheduleSize := nIDs[i], size[i]
		nodeIDMap[int(nID)] = i
		slot := paradigm.NewSlot(computeSlotHash(int(nID)), task.TaskID, paradigm.ScheduleHash(scheduleIndex), scheduleSize)
		//slots = append(slots, slot)
		slots = append(slots, slot)
	}
	schedule.NodeIDMap = nodeIDMap
	schedule.Slots = slots
	//schedule.Print()
	return schedule
}

//func (s *Scheduler) Setup(config *Config.BHLayer2NodeConfig) {
//	s.config = config
//	//s.pendingRequestPool = make(chan HTTP.UnprocessedTask, config.MaxHttpRequestPoolSize)
//	//s.pendingSlotPool = make(chan PendingSlotItem, config.MaxGrpcRequestPoolSize)
//	//s.pendingSlotRecord = make(chan SlotRecord, config.MaxGrpcRequestPoolSize)
//}

//	func (s *Scheduler) SetGrpc(grpc Grpc.GrpcInterface) {
//		s.grpcEngine = grpc
//	}

// NewScheduler 创建新的 Scheduler todo
func NewScheduler(channel *paradigm.RappaChannel) *Scheduler {
	return &Scheduler{
		//grpcEngine:       nil,
		channel:   channel,
		schedules: make(map[paradigm.TaskHash]int),
		mu:        sync.Mutex{},
	}
}
