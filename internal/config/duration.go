package config

import "time"

// Duration exists so that we can easily parse text durations like 3m or 7d
type Duration time.Duration

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
