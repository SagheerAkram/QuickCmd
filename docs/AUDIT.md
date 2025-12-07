# Audit Logging System

## Overview

QuickCMD maintains a complete audit trail of all command executions in an SQLite database, providing accountability, traceability, and the ability to review past actions.

## Database Schema

### Runs Table

```sql
CREATE TABLE IF NOT EXISTS runs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TEXT NOT NULL,
    user TEXT,
    prompt TEXT NOT NULL,
    selected_command TEXT NOT NULL,
    sandbox_id TEXT,
    exit_code INTEGER,
    stdout BLOB,
    stderr BLOB,
    risk_level TEXT NOT NULL,
    snapshot TEXT,
    executed BOOLEAN DEFAULT 0,
    duration_ms INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Indexes

```sql
CREATE INDEX idx_runs_timestamp ON runs(timestamp);
CREATE INDEX idx_runs_user ON runs(user);
CREATE INDEX idx_runs_executed ON runs(executed);
CREATE INDEX idx_runs_risk_level ON runs(risk_level);
```

## Audit Record Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | INTEGER | Auto-incrementing primary key |
| `timestamp` | TEXT | ISO 8601 timestamp (RFC3339) |
| `user` | TEXT | Username who executed the command |
| `prompt` | TEXT | Original natural language prompt |
| `selected_command` | TEXT | The actual shell command executed |
| `sandbox_id` | TEXT | Docker container ID (short form) |
| `exit_code` | INTEGER | Command exit code (0 = success) |
| `stdout` | BLOB | Standard output (redacted) |
| `stderr` | BLOB | Standard error (redacted) |
| `risk_level` | TEXT | Risk classification (safe/medium/high) |
| `snapshot` | TEXT | JSON-encoded snapshot metadata |
| `executed` | BOOLEAN | Whether command was actually executed |
| `duration_ms` | INTEGER | Execution duration in milliseconds |
| `created_at` | DATETIME | Database insertion timestamp |

## Secrets Redaction

### Automatic Redaction

All sensitive data is automatically redacted before storage:

#### Environment Variables

```bash
# Original
PASSWORD=secret123 ./deploy.sh

# Redacted
PASSWORD=***REDACTED*** ./deploy.sh
```

#### API Keys and Tokens

```bash
# Original
curl -H "Authorization: Bearer eyJhbGc..."

# Redacted
curl -H "Authorization: ***REDACTED***"
```

#### Patterns Redacted

- `PASSWORD`, `PASSWD`, `PWD`
- `TOKEN`, `AUTH_TOKEN`, `ACCESS_TOKEN`
- `API_KEY`, `APIKEY`
- `SECRET`, `SECRET_KEY`
- `Bearer` tokens
- `Basic` auth credentials
- AWS credentials
- Private keys
- Database connection strings

### Custom Redaction Patterns

Add custom patterns to the redactor:

```go
redactor := policy.NewSecretRedactor()
redactor.AddPattern(`my_secret_pattern`)
```

## Snapshot Metadata

Snapshots are stored as JSON in the `snapshot` field:

```json
{
  "type": "git",
  "location": "quickcmd/backup/20250107-093000",
  "timestamp": "2025-01-07T09:30:00Z",
  "reversible": true,
  "restore_cmd": "git checkout quickcmd/backup/20250107-093000"
}
```

### Snapshot Types

#### Git Snapshot

```json
{
  "type": "git",
  "location": "quickcmd/backup/<timestamp>",
  "reversible": true,
  "restore_cmd": "git checkout <branch>"
}
```

#### Filesystem Snapshot

```json
{
  "type": "filesystem",
  "location": "/tmp/quickcmd/backups/snapshot-<timestamp>",
  "affected_paths": ["file1.txt", "file2.txt"],
  "reversible": true,
  "restore_cmd": "cp -r <backup>/* ./"
}
```

## Database Location

### Default Path

```
~/.quickcmd/audit.db
```

**Windows:** `C:\Users\<username>\.quickcmd\audit.db`
**macOS/Linux:** `/home/<username>/.quickcmd/audit.db`

### Custom Path

Set via environment variable:

```bash
export QUICKCMD_AUDIT_DB=/path/to/custom/audit.db
```

## Querying History

### CLI Commands

#### View Recent History

```bash
quickcmd history
```

Shows last 20 executions.

#### Limit Results

```bash
quickcmd history -n 50
```

Shows last 50 executions.

#### Filter by Keyword

```bash
quickcmd history -f "docker"
```

Shows only executions containing "docker".

#### View Statistics

```bash
quickcmd history --stats
```

Shows aggregate statistics:
- Total executions
- Success rate
- Breakdown by risk level

### Programmatic Access

```go
store, _ := audit.NewSQLiteStore("~/.quickcmd/audit.db")
defer store.Close()

// Get history
records, _ := store.GetHistory(20, "")

// Get specific record
record, _ := store.GetRecordByID(123)

