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
	Name          string
	Servers       []string `toml:"servers"`
	MaxDopCores   int      `toml:"maxdop_cores"`
	MaxDopPercent int      `toml:"maxdop_percent"`
	CheckDB       struct {
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

// MaxDop determines the MAXDOP for a particular server
func (p Plan) MaxDop(cores, maxdop int) (int, error) {
	if p.MaxDopCores < 0 {
		return 0, fmt.Errorf("invalid maxdop_cores: %d", p.MaxDopCores)
	}
	if p.MaxDopPercent < 0 || p.MaxDopPercent > 100 {
		return 0, fmt.Errorf("invalid maxdop_percent: %d", p.MaxDopPercent)
	}
	if p.MaxDopCores == 0 && p.MaxDopPercent == 0 {
		return 0, nil
	}
	corespct := cores // start with cores and reduce if needed
	if p.MaxDopPercent > 0 {
		val := float64(cores) * (float64(p.MaxDopPercent) / 100.0)
		corespct = int(val)
		if corespct == 0 {
			corespct = 1
		}
	}
	coresnum := cores // start with cores and reduce if needed
	if p.MaxDopCores > 0 {
		if coresnum > p.MaxDopCores {
			coresnum = p.MaxDopCores
		}
	}
	value := lowest(coresnum, corespct)
	if value >= cores || value >= maxdop {
		return 0, nil
	}
	// MAYBE: if >= 3 and odd, subtract 1; based on `maxdop_even=true`
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
