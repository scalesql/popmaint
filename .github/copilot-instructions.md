# Copilot Instructions for PopMaint

## Language & Framework
- Go 1.21+
- SQL Server (via `github.com/microsoft/go-mssqldb`)
- TOML configuration (via `pelletier/go-toml/v2`)

## Code Standards
- Use `sql.Named` parameters for all SQL queries (never string concatenation)
- All exported functions must have doc comments
- Error messages should be lowercase and not end with punctuation
- Wrap errors with `fmt.Errorf("context: %w", err)`
- Use `context.Context` as the first parameter for any function that does I/O

## Naming Conventions
- Use camelCase for unexported identifiers
- Use PascalCase for exported identifiers
- SQL column names use snake_case
- Struct tags: use `db:"column_name"` for database mapping

## Project Structure
- `internal/app/` - Main application logic
- `internal/config/` - Configuration parsing
- `internal/state/` - Database state repository
- `internal/mssqlz/` - SQL Server helpers
- `assets/migrations/` - Goose SQL migrations

## SQL Migrations
- Use goose format with `-- +goose Up` and `-- +goose Down` markers
- Always include a down migration
- Table/index names use snake_case with prefixes (e.g., `ix_`, `pk_`)

## Testing
- Table-driven tests preferred
- Test files next to source files (`_test.go` suffix)
- Tests should use github.com/stretchr/testify/assert
- Tests should prefer `assert := assert.New(t)` format

## Security
- Never log passwords or connection strings
- Use environment variable substitution for secrets
- Use `sql.Named` parameters to prevent SQL injection

## Documentation
- The documentation to use the application is at `docs/DOCS.md`

## Logging 
- Logging uses a custom framework defined in `internal/lx`
