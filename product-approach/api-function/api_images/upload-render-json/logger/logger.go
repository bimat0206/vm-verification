package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Logger struct {
	s3Client *s3.Client
	bucket   string
}

type LogEntry struct {
	Timestamp string      `json:"timestamp"`
	Level     string      `json:"level"`
	LayoutID  int64       `json:"layoutId,omitempty"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
}

func NewLogger(s3Client *s3.Client, bucket string) *Logger {
	return &Logger{
		s3Client: s3Client,
		bucket:   bucket,
	}
}

func (l *Logger) Info(ctx context.Context, layoutID int64, message string) {
	l.log(ctx, "INFO", layoutID, message, nil)
}

func (l *Logger) Error(ctx context.Context, layoutID int64, message string) {
	l.log(ctx, "ERROR", layoutID, message, nil)
}

func (l *Logger) InfoWithData(ctx context.Context, layoutID int64, message string, data interface{}) {
	l.log(ctx, "INFO", layoutID, message, data)
}

func (l *Logger) ErrorWithData(ctx context.Context, layoutID int64, message string, data interface{}) {
	l.log(ctx, "ERROR", layoutID, message, data)
}

func (l *Logger) log(ctx context.Context, level string, layoutID int64, message string, data interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		LayoutID:  layoutID,
		Message:   message,
		Data:      data,
	}

	// For now, just print to stdout (CloudWatch logs)
	// In production, you might want to also store logs in S3
	jsonData, _ := json.Marshal(entry)
	fmt.Printf("%s\n", string(jsonData))
}
