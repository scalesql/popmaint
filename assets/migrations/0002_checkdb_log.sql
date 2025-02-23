-- +goose Up
SET ANSI_NULLS ON;
SET QUOTED_IDENTIFIER ON;

CREATE TABLE [dbo].[checkdb_log](
    log_id BIGINT IDENTITY(1,1) NOT NULL,
    job_id varchar(128) NOT NULL, 
    plan_name [nvarchar](1024) NOT NULL,
    [domain_name] [nvarchar](128) NOT NULL,
    [server_name] [nvarchar](128) NOT NULL,
    [database_name] [nvarchar](128) NOT NULL,
    [completed_at] [datetimeoffset](0) NOT NULL,
    [duration_sec] [int] NOT NULL,
    data_mb BIGINT NOT NULL,
    CONSTRAINT pk_checkdb_log PRIMARY KEY CLUSTERED (log_id)
) ON [PRIMARY];

SET ANSI_PADDING ON;

CREATE NONCLUSTERED INDEX [ix_checkdb_log_job_id] ON [dbo].[checkdb_log] (job_id ASC)
WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, SORT_IN_TEMPDB = OFF,
DROP_EXISTING = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON) ON [PRIMARY];

CREATE NONCLUSTERED INDEX [ix_checkdb_log_domain_completed] ON [dbo].[checkdb_log]
(
	[domain_name] ASC,
	[completed_at] ASC
)WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, SORT_IN_TEMPDB = OFF, DROP_EXISTING = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON) ON [PRIMARY];

CREATE NONCLUSTERED INDEX [ix_checkdb_log_server_completed] ON [dbo].[checkdb_log]
(
	[server_name] ASC,
	[completed_at] ASC
)WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, SORT_IN_TEMPDB = OFF, DROP_EXISTING = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON) ON [PRIMARY];

CREATE NONCLUSTERED INDEX [ix_checkdb_log_database_completed] ON [dbo].[checkdb_log]
(
	[database_name] ASC,
	[completed_at] ASC
)WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, SORT_IN_TEMPDB = OFF, DROP_EXISTING = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON) ON [PRIMARY];

CREATE NONCLUSTERED INDEX [ix_checkdb_log_completed] ON [dbo].[checkdb_log]
(	[completed_at] ASC )
WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, SORT_IN_TEMPDB = OFF, DROP_EXISTING = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON) ON [PRIMARY];

-- +goose Down
DROP TABLE [dbo].[checkdb_log];