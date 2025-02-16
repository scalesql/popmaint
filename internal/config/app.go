package config

import (
	"errors"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type AppConfig struct {
	Log struct {
		Level            string         `toml:"level"`
		LogRetentionDays int            `toml:"retention_days"`
		Fields           map[string]any `toml:"fields"`
	} `toml:"log"`
	Repository struct {
		Server   string `toml:"server"`
		Database string `toml:"database"`
		UserName string `toml:"username"`
		Password string `toml:"password"`
	} `toml:"repository"`
}

// ReadConfig reads popmaint.toml and returns an AppConfig structure.  It also
// accepts a mapper function to replace environment variables in certain fields.
func ReadConfig(getenv func(string) string) (AppConfig, error) {
	if _, err := os.Stat("popmaint.toml"); errors.Is(err, os.ErrNotExist) {
		appconfig := AppConfig{}
		appconfig.Log.LogRetentionDays = 30
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
	appconfig.Repository.UserName = os.Expand(appconfig.Repository.UserName, getenv)
	appconfig.Repository.Password = os.Expand(appconfig.Repository.Password, getenv)
	return appconfig, nil
}
