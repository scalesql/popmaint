package zerr

import "fmt"

// Wrap an error with a message
func Wrap(err error, msg string) error {
	return fmt.Errorf("%s: %w", msg, err)
}

// Wrapf wraps an error with a formatted message
func Wrapf(err error, format string, args ...any) error {
	str := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s: %w", str, err)
}
