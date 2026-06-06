package paradigm

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	ecdsa_secp "github.com/consensys/gnark-crypto/ecc/secp256k1/ecdsa"
	"os"
	"strconv"
)

// 定义通信节点的地址配置
type BHNodeAddress struct {
	NodeIPAddress string //节点IP
	NodeGrpcPort  int    //节点grpc端口
	nodeUrl       string //节点访问url
}

// 定义节点公钥配置
type BHNodeKey struct {
	SecpKey ecdsa_secp.PublicKey
	BlSKey  BLS12381PublicKey
}

// 自定义序列化
func (nk *BHNodeKey) UnmarshalJSON(data []byte) error {
	// 定义临时结构体用于解析原始 JSON
	var raw struct {
		SecpKey string `json:"spKey"`
		BlsKey  string `json:"blsKey"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// 解析 secpKey
	secpKey, err := DecodeSecpPublicKey(raw.SecpKey)
	if err != nil {
		return fmt.Errorf("invalid secpKey: %w", err)
	}

	nk.SecpKey = secpKey
	// 解析 blsKey
	blsKey, err := DecodeBLS12381PublicKey(raw.BlsKey)
	if err != nil {
		return fmt.Errorf("invalid blsKey: %w", err)
	}
	nk.BlSKey = blsKey

	return nil
}

// 返回地址字符串
func (b BHNodeAddress) GetAddrStr() string {
	if b.nodeUrl == "" {
		b.nodeUrl = b.NodeIPAddress + ":" + strconv.Itoa(b.NodeGrpcPort)
	}
	return b.nodeUrl
}

type DatabaseConfig struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Dbname       string `json:"dbname"`
	Timeout      string `json:"timeout"`
	MaxIdleConns int    `json:"maxIdleConns"`
	MaxOpenConns int    `json:"maxOpenConns"`
	MaxLifetime  string `json:"maxLifetime"`
}

// 返回dsn字符串
func (d DatabaseConfig) BuildDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=%s",
		d.Username,
		d.Password,
		d.Host,
		d.Port,
		d.Dbname,
		d.Timeout)
}

// BHLayer2NodeConfig 定义 Layer2 节点的配置
type BHLayer2NodeConfig struct {
	GrpcPort                   int // gRPC 服务端口
	HttpPort                   int // HTTP 服务端口
	MaxEpochDelay              int //MaxEpochDelay 对应proto里的timeout
	MaxUnprocessedTaskPoolSize int // HTTP 请求池的最大大小
	MaxPendingSchedulePoolSize int
	MaxScheduledTasksPoolSize  int
	MaxCommitSlotItemPoolSize  int
	MaxGrpcRequestPoolSize     int                    // gRPC 请求池的最大大小
	DefaultSlotSize            int                    // 默认的slot大小
	LogPath                    string                 // 日志路径
	BHNodeAddressMap           map[int]*BHNodeAddress //节点的grpc端口配置，id->nodeIdaddress todo 这里是不是可以改成数组
	DEBUG                      bool
	// 纠删码配置
	// todo 这里考虑可以加一个枚举作为纠删码的选项
	ErasureCodeParamN int
	ErasureCodeParamK int
	//密钥管理配置
	CertPath       string                // 主机密钥路径
	HostPrivateKey ecdsa_secp.PrivateKey // 主机公钥
	BHNodeKeyMap   map[int]*BHNodeKey    // 节点公钥存储

	// ChainUpper 配置
	FiscoBcosHost string // FISCO-BCOS 节点地址
	FiscoBcosPort int    // FISCO-BCOS 节点端口
	GroupID       string // FISCO-BCOS 群组 ID
	PrivateKey    string // 用于签名的私钥
	TLSCaFile     string // TLS CA 证书文件路径
	TLSCertFile   string // TLS 客户端证书路径
	TLSKeyFile    string // TLS 客户端密钥路径
	// ContractAddress       string // 链上合约地址
	QueueBufferSize int // 上链队列缓冲区大小
	WorkerCount     int // Worker 的数量
	IsRecovery      bool
	IsAutoMigrate   bool
	BatchSize       int

	Database                      *DatabaseConfig
	AbmParameters                 map[string]interface{} `json:"-"` // 不直接从 config.json 解析，从独立文件加载
	ABMStockDataDir               string                 // ABM 真实股票输入数据目录
	ABMStockParamDir              string                 // ABM 离线调参参数目录
	ABMV2MaxConcurrency           int                    // ABM_V2 全局最大并发子任务数
	ABMStockDataSource            string                 // ABM 真实行情数据源：local/dolphindb/auto
	ABMRemoteDBHost               string                 // ABM 远程行情库地址
	ABMRemoteDBPort               int                    // ABM 远程行情库端口
	ABMRemoteDBUser               string                 // ABM 远程行情库用户
	ABMRemoteDBPassword           string                 // ABM 远程行情库密码
	ABMRemoteDBName               string                 // ABM 远程行情库 database
	ABMRemoteTableName            string                 // ABM 远程行情库 table
	ABMRemoteCheckPython          string                 // ABM 远程行情校验 Python 可执行文件
	ABMRemoteCheckScript          string                 // ABM 远程行情校验脚本
	ABMRemoteDefaultEndOffsetDays int                    // 默认查询 T-N 日
	ABMUniverseSnapshotRoot       string                 // ABM 指数成分股季度快照目录
}

// DefaultBHLayer2NodeConfig 定义默认的配置值
var DefaultBHLayer2NodeConfig = BHLayer2NodeConfig{
	GrpcPort:                   50051, // 默认 gRPC 端口
	HttpPort:                   8080,  // 默认 HTTP 端口
	MaxEpochDelay:              1,
	MaxUnprocessedTaskPoolSize: 100,
	MaxPendingSchedulePoolSize: 100,
	MaxScheduledTasksPoolSize:  100,
	MaxCommitSlotItemPoolSize:  100,
	MaxGrpcRequestPoolSize:     200, // 默认 gRPC 请求池大小
	DefaultSlotSize:            100,
	LogPath:                    "logs/",
	BHNodeAddressMap:           make(map[int]*BHNodeAddress, 0),
	DEBUG:                      false,

	ErasureCodeParamN: 9,
	ErasureCodeParamK: 6, // 默认配置

	CertPath:       "./cert",
	HostPrivateKey: ecdsa_secp.PrivateKey{},
	BHNodeKeyMap:   make(map[int]*BHNodeKey, 0),

	// 默认 ChainUpper 配置
	FiscoBcosHost: "127.0.0.1",
	FiscoBcosPort: 20200,
	GroupID:       "group0",
	PrivateKey:    "145e247e170ba3afd6ae97e88f00dbc976c2345d511b0f6713355d19d8b80b58",
	TLSCaFile:     "./ChainUpper/ca.crt",
	TLSCertFile:   "./ChainUpper/sdk.crt",
	TLSKeyFile:    "./ChainUpper/sdk.key",
	// ContractAddress: "ChainUpper/contract_address.txt",
	QueueBufferSize: 100000,
	WorkerCount:     3, // 256
	BatchSize:       1,
	IsRecovery:      true,
	IsAutoMigrate:   true,

	Database: &DatabaseConfig{
		Username:     "root",
		Password:     "bassword",
		Host:         "127.0.0.1",
		Port:         3306,
		Dbname:       "db_rappa",
		Timeout:      "5s",
		MaxIdleConns: 10,
		MaxOpenConns: 100,
		MaxLifetime:  "1h",
	},
	ABMStockDataDir:               "/root/rappa/stockdata",
	ABMStockParamDir:              "/root/rappa/stockdata/params",
	ABMV2MaxConcurrency:           4,
	ABMStockDataSource:            "auto",
	ABMRemoteDBHost:               "192.168.198.76",
	ABMRemoteDBPort:               8904,
	ABMRemoteDBUser:               "maoshuoyu",
	ABMRemoteDBPassword:           "Swhy1234!@#$",
	ABMRemoteDBName:               "dfs://ods_tsdb_d_hash20_csmar",
	ABMRemoteTableName:            "l1_trdmin1_sh",
	ABMRemoteCheckPython:          "python3",
	ABMRemoteCheckScript:          "tools/abm_remote_data_check.py",
	ABMRemoteDefaultEndOffsetDays: 1,
	ABMUniverseSnapshotRoot:       "resources/abm_universe",
}

//var (
//	// GlobalConfig 全局配置实例
//	GlobalConfig *BHLayer2NodeConfig
//	once         sync.Once
//)

// LoadBHLayer2NodeConfig 从指定路径加载配置文件，覆盖默认值
// 如果文件不存在或加载失败，则使用默认配置
func LoadBHLayer2NodeConfig(path string) *BHLayer2NodeConfig {
	//once.Do(func() {
	config := DefaultBHLayer2NodeConfig
	InitGlobalLogWriter(config.LogPath, config.DEBUG)
	loadPKI(&config)
	// 尝试从配置文件加载
	if path != "" {
		file, err := os.Open(path)
		if err == nil {
			defer file.Close()
			decoder := json.NewDecoder(file)
			err = decoder.Decode(&config)
			if err != nil {
				// 配置文件解析失败时保留默认值
				println("Failed to parse config file, using default values:", err.Error())
			}
			//加载公私钥

		} else {
			// 文件打开失败时保留默认值
			println("Failed to open config file, using default values:", err.Error())
		}
	}

	// 加载默认 ABM 参数
	abmPath := "Config/abm_parameters.json"
	abmFile, err := os.Open(abmPath)
	if err == nil {
		defer abmFile.Close()
		abmDecoder := json.NewDecoder(abmFile)
		err = abmDecoder.Decode(&config.AbmParameters)
		if err != nil {
			println("Failed to parse ABM config file:", err.Error())
		}
	} else {
		println("Failed to open ABM config file, /api/simulation/abm-parameters may return empty:", err.Error())
	}

	// 设置全局配置
	return &config
}

func loadPKI(config *BHLayer2NodeConfig) {
	// 加载主机公私钥
	hostSKPath := fmt.Sprintf("%s/host_sk.key", config.CertPath)
	// 加载文件，公私钥bas64解码
	hostSKBytes, err := os.ReadFile(hostSKPath)
	if err != nil {
		println("Failed to read host secret key file, using default values:", err.Error())
	}
	// 解码公私钥
	hostSK, err := base64.StdEncoding.DecodeString(string(hostSKBytes))
	if err != nil {
		println("Failed to decode host secret key, using default values:", err.Error())
	}
	_, err = config.HostPrivateKey.SetBytes(hostSK)
	if err != nil {
		return
	}
}
