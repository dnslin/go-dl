package main

import (
	"log"
	"time"
)

func worker(id int, taskChan <-chan DownloadTask, resultChan chan<- DownloadResult) error {
	log.Printf("工作协程 %d 启动", id)
	for task := range taskChan {
		result := DownloadResult{ID: task.ID}

		for retry := 0; retry < maxRetries; retry++ {
			log.Printf("工作协程 %d 开始处理任务 (ID: %d, 重试: %d)", id, task.ID, retry)
			err := downloadFile(task)
			if err == nil {
				result.Status = 1 // 下载成功
				log.Printf("工作协程 %d 成功完成任务 (ID: %d)", id, task.ID)
				break
			}

			log.Printf("工作协程 %d 下载失败 (ID: %d, 重试: %d): %v", id, task.ID, retry+1, err)
			result.Err = err

			if retry == maxRetries-1 {
				result.Status = 2 // 下载失败
				log.Printf("工作协程 %d 任务失败 (ID: %d)，达到最大重试次数", id, task.ID)
			}

			retryDelay := time.Second * time.Duration(retry+1)
			log.Printf("工作协程 %d 等待 %v 后重试任务 (ID: %d)", id, retryDelay, task.ID)
			time.Sleep(retryDelay)
		}

		resultChan <- result
	}

	log.Printf("工作协程 %d 结束", id)
	return nil
}
