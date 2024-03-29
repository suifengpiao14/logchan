package logchan

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/pkg/errors"
	"github.com/suifengpiao14/funcs"
)

type LogName interface {
	String() (name string)
}

type LogInforInterface interface {
	SetContext(ctx context.Context)    // 增加上下文，方便统一增加协程ID、链路追踪id等通用变量
	GetContext() (ctx context.Context) // 增加上下文，方便统一增加协程ID、链路追踪id等通用变量
	GetName() (logName LogName)        // 当logchan 应用广泛后，字符串 name 容易冲突，建议在具体包内定义 string 别名方式解决该问题
	Error() (err error)
	BeforeSend() // 在发送前,整理数据,如运行函数填充数据
}

var (
	ERROR_NOT_IMPLEMENTED = errors.New("not implemented")
)

// LogWriter 外部可以指定日志写入句柄,默认标准输出
var LogWriter io.WriteCloser = os.Stdout

// MakeTypeError 生成类型错误
func MakeTypeError(l LogInforInterface) (err error) {
	err = errors.Errorf("type error: excetp:%s,got:%T", l.GetName().String(), l)
	return err
}

type EmptyLogInfo struct {
	ctx context.Context
}

func (l *EmptyLogInfo) SetContext(ctx context.Context) {
	l.ctx = ctx
}
func (l *EmptyLogInfo) GetContext() (ctx context.Context) {
	if l.ctx == nil {
		l.ctx = context.Background()
	}
	return l.ctx
}

func (l *EmptyLogInfo) GetName() (name LogName) {
	err := errors.WithMessage(ERROR_NOT_IMPLEMENTED, "GetName")
	panic(err)
}

func (l *EmptyLogInfo) Error() (err error) {
	err = errors.WithMessage(ERROR_NOT_IMPLEMENTED, "Error")
	panic(err)
}
func (l *EmptyLogInfo) BeforeSend() {
}

// LogInfoChainBuffer 日志缓冲区,减少并发日志丢失情况
var LogInfoChainBuffer int = 50

var doneChan = make(chan struct{}, 1) //设置缓冲后，不会阻塞当前协程
// logInfoChain 日志传送通道，缓冲区满后,会丢弃日志
var logInfoChain = make(chan LogInforInterface, LogInfoChainBuffer)

type LogInfoHandlerFn func(logInfo LogInforInterface, logName LogName, err error)

// 处理函数queue
var _handlerLogInfoFns = make([]LogInfoHandlerFn, 0)

func init() {
	// 初始化监听
	go func() {
		defer func() {
			LogWriter.Close()
			doneChan <- struct{}{} // 通知日志处理完成
			close(doneChan)
			result := recover() // 此处由错误，直接丢弃，无法输出，可探讨是否可以输出到标准输出
			if result != nil {
				msg := fmt.Sprintf("logchan.setLoggerWriteOnce.Do recover result:%v", result)
				fmt.Println(msg)
			}
		}()
		for logInfo := range logInfoChain {
			for _, fn := range _handlerLogInfoFns {
				fn(logInfo, logInfo.GetName(), logInfo.Error())
			}
		}
	}()
}

// SetLoggerWriter 设置日志处理函数，同时返回发送日志函数
func SetLoggerWriter(handlerLogInfoFns ...LogInfoHandlerFn) {
	for _, h := range handlerLogInfoFns {
		if funcs.IsNil(h) {
			continue
		}
		_handlerLogInfoFns = append(_handlerLogInfoFns, h)
	}
}

type LogVaraible string

func SendLogInfo(logInfo LogInforInterface) {
	setGoroutineID(logInfo)
	setProcessSessionID(logInfo)
	setCallersFrames(logInfo)
	setRunTime(logInfo)
	logInfo.BeforeSend() // 发送前执行格式化（在当前携程执行,方便调试，符合通道仅仅传递消息的原则,即便消息被序列化为字符串，也能执行）
	select {
	case logInfoChain <- logInfo:
		return
	default:
		return

	}
}

func CloseLogChan() {
	close(logInfoChain)
	<-doneChan
}

// GetCallStackInfoFromFrames 获取运行时文件、函数、行号信息
func GetCallStackInfoFromFrames(frames *runtime.Frames, filterFn func(filename, fullFuncName string, line int, frame runtime.Frame) (ok bool)) (filename string, fullFuncName string, line int) {
	for {
		frame, hasNext := frames.Next()
		if !hasNext {
			break
		}
		filename = frame.File
		fullFuncName = frame.Function
		line = frame.Line
		if filterFn(filename, fullFuncName, line, frame) {
			break
		}
	}
	return filename, fullFuncName, line
}

// DefaultFramesFilterFn  默认调用栈过滤函数，返回发起 logchan.SendLogInfo 的函数信息
func DefaultFramesFilterFn(filename, fullFuncName string, line int, frame runtime.Frame) (ok bool) {
	return true
}
