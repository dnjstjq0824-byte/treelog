#!/usr/bin/env bash
# TreeBlog Point-in-Time Recovery (PITR) & Snapshot Restoration Script
set -euo pipefail

TARGET_TIMESTAMP="${1:-}"
BASE_BACKUP_FILE="${2:-}"

if [ -z "${BASE_BACKUP_FILE}" ]; then
    echo "Usage: $0 [TARGET_TIMESTAMP_ISO] [BASE_BACKUP_FILE]"
    echo "Example: $0 '2026-10-31 12:00:00 UTC' /var/backups/treeblog/treeblog_dump_20261031_000000Z.sql.zst"
    exit 1
fi

echo "[INFO] Restoring base backup from ${BASE_BACKUP_FILE}..."
zstd -d -c "${BASE_BACKUP_FILE}" | pg_restore \
    -h "${TREEBLOG_DB_HOST:-localhost}" \
    -p "${TREEBLOG_DB_PORT:-5432}" \
    -U "${TREEBLOG_DB_USER:-treeblog}" \
    -d "${TREEBLOG_DB_NAME:-treeblog}" \
    --clean --if-exists -v

if [ -n "${TARGET_TIMESTAMP}" ]; then
    echo "[INFO] Setting recovery target time: ${TARGET_TIMESTAMP} in recovery configuration..."
    # Recovery target configuration for PostgreSQL WAL replay
    echo "restore_command = 'cp /var/lib/postgresql/wal_archive/%f %p'" > /var/lib/postgresql/data/recovery.signal
    echo "recovery_target_time = '${TARGET_TIMESTAMP}'" >> /var/lib/postgresql/data/postgresql.auto.conf
    echo "recovery_target_action = 'promote'" >> /var/lib/postgresql/data/postgresql.auto.conf
    echo "[INFO] PostgreSQL ready for PITR replay."
fi
