package config

import (
	"errors"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type AppConfig struct {
	Logging struct {
		LogRetentionDays int            `toml:"log_retention_days"`
		Fields           map[string]any `toml:"fields"`
	} `toml:"logging"`
}

func ReadConfig() (AppConfig, error) {
	if _, err := os.Stat("popmaint.toml"); errors.Is(err, os.ErrNotExist) {
		appconfig := AppConfig{}
		appconfig.Logging.LogRetentionDays = 30
		return appconfig, nil
	}
	bb, err := os.ReadFile("popmaint.toml")
	if err != nil {
		return AppConfig{}, err
	}
	appconfig := AppConfig{}
	err = toml.Unmarshal(bb, &appconfig)
	if err != nil {
		return AppConfig{}, err
	}
	return appconfig, nil
}
