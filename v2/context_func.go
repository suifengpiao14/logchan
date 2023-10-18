package logchan

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"github.com/suifengpiao14/funcs"
)

const (
	context_Name_GoroutineID   LogVaraible = "GoroutineID"
	context_Name_SessionID     LogVaraible = "SessionID"
	context_Name_CallersFrames LogVaraible = "CallersFrames"
	context_Name_Run_Time      LogVaraible = "run_time"
)

func setProcessSessionID(logInfo LogInforInterface) {
	ctx := logInfo.GetContext()
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = context.WithValue(ctx, context_Name_SessionID, SessionID())
	logInfo.SetContext(ctx)
}
func setGoroutineID(logInfo LogInforInterface) {
	ctx := logInfo.GetContext()
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = context.WithValue(ctx, context_Name_GoroutineID, funcs.GoroutineID())
	logInfo.SetContext(ctx)
}

// setCallersFrames 设置运行时信息
func setCallersFrames(logInfo LogInforInterface) {
	ctx := logInfo.GetContext()
	if ctx == nil {
		ctx = context.Background()
	}
	var pcArr [32]uintptr // at least 1 entry needed
	var frames *runtime.Frames
	n := 0
	n = runtime.Callers(3, pcArr[:]) // 去除SetCallersFrames、SendLogInfo 2层
	frames = runtime.CallersFrames(pcArr[:n])
	ctx = context.WithValue(ctx, context_Name_CallersFrames, frames)
	logInfo.SetContext(ctx)
}

func setRunTime(logInfo LogInforInterface) {
	ctx := logInfo.GetContext()
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = context.WithValue(ctx, context_Name_Run_Time, time.Now().Local())
	logInfo.SetContext(ctx)
}

//GetGoroutineID 从日志记录中获取协程ID
func GetGoroutineID(logInfo LogInforInterface) (goroutineID string) {
	ctx := logInfo.GetContext()
	i := ctx.Value(context_Name_GoroutineID)
	goroutineID = cast.ToString(i)
	return goroutineID
}

//GetGoroutineID 从日志记录中获取ip、进程、协程ID
func GetSessionID(logInfo LogInforInterface) (sessionID string) {
	ctx := logInfo.GetContext()
	i := ctx.Value(context_Name_SessionID)
	sessionID = cast.ToString(i)
	return sessionID
}

//GetCallersFrames 获取运行栈信息
func GetCallersFrames(logInfo LogInforInterface) (frames *runtime.Frames) {
	ctx := logInfo.GetContext()
	value := ctx.Value(context_Name_CallersFrames)
	frames, ok := value.(*runtime.Frames)
	if !ok {
		err := errors.Errorf("exept *runtime.Frames,got:%T", value)
		panic(err)
	}
	return frames
}

func GetTime(logInfo LogInforInterface) (t time.Time) {
	ctx := logInfo.GetContext()
	value := ctx.Value(context_Name_Run_Time)
	t, ok := value.(time.Time)
	if !ok {
		err := errors.Errorf("exept time.Time,got:%T", value)
		panic(err)
	}
	return t
}

//GetCallInfo 获取调用函数信息
func GetCallInfo(logInfo LogInforInterface) (filename string, fullFuncName string, line int) {
	frames := GetCallersFrames(logInfo)
	filename, fullFuncName, line = GetCallStackInfoFromFrames(frames, DefaultFramesFilterFn)
	return filename, fullFuncName, line
}

//DefaultPrintLog 输出默认日志
func DefaultPrintLog(logInfo LogInforInterface) (s string) {
	processSessionID := GetSessionID(logInfo)
	t := GetTime(logInfo).Format("20060102150405")
	file, funcName, line := GetCallInfo(logInfo)
	s = fmt.Sprintf(`time:%s|coroutineID:%s|file:%s|func:%s|line:%d`, t, processSessionID, file, funcName, line)
	return s
}

func SessionID() string {
	goid := funcs.GoroutineID()
	ip, _ := funcs.GetIp()
	s := fmt.Sprintf("%s-%d-%d-%d", ip, os.Getppid(), os.Getpid(), goid)
	digestBytes := md5.Sum([]byte(s))
	md5Str := fmt.Sprintf("%x", digestBytes)
	return md5Str[0:16]
}
