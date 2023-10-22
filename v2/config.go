package logchan

import (
	"os"
	"path/filepath"
	"time"

	"github.com/go-playground/validator"
	"github.com/pkg/errors"
	"github.com/suifengpiao14/funcs"
)

type Config struct {
	FileFormat string `mapstructure:"fileFormat" json:"fileFormat"`
	AppName    string `mapstructure:"appName" json:"appName" validate:"required"`
	ModePerm   *os.FileMode
}

var (
	configInstance *Config

	ERROR_Init_Config = errors.New("use SetConfig init Config")
)

func SetConfig(c *Config) (err error) {
	validate := validator.New()
	err = validate.Struct(c)
	if err != nil {
		return err
	}
	if c.ModePerm == nil {
		m := os.ModePerm
		c.ModePerm = &m
	}
	configInstance = c
	if configInstance.FileFormat != "" {
		logFile := funcs.Strtr(configInstance.FileFormat, map[string]string{
			"{appName}": c.AppName,
			"{date}":    time.Now().Format("20060102"),
		})
		dir := filepath.Dir(logFile)
		err = os.MkdirAll(dir, *configInstance.ModePerm)
		if err != nil {
			return err
		}
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, *configInstance.ModePerm)
		if err != nil {
			return err
		}
		LogWriter = f
	}
	return nil
}
