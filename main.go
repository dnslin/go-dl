package main

import (
	"log"
	"runtime"

	"golang.org/x/sync/errgroup"
)

const (
	dbPath       = "mv.db"
	batchSize    = 100
	maxWorkers   = 10
	maxRetries   = 3
	proxyAddress = "http://127.0.0.1:17890"
)

func main() {
	initLogger()
	log.Println("开始下载任务")

	db, err := openDatabase(dbPath)
	if err != nil {
		log.Fatalf("无法打开数据库: %v", err)
	}
	defer db.Close()

	taskChan := make(chan DownloadTask, batchSize)
	resultChan := make(chan DownloadResult, batchSize)

	var eg errgroup.Group

	for i := 0; i < maxWorkers; i++ {
		workerID := i
		eg.Go(func() error {
			return worker(workerID, taskChan, resultChan)
		})
	}

	eg.Go(func() error {
		return updateStatus(db, resultChan)
	})

	offset := 0
	for {
		tasks, err := getDownloadTasks(db, offset, batchSize)
		if err != nil {
			log.Printf("获取下载任务失败: %v", err)
			break
		}
		if len(tasks) == 0 {
			log.Println("没有更多任务，结束获取")
			break
		}

		log.Printf("获取到 %d 个任务，开始分发", len(tasks))
		for _, task := range tasks {
			taskChan <- task
		}

		offset += len(tasks)
	}

	close(taskChan)
	log.Println("所有任务已分发，等待工作协程完成")

	if err := eg.Wait(); err != nil {
		log.Printf("工作协程出错: %v", err)
	}

	close(resultChan)

	log.Println("所有下载任务已完成")
	printStats()
}

func printStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Printf("内存使用统计：使用 = %v MB, 系统 = %v MB, GC次数 = %v",
		m.Alloc/1024/1024, m.Sys/1024/1024, m.NumGC)
}
