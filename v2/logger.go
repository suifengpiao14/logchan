package logchan

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
)

const (
	LOG_LEVEL_FATAL = "fatal"
	LOG_LEVEL_ERROR = "error"
	LOG_LEVEL_WARN  = "warn"
	LOG_LEVEL_INFO  = "info"
	LOG_LEVEL_DEBUG = "debug"
	LOG_LEVEL_TRACE = "trace"
)

type LogInforInterface interface {
	GetName() (name string)
	Error() (err error)
	GetLevel() (level string)
}

var (
	ERROR_NOT_IMPLEMENTED = errors.New("not implemented")
)

type EmptyLogInfo struct {
}

func (l *EmptyLogInfo) GetName() (name string) {
	err := errors.WithMessage(ERROR_NOT_IMPLEMENTED, "GetName")
	panic(err)
}

func (l *EmptyLogInfo) Error() (err error) {
	err = errors.WithMessage(ERROR_NOT_IMPLEMENTED, "Error")
	panic(err)
}

func (l *EmptyLogInfo) GetLevel() (name string) {
	err := errors.WithMessage(ERROR_NOT_IMPLEMENTED, "GetLevel")
	panic(err)
}

// LogInfoChainBuffer 日志缓冲区,减少并发日志丢失情况
var LogInfoChainBuffer int = 50

// logInfoChain 日志传送通道，缓冲区满后,会丢弃日志
var logInfoChain = make(chan LogInforInterface, LogInfoChainBuffer)

var setLoggerWrite sync.Once
var count int64 // 原子计数,用来支持优雅退出
// SetLoggerWriter 设置日志输出逻辑
func SetLoggerWriter(fn func(logInfo LogInforInterface, typeName string, err error)) {
	if fn == nil {
		return
	}
	setLoggerWrite.Do(func() {
		go func() {
			defer func() {
				result := recover() // 此处由错误，直接丢弃，无法输出，可探讨是否可以输出到标准输出
				if result != nil {
					msg := fmt.Sprintf("logchan.setLoggerWrite.Do recover result:%v", result)
					fmt.Println(msg)
				}
			}()
			for logInfo := range logInfoChain {
				func(logInfo LogInforInterface) {
					defer atomic.AddInt64(&count, -1) // 确保函数执行后操作计数
					fn(logInfo, logInfo.GetName(), logInfo.Error())
				}(logInfo)

			}
		}()
	})

}

func SendLogInfo(info LogInforInterface) {
	select { // 不阻塞写入,避免影响主程序
	case logInfoChain <- info:
		atomic.AddInt64(&count, 1)
		return
	default:
		return
	}
}

// IsFinished 检测管道日志是否全部输出
func IsFinished() (yes bool) {
	return atomic.LoadInt64(&count) <= 0
}

// UntilFinished 阻塞，直到所有日志处理完,timeout 等待处理超时时间
func UntilFinished(timeout time.Duration) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(10 * time.Millisecond)
			if IsFinished() {
				return
			}
		}
	}
}
