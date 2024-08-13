package slogx

import (
	"log/slog"
	"strings"
)

/*
TODO func Error(err error)
TODO func Error(logger, err, []attr|[]any)
NestedAttr("popmaint.target.fqdn", value)
NestedGroup("popmaint.target", []any)
*/

func GroupAttrs(group string, vals []any) []slog.Attr {
	_ = strings.Split(group, ".")
	return []slog.Attr{}
}
