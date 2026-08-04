package test

import (
	"encoding/json"
	"fmt"
	"github.com/go-gota/gota/dataframe"
	"github.com/klauspost/reedsolomon"
	"os"
	"path/filepath"
	"testing"
)

// 这里测试纠删码
func ReadChunkFromFile(t *testing.T, filename string) []byte {
	t.Helper()
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read chunk %s: %v", filename, err)
	}
	return data
}

func TestECRecover(t *testing.T) {
	// 读取所有的文件块
	chunks := make([][]byte, 0)
	dec, err := reedsolomon.New(6, 3)
	for index := 0; index <= 8; index++ {
		filename := fmt.Sprintf("/root/zkml_test/BHLayer2Node/test/FakeSign-1736836377/0/FakeSign-1736836377_0_0-row-0-%d-chunk.slot", index)
		if _, err := os.Stat(filename); err != nil {
			t.Skipf("external EC fixture missing: %s", filename)
		}
		chunk := ReadChunkFromFile(t, filename)
		chunks = append(chunks, chunk)
	}
	fmt.Println(len(chunks))
	//for _, chunk := range chunks {
	//	fmt.Println(chunk)
	//}
	err = dec.Reconstruct(chunks)
	if err != nil {
		fmt.Printf("Failed to reconstruct data: %v\n", err)
		return
	}
	toBytes := func(chunks [][]byte) []byte {
		var result []byte
		for i := 0; i < 6; i++ {
			chunk := chunks[i]
			result = append(result, chunk...)
		}
		return result
	}
	jsonStr := toBytes(chunks)
	err = os.WriteFile(filepath.Join(t.TempDir(), "FakeSign-1736758379_0_0-row-0.chunk"), jsonStr, 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}
	jsonStr = jsonStr[:len(jsonStr)-3]
	// 解析原始 JSON 字符串到一个 map
	var rawData map[string]map[string]interface{}
	err = json.Unmarshal(jsonStr, &rawData)
	if err != nil {
		t.Fatalf("Error parsing JSON: %v", err)
	}
	numRows := 1
	for key, _ := range rawData {
		numRows = len(rawData[key])
		// 计算数据的行数，假设所有列的长度相同, python应该能保证
		break
	}

	// 创建一个空的切片来存储每行数据
	var result []map[string]interface{}

	// 遍历每一行数据，构造每一行的 map
	for i := 0; i < numRows; i++ {
		// 创建一个新的 map 来存储当前行的数据
		row := make(map[string]interface{})

		// 填充每一列的数据到当前行
		for key, values := range rawData {
			// 将列的索引（i）作为字符串来获取对应的数据
			row[key] = values[fmt.Sprintf("%d", i)]
		}

		// 将当前行数据添加到结果中
		result = append(result, row)
	}

	// 输出结果
	fmt.Println(result)
	// 转换为 DataFrame
	df := dataframe.LoadMaps(result)

	// 输出 DataFrame
	fmt.Println(df)
}
