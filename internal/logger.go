package internal

import (
	"fmt"
	"io"
	"os"
	"time"
)

type Logger interface {
	Print(level string, content string) error
	Close() error
}

// yinglong专用日志
type SimpleLog struct {
	filepath string   //日志文件路径
	fp       *os.File //日志文件指针
}

// 定义选项类型：一个能修改 SimpleLog 指针的函数
type LoggerConfig struct {
	Filepath string
}

func NewSimpleLog(cfg LoggerConfig) (*SimpleLog, error) {
	// 设置默认值
	logObj := &SimpleLog{
		filepath: cfg.Filepath,
		fp:       os.Stderr,
	}
	if cfg.Filepath != "" {
		fp, err := os.OpenFile(cfg.Filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("打开日志文件失败：%w", err)
		}
		logObj.fp = fp
	}
	return logObj, nil
}

func (obj *SimpleLog) Print(level, content string) error {
	if _, err := io.WriteString(obj.fp, fmt.Sprintf("[%s %s] %s\n", time.Now().Format("2006-01-02 15:04:05"), level, content)); err != nil {
		return err
	}
	return nil
}

func (obj *SimpleLog) Close() error {
	if err := obj.fp.Close(); err != nil {
		return err
	}
	return nil
}