// Get statistics
stats, _ := store.GetStats()
```

## Audit Log Security

### Write-Ahead Logging (WAL)

The database uses WAL mode for better concurrency and crash recovery:

```sql
PRAGMA journal_mode=WAL;
```

**Benefits:**
- Concurrent reads during writes
- Better crash recovery
- Improved performance

### Append-Only Design

Records are never updated or deleted, only inserted:

```go
// Insert only - no UPDATE or DELETE
_, err := s.db.Exec("INSERT INTO runs (...) VALUES (...)")
```

**Benefits:**
- Tamper-evident
- Complete history
- Audit compliance

### File Permissions

The audit database should have restricted permissions:

```bash
chmod 600 ~/.quickcmd/audit.db
```

Only the owner can read/write.

## Data Retention

### Manual Cleanup

Currently, records are retained indefinitely. Future versions will support:

```yaml
# config.yaml
audit:
  retention_days: 90
  archive_location: "/var/log/quickcmd/archive"
```

### Backup

Backup the audit database regularly:

```bash
# SQLite backup
sqlite3 ~/.quickcmd/audit.db ".backup /path/to/backup.db"

# Or simple copy
cp ~/.quickcmd/audit.db /path/to/backup.db
```

## Statistics and Reporting

### Available Statistics

```go
stats := {
    "total_executions": 1234,
    "success_rate": 87.5,
    "by_risk_level": {
        "safe": 800,
        "medium": 300,
        "high": 134
    }
}
```

### Future Analytics

Planned features:
- Most used commands
- Failure patterns
- Execution time trends
- User activity reports
- Risk level distribution over time

## Compliance and Governance

### Audit Requirements

QuickCMD audit logs satisfy common compliance requirements:

- ✅ **Who:** User attribution
- ✅ **What:** Command executed
- ✅ **When:** Timestamp
- ✅ **Where:** Sandbox ID
- ✅ **Result:** Exit code and output
- ✅ **Integrity:** Append-only, tamper-evident

### Export for Compliance

Export audit logs to standard formats:

```bash
# Export to JSON
sqlite3 ~/.quickcmd/audit.db \
  "SELECT json_object('id', id, 'timestamp', timestamp, ...) FROM runs" \
  > audit_export.json

# Export to CSV
sqlite3 -header -csv ~/.quickcmd/audit.db \
  "SELECT * FROM runs" \
  > audit_export.csv
```

## Troubleshooting

### Database Locked

**Error:**
```
database is locked
```

**Solution:**
- Close other QuickCMD instances
- Check for stale lock files
- Use WAL mode (enabled by default)

### Disk Space

Monitor database size:

```bash
du -h ~/.quickcmd/audit.db
```

Large databases (>100MB) may need cleanup or archival.

### Corrupted Database

**Recovery:**
```bash
# Dump to SQL
sqlite3 ~/.quickcmd/audit.db .dump > backup.sql

# Recreate database
rm ~/.quickcmd/audit.db
sqlite3 ~/.quickcmd/audit.db < backup.sql
```

## Privacy Considerations

### Sensitive Data

Even with redaction, audit logs may contain:
- File paths
- Command arguments
- Output data

**Recommendations:**
- Restrict database file permissions
- Encrypt the database file system
- Implement log rotation and archival
- Review logs before sharing

### GDPR Compliance

For GDPR compliance:
- Implement data retention policies
- Provide data export functionality
- Support data deletion requests
- Document data processing

## Future Enhancements

### Planned Features

- **Encryption at rest:** AES-256 encryption for audit database
- **Remote logging:** Send logs to centralized server
- **Log rotation:** Automatic archival of old logs
- **Advanced search:** Full-text search across all fields
- **Alerting:** Notifications for high-risk executions
- **Compliance reports:** Pre-built reports for auditors
- **Undo automation:** One-click restore from snapshots

### Integration Points

- **SIEM systems:** Export to Splunk, ELK, etc.
- **Slack/Teams:** Notifications for executions
- **Webhooks:** Real-time event streaming
- **Prometheus:** Metrics export

## Best Practices

1. **Regular backups:** Backup audit database weekly
2. **Monitor size:** Set up alerts for large databases
3. **Review logs:** Periodically review for anomalies
4. **Restrict access:** Limit who can read audit logs
5. **Test recovery:** Verify backup restoration works
6. **Document retention:** Define and enforce retention policies
7. **Encrypt storage:** Use encrypted filesystems
8. **Audit the auditors:** Log access to audit logs

## Example Queries

### Find Failed Commands

```sql
SELECT timestamp, selected_command, exit_code
FROM runs
WHERE executed = 1 AND exit_code != 0
ORDER BY timestamp DESC
LIMIT 10;
```

### High-Risk Executions

```sql
SELECT timestamp, user, selected_command, exit_code
FROM runs
WHERE risk_level = 'high' AND executed = 1
ORDER BY timestamp DESC;
```

### Commands by User

```sql
SELECT user, COUNT(*) as count
FROM runs
WHERE executed = 1
GROUP BY user
ORDER BY count DESC;
```

### Average Execution Time

```sql
SELECT AVG(duration_ms) as avg_duration_ms
FROM runs
WHERE executed = 1 AND duration_ms > 0;
```

---

**Remember: Audit logs are only useful if they're reviewed. Set up regular reviews and automated alerts for suspicious activity.**
