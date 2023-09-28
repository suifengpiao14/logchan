package logchan

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cast"
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
	ctx context.Context
}

func (l *EmptyLogInfo) SetContext(ctx context.Context) {
	l.ctx = ctx
}
func (l *EmptyLogInfo) GetContext() (ctx context.Context) {
	if l.ctx == nil {
		ctx = context.Background()
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
func (l *EmptyLogInfo) BeforSend() {
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
	_handlerLogInfoFns = append(_handlerLogInfoFns, handlerLogInfoFns...)
}

type LogVaraible string

const (
	Context_Name_GoroutineID LogVaraible = "GoroutineID"
	Context_Name_SessionID   LogVaraible = "SessionID"
)

func SendLogInfo(logInfo LogInforInterface) {
	ctx := logInfo.GetContext()
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = context.WithValue(ctx, Context_Name_GoroutineID, funcs.GoroutineID())
	ctx = context.WithValue(ctx, Context_Name_SessionID, SessionID())
	logInfo.SetContext(ctx)
	select {
	case logInfoChain <- logInfo:
		return
	default:
		return

	}
}

//GetGoroutineID 从日志记录中获取协程ID
func GetGoroutineID(logInfo LogInforInterface) (goroutineID string) {
	ctx := logInfo.GetContext()
	i := ctx.Value(Context_Name_GoroutineID)
	goroutineID = cast.ToString(i)
	return goroutineID
}

//GetGoroutineID 从日志记录中获取ip、进程、协程ID
func GetSessionID(logInfo LogInforInterface) (sessionID string) {
	ctx := logInfo.GetContext()
	i := ctx.Value(Context_Name_SessionID)
	sessionID = cast.ToString(i)
	return sessionID
}

func CloseLogChan() {
	close(logInfoChain)
	<-doneChan
}

func SessionID() string {
	goid := funcs.GoroutineID()
	ip, _ := funcs.GetIp()
	s := fmt.Sprintf("%s-%d-%d-%d", ip, os.Getppid(), os.Getpid(), goid)
	digestBytes := md5.Sum([]byte(s))
	md5Str := fmt.Sprintf("%x", digestBytes)
	return md5Str[0:16]
}
