package logchan

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
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
	GetName() string
	Error() error
	GetLevel() string
}

//LogInfoChainBuffer 日志缓冲区,减少并发日志丢失情况
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
				recover() // 此处由错误，直接丢弃，无法输出，可探讨是否可以输出到标准输出
			}()
			for logInfo := range logInfoChain {
				atomic.AddInt64(&count, -1)
				fn(logInfo, logInfo.GetName(), logInfo.Error())
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

//IsFinished 检测管道日志是否全部输出
func IsFinished() (yes bool) {
	return atomic.LoadInt64(&count) <= 0
}

//UntilFinished 阻塞，直到所有日志处理完,maxInterval 处理日志最长时间
func UntilFinished(maxInterval time.Duration) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, maxInterval)
	defer cancel()
	select {
	case <-time.After(1 * time.Microsecond):
		if IsFinished() {
			cancel()
		}
	case <-ctx.Done():
	}
}
