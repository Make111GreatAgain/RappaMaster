package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

func validateUploadResponse(responseText string) error {
	if responseText == "" {
		return nil
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(responseText), &response); err != nil {
		return nil
	}

	codeValue, exists := response["code"]
	if !exists {
		return nil
	}

	var code float64
	switch v := codeValue.(type) {
	case float64:
		code = v
	case int:
		code = float64(v)
	default:
		return nil
	}
	if code == 0 || code == http.StatusOK {
		return nil
	}

	message, _ := response["message"].(string)
	return fmt.Errorf("upload failed: platform code=%v message=%s body=%s", codeValue, message, responseText)
}

func UploadFile(uploadURL string, params map[string]string, fileBytes []byte, fileName, fieldName string) (string, error) {
	// 1. 构造完整的 URL（添加查询参数）
	u, err := url.Parse(uploadURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	u.RawQuery = q.Encode()

	// 2. 创建 multipart form 数据
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 3. 添加文件部分
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(part, bytes.NewReader(fileBytes))
	if err != nil {
		return "", err
	}

	// 4. 关闭 multipart writer，确保写入尾部边界
	err = writer.Close()
	if err != nil {
		return "", err
	}

	// 5. 创建请求
	req, err := http.NewRequest("POST", u.String(), body)
	if err != nil {
		return "", err
	}

	// 6. 设置请求头 Content-Type
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 7. 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	responseText := strings.TrimSpace(string(responseBody))
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return responseText, fmt.Errorf("upload failed: status=%s body=%s", resp.Status, responseText)
	}
	if err := validateUploadResponse(responseText); err != nil {
		return responseText, err
	}

	return responseText, nil
}
