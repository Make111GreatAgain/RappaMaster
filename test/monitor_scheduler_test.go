package test

import (
	"BHLayer2Node/Monitor"
	"BHLayer2Node/paradigm"
	"testing"
)

func TestMonitorSelectLeastLoadedNodeWithReservationsBalancesBatch(t *testing.T) {
	monitor := Monitor.NewMonitor(newMonitorTestChannel())
	reserved := map[int32]int{}
	for i := 0; i < 51; i++ {
		nodeID := monitor.SelectLeastLoadedNodeWithReservations(reserved)
		reserved[nodeID]++
	}

	if len(reserved) != 4 {
		t.Fatalf("expected all 4 nodes to receive reservations, got %#v", reserved)
	}
	for nodeID, count := range reserved {
		if count < 12 || count > 13 {
			t.Fatalf("node %d should receive 12 or 13 tasks, got %d; all=%#v", nodeID, count, reserved)
		}
	}
}

func TestMonitorAdviceReturnsOnlyIdleNodeForSingleSlot(t *testing.T) {
	channel := newMonitorTestChannel()
	monitor := Monitor.NewMonitor(channel)
	monitor.Start()

	counts := map[int32]int{}
	for i := 0; i < 4; i++ {
		resp := requestMonitorAdvice(channel, paradigm.NewAdviceRequest(1, 1))
		if len(resp.NodeIDs) != 1 || len(resp.ScheduleSize) != 1 {
			t.Fatalf("single-slot advice should return exactly one node, got %#v", resp)
		}
		counts[resp.NodeIDs[0]]++
	}

	for nodeID := int32(0); nodeID < 4; nodeID++ {
		if counts[nodeID] != 1 {
			t.Fatalf("node %d should receive exactly one single-slot reservation, all=%#v", nodeID, counts)
		}
	}

	resp := requestMonitorAdvice(channel, paradigm.NewAdviceRequest(1, 1))
	if len(resp.NodeIDs) != 0 || len(resp.ScheduleSize) != 0 {
		t.Fatalf("expected no advice when all nodes already have reserved work, got %#v", resp)
	}
}
