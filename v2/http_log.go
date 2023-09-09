package logchan

import (
	"context"
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
	Name    string          `json:"name"`
	Method  string          `json:"method"`
	Url     string          `json:"url"`
	Input   string          `json:"input"`
	Output  string          `json:"output"`
	Context context.Context `json:"context"`
	Err     error
}
type DefaultLogName string

func (l DefaultLogName) String() string {
	return string(l)
}

func (h HttpLogInfo) GetName() (logName LogName) {
	return HttpLogInfoName
}

func (h HttpLogInfo) Error() (err error) {
	return err
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
