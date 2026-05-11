package HTTP

import (
	"BHLayer2Node/Database"
	"BHLayer2Node/Monitor"
	"BHLayer2Node/Network/Grpc"
	"BHLayer2Node/Network/HTTP/abm"
	"BHLayer2Node/PKI"
	"BHLayer2Node/paradigm"
	"fmt"
	"time"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
)

// HttpEngine 定义模拟的 HTTP 引擎,这里引入gin框架提高编码效率
type HttpEngine struct {
	channel        *paradigm.RappaChannel
	taskIDConsumer chan int // 这里暂时用这个方法获取TaskID
	taskIDProvider chan paradigm.TaskHash
	config         paradigm.BHLayer2NodeConfig
	ip             string // IP 地址
	port           int    // 端口
	pkiManager     *PKI.PKIManager
	dbService      *Database.DatabaseService
	monitor        *Monitor.Monitor
	grpcManager    *Grpc.NodeGrpcManager // 新增：用于与节点进行 gRPC 通信
	// 服务器
	r *gin.Engine
}

// AccumulateTaskID 不断累加即可
func (e *HttpEngine) AccumulateTaskID() {
	taskID := 0
	// 这里是一个额外的协程，在获取ID的时候阻塞
	for {
		<-e.taskIDConsumer
		e.taskIDProvider <- fmt.Sprintf("SynthTask-%d-%d", taskID, time.Now().Unix())
		taskID++
	}

}

func (e *HttpEngine) Start() {
	go e.AccumulateTaskID()
	abm.StartABMParameterIndexRefresher(e.config.AbmParameters, &e.config)
	paradigm.Print("INFO", fmt.Sprintf("Http server run on port %s:%d", e.ip, e.port))
	err := e.r.Run(fmt.Sprintf(":%d", e.port))
	if err != nil {
		paradigm.Error(paradigm.NetworkError, "Faild to start http engine because of"+err.Error())
	}
}

// Setup 配置 HTTP 引擎
func (e *HttpEngine) Setup(config paradigm.BHLayer2NodeConfig) {
	e.config = config
	e.port = config.HttpPort
	e.ip = "127.0.0.1" // 默认绑定到本地地址
	if e.config.DEBUG {
		gin.SetMode(gin.DebugMode)
	} else {
		// 设置 Gin 为 release 模式
		gin.SetMode(gin.ReleaseMode)
	}
	e.r = gin.Default()
	e.r.Use(cors.Default())
	abm.InitABMParameterIndex(e.config.AbmParameters, &e.config)

	// 注册url
	//e.r.POST("/task", e.HandleRequest)
	for _, s := range e.SupportUrl() {
		service, err := e.GetHttpService(s)
		if err != nil {
			paradigm.Error(paradigm.RuntimeError, "url service not impl")
			continue
		}
		if service.Method == "POST" {
			e.r.POST(service.Url, service.Handler)
		} else if service.Method == "GET" {
			e.r.GET(service.Url, service.Handler)
		} else {
			// TODO
		}
	}
}

// NewHttpEngine 创建并返回一个新的 HttpEngine 实例
func NewHttpEngine(channel *paradigm.RappaChannel, pkiManager *PKI.PKIManager, dbService *Database.DatabaseService, monitor *Monitor.Monitor) *HttpEngine {
	http := HttpEngine{
		channel:        channel,
		taskIDProvider: make(chan paradigm.TaskHash, 100),
		taskIDConsumer: make(chan int, 100),
		pkiManager:     pkiManager,
		dbService:      dbService,
		monitor:        monitor,
		grpcManager:    Grpc.NewNodeGrpcManager(channel.Config.BHNodeAddressMap),
	}
	http.Setup(*channel.Config)
	return &http
}
