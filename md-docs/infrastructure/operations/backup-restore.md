# Backup and restore runbook

This runbook becomes active only after self-hosted PostgreSQL is approved.

WAL-G archives WAL continuously and the systemd timer creates a full backup
daily, retaining seven full backups by default. Store credentials only in
`/etc/mantrixflow/backup.env` with mode `0600`. Alert on archive failures,
missing daily backups, retention failures, and bucket access errors.

Quarterly restore drill:

1. Create a disposable isolated database target.
2. Set the backup environment and an empty directory below
   `/var/lib/mantrixflow-restore/`.
3. Run `restore-latest` with the exact confirmation variable.
4. Start PostgreSQL on a non-production port against the restored directory.
5. Run schema, PGMQ, row-count, checkpoint, and application smoke checks.
6. Record recovery point and recovery time, then destroy the disposable target.

Never test restore over the production data directory. VPS snapshots do not
replace WAL-G.
