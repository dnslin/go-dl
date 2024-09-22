package main

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

func openDatabase(dbPath string) (*sql.DB, error) {
	log.Printf("正在打开数据库: %s", dbPath)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}
	log.Println("数据库连接成功")
	return db, nil
}

func getDownloadTasks(db *sql.DB, offset, limit int) ([]DownloadTask, error) {
	log.Printf("正在获取下载任务: offset=%d, limit=%d", offset, limit)
	query := `
		SELECT id, path, purity
		FROM data
		WHERE status = null
		LIMIT ? OFFSET ?
	`
	rows, err := db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询下载任务失败: %w", err)
	}
	defer rows.Close()

	var tasks []DownloadTask
	for rows.Next() {
		var task DownloadTask
		err := rows.Scan(&task.ID, &task.URL, &task.Purity)
		if err != nil {
			return nil, fmt.Errorf("扫描下载任务失败: %w", err)
		}
		task.FilePath = filepath.Join(task.Purity, filepath.Base(task.URL))
		tasks = append(tasks, task)
	}

	log.Printf("获取到 %d 个下载任务", len(tasks))
	return tasks, nil
}

func updateStatus(db *sql.DB, resultChan <-chan DownloadResult) error {
	log.Println("启动状态更新协程")
	stmt, err := db.Prepare("UPDATE data SET status = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("准备更新语句失败: %w", err)
	}
	defer stmt.Close()

	var mu sync.Mutex
	var updatedCount int
	for result := range resultChan {
		mu.Lock()
		_, err := stmt.Exec(result.Status, result.ID)
		mu.Unlock()
		if err != nil {
			log.Printf("更新状态失败 (ID: %d): %v", result.ID, err)
		} else {
			log.Printf("更新状态成功 (ID: %d, 状态: %d)", result.ID, result.Status)
			updatedCount++
		}
	}

	log.Printf("状态更新协程结束，共更新 %d 条记录", updatedCount)
	return nil
}
