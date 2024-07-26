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
	Name          string
	Servers       []string `toml:"servers"`
	MaxDopCores   int      `toml:"maxdop_cores"`
	MaxDopPercent int      `toml:"maxdop_percent"`
	CheckDB       struct {
		TimeLimit             Duration `toml:"time_limit"`
		NoIndex               bool     `toml:"no_index"`
		MaxSizeMB             int      `toml:"max_size_mb"`
		PhysicalOnly          bool     `toml:"physical_only"`
		InfoMessages          bool     `toml:"info_messages"`
		ExtendedLogicalChecks bool     `toml:"extended_logical_checks"`
		DataPurity            bool     `toml:"data_purity"`
		EstimateOnly          bool     `toml:"estimate_only"`
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
	if plan.MaxDopCores < 0 {
		return Plan{}, fmt.Errorf("invalid maxdop_cores: %d", plan.MaxDopCores)
	}
	if plan.MaxDopPercent < 0 || plan.MaxDopPercent > 100 {
		return Plan{}, fmt.Errorf("invalid maxdop_percent: %d", plan.MaxDopPercent)
	}
	if plan.CheckDB.DataPurity && plan.CheckDB.PhysicalOnly {
		return Plan{}, fmt.Errorf("can't set data_purity and physical_only")
	}
	return plan, nil
}

func (p Plan) MaxDop(cores int) (int, error) {
	if p.MaxDopCores < 0 {
		return 0, fmt.Errorf("invalid maxdop_cores: %d", p.MaxDopCores)
	}
	if p.MaxDopPercent < 0 || p.MaxDopPercent > 100 {
		return 0, fmt.Errorf("invalid maxdop_percent: %d", p.MaxDopPercent)
	}
	if p.MaxDopCores == 0 && p.MaxDopPercent == 0 {
		return 0, nil
	}
	pctcores := 0
	if p.MaxDopPercent > 0 {
		val := float64(cores) * (float64(p.MaxDopPercent) / 100.0)
		pctcores = int(val)
	}
	value := lowest(p.MaxDopCores, pctcores)
	if value > cores {
		return 0, nil
	}
	// MAYBE: if > 3 and odd, subtract 1
	// based on `maxdop_even=true`
	return value, nil
}

// lowest returns the lowest non-zero value
// it returns zero if no values are > 0
func lowest(vals ...int) int {
	maxint := int(^uint(0) >> 1)
	lowest := maxint
	for _, j := range vals {
		if j <= 0 {
			continue
		}
		if j < lowest {
			lowest = j
		}
	}
	if lowest == maxint {
		return 0
	}
	return lowest
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
