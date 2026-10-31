#!/usr/bin/env bash
# TreeBlog Database Automated Backup Script
# Performs pg_dump with zstd compression and uploads to S3/GCS backup bucket
set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-/var/backups/treeblog}"
TIMESTAMP=$(date -u +"%Y%m%d_%H%M%SZ")
BACKUP_FILE="${BACKUP_DIR}/treeblog_dump_${TIMESTAMP}.sql.zst"
S3_BUCKET="${TREEBLOG_BACKUP_S3_BUCKET:-s3://treeblog-backups/postgres}"

mkdir -p "${BACKUP_DIR}"

echo "[INFO] Starting database backup at ${TIMESTAMP}..."

PGPASSWORD="${TREEBLOG_DB_PASSWORD:-postgres}" pg_dump \
    -h "${TREEBLOG_DB_HOST:-localhost}" \
    -p "${TREEBLOG_DB_PORT:-5432}" \
    -U "${TREEBLOG_DB_USER:-treeblog}" \
    -d "${TREEBLOG_DB_NAME:-treeblog}" \
    -F c -b -v \
    | zstd -T0 -3 > "${BACKUP_FILE}"

echo "[INFO] Backup created: ${BACKUP_FILE} ($(du -h "${BACKUP_FILE}" | cut -f1))"

if command -v aws &> /dev/null && [ -n "${TREEBLOG_BACKUP_S3_BUCKET:-}" ]; then
    echo "[INFO] Uploading backup to ${S3_BUCKET}..."
    aws s3 cp "${BACKUP_FILE}" "${S3_BUCKET}/" --storage-class STANDARD_IA
    echo "[INFO] Upload completed."
fi

# Retention policy: remove backups older than 14 days locally
find "${BACKUP_DIR}" -type f -name "treeblog_dump_*.sql.zst" -mtime +14 -delete
echo "[INFO] Local cleanup finished."
