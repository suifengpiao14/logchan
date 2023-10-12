package logchan

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	HttpLogInfoName DefaultLogName = "HttpLogInfo"
)

type HttpLogInfo struct {
	Name   string `json:"name"`
	Method string `json:"method"`
	Url    string `json:"url"`
	Input  string `json:"input"`
	Output string `json:"output"`
	Curl   string `json:"curl"`
	Err    error
	EmptyLogInfo
}
type DefaultLogName string

func (l DefaultLogName) String() string {
	return string(l)
}

func (h *HttpLogInfo) GetName() (logName LogName) {
	return HttpLogInfoName
}

func (h *HttpLogInfo) Error() (err error) {
	return err
}
func (h *HttpLogInfo) BeforSend() {
	h.Curl, _ = h.CURLCli() // 此处的err不能影响业务error
}

//DefaultPrintWatchFileLog 默认日志打印函数
func DefaultPrintWatchFileLog(logInfo LogInforInterface, typeName LogName, err error) {
	if typeName != HttpLogInfoName {
		return
	}
	httpLogInfo, ok := logInfo.(*HttpLogInfo)
	if !ok {
		return
	}
	if err != nil {
		fmt.Fprintf(LogWriter, "loginInfo:%s,error:%s", httpLogInfo.GetName(), err.Error())
		return
	}
	curlcmd, _ := httpLogInfo.CURLCli()
	fmt.Fprintf(LogWriter, "curl:%s", curlcmd)
}

// CURLCli 生成curl 命令
func (h HttpLogInfo) CURLCli() (curlCli string, err error) {
	switch strings.ToUpper(h.Method) {
	case http.MethodPost:
		curlCli = fmt.Sprintf(`curl -X%s -d'%s' '%s'`, strings.ToUpper(h.Method), h.Input, h.Url)
	case http.MethodGet:
		params := make(map[string]string)
		u, err := url.Parse(h.Url)
		if err != nil {
			return "", err
		}
		values := u.Query()
		if h.Input != "" {
			err = json.Unmarshal([]byte(h.Input), &params)
			if err != nil {
				return "", err
			}
		}
		for k, v := range params {
			if values.Has(k) {
				continue
			}
			values.Add(k, v)
		}
		u.RawQuery = values.Encode()
		curlCli = fmt.Sprintf(`curl -X%s  '%s'`, strings.ToUpper(h.Method), u.String())
	}

	return curlCli, nil
}
