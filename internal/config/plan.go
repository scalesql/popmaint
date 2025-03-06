package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
)

type Duration time.Duration

type Plan struct {
	Name                string
	Servers             []string `toml:"servers"`
	MaxDopCores         int      `toml:"maxdop_cores"`
	MaxDopPercent       int      `toml:"maxdop_pct"`
	MaxDopPercentMaxDop int      `toml:"maxdop_pct_maxdop"`
	Log                 struct {
		Level string `toml:"level"`
	} `toml:"log"`
	CheckDB struct {
		TimeLimit             Duration `toml:"time_limit" json:"-"`
		NoIndex               bool     `toml:"no_index" json:"no_index"`
		MaxSizeMB             int      `toml:"max_size_mb" json:"-"`
		PhysicalOnly          bool     `toml:"physical_only" json:"physical_only"`
		InfoMessages          bool     `toml:"info_messages" json:"info_messages"`
		ExtendedLogicalChecks bool     `toml:"extended_logical_checks" json:"extended_logical_checks"`
		DataPurity            bool     `toml:"data_purity" json:"data_purity"`
		EstimateOnly          bool     `toml:"estimate_only" json:"estimate_only"`
		Included              []string `toml:"included" json:"-"`
		Excluded              []string `toml:"excluded" json:"-"`
		MinIntervalDays       int      `toml:"min_interval_days" json:"-"`
	} `toml:"checkdb" json:"checkdb"`
}

func ReadPlan(name string) (Plan, error) {
	plan := Plan{Name: name}
	// TODO -- sort the sections and run them in order
	fileName := fmt.Sprintf("%s.toml", name)
	fileName = filepath.Join(".", "plans", fileName)
	if _, err := os.Stat(fileName); errors.Is(err, os.ErrNotExist) {
		//return plan, err
		return plan, fmt.Errorf("plan file not found: %s", fileName)
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
		return Plan{}, fmt.Errorf("invalid maxdop_pct: %d", plan.MaxDopPercent)
	}
	if plan.CheckDB.DataPurity && plan.CheckDB.PhysicalOnly {
		return Plan{}, fmt.Errorf("can't set data_purity and physical_only")
	}
	return plan, nil
}

// RemoveDupes removes duplicate servers from the Plan using a case-insensitive
// comparison.  It returns any duplicates for logging.
func (p *Plan) RemoveDupes() []string {
	servers := make(map[string]bool)
	dupes := make([]string, 0)
	uniques := make([]string, 0)
	for _, srv := range p.Servers {
		key := strings.ToLower(srv)
		_, ok := servers[key]
		if ok {
			dupes = append(dupes, srv)
		} else {
			uniques = append(uniques, srv)
			servers[key] = true
		}
	}
	p.Servers = uniques
	return dupes
}

func (p Plan) MaxDop(serverCores, serverMaxdop int) (int, error) {
	if p.MaxDopCores < 0 {
		return 0, fmt.Errorf("invalid maxdop_cores: %d", p.MaxDopCores)
	}
	if p.MaxDopPercent < 0 || p.MaxDopPercent > 100 {
		return 0, fmt.Errorf("invalid maxdop_pct: %d", p.MaxDopPercent)
	}
	if p.MaxDopPercentMaxDop < 0 || p.MaxDopPercentMaxDop > 100 {
		return 0, fmt.Errorf("invalid maxdop_pct_maxdop: %d", p.MaxDopPercent)
	}

	// if we didn't set any values, use the default
	if p.MaxDopCores == 0 && p.MaxDopPercent == 0 && p.MaxDopPercentMaxDop == 0 {
		return 0, nil
	}

	// figure out the lowest value from our settings
	// one of these will be set since we exited above if they were all zero
	ceiling := p.maxdopCeiling(serverCores, serverMaxdop)
	if serverMaxdop == 0 {
		if ceiling < serverCores {
			return ceiling, nil // maxdop is not set and we have a ceiling below cores
		} else {
			return 0, nil // just use all the cores
		}
	} else {
		if ceiling < serverMaxdop {
			return ceiling, nil // maxdop is set and we have a ceiling below that
		}
		return 0, nil // just use the server maxdop setting
	}
}

// calcMaxdopCeiling calculates the ceiling imposed on MAXDOP by the settings
// from the TOML file. It uses the lowest of the calculated values
// ceiling will always be 1 or higher
func (p Plan) maxdopCeiling(serverCores, serverMaxdop int) int {
	ceiling := serverCores
	if p.MaxDopCores > 0 { // we have a setting
		if p.MaxDopCores < ceiling { // it's less than we already have
			ceiling = p.MaxDopCores // plan cores
		}
	}

	if p.MaxDopPercent > 0 { // we have a setting
		val := int(float64(serverCores) * (float64(p.MaxDopPercent) / 100.0))
		if val == 0 {
			val = 1
		}
		if val < ceiling {
			ceiling = val // plan cores from percent
		}
	}

	// calculate as a percent of maxdop
	if p.MaxDopPercentMaxDop > 0 && serverMaxdop > 0 { // we have a setting
		val := int(float64(serverMaxdop) * (float64(p.MaxDopPercentMaxDop) / 100.0))
		if val == 0 {
			val = 1
		}
		if val < ceiling {
			ceiling = val // plan cores from percent
		}
	}

	if ceiling < 1 { // I think this should never hit
		ceiling = 1
	}
	return ceiling
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
