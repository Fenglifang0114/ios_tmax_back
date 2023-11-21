package svc

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"tmaxsrv/comm"
	"tmaxsrv/log"

	"github.com/spf13/viper"
)

const (
	DEFAULT_UI_CFG_STR = `{
		"dateformat": "1",
		"recmode": "auto",
		"stabletimetorec": "2",
		"zerorange": "1",
		"dateseparator":"-",
		"scalemode":"1"
	  }`
	CFG_FILE_NAME = "config"
)

type UiConfig struct {
	Config *Config
}

type Config struct {
	RecMode         string
	StableTimeToRec string
	ZeroRange       string
	DateFormat      string
	DateSeparator   string
	ScaleMode       string
}

func NewUiConfig() *UiConfig {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))
	currentPath := path[:index]
	currentPath = filepath.Join(currentPath, comm.SRV_DATA_PATH)
	cnf := viper.New()
	cnf.AddConfigPath(currentPath + "/")
	cnf.SetConfigName(CFG_FILE_NAME)
	cnf.SetConfigType("json")

	var config Config
	if err := cnf.ReadInConfig(); err != nil {
		// create a config.json file with default
		os.WriteFile(currentPath+"/"+CFG_FILE_NAME+".json", []byte(DEFAULT_UI_CFG_STR), 0o644)
		config := Config{RecMode: "auto", StableTimeToRec: "2", ZeroRange: "1", DateFormat: "1", DateSeparator: "-", ScaleMode: "1"}
		uiCfg := &UiConfig{Config: &config}
		uiCfg.UpdateConfig(&config)

		return uiCfg
	}

	if err := cnf.Unmarshal(&config); err != nil {
		log.Log.Error(err)
	}

	return &UiConfig{Config: &config}
}

func (c *UiConfig) GetConfig() (*Config, error) {
	return c.Config, nil
}

func (c *UiConfig) UpdateConfig(config *Config) error {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))
	currentPath := path[:index]
	currentPath = filepath.Join(currentPath, comm.SRV_DATA_PATH)
	cnf := viper.New()
	cnf.AddConfigPath(currentPath + "/")
	cnf.SetConfigName("config")
	cnf.SetConfigType("json")
	c.Config = config
	cnf.Set("RecMode", config.RecMode)
	cnf.Set("StableTimeToRec", config.StableTimeToRec)
	cnf.Set("ZeroRange", config.ZeroRange)
	cnf.Set("DateFormat", config.DateFormat)
	cnf.Set("DateSeparator", config.DateSeparator)
	cnf.Set("ScaleMode", config.ScaleMode)
	return cnf.WriteConfig()
}
