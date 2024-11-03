package mssqlz

import "strings"

// DupeCheck is used for find duplicate Servers
type DupeCheck struct {
	servers map[string]bool
}

func NewDupeCheck() DupeCheck {
	return DupeCheck{
		servers: make(map[string]bool),
	}
}

// IsDupe indicates whether we have seen this server before.
// It is a case-insenstive comparison.
func (dc *DupeCheck) IsDupe(srv Server) bool {
	path := strings.ToLower(srv.Path())
	_, ok := dc.servers[path]
	if ok {
		return true
	} else {
		dc.servers[path] = true
		return false
	}
}
