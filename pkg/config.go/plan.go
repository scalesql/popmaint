package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pelletier/go-toml/v2"
)

type Duration time.Duration

type Plan struct {
	Name    string
	Servers []string `toml:"servers"`
	CheckDB struct {
		TimeLimit    Duration `toml:"time_limit"`
		NoIndex      bool     `toml:"no_index"`
		MaxSizeMB    int      `toml:"max_size_mb"`
		PhysicalOnly bool     `toml:"physical_only"`
		InfoMessages bool     `toml:"info_messages"`
	} `toml:"checkdb"`
}

func ReadPlan(name string) (Plan, error) {
	plan := Plan{Name: name}
	// TODO -- sort the sections and run them in order
	fileName := fmt.Sprintf("%s.toml", name)
	fileName = filepath.Join(".", "plans", fileName)
	if _, err := os.Stat(fileName); errors.Is(err, os.ErrNotExist) {
		return plan, err
	}
	bb, err := os.ReadFile(fileName)
	if err != nil {
		return plan, err
	}
	err = toml.Unmarshal(bb, &plan)
	if err != nil {
		return plan, err
	}
	return plan, nil
}

func (d *Duration) UnmarshalText(b []byte) error {
	x, err := time.ParseDuration(string(b))
	if err != nil {
		return err
	}
	*d = Duration(x)
	return nil
}

func (d Duration) String() string {
	dur := time.Duration(d)
	return dur.String()
}
