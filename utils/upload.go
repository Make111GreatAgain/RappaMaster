package utils

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

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

	return responseText, nil
}
