package paradigm

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type BlockInfo struct {
	//BlockHash:      block.Hash,
	//"parentHash":     parentHash,
	//"blockHeight":    block.Number,
	//"merkleRoot":     block.StateRoot,
	//"nbTransactions": len(block.Transactions),
	//"transactions":   txDetails,
	BlockHash       string
	ParentHash      string
	BlockHeight     int32
	TransactionRoot string
	Txs             []TransactionInfo
	Invalid         bool
	ErrorMessage    string
}

func (b *BlockInfo) IsValid() bool {
	return b.Invalid
}
func (b *BlockInfo) Error() string {
	return b.ErrorMessage
}
func NewInvalidBlockInfo(e string) BlockInfo {
	return BlockInfo{
		BlockHash:       "",
		ParentHash:      "",
		BlockHeight:     -1,
		TransactionRoot: "",
		Txs:             make([]TransactionInfo, 0),
		Invalid:         false,
		ErrorMessage:    e,
	}
}

func NewMockerBlockInfo() BlockInfo {
	return BlockInfo{
		BlockHash:       "0x123456",
		ParentHash:      "0x123455",
		BlockHeight:     1,
		TransactionRoot: "0x111111",
		Txs:             []TransactionInfo{NewMockerTransactionInfo(), NewMockerTransactionInfo(), NewMockerTransactionInfo()},
		Invalid:         true,
		ErrorMessage:    "",
	}
}

type TransactionInfo struct {
	TxHash       string
	Contract     string
	Abi          string
	BlockHash    string
	UpchainTime  time.Time
	Invalid      bool
	ErrorMessage string
	// TODO
}

type Receipt struct {
	TransactionHash string   `json:"transactionHash"`
	BlockNumber     int64    `json:"blockNumber"`
	To              string   `json:"to"`
	ReceiptProof    []string `json:"receiptProof"`
}

type Block struct {
	Hash         string `json:"hash"`
	Timestamp    uint64 `json:"timestamp"`
	ReceiptsRoot string `json:"receiptsRoot"`
}

func NewMockerTransactionInfo() TransactionInfo {
	return TransactionInfo{
		TxHash:       "",
		Contract:     "",
		Abi:          "STORE",
		BlockHash:    "",
		UpchainTime:  time.Now(),
		Invalid:      true,
		ErrorMessage: "",
	}
}

func NewInvalidTransactionInfo(e string) TransactionInfo {
	return TransactionInfo{
		TxHash:       "",
		Contract:     "",
		Abi:          "",
		BlockHash:    "",
		UpchainTime:  time.Now(),
		Invalid:      false,
		ErrorMessage: e,
	}
}

//	func normalizeHash(hash string) string {
//		if strings.HasPrefix(hash, "0x") {
//			return hash[2:] // 去掉 0x
//		}
//		return hash
//	}
//
// 计算 SHA256 哈希
func sha256Hash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// 计算 Merkle Root
func CalculateMerkleRoot(receipt *Receipt) string {
	if len(receipt.ReceiptProof) == 0 {
		return "" // 没有回执时返回空 Merkle Root
	}
	// var layer []string
	// for _, hash := range receipt.ReceiptProof {
	// 	// layer = append(layer, sha256Hash(hash)) // 叶子节点
	// 	layer = append(layer, hash)
	// }
	layer := receipt.ReceiptProof
	// 2. 构建 Merkle 树
	for len(layer) > 1 {
		var newLayer []string
		// 如果是奇数个哈希，复制最后一个
		if len(layer)%2 == 1 {
			layer = append(layer, layer[len(layer)-1])
		}
		// 计算父节点
		for i := 0; i < len(layer); i += 2 {
			parent := sha256Hash(layer[i] + layer[i+1])
			newLayer = append(newLayer, parent)
		}
		layer = newLayer // 进入下一层
	}

	// 3. 返回最终的 Merkle Root
	return layer[0] // 只有1个回执hash 则直接返回
}

// 验证 Merkle Root
func VerifyMerkleRoot(receipt *Receipt, block *Block) bool {
	// LogWriter.Log("DBEUG", fmt.Sprintf("verify receipt:%s", receipt))
	// LogWriter.Log("DBEUG", fmt.Sprintf("verify block:%+v", block))
	calculatedRoot := CalculateMerkleRoot(receipt)
	// LogWriter.Log("DBEUG", fmt.Sprintf("verify Result %s block %s", calculatedRoot, block.ReceiptsRoot))
	return calculatedRoot == block.ReceiptsRoot[2:]
}
