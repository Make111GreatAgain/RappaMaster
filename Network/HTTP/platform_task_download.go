package HTTP

import (
	"BHLayer2Node/Network/HTTP/abm"
	"BHLayer2Node/paradigm"
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// HandlePlatformTaskDownload 处理平台任务打包下载
func (e *HttpEngine) HandlePlatformTaskDownload(c *gin.Context) {
	platformTaskID := c.Query("taskId")
	if platformTaskID == "" {
		c.JSON(http.StatusBadRequest, paradigm.HttpResponse{Message: "taskId required", Code: "E100015", Data: nil})
		return
	}

	platformTask, err := e.dbService.GetPlatformTaskByID(platformTaskID)
	if err != nil || platformTask == nil {
		c.JSON(http.StatusNotFound, paradigm.HttpResponse{Message: "平台任务未找到", Code: "E100016", Data: nil})
		return
	}

	// 创建临时工作目录
	tempDir, err := os.MkdirTemp("", "platform_task_"+platformTaskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, paradigm.HttpResponse{Message: "创建临时目录失败", Code: "E100017", Data: nil})
		return
	}
	defer os.RemoveAll(tempDir) // 请求结束清理临时目录

	// 遍历子任务并收集数据
	for _, subTask := range platformTask.SubTasks {
		// 这里只处理已完成的任务
		if subTask.Status != paradigm.Finished {
			continue
		}

		// 恢复 Collector
		if err := e.dbService.RecoverCollector(&subTask, e.pkiManager); err != nil {
			paradigm.Log("ERROR", fmt.Sprintf("Failed to recover collector for subtask %s: %v", subTask.Sign, err))
			continue
		}

		// 执行收集
		reader, err := subTask.Collector.ProcessCollect(paradigm.HttpCollectRequest{
			Sign: subTask.Sign,
			Size: subTask.Size,
		})
		if err != nil || reader == nil {
			paradigm.Log("ERROR", fmt.Sprintf("Failed to collect subtask %s: %v", subTask.Sign, err))
			continue
		}

		// 读取数据并写入文件
		stockCode := abm.ExtractTaskStockCode(&subTask)
		if stockCode == "" {
			stockCode = subTask.Sign
		}
		ext := paradigm.ModelOutputTypeToFileExt(subTask.Collector.OutputType())
		fileName := fmt.Sprintf("%s.%s", stockCode, ext)
		filePath := filepath.Join(tempDir, fileName)

		f, err := os.Create(filePath)
		if err != nil {
			paradigm.Log("ERROR", fmt.Sprintf("Failed to create file %s: %v", filePath, err))
			continue
		}
		_, err = io.Copy(f, reader)
		f.Close()
		if err != nil {
			paradigm.Log("ERROR", fmt.Sprintf("Failed to write data to %s: %v", filePath, err))
			continue
		}
	}

	// 检查是否有成功收集的文件
	files, err := os.ReadDir(tempDir)
	if err != nil || len(files) == 0 {
		c.JSON(http.StatusNotFound, paradigm.HttpResponse{Message: "没有找到可下载的子任务数据", Code: "E100018", Data: nil})
		return
	}

	// 创建压缩包并通过 io.Pipe 返回 reader
	zipFileName := platformTaskID + ".zip"
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		archive := zip.NewWriter(pw)
		for _, file := range files {
			f, err := os.Open(filepath.Join(tempDir, file.Name()))
			if err != nil {
				continue
			}

			w, err := archive.Create(file.Name())
			if err != nil {
				f.Close()
				continue
			}
			_, err = io.Copy(w, f)
			f.Close()
		}
		archive.Close()
	}()

	// 发送压缩包
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", zipFileName))
	c.DataFromReader(http.StatusOK, -1, "application/octet-stream", pr, nil)
}
