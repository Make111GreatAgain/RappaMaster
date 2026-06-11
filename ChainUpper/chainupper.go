//go:build ignore

package ChainUpper

import (
	"BHLayer2Node/ChainUpper/service"
	"BHLayer2Node/paradigm"
	"fmt"
	"sync"
	"time"

	Store "BHLayer2Node/ChainUpper/contract/storeData"
	"context"
	"encoding/hex"
	"log"

	"github.com/FISCO-BCOS/go-sdk/v3/client"
)

type ChainUpper struct {
	channel *paradigm.RappaChannel
	//pendingTransactions chan paradigm.Transaction // 交易channel
	transactionPool  []paradigm.Transaction // 所有的交易
	unprocessedIndex int                    // 未处理的交易index，包括这一index
	mu               sync.Mutex
	//queue               chan map[string]interface{} // 用于异步上链的队列
	queue    chan paradigm.Transaction // 用于异步上链的队列 modified by zhmye
	client   *client.Client            // FISCO-BCOS 客户端
	instance *Store.StoreData
	count    int // add by zhmye, 这里是用来给每笔交易赋予一个id的
}

func (c *ChainUpper) Start() {
	go c.handleQuery()
	timeStart := time.Now()
	go func() {
		for {
			if time.Since(timeStart) >= 10*time.Second {
				timeStart = time.Now()
				c.UpChain()
			}
		}
	}()
	for {
		select {
		case transaction := <-c.channel.PendingTransactions:
			// 先简单写一下
			c.mu.Lock()
			c.transactionPool = append(c.transactionPool, transaction)
			c.mu.Unlock()
		default:
			continue
		}
	}
}
func (c *ChainUpper) UpChain() {
	// 这里简单写一下，应该是用异步上链组件接入这里
	pack := func() []paradigm.Transaction {
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.unprocessedIndex == len(c.transactionPool) {
			return []paradigm.Transaction{}
		}
		packedTransactions := c.transactionPool[c.unprocessedIndex:]
		c.unprocessedIndex = len(c.transactionPool)
		return packedTransactions
	}
	packedTransactions := pack()
	if len(packedTransactions) > 0 {
		// 将交易打包为链上合约的参数
		for _, tx := range packedTransactions {
			// modify by zhmye
			check := func(tx paradigm.Transaction) error {
				calldata := tx.CallData()
				switch tx.(type) {
				// 这里可以写在Transaction的interface里，加一个Check()，然后下面统一tx.Check()
				case *paradigm.InitTaskTransaction:
					// todo
					return nil
				case *paradigm.TaskProcessTransaction:
					if calldata["Process"].(int32) < 0 {
						return fmt.Errorf("TaskProcessTransaction Process <0")
					}
					// todo
					return nil
				case *paradigm.EpochRecordTransaction:
					// todo
					return nil
				default:
					return nil
				}
			}
			if err := check(tx); err != nil {
				//panic(err)
				paradigm.Log("ERROR", err.Error())
				continue
			} else {
				c.queue <- tx
			}
			//result := tx.CallData()
			//if result["Process"] == -1 {
			//	LogWriter.Log("ERROR", fmt.Sprintf("Transaction %d BUG: %v", id, result))
			//	continue
			//} else {
			//	c.queue <- result // 推送到异步队列
			//	LogWriter.Log("CHAINUP", fmt.Sprintf("Transaction %d pushed to queue: %v", id, result))
			//}
		}
		// LogWriter.Log("CHAINUP", fmt.Sprintf("%d Transactions pushed to queue for async processing", len(packedTransactions)))
		paradigm.Log("CHAINUP", fmt.Sprintf("up %d transactions to blockchain...", len(packedTransactions)))

	} else {
		//paradigm.Log("CHAINUP", "Nothing to up to Blockchain..., len(transactionPool) = 0")
	}
}

func NewChainUpper(channel *paradigm.RappaChannel, config *paradigm.BHLayer2NodeConfig) (*ChainUpper, error) {
	// 初始化 FISCO-BCOS 客户端
	privateKey, _ := hex.DecodeString(config.PrivateKey)
	client, err := client.DialContext(context.Background(), &client.Config{
		IsSMCrypto:  false,
		GroupID:     config.GroupID,
		PrivateKey:  privateKey,
		Host:        config.FiscoBcosHost,
		Port:        config.FiscoBcosPort,
		TLSCaFile:   config.TLSCaFile,
		TLSCertFile: config.TLSCertFile,
		TLSKeyFile:  config.TLSKeyFile,
	})
	if err != nil {
		e := paradigm.Error(paradigm.RuntimeError, fmt.Sprintf("Failed to initialize FISCO-BCOS client: %v", err))
		//paradigm.Log("ERROR", fmt.Sprintf("failed to initialize FISCO-BCOS client: %v", err))
		return nil, fmt.Errorf(e.Error())
	}

	// 部署或加载合约
	// instance, err := Store.NewStore(common.HexToAddress(config.ContractAddress), client)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to load contract: %v", err)
	// }
	address, receipt, instance, err := Store.DeployStoreData(client.GetTransactOpts(), client)
	if err != nil {
		log.Fatal(err)
	}
	paradigm.Print("INFO", fmt.Sprintf("Deploy Contract on Blockchain Success, contract adddress: %s, transaction hash: %s", address.Hex(), receipt.TransactionHash))
	//paradigm.Print("INFO", fmt.Sprintf("contract address: %s", address.Hex())) // the address should be saved, will use in next example
	//paradigm.Log("INFO", fmt.Sprintf("transaction hash: %s", receipt.TransactionHash))

	// 初始化队列和 Worker
	queue := make(chan paradigm.Transaction, config.QueueBufferSize)
	for i := 0; i < config.WorkerCount; i++ {
		worker := service.NewUpchainWorker(i, queue, channel.DevTransactionChannel, instance, client)
		go worker.Process()
		//go service. (i, queue, instance, client)
	}
	paradigm.Log("INFO", "Chainupper initialized successfully, workers waiting for transactions...")

	return &ChainUpper{
		channel: channel,
		//pendingTransactions: channel.PendingTransactions,
		transactionPool:  make([]paradigm.Transaction, 0),
		unprocessedIndex: 0,
		mu:               sync.Mutex{},
		queue:            queue,
		client:           client,
		instance:         instance,
	}, nil
}
