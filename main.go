package main

import (
	"BHLayer2Node/Coordinator"
	"BHLayer2Node/Database"
	"BHLayer2Node/Epoch"
	"BHLayer2Node/Event"
	"BHLayer2Node/LightChainUpper"
	"BHLayer2Node/Monitor"
	"BHLayer2Node/Network/HTTP"
	"BHLayer2Node/Oracle"
	"BHLayer2Node/PKI"
	"BHLayer2Node/Recovery"
	"BHLayer2Node/Schedule"
	"BHLayer2Node/paradigm"
	"fmt"
)

func catchPanic() {
	if r := recover(); r != nil {
		// 捕获 panic 错误并输出
		paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Recovered from panic: %v", r))
	}
}
func main() {
	defer catchPanic()
	config := paradigm.LoadBHLayer2NodeConfig("config.json")
	rappaChannel := paradigm.NewRappaChannel(config)
	// 加载数据库
	dbService, err := Database.NewDatabaseService(config, rappaChannel)
	if err != nil {
		paradigm.Error(paradigm.DatabaseError, fmt.Sprintf("Failed to init database: %v", err))
	}
	// 恢复数据
	recovery := Recovery.RecoverFromDataBase(config, dbService)
	pkiManager := PKI.NewPKIManager(config)
	// 初始化各个组件
	//grpcEngine := Grpc.NewFakeGrpcEngine(pendingSlotPool, pendingSlotRecord)
	//grpcEngine.Setup(*config)
	//httpEngine := HTTP.NewFakeHttpEngine(rappaChannel)
	//httpEngine.Setup(*config)
	event := Event.NewEvent(rappaChannel)
	coordinator := Coordinator.NewCoordinator(rappaChannel)
	epochManager := Epoch.NewEpochManager(rappaChannel, recovery, pkiManager)
	//chainUpper, _ := ChainUpper.NewMockerChainUpper(rappaChannel) // todo @XQ 测试的时候用的是这个mocker
	oracle := Oracle.NewPersistedOracle(rappaChannel, dbService, pkiManager)
	monitor := Monitor.NewMonitor(rappaChannel)
	httpEngine := HTTP.NewHttpEngine(rappaChannel, pkiManager, dbService, monitor)
	chainUpper, err := LightChainUpper.NewLightChainUpper(rappaChannel)
	if err != nil {
		paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Failed to initialize ChainUpper: %v", err))
	}

	// 初始化 Scheduler
	scheduler := Schedule.NewScheduler(rappaChannel)
	// 配置 Scheduler
	//scheduler.Setup(config)
	//scheduler.SetGrpc(grpcEngine)
	//scheduler.SetTaskManager(taskManager)

	// 启动各个组件
	//go grpcEngine.Start()
	go httpEngine.Start()
	go epochManager.Start()
	go coordinator.Start()
	//定时，如果大于10s,EpochEvent队列里放置一个true
	go event.Start()
	go monitor.Start()
	go chainUpper.Start()
	//go collector.Start()
	// 启动 Scheduler
	if err := scheduler.Start(); err != nil {
		paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Failed to start scheduler: %v", err))
		return
	}

	// 主程序保持运行，等待任务完成
	//paradigm.Log("INFO", "Main program is running...")
	//time.Sleep(200 * time.Second)
	oracle.Start()
}
