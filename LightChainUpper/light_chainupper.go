package LightChainUpper

import (
	"BHLayer2Node/paradigm"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"
)

type LightChainUpper struct {
	channel *paradigm.RappaChannel
	nextID  uint64
}

func NewLightChainUpper(channel *paradigm.RappaChannel) (*LightChainUpper, error) {
	return &LightChainUpper{channel: channel}, nil
}

func (c *LightChainUpper) Start() {
	paradigm.Log("INFO", "LightChainUpper started: blockchain persistence is disabled")
	for tx := range c.channel.PendingTransactions {
		c.publishLocalTransaction(tx)
	}
}

func (c *LightChainUpper) publishLocalTransaction(tx paradigm.Transaction) {
	id := atomic.AddUint64(&c.nextID, 1)
	now := time.Now()
	txHash := c.localHash(id, tx, now)
	receipt := &paradigm.Receipt{
		TransactionHash: txHash,
		BlockNumber:     int64(id),
		To:              "light-chainupper",
		ReceiptProof:    []string{},
	}
	ptx := paradigm.NewPackedTransaction(tx, receipt, fmt.Sprintf("light-block-%d", id))
	ptx.SetID(int(id))
	ptx.SetUpchainTime(now)
	c.channel.DevTransactionChannel <- []*paradigm.PackedTransaction{ptx}
}

func (c *LightChainUpper) localHash(id uint64, tx paradigm.Transaction, ts time.Time) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%d:%s:%v:%d", id, tx.Call(), tx.CallData(), ts.UnixNano())))
	return "0x" + hex.EncodeToString(sum[:])
}
