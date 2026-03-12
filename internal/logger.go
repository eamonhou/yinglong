package internal

import (
	"fmt"
	"io"
	"os"
	"time"
)

type Logger interface {
	Print(level string, content string)
	SetFile(filepath string) Logger
}

func NewLogger(kind string) Logger {
	var yllog Logger

	switch kind {
	case "simple":
		yllog = NewSimpleLog()
	default:
		yllog = NewSimpleLog()
	}
	return yllog
}

// yinglong专用日志
type SimpleLog struct {
	filepath string   //日志文件路径
	fp       *os.File //日志文件指针
}

func NewSimpleLog() *SimpleLog {
	return &SimpleLog{}
}

func (obj *SimpleLog) SetFile(filepath string) Logger {
	fp, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil
	}
	defer fp.Close()

	obj.filepath = filepath
	obj.fp = fp
	return obj
}

func (obj *SimpleLog) Print(level, content string) {
	io.WriteString(obj.fp, fmt.Sprintf("[%s] %s %s", time.Now().Format("2006-01-02 15:04:05"), level, content))
}
