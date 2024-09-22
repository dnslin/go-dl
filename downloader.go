package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type DownloadTask struct {
	ID       string
	URL      string
	Purity   string
	FilePath string
}

type DownloadResult struct {
	ID     string
	Status int
	Err    error
}

func downloadFile(task DownloadTask) error {
	log.Printf("开始下载文件 (ID: %d, URL: %s)", task.ID, task.URL)

	proxyURL, err := url.Parse(proxyAddress)
	if err != nil {
		return fmt.Errorf("解析代理地址失败: %w", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	resp, err := client.Get(task.URL)
	if err != nil {
		return fmt.Errorf("HTTP GET 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP 状态码错误: %d", resp.StatusCode)
	}

	log.Printf("创建目录: %s", filepath.Dir(task.FilePath))
	err = os.MkdirAll(filepath.Dir(task.FilePath), 0755)
	if err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	log.Printf("创建文件: %s", task.FilePath)
	out, err := os.Create(task.FilePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer out.Close()

	startTime := time.Now()
	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}
	duration := time.Since(startTime)

	log.Printf("文件下载完成 (ID: %d, 大小: %d bytes, 耗时: %v)", task.ID, written, duration)
	return nil
}
