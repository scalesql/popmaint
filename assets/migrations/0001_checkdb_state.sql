-- +goose Up
SET ANSI_NULLS ON;
SET QUOTED_IDENTIFIER ON;

CREATE TABLE [dbo].[checkdb_state](
	[domain_name] [nvarchar](128) NOT NULL,
	[server_name] [nvarchar](128) NOT NULL,
	[database_name] [nvarchar](128) NOT NULL,
	[last_checkdb] [datetimeoffset](7) NOT NULL,
 CONSTRAINT [PK_checkdb_state] PRIMARY KEY CLUSTERED 
(
	[domain_name] ASC,
	[server_name] ASC,
	[database_name] ASC
)WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, IGNORE_DUP_KEY = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON) ON [PRIMARY]
) ON [PRIMARY];

-- +goose Down
DROP TABLE [dbo].[checkdb_state];

