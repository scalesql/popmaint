# Extras

## Script to manually clean up history
```sql
SELECT DATEDIFF(d, MIN(sent_date), GETDATE()) AS [mail_days], 
    count(*) AS [mail_rows] 
FROM msdb.dbo.sysmail_mailitems WITH (NOLOCK)

SELECT DATEDIFF(d, MIN(log_date), GETDATE()) AS [mail_log_days], 
    count(*) AS [mail_log_rows] 
from msdb.dbo.sysmail_log WITH(NOLOCK) 

SELECT DATEDIFF(d, MIN(backup_start_date), GETDATE()) AS [backup_days], 
    count(*) AS [backup_rows]
FROM msdb.dbo.backupset bs
JOIN msdb.dbo.backupfile bf ON bf.backup_set_id = bs.backup_set_id

SELECT DATEDIFF(d, MIN(msdb.dbo.agent_datetime(run_date, run_time)), GETDATE()) AS [job_days], 
    count(*) AS [job_rows] 
FROM msdb.dbo.sysjobhistory

/*

DECLARE @limit DATE;
SET @limit = DATEADD(dd, -365, GETDATE()); 
EXEC msdb.dbo.sysmail_delete_log_sp @logged_before = @limit;
EXEC msdb.dbo.sysmail_delete_mailitems_sp	@sent_before = @limit

SET @limit = DATEADD(dd, -90, GETDATE()); 
EXEC msdb.dbo.sp_delete_backuphistory @oldest_date = @limit;

SET @limit =  DATEADD(dd, -90, GETDATE()); 
EXEC msdb.dbo.sp_purge_jobhistory @oldest_date = @limit;

*/
```

