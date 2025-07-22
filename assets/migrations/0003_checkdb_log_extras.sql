-- +goose Up
SET ANSI_NULLS ON;
SET QUOTED_IDENTIFIER ON;
SET XACT_ABORT ON;

ALTER TABLE dbo.checkdb_log ADD [server_cores]    INT NULL;
ALTER TABLE dbo.checkdb_log ADD [server_maxdop]   INT NULL;
ALTER TABLE dbo.checkdb_log ADD [stmt_maxdop]     INT NULL;

ALTER TABLE dbo.checkdb_log ADD [no_index]                BIT NULL;
ALTER TABLE dbo.checkdb_log ADD [physical_only]           BIT NULL;
ALTER TABLE dbo.checkdb_log ADD [extended_logical_checks] BIT NULL;
ALTER TABLE dbo.checkdb_log ADD [data_purity]             BIT NULL;


-- +goose Down
SET ANSI_NULLS ON;
SET QUOTED_IDENTIFIER ON;
SET XACT_ABORT ON;

ALTER TABLE dbo.checkdb_log DROP COLUMN [server_cores];
ALTER TABLE dbo.checkdb_log DROP COLUMN [server_maxdop];
ALTER TABLE dbo.checkdb_log DROP COLUMN [stmt_maxdop];

ALTER TABLE dbo.checkdb_log DROP COLUMN [no_index];
ALTER TABLE dbo.checkdb_log DROP COLUMN [physical_only];
ALTER TABLE dbo.checkdb_log DROP COLUMN [extended_logical_checks];
ALTER TABLE dbo.checkdb_log DROP COLUMN [data_purity];
