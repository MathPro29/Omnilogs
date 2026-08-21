# Retention backup flow

## Scheduling and time boundaries

- Each retention policy stores a five-field `cron_schedule` and an IANA
  `schedule_timezone` (default `Asia/Bangkok`).
- `next_purge_at` is persisted as UTC (`timestamptz`), while the cron is
  evaluated in the policy timezone. For example, `0 1 * * *` in
  `Asia/Bangkok` runs at 01:00 local time (18:00 UTC on the previous date).
- Archive ranges are half-open: `[date_from, date_to)`. A backup for August
  16-23 therefore sends `date_from=2026-08-16T00:00:00+07:00` and
  `date_to=2026-08-24T00:00:00+07:00`.

## Policy and manual backups

- The worker creates `POLICY` backups.
- `POST /products/{productId}/log-archives` creates a `MANUAL` backup by
  default and requires an explicit range. The UI also sends
  `backup_type=MANUAL`.
- Both modes run the same export, checksum, optional R2 upload, verification,
  and restore pipeline.
- Every archive records `backup_type`, `backup_tag`, and a unique daily
  `coverage_key`. Coverage does not include backup type, so a policy run
  reuses a valid manual backup for the same product/environment/scope/day.

## Cloudflare R2

Select `R2` as the policy storage provider and configure:

```text
R2_ENDPOINT
R2_REGION=auto
R2_ACCESS_KEY_ID
R2_SECRET_ACCESS_KEY
R2_BUCKET
R2_OBJECT_PREFIX
```

The database stores the R2 bucket/object key. Verify, ZIP download, expiry
packaging, and rollback materialize the object from R2 when needed.

## Rollback lookup

Archive listing accepts `date_from` and `date_to` filters. The UI sends the
selected day, week, month, or year as an offset-aware half-open range, then
restores only the returned verified archives.

## Worker split

- `omnilogs-worker` is responsible for ingestion only.
- `omnilogs-retention-worker` owns archive/verify/purge execution. It reads
  the persisted `next_purge_at` handoff and sleeps until shortly before the
  next deadline (`RETENTION_WORKER_LEAD_SECONDS`, default 30 seconds).
- The retention worker polls only the nearest deadline every
  `RETENTION_WORKER_POLL_SECONDS` (default 30 seconds), so it does not scan or
  run the full retention flow continuously.
- Both workers can restart safely because the handoff is in PostgreSQL, not
  process memory. Run one retention-worker replica per database unless a
  database-level lease is added for multi-replica deployment.
