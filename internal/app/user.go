//go:build !windows
// +build !windows

package app

import (
	"os/user"

	"github.com/pkg/errors"
)

// currentUserName returns the current user name in DOMAIN\USER format
// TODO need a better solution: https://github.com/golang/go/issues/37348
func currentUserName() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return errors.Wrap(err, "user.current").Error(), nil
		//return "", errors.Wrap(err, "user.current")
	}
	return usr.Username, nil
}
