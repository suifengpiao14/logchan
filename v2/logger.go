package logchan

import (
	"context"
	"fmt"
	"sync"
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
var setLoggerWriteOnce sync.Once
var sendLogInfoFn func(logInfo LogInforInterface) //只能初始化一次
var doneChan chan struct{}

// SetLoggerWriter 设置日志处理函数，同时返回发送日志函数
func SetLoggerWriter(ctx context.Context, handlerLogInfoFn func(logInfo LogInforInterface, typeName string, err error)) {
	if handlerLogInfoFn == nil {
		return
	}
	setLoggerWriteOnce.Do(func() {
		go func() {
			defer func() {
				doneChan <- struct{}{} // 通知日志写入结束
				close(doneChan)
				result := recover() // 此处由错误，直接丢弃，无法输出，可探讨是否可以输出到标准输出
				if result != nil {
					msg := fmt.Sprintf("logchan.setLoggerWriteOnce.Do recover result:%v", result)
					fmt.Println(msg)
				}
			}()
			for logInfo := range logInfoChain {
				func(logInfo LogInforInterface) {
					handlerLogInfoFn(logInfo, logInfo.GetName(), logInfo.Error())
				}(logInfo)

			}
		}()

		go func() {
			select {
			case <-ctx.Done():
				close(logInfoChain)
				return
			}
		}()

		sendLogInfoFn = func(logInfo LogInforInterface) {
			select {
			case logInfoChain <- logInfo:
				return
			default:
				return

			}
		}
	})
	return
}

func SendLogInfo(logInfo LogInforInterface) {
	if sendLogInfoFn == nil {
		return
	}
	sendLogInfoFn(logInfo)
}

// UntilFinished 阻塞，直到所有日志处理完,timeout 等待处理超时时间
func UntilFinished(timeout time.Duration, closeLogInfoChan bool) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if closeLogInfoChan && logInfoChain != nil {
		close(logInfoChain)
	}
	select {
	case <-ctx.Done():
		if !closeLogInfoChan && logInfoChain != nil {
			close(logInfoChain)
		}
		<-doneChan //阻塞 确保已有的日志被处理完毕
	case <-doneChan:
	}
}
