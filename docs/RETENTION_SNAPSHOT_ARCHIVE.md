# Retention archive with Elasticsearch Snapshot

The retention lifecycle now supports two archive artifacts for each daily
range:

1. Elasticsearch Snapshot is the primary restore source.
2. NDJSON.GZ is kept as a portable export and can be downloaded as ZIP.

The backend creates the daily NDJSON export first, creates a snapshot for the
same daily Elasticsearch indices, and only exposes the archive for Verify after
the snapshot completes successfully. Verify checks the portable file and the
snapshot repository before `delete_active_after_archive` is allowed to delete
active logs.

## Elasticsearch setup

Register a repository in Elasticsearch before enabling the application. The
repository storage should be outside the Elasticsearch data path and should be
shared by all nodes. The exact repository body depends on the storage backend
(S3, shared filesystem, MinIO, and so on).

Example filesystem repository for a single-node development environment:

```http
PUT /_snapshot/omnilogs-archive
{
  "type": "fs",
  "settings": {
    "location": "/usr/share/elasticsearch/snapshots"
  }
}
```

Verify it from Elasticsearch before starting the worker:

```http
POST /_snapshot/omnilogs-archive/_verify
```

The repository must be mounted/configured for Elasticsearch itself. Do not
point it at the application `data` directory.

## Application environment

Set these variables in `backend/api/.env` for API and worker processes:

```env
ELASTICSEARCH_SNAPSHOT_ENABLED=true
ELASTICSEARCH_SNAPSHOT_REPOSITORY=omnilogs-archive
ELASTICSEARCH_SNAPSHOT_TIMEOUT_SECONDS=1800
```

Restart both API and worker after changing the environment. Database startup
migration adds the snapshot metadata columns to `log_archives`.

## Restore flow

`POST /api/v1/products/:productId/log-archives/:archiveId/restore` with:

```json
{
  "mode": "SNAPSHOT_SEARCH_ONLY"
}
```

restores the selected snapshot into a read-only temporary index pattern. The
response contains `index_name` and `archive_id`. Opening
`/logs-explorer?product=<productId>&environment_id=<environmentId>&archive_id=<archiveId>`
uses that restored pattern and does not search active indices.

`RESTORE_TO_ACTIVE` remains available for rollback and continues to use the
portable NDJSON file so it can safely skip existing documents.

When portable files reach archive retention, the worker removes only the
NDJSON file and changes the record to `SNAPSHOT_ONLY`; the Elasticsearch
snapshot metadata remains available for restore.

## Portable ZIP import

Snapshot is optional. The portable flow can be used by itself with
`ELASTICSEARCH_SNAPSHOT_ENABLED=false`.

The generated ZIP keeps each compressed file in its scope path:

```text
product-77/environment-137/category-10/feature-12/sub-feature-15/2026/08/18.ndjson.gz
```

The ZIP also contains `manifest.json`. Import validates product, environment,
feature scope, sub-feature scope, date range, checksum, and document count
before copying the archive into the local archive storage.

```http
POST /api/v1/products/77/log-archives/import?environment_id=137
Content-Type: multipart/form-data
file=<portable-archive.zip>
```

An optional `category_id`, `feature_id`, and `sub_feature_id` query parameter
restricts the import to one scope. Imported archives become `VERIFIED` but do
not delete active logs automatically. Use the existing Restore action when
you want to restore them into Active Elasticsearch.
