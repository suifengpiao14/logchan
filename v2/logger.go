package logchan

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"
)

type LogName interface {
	String() (name string)
}

type LogInforInterface interface {
	GetName() (logName LogName) // 当logchan 应用广泛后，字符串 name 容易冲突，建议在具体包内定义 string 别名方式解决该问题
	Error() (err error)
	BeforSend() // 在发送前,整理数据,如运行函数填充数据
}

var (
	ERROR_NOT_IMPLEMENTED = errors.New("not implemented")
)

// MakeTypeError 生成类型错误
func MakeTypeError(l LogInforInterface) (err error) {
	err = errors.Errorf("type error: excetp:%s,got:%T", l.GetName().String(), l)
	return err
}

type EmptyLogInfo struct {
}

func (l *EmptyLogInfo) GetName() (name LogName) {
	err := errors.WithMessage(ERROR_NOT_IMPLEMENTED, "GetName")
	panic(err)
}

func (l *EmptyLogInfo) Error() (err error) {
	err = errors.WithMessage(ERROR_NOT_IMPLEMENTED, "Error")
	panic(err)
}
func (l *EmptyLogInfo) BeforSend() {
}

// LogInfoChainBuffer 日志缓冲区,减少并发日志丢失情况
var LogInfoChainBuffer int = 50

var setLoggerWriteOnce sync.Once
var sendLogInfoFn func(logInfo LogInforInterface) //只能初始化一次
var closeLogChan func()

// SetLoggerWriter 设置日志处理函数，同时返回发送日志函数
func SetLoggerWriter(handlerLogInfoFn func(logInfo LogInforInterface, logName LogName, err error)) {
	if handlerLogInfoFn == nil {
		return
	}

	setLoggerWriteOnce.Do(func() {
		var doneChan = make(chan struct{}, 1) //设置缓冲后，不会阻塞当前协程
		// logInfoChain 日志传送通道，缓冲区满后,会丢弃日志
		var logInfoChain = make(chan LogInforInterface, LogInfoChainBuffer)
		// 启动监听
		go func() {
			defer func() {
				doneChan <- struct{}{} // 通知日志处理完成
				close(doneChan)
				result := recover() // 此处由错误，直接丢弃，无法输出，可探讨是否可以输出到标准输出
				if result != nil {
					msg := fmt.Sprintf("logchan.setLoggerWriteOnce.Do recover result:%v", result)
					fmt.Println(msg)
				}
			}()
			for logInfo := range logInfoChain {
				handlerLogInfoFn(logInfo, logInfo.GetName(), logInfo.Error())
			}
		}()

		//封装发送句柄
		sendLogInfoFn = func(logInfo LogInforInterface) {
			select {
			case logInfoChain <- logInfo:
				return
			default:
				return

			}
		}

		//封装关闭句柄
		closeLogChan = func() {
			close(logInfoChain)
			<-doneChan
		}

	})

}

func SendLogInfo(logInfo LogInforInterface) {
	if sendLogInfoFn == nil {
		return
	}
	logInfo.BeforSend()
	sendLogInfoFn(logInfo)
}

func CloseLogChan() {
	if closeLogChan == nil {
		return
	}
	closeLogChan()
}
