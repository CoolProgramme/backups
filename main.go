package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 定义一个通道用于异步日志记录
var logChan = make(chan string, 100)

func main() {

	// 启动日志记录协程
	go asyncLogger()

	// 设置日志输出到文件
	logFilePath := filepath.Join("./logs", "backup.log")

	// 创建目录（如果不存在）
	if err := os.MkdirAll(filepath.Dir(logFilePath), 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	// 设置日志输出到文件
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer closeWithErrorHandling(logFile)
	log.SetOutput(logFile)
	logChan <- "备份程序启动"

	// 启动定时任务，每天凌晨1点执行
	for {
		now := time.Now()
		nextBackup := time.Date(now.Year(), now.Month(), now.Day(), 1, 0, 0, 0, now.Location())

		// 如果当前时间已经超过1点，则将备份时间设为第二天的凌晨1点
		if now.After(nextBackup) {
			nextBackup = nextBackup.Add(24 * time.Hour)
		}

		durationUntilNextBackup := nextBackup.Sub(now)
		logChan <- fmt.Sprintf("距离下次备份还有 %v", durationUntilNextBackup)

		// 等待到下一个执行时间
		time.Sleep(durationUntilNextBackup)

		// 启动备份操作协程
		go performBackup()
	}
}

// asyncLogger 日志记录的异步处理函数
func asyncLogger() {
	for logMsg := range logChan {
		log.Println(logMsg)
	}
}

// performBackup 执行实际的备份操作并记录日志
func performBackup() {
	logChan <- "开始备份操作"

	endpoint := "Your own Endpoint"
	accessKeyID := "Your own AccessKey ID"
	accessKeySecret := "Your own AccessKey Secret"

	client, err := oss.New(endpoint, accessKeyID, accessKeySecret)
	if err != nil {
		logChan <- fmt.Sprintf("创建OSS客户端失败: %v", err)
		return
	}

	bucket, err := client.Bucket("Your own Bucket Name")
	if err != nil {
		logChan <- fmt.Sprintf("获取OSS Bucket失败: %v", err)
		return
	}

	oss.BuildLifecycleRuleByDays("", "", false, 7)

	pr, pw := io.Pipe()

	// 启动压缩的协程
	go func() {
		defer closeWithErrorHandling(pw)
		err := createTarGz("/app/docker", pw)
		if err != nil {
			logChan <- fmt.Sprintf("创建压缩文件失败: %v", err)
			return
		}
	}()

	now := time.Now().Format("20060102_150405")

	fileName := fmt.Sprintf("backups/docker_%s.tar.gz", now)

	// 上传操作
	err = bucket.PutObject(fileName, pr)
	if err != nil {
		logChan <- fmt.Sprintf("上传文件失败: %v", err)
		return
	}

	logChan <- "备份成功"
}

// createTarGz 创建tar.gz文件并返回压缩文件的文件流
func createTarGz(sourceDir string, writer io.Writer) error {
	gw := gzip.NewWriter(writer)
	defer closeWithErrorHandling(gw)

	tw := tar.NewWriter(gw)
	defer closeWithErrorHandling(tw)

	return filepath.Walk(sourceDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(strings.TrimPrefix(file, sourceDir))

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer closeWithErrorHandling(f)

		_, err = io.Copy(tw, f)
		return err
	})
}

// closeWithErrorHandling 关闭资源并处理错误
func closeWithErrorHandling(c io.Closer) {
	err := c.Close()
	if err != nil {
		logChan <- fmt.Sprintf("关闭资源时出错: %v", err)
	}
}
