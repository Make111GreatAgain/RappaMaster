//go:build ignore

package service

import (
	"BHLayer2Node/paradigm"
	"BHLayer2Node/utils"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"strings"
)

// --- JSON 结构体 ---

type InitTaskJSON struct {
	Sign       string                 `json:"sign"`
	Name       string                 `json:"name"`
	Size       uint32                 `json:"size"`
	Model      string                 `json:"model"`
	IsReliable bool                   `json:"is_reliable"`
	Params     map[string]interface{} `json:"params"`
}

type TaskProcessJSON struct {
	Sign       string   `json:"sign"`
	Hash       string   `json:"hash"`
	Slot       uint32   `json:"slot"`
	Process    uint32   `json:"process"`
	Id         uint32   `json:"id"`
	Epoch      uint32   `json:"epoch"`
	Commitment string   `json:"commitment"` // hex
	Proof      string   `json:"proof"`      // hex/base64 as you like
	Signatures []string `json:"signatures"` // hex
}

type EpochRecordJSON struct {
	Id            uint64           `json:"id"`
	JustifiedHash []string         `json:"justified_hash"`
	CommitsHash   []string         `json:"commits_hash"`
	Invalids      map[string]uint8 `json:"invalids"` // hash -> reason
}

// GetInitTaskJSON 查询 InitTask 并返回 JSON
func (q *UpChainWorker) GetInitTaskJSON(sign string) (string, error) {
	// 准备 sign
	var sign32 [32]byte
	copy(sign32[:], []byte(sign))

	// 调用
	res, err := q.instance.GetInitTask(q.client.GetCallOpts(), sign32)
	if err != nil {
		return "", err
	}

	// 解析 JSON params
	var params map[string]interface{}
	if err := json.Unmarshal(res.Params, &params); err != nil {
		params = map[string]interface{}{}
	}

	// model: trim zero bytes
	model := paradigm.ModelTypeToString(paradigm.SupportModelType(res.Model[31]))

	out := InitTaskJSON{
		Sign:       sign,
		Name:       res.Name,
		Size:       res.Size,
		Model:      model,
		IsReliable: res.IsReliable,
		Params:     params,
	}
	b, err := json.MarshalIndent(out, "", "  ")
	return string(b), err
}

// GetTaskProcessJSON 查询 TaskProcess 并返回 JSON
func (q *UpChainWorker) GetTaskProcessJSON(hashHex string) (string, error) {
	// 准备 hash
	var hash32 [32]byte
	clean := strings.TrimPrefix(hashHex, "0x")
	if bs, err := hex.DecodeString(clean); err == nil {
		copy(hash32[:], bs)
	} else {
		copy(hash32[:], []byte(hashHex))
	}

	// 调用
	res, err := q.instance.GetTaskProcess(q.client.GetCallOpts(), hash32)
	if err != nil {
		return "", err
	}

	out := TaskProcessJSON{
		Sign:       utils.TrimZero(res.Sign[:]),
		Hash:       hashHex,
		Slot:       res.Slot,
		Process:    res.Process,
		Id:         res.Id,
		Epoch:      res.Epoch,
		Commitment: "0x" + hex.EncodeToString(res.Commitment[:]),
		// Proof 可以是任意格式，Hex 保底
		Proof:      "0x" + hex.EncodeToString(res.Proof),
		Signatures: utils.HexList(res.Signatures),
	}

	b, err := json.MarshalIndent(out, "", "  ")
	return string(b), err
}

// GetEpochRecordJSON 查询 EpochRecord 并返回 JSON
func (q *UpChainWorker) GetEpochRecordJSON(id uint64) (string, error) {
	// 准备 id
	idBig := big.NewInt(int64(id))

	// 调用
	res, err := q.instance.GetEpochRecordFull(q.client.GetCallOpts(), idBig)
	if err != nil {
		return "", err
	}

	// 解析 invalids
	invalids := make(map[string]uint8, len(res.Invalids))
	for _, e := range res.Invalids {
		// 把 [32]byte 作为原始文本，去掉尾部的 '\x00'
		slotHash := strings.TrimRight(string(e.Hash[:]), "\x00")
		invalids[slotHash] = e.Reason
	}

	out := EpochRecordJSON{
		Id:            id,
		JustifiedHash: utils.Bytes32ListToStrings(res.Justified),
		CommitsHash:   utils.Bytes32ListToStrings(res.Commits),
		Invalids:      invalids,
	}
	b, err := json.MarshalIndent(out, "", "  ")
	return string(b), err
}
