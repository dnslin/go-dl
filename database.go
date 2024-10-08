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

	// 检查总记录数
	var totalCount int
	err = db.QueryRow("SELECT COUNT(*) FROM data").Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("查询总记录数失败: %w", err)
	}
	log.Printf("数据库中总共有 %d 条记录", totalCount)

	// 检查 status = 0 的记录数
	var pendingCount int
	err = db.QueryRow("SELECT COUNT(*) FROM data WHERE status = 0").Scan(&pendingCount)
	if err != nil {
		return nil, fmt.Errorf("查询待处理记录数失败: %w", err)
	}
	log.Printf("数据库中有 %d 条待处理记录 (status = 0)", pendingCount)

	// 如果有待处理记录，查看一条示例
	if pendingCount > 0 {
		var id string
		var path, purity string
		err = db.QueryRow("SELECT id, path, purity FROM data WHERE status = 0 LIMIT 1").Scan(&id, &path, &purity)
		if err != nil {
			return nil, fmt.Errorf("查询示例记录失败: %w", err)
		}
		log.Printf("示例待处理记录: ID=%d, Path=%s, Purity=%s", id, path, purity)
	}

	return db, nil
}

func getDownloadTasks(db *sql.DB, offset, limit int) ([]DownloadTask, error) {
	log.Printf("正在获取下载任务: offset=%d, limit=%d", offset, limit)
	query := `
        SELECT id, path, purity
        FROM data
        WHERE status = 0
        LIMIT ? OFFSET ?
    `
	log.Printf("执行查询: %s", query)
	rows, err := db.Query(query, limit, offset)
	if err != nil {
		log.Printf("查询失败: %v", err)
		return nil, fmt.Errorf("查询下载任务失败: %w", err)
	}
	defer rows.Close()

	var tasks []DownloadTask
	for rows.Next() {
		var task DownloadTask
		err := rows.Scan(&task.ID, &task.URL, &task.Purity)
		if err != nil {
			log.Printf("扫描行失败: %v", err)
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
			log.Printf("更新状态失败 (ID: %s): %v", result.ID, err)
		} else {
			log.Printf("更新状态成功 (ID: %s, 状态: %d)", result.ID, result.Status)
			updatedCount++
		}
	}

	log.Printf("状态更新协程结束，共更新 %d 条记录", updatedCount)
	return nil
}
