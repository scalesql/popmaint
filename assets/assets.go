package assets

import "embed"

//go:embed migrations/*.sql
var DBMigrationsFS embed.FS
