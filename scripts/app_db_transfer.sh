#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="${ENV_FILE:-${ROOT_DIR}/.env}"
COMPOSE_FILE="${COMPOSE_FILE:-${ROOT_DIR}/docker-compose.yml}"
COMPOSE_PROJECT="${COMPOSE_PROJECT:-shushu-internal}"

if [ -f "${ENV_FILE}" ]; then
  set -a
  # shellcheck disable=SC1090
  . "${ENV_FILE}"
  set +a
fi

MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-changeme}"
MYSQL_DATABASE="${MYSQL_DATABASE:-shushu_photo}"

SOURCE_MYSQL_HOST="${SOURCE_MYSQL_HOST:-127.0.0.1}"
SOURCE_MYSQL_PORT="${SOURCE_MYSQL_PORT:-3306}"
SOURCE_MYSQL_USER="${SOURCE_MYSQL_USER:-root}"
SOURCE_MYSQL_PASSWORD="${SOURCE_MYSQL_PASSWORD:-${MYSQL_ROOT_PASSWORD}}"
SOURCE_MYSQL_DB="${SOURCE_MYSQL_DB:-${MYSQL_DATABASE}}"

ACTION="${1:-}"
DUMP_FILE="${2:-${ROOT_DIR}/app_db_only.sql}"
TRUNCATE_SQL="${ROOT_DIR}/app_db_truncate.sql"

usage() {
  cat <<EOF
Usage:
  $(basename "$0") export [dump.sql]
  $(basename "$0") import [dump.sql]
  $(basename "$0") clean

Env overrides:
  ENV_FILE, COMPOSE_FILE, COMPOSE_PROJECT
  SOURCE_MYSQL_HOST, SOURCE_MYSQL_PORT, SOURCE_MYSQL_USER, SOURCE_MYSQL_PASSWORD, SOURCE_MYSQL_DB
  MYSQL_ROOT_PASSWORD, MYSQL_DATABASE
EOF
}

export_data() {
  local tables
  tables="$(mysql -h "${SOURCE_MYSQL_HOST}" -P "${SOURCE_MYSQL_PORT}" -u "${SOURCE_MYSQL_USER}" -p"${SOURCE_MYSQL_PASSWORD}" -N -B \
    -e "SELECT table_name FROM information_schema.tables WHERE table_schema='${SOURCE_MYSQL_DB}' AND table_name LIKE 'app_db_%';")"

  if [ -z "${tables}" ]; then
    echo "No app_db_ tables found in ${SOURCE_MYSQL_DB}" >&2
    exit 1
  fi

  printf '%s\n' ${tables} | xargs mysqldump -h "${SOURCE_MYSQL_HOST}" -P "${SOURCE_MYSQL_PORT}" -u "${SOURCE_MYSQL_USER}" -p"${SOURCE_MYSQL_PASSWORD}" \
    --single-transaction --quick --set-gtid-purged=OFF --no-create-info "${SOURCE_MYSQL_DB}" > "${DUMP_FILE}"

  echo "Exported app_db_ tables to ${DUMP_FILE}"
}

clean_data() {
  local statements
  local sql
  sql="SELECT CONCAT('TRUNCATE TABLE ', CHAR(96), table_name, CHAR(96), ';') FROM information_schema.tables WHERE table_schema='${MYSQL_DATABASE}' AND table_name LIKE 'app_db_%';"
  statements="$(docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}" -p "${COMPOSE_PROJECT}" exec -T mysql \
    mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" -N -B -e "${sql}" "${MYSQL_DATABASE}")"

  if [ -z "${statements}" ]; then
    echo "No app_db_ tables found in ${SOURCE_MYSQL_DB}" >&2
    exit 1
  fi

  printf '%s\n' "${statements}" > "${TRUNCATE_SQL}"

  docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}" -p "${COMPOSE_PROJECT}" exec -T mysql \
    mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" "${MYSQL_DATABASE}" < "${TRUNCATE_SQL}"

  echo "Truncated app_db_ tables in ${MYSQL_DATABASE}"
}

import_data() {
  if [ ! -f "${DUMP_FILE}" ]; then
    echo "Dump file not found: ${DUMP_FILE}" >&2
    exit 1
  fi

  docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}" -p "${COMPOSE_PROJECT}" exec -T mysql \
    mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" "${MYSQL_DATABASE}" < "${DUMP_FILE}"

  echo "Imported app_db_ tables into ${MYSQL_DATABASE}"
}

case "${ACTION}" in
  export)
    export_data
    ;;
  clean)
    clean_data
    ;;
  import)
    import_data
    ;;
  *)
    usage
    exit 1
    ;;
esac
