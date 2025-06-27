package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"vending-machine-layout-generator/s3utils"
)

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	LayoutID  int64  `json:"layoutId"`
	Message   string `json:"message"`
}

type Logger struct {
	s3Client  *s3.Client
	bucket    string
	logs      []LogEntry
	mutex     sync.Mutex
	startTime time.Time
}

func NewLogger(s3Client *s3.Client, bucket string) *Logger {
	return &Logger{
		s3Client:  s3Client,
		bucket:    bucket,
		logs:      []LogEntry{},
		startTime: time.Now(),
	}
}

func (l *Logger) Info(ctx context.Context, layoutID int64, message string) {
	l.addLog(ctx, "INFO", layoutID, message)
}

func (l *Logger) Error(ctx context.Context, layoutID int64, message string) {
	l.addLog(ctx, "ERROR", layoutID, message)
}

func (l *Logger) addLog(ctx context.Context, level string, layoutID int64, message string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	logEntry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		LayoutID:  layoutID,
		Message:   message,
	}
	l.logs = append(l.logs, logEntry)

	// Upload logs to S3
	logData, err := json.MarshalIndent(l.logs, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal logs: %v\n", err)
		return
	}

	// Use the same structure as the image path
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	date := now.Format("02")
	
	// Create log key in the same format as the processed image
	// Format: logs/{year}/{month}/{date}/{layoutId}_log.json
	logKey := fmt.Sprintf("logs/%s/%s/%s/%d_log.json", year, month, date, layoutID)
	
	err = s3utils.UploadLog(ctx, l.s3Client, l.bucket, logKey, logData)
	if err != nil {
		fmt.Printf("Failed to upload logs: %v\n", err)
	}
}