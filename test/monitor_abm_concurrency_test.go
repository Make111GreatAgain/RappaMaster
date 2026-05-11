package test

import (
	"BHLayer2Node/Monitor"
	"BHLayer2Node/paradigm"
	"testing"
)

func TestMonitorAdviceLimitsABMV2Concurrency(t *testing.T) {
	channel := newMonitorTestChannel()
	channel.Config.ABMV2MaxConcurrency = 2
	monitor := Monitor.NewMonitor(channel)
	monitor.Start()

	for i := 0; i < 2; i++ {
		resp := requestMonitorAdvice(channel, paradigm.NewModelAdviceRequest(1, 1, paradigm.ABM_V2))
		if len(resp.NodeIDs) != 1 || len(resp.ScheduleSize) != 1 {
			t.Fatalf("ABM_V2 advice %d should return one node, got %#v", i, resp)
		}
	}

	resp := requestMonitorAdvice(channel, paradigm.NewModelAdviceRequest(1, 1, paradigm.ABM_V2))
	if len(resp.NodeIDs) != 0 || len(resp.ScheduleSize) != 0 {
		t.Fatalf("expected no advice after ABM_V2 concurrency cap is reached, got %#v", resp)
	}
}

func requestMonitorAdvice(channel *paradigm.RappaChannel, request *paradigm.AdviceRequest) paradigm.AdviceResponse {
	channel.MonitorAdviceChannel <- request
	return request.ReceiveResponse()
}

func newMonitorTestChannel() *paradigm.RappaChannel {
	return &paradigm.RappaChannel{
		Config: &paradigm.BHLayer2NodeConfig{
			BHNodeAddressMap: map[int]*paradigm.BHNodeAddress{
				0: {NodeIPAddress: "127.0.0.1", NodeGrpcPort: 9000},
				1: {NodeIPAddress: "127.0.0.1", NodeGrpcPort: 9001},
				2: {NodeIPAddress: "127.0.0.1", NodeGrpcPort: 9002},
				3: {NodeIPAddress: "127.0.0.1", NodeGrpcPort: 9003},
			},
		},
		MonitorAdviceChannel: make(chan *paradigm.AdviceRequest, 10),
	}
}
