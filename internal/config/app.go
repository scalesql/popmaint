package config

import (
	"errors"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type AppConfig struct {
	Log struct {
		Level            string         `toml:"level"`
		Advanced         bool           `toml:"advanced"`
		Folder           string         `toml:"folder"`
		FileNameTemplate string         `toml:"file_name_template"`
		RetainDays       int            `toml:"retain_days"`
		PurgeGlob        string         `toml:"purge_glob"`
		UseUTC           bool           `toml:"use_utc"`
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
	err = appconfig.setlogsettings()
	if err != nil {
		return AppConfig{}, err
	}
	appconfig.Repository.UserName = os.Expand(appconfig.Repository.UserName, getenv)
	appconfig.Repository.Password = os.Expand(appconfig.Repository.Password, getenv)
	return appconfig, nil
}

func (ac *AppConfig) setlogsettings() error {
	if ac.Log.RetainDays < 0 {
		return errors.New("retain_days must be >= 0")
	}

	// validate advanced - required fields
	if ac.Log.Advanced {
		if ac.Log.Folder == "" {
			return errors.New("advanced logging requires log.folder")
		}
		if ac.Log.FileNameTemplate == "" {
			return errors.New("advanced logging requires log.file_name_template")
		}
		if ac.Log.RetainDays > 0 && ac.Log.PurgeGlob == "" {
			return errors.New("if retain_days > 0, log.purge_glob required")
		}
		return nil
	}
	// not advanced (Basic Logging) -- all should be empty
	if ac.Log.Folder != "" {
		return errors.New("setting log.folder requires advanced=true")
	}
	if ac.Log.FileNameTemplate != "" {
		return errors.New("setting log.log_file_template requires advanced=true")
	}
	if ac.Log.RetainDays > 0 {
		return errors.New("setting log.retain_days requires advanced=true")
	}
	if ac.Log.PurgeGlob != "" {
		return errors.New("settings log.purge_glob requires advanced=true")
	}

	// set defaults
	ac.Log.Folder = "./logs/json"
	ac.Log.FileNameTemplate = "{{.job_id}}.ndjson"
	ac.Log.RetainDays = 30
	ac.Log.PurgeGlob = "*.ndjson"

	return nil
}
