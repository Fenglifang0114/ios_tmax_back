package svc

import (
	"os"

	"tmaxsrv/log"

	"github.com/spf13/viper"
)

const (
	DEFAULT_UI_CFG_STR = `{
		"dateformat": "1",
		"recmode": "auto",
		"stabletimetorec": "2",
		"zerorange": "1"
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
}

func NewUiConfig() *UiConfig {
	cnf := viper.New()
	cnf.AddConfigPath("./")
	cnf.SetConfigName(CFG_FILE_NAME)
	cnf.SetConfigType("json")

	var config Config
	if err := cnf.ReadInConfig(); err != nil {
		// create a config.json file with default
		os.WriteFile(CFG_FILE_NAME+".json", []byte(DEFAULT_UI_CFG_STR), 0o644)
		config := Config{RecMode: "auto", StableTimeToRec: "2", ZeroRange: "1", DateFormat: "1"}
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
	cnf := viper.New()
	cnf.AddConfigPath("./")
	cnf.SetConfigName("config")
	cnf.SetConfigType("json")
	c.Config = config
	cnf.Set("RecMode", config.RecMode)
	cnf.Set("StableTimeToRec", config.StableTimeToRec)
	cnf.Set("ZeroRange", config.ZeroRange)
	cnf.Set("DateFormat", config.DateFormat)
	return cnf.WriteConfig()
}
