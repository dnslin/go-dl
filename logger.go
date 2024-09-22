package main

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

func initLogger() {
	logDir := "logs"
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		log.Fatalf("无法创建日志目录: %v", err)
	}

	logFile := filepath.Join(logDir, time.Now().Format("2006-01-02_15-04-05")+".log")
	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("无法创建日志文件: %v", err)
	}

	log.SetOutput(f)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	log.Println("日志系统初始化完成")
}
