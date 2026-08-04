package Mocker

import (
	"BHLayer2Node/paradigm"
	"BHLayer2Node/pb/service"
	"context"
	"fmt"
	"math/rand"
	"strconv"
)

//
//func (m *MockerExecutionNode) mustEmbedUnimplementedCoordinatorServer() {
//	//TODO implement me
//	panic("implement me")
//}

// Schedule 实现 gRPC Schedule 方法
func (m *MockerExecutionNode) Schedule(ctx context.Context, req *service.ScheduleRequest) (*service.ScheduleResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	paradigm.Log("DEBUG", fmt.Sprintf("Node %d received ScheduleRequest: %+v", m.nodeID, req))

	// 模拟随机接受或拒绝任务
	accept := m.nodeID != 0 || rand.Intn(4) == 0

	if !accept {
		return &service.ScheduleResponse{
			Sign:         req.Sign,
			NodeId:       strconv.Itoa(m.nodeID),
			Accept:       false,
			ErrorMessage: "Epoch rejected due to resource constraints",
		}, nil
	}

	// 如果接受任务，将其存储到 slotData
	m.slotData[req.Sign] = strconv.Itoa(int(req.Slot))

	paradigm.Log("DEBUG", fmt.Sprintf("Node %d accepted task %s for slot %d", m.nodeID, req.Sign, req.Slot))
	id, _ := strconv.Atoi(req.Sign)
	size := req.Size
	slot := req.Slot
	go func() {
		//time.Sleep(2 * time.Second)
		process := size - 1
		if process <= 0 {
			process = 1
		}
		paradigm.Log("DEBUG", fmt.Sprintf("Node %d finished %d in Epoch %s Slot %d", id, size-1, req.Sign, slot))
		// todo 这里现在是直接提交的，没有走grpc
		m.commitSlot <- paradigm.NewCommitSlotItem(&service.JustifiedSlot{
			Nid:     int32(id),
			Process: process,
			Sign:    req.Sign,
			Slot:    slot,
			Epoch:   -1,
		})
	}()
	return &service.ScheduleResponse{
		Sign:   req.Sign,
		NodeId: strconv.Itoa(m.nodeID),
		Accept: true,
	}, nil
}

func (m *MockerExecutionNode) Heartbeat(ctx context.Context, req *service.HeartbeatRequest) (*service.HeartbeatResponse, error) {
	// todo
	votes := make([]*service.Vote, 0)
	for hash, _ := range req.Commits {
		votes = append(votes, &service.Vote{
			//Slot:   slot,
			Hash:   hash,
			NodeId: int32(m.nodeID),
			State:  true,
			Desp:   "",
		})
	}
	return &service.HeartbeatResponse{
		NodeId:     int32(m.nodeID),
		NodeStatus: make(map[string]string),
		Votes:      votes,
	}, nil
}

//func (m *MockerExecutionNode) EpochVote(ctx context.Context, req *service.EpochVoteRequest) (*service.EpochVoteResponse, error) {
//	return &service.EpochVoteResponse{
//		NodeId:     "",
//		VoteBitmap: nil,
//	}, nil
//}

func (m *MockerExecutionNode) CommitSlot(ctx context.Context, req *service.SlotCommitRequest) (*service.SlotCommitResponse, error) {
	return &service.SlotCommitResponse{
		Seed:    nil,
		Hash:    "",
		Timeout: 0,
	}, nil
}
