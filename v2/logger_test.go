package logchan_test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/suifengpiao14/logchan/v2"
)

type LogName string

func (l LogName) String() string {
	return string(l)
}

const (
	LOG_INFO_TEST LogName = "test"
)

type LogInfoTest struct {
	err error
	logchan.EmptyLogInfo
}

func (l *LogInfoTest) GetName() (logName logchan.LogName) {
	return LOG_INFO_TEST
}
func (l *LogInfoTest) Error() (err error) {
	return l.err
}

func (l *LogInfoTest) BeforeSend() {}

func TestMakeTypeError(t *testing.T) {
	logInfo := LogInfoTest{}
	err := logchan.MakeTypeError(&logInfo)
	fmt.Println(err.Error())
}

func TestSessionID(t *testing.T) {
	s := logchan.SessionID()
	fmt.Println(s)
}

func TestGetRunInfoFromFrames(t *testing.T) {
	logInfo := &LogInfoTest{}
	FnA(logInfo)
	frams := logchan.GetCallersFrames(logInfo)
	filename, fullFuncName, line := logchan.GetCallStackInfoFromFrames(frams, func(filename, fullFuncName string, line int, frame runtime.Frame) (ok bool) {
		return true
	})
	fmt.Println(filename, fullFuncName, line)
}

func FnA(logInfo *LogInfoTest) {
	defer func() {
		logchan.SendLogInfo(logInfo)
	}()
}
