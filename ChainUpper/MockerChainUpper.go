//go:build ignore

package ChainUpper

import (
	"BHLayer2Node/ChainUpper/service"
	"BHLayer2Node/Query"
	"BHLayer2Node/paradigm"
	"fmt"
	"sync"
	"time"
)

type MockerChainUpper struct {
	channel *paradigm.RappaChannel
	//pendingTransactions chan paradigm.Transaction // 交易channel
	transactionPool  []paradigm.Transaction // 所有的交易
	unprocessedIndex int                    // 未处理的交易index，包括这一index
	mu               sync.Mutex
	//queue               chan map[string]interface{} // 用于异步上链的队列
	queue chan paradigm.Transaction // 用于异步上链的队列 modified by zhmye
	//client   *client.Client            // FISCO-BCOS 客户端
	//instance *SlotCommit.SlotCommit    // 合约实例
	count int // add by zhmye, 这里是用来给每笔交易赋予一个id的
}

func (c *MockerChainUpper) Start() {
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
func (c *MockerChainUpper) UpChain() {
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
		paradigm.Log("DEBUG", fmt.Sprintf("Pending %d transactions, prepare to up chain...", len(packedTransactions)))
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
					//fmt.Println(444, tx.Blob().(*paradigm.EpochRecord))
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
		paradigm.Log("CHAINUP", "Nothing to up to Blockchain..., len(transactionPool) = 0")
	}
}

func (c *MockerChainUpper) handleQuery() {
	for query := range c.channel.BlockchainQueryChannel {
		c.handle(query)
	}
}

func (c *MockerChainUpper) handle(query paradigm.Query) {
	//ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()
	switch query.(type) {
	// 交易和最新信息的query都在oracle处理了,TODO @XQ 判断一下在oracle里的信息是否完整/正确
	case *Query.BlockchainBlockHashQuery:
		// 通过client获取到block
		item := query.(*Query.BlockchainBlockHashQuery)
		//blockHash := item.BlockHash
		//paradigm.Log("DEBUG", fmt.Sprintf("receive BlockchainBlockHashQuery, blockHash: %s", blockHash))
		//blockInfo := c.getBlockInfo(*block)
		item.SendInfo(paradigm.NewMockerBlockInfo())

	case *Query.BlockchainBlockNumberQuery:
		// 通过client获取到block
		item := query.(*Query.BlockchainBlockNumberQuery)
		//blockNumber := item.BlockNumber
		//paradigm.Log("DEBUG", fmt.Sprintf("receive BlockchainBlockNumberQuery, blockNumber: %d", blockNumber))
		//blockInfo := c.getBlockInfo(*block)
		item.SendInfo(paradigm.NewMockerBlockInfo())
	default:
		paradigm.Error(paradigm.RuntimeError, "Unsupported Query Type In ChainUpper")
	}
}

func NewMockerChainUpper(channel *paradigm.RappaChannel) (*MockerChainUpper, error) {
	queue := make(chan paradigm.Transaction, 10000)
	for i := 0; i < 1; i++ {
		worker := service.NewMockerUpChainWorker(i, queue, channel.DevTransactionChannel)
		go worker.Process()
		//go service. (i, queue, instance, client)
	}
	paradigm.Print("INFO", "Chainupper initialized successfully, workers waiting for transactions...")
	return &MockerChainUpper{
		channel: channel,
		//pendingTransactions: channel.PendingTransactions,
		transactionPool:  make([]paradigm.Transaction, 0),
		unprocessedIndex: 0,
		mu:               sync.Mutex{},
		queue:            queue,
		//client:              client,
		//instance:            instance,
	}, nil
}
