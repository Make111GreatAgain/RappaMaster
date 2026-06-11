package Query

import (
	"BHLayer2Node/paradigm"
)

func safeReceiptTxHash(receipt *paradigm.Receipt) string {
	if receipt == nil {
		return ""
	}
	return receipt.TransactionHash
}

func safeReceiptBlockHeight(receipt *paradigm.Receipt) interface{} {
	if receipt == nil {
		return nil
	}
	return receipt.BlockNumber
}

func safeReceiptContract(receipt *paradigm.Receipt) interface{} {
	if receipt == nil {
		return nil
	}
	return receipt.To
}

func safeReceiptMerkleRoot(receipt *paradigm.Receipt) string {
	if receipt == nil {
		return ""
	}
	return paradigm.CalculateMerkleRoot(receipt)
}

func safeReceiptProof(receipt *paradigm.Receipt) interface{} {
	if receipt == nil {
		return nil
	}
	return receipt.ReceiptProof
}
