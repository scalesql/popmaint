//go:build windows
// +build windows

package app

import (
	"os"
	"strings"
)

// currentUserName returns the current user name in DOMAIN\USER format
func currentUserName() (string, error) {
	// See https://github.com/kubernetes/klog/blob/main/klog_file_windows.go
	// On Windows, the Go 'user' package requires netapi32.dll.
	// This affects Windows Nano Server: https://github.com/golang/go/issues/21867
	// Fallback to using environment variables.
	u := os.Getenv("USERNAME")
	if len(u) == 0 {
		return "(unknown)", nil
	}
	// Sanitize the USERNAME since it may contain filepath separators.
	u = strings.Replace(u, `\`, "_", -1)

	d := os.Getenv("USERDOMAIN")
	if len(d) != 0 {
		return d + "\\" + u, nil
	}
	return u, nil
}
