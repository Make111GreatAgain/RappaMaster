package test

import (
	"BHLayer2Node/paradigm"
	"bytes"
	"fmt"
	"github.com/goccy/go-json"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestHttpClient(t *testing.T) {
	if os.Getenv("RAPPA_RUN_INTERACTIVE_HTTP_CLIENT_TEST") != "1" {
		t.Skip("set RAPPA_RUN_INTERACTIVE_HTTP_CLIENT_TEST=1 to run the interactive HTTP client")
	}
	//fmt.Println("Welcome to the HTTP Client Shell!")
	//fmt.Println("Type 'create' to create a new task or 'exit' to quit.")

	for {
		// 读取用户输入
		var command string
		fmt.Print("> ")
		_, err := fmt.Scanln(&command)
		if err != nil {
			if err.Error() == "unexpected newline" {
				// 用户输入了空行，继续等待命令
				continue
			}
			log.Fatalf("Error reading input: %v", err)
		}

		// 转换为小写，处理大小写不敏感
		command = strings.ToLower(command)

		// 根据命令执行不同的功能
		switch command {
		case "create":
			createTask()
		case "exit":
			fmt.Println("Exiting the client...")
			os.Exit(0)
		default:
			fmt.Println("Unknown command:", command)
			fmt.Println("Valid commands: create, exit")
		}
	}
}

// createTask 发送 POST 请求到 /create
func createTask() {
	// 请求体数据
	requestData := paradigm.Task{
		Model: paradigm.ABM,
		Params: map[string]interface{}{
			"param1": "value1",
			"param2": "value2",
		},
		Size: 10,
	}

	// 将请求数据编码为 JSON
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		log.Fatalf("Error marshalling request body: %v", err)
	}

	// 发送 POST 请求
	url := "http://127.0.0.1:8080/create" // 修改为你的实际服务器地址和端口
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalf("Error sending POST request: %v", err)
	}
	defer resp.Body.Close()

	// 打印响应状态码
	fmt.Printf("Response Status: %s\n", resp.Status)

	// 解析响应
	var response paradigm.HttpResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Fatalf("Error decoding response body: %v", err)
	}

	// 打印响应内容
	fmt.Printf("Response Message: %s\n", response.Message)
	fmt.Printf("Response Code: %s\n", response.Code)
}
