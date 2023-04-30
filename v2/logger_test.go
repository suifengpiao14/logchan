package logchan_test

import (
	"fmt"
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
}

func (l *LogInfoTest) GetName() (logName logchan.LogName) {
	return LOG_INFO_TEST
}
func (l *LogInfoTest) Error() (err error) {
	return l.err
}

func TestMakeTypeError(t *testing.T) {
	logInfo := LogInfoTest{}
	err := logchan.MakeTypeError(&logInfo)
	fmt.Println(err.Error())
}