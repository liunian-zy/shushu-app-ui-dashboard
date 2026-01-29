package handlers

import (
  "database/sql"
  "encoding/json"
  "net/http"
  "strings"

  "github.com/gin-gonic/gin"
  "github.com/redis/go-redis/v9"

  "shushu-app-ui-dashboard/internal/config"
  "shushu-app-ui-dashboard/internal/services"
)

type HistoryHandler struct {
  cfg   *config.Config
  db    *sql.DB
  redis *redis.Client
}

// NewHistoryHandler creates a handler for audit/history queries.
// Args:
//   cfg: App config instance.
//   db: Database connection.
//   redis: Redis client for OSS cache.
// Returns:
//   *HistoryHandler: Initialized handler.
func NewHistoryHandler(cfg *config.Config, db *sql.DB, redis *redis.Client) *HistoryHandler {
  return &HistoryHandler{cfg: cfg, db: db, redis: redis}
}

// ListAuditLogs returns audit log entries.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *HistoryHandler) ListAuditLogs(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  draftID := parseInt64Query(c, "draft_version_id")
  entityTable := strings.TrimSpace(c.Query("entity_table"))
  entityID := parseInt64Query(c, "entity_id")
  action := strings.TrimSpace(c.Query("action"))
  limit, offset := parsePagination(c)

  query := `SELECT l.id, l.draft_version_id, l.entity_table, l.entity_id, l.action, l.actor_id, l.detail_json, l.created_at,
    u.display_name, u.username
    FROM app_db_audit_logs l
    LEFT JOIN app_db_users u ON u.id = l.actor_id
    WHERE 1=1`
  args := make([]interface{}, 0)

  if draftID > 0 {
    query += " AND l.draft_version_id = ?"
    args = append(args, draftID)
  }
  if entityTable != "" {
    query += " AND l.entity_table = ?"
    args = append(args, entityTable)
  }
  if entityID > 0 {
    query += " AND l.entity_id = ?"
    args = append(args, entityID)
  }
  if action != "" {
    query += " AND l.action = ?"
    args = append(args, action)
  }

  query += " ORDER BY l.id DESC LIMIT ? OFFSET ?"
  args = append(args, limit, offset)

  rows, err := h.db.Query(query, args...)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }
  defer rows.Close()

  items := make([]gin.H, 0)
  for rows.Next() {
    var (
      id            int64
      draftVersion  sql.NullInt64
      entityTableV  sql.NullString
      entityIDV     sql.NullInt64
      actionV       sql.NullString
      actorID       sql.NullInt64
      detailJSON    sql.NullString
      createdAt     sql.NullTime
      displayName   sql.NullString
      username      sql.NullString
    )

    if err := rows.Scan(
      &id,
      &draftVersion,
      &entityTableV,
      &entityIDV,
      &actionV,
      &actorID,
      &detailJSON,
      &createdAt,
      &displayName,
      &username,
    ); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed"})
      return
    }

    items = append(items, gin.H{
      "id":               id,
      "draft_version_id": nullableInt64Pointer(draftVersion),
      "entity_table":     nullableStringValue(entityTableV),
      "entity_id":        nullableInt64Pointer(entityIDV),
      "action":           nullableStringValue(actionV),
      "actor_id":         nullableInt64Pointer(actorID),
      "actor_name":       nullableStringValue(displayName),
      "actor_username":   nullableStringValue(username),
      "detail":           decodeJSON(detailJSON),
      "created_at":       nullableTimePointer(createdAt),
    })
  }

  c.JSON(http.StatusOK, gin.H{"data": items})
}

// ListFieldHistory returns field-level history records.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *HistoryHandler) ListFieldHistory(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  draftID := parseInt64Query(c, "draft_version_id")
  entityTable := strings.TrimSpace(c.Query("entity_table"))
  entityID := parseInt64Query(c, "entity_id")
  fieldName := strings.TrimSpace(c.Query("field_name"))
  if draftID <= 0 || entityTable == "" {
    c.JSON(http.StatusBadRequest, gin.H{"error": "draft_version_id and entity_table are required"})
    return
  }

  limit, offset := parsePagination(c)

  query := `SELECT h.id, h.draft_version_id, h.entity_table, h.entity_id, h.field_name, h.old_value, h.new_value,
    h.submit_id, h.changed_by, h.created_at, u.display_name, u.username
    FROM app_db_field_history h
    LEFT JOIN app_db_users u ON u.id = h.changed_by
    WHERE h.draft_version_id = ? AND h.entity_table = ?`
  args := []interface{}{draftID, entityTable}

  if entityID > 0 {
    query += " AND h.entity_id = ?"
    args = append(args, entityID)
  }
  if fieldName != "" {
    query += " AND h.field_name = ?"
    args = append(args, fieldName)
  }

  query += " ORDER BY h.id DESC LIMIT ? OFFSET ?"
  args = append(args, limit, offset)

  rows, err := h.db.Query(query, args...)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }
  defer rows.Close()

  items := make([]gin.H, 0)
  for rows.Next() {
    var (
      id            int64
      draftVersion  sql.NullInt64
      entityTableV  sql.NullString
      entityIDV     sql.NullInt64
      fieldNameV    sql.NullString
      oldValue      sql.NullString
      newValue      sql.NullString
      submitID      sql.NullInt64
      changedBy     sql.NullInt64
      createdAt     sql.NullTime
      displayName   sql.NullString
      username      sql.NullString
    )

    if err := rows.Scan(
      &id,
      &draftVersion,
      &entityTableV,
      &entityIDV,
      &fieldNameV,
      &oldValue,
      &newValue,
      &submitID,
      &changedBy,
      &createdAt,
      &displayName,
      &username,
    ); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed"})
      return
    }

    items = append(items, gin.H{
      "id":               id,
      "draft_version_id": nullableInt64Pointer(draftVersion),
      "entity_table":     nullableStringValue(entityTableV),
      "entity_id":        nullableInt64Pointer(entityIDV),
      "field_name":       nullableStringValue(fieldNameV),
      "old_value":        nullableStringValue(oldValue),
      "new_value":        nullableStringValue(newValue),
      "submit_id":        nullableInt64Pointer(submitID),
      "changed_by":       nullableInt64Pointer(changedBy),
      "changed_name":     nullableStringValue(displayName),
      "changed_username": nullableStringValue(username),
      "created_at":       nullableTimePointer(createdAt),
    })
  }

  c.JSON(http.StatusOK, gin.H{"data": items})
}

// ListMediaVersions returns media version history.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *HistoryHandler) ListMediaVersions(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  draftID := parseInt64Query(c, "draft_version_id")
  moduleKey := strings.TrimSpace(c.Query("module_key"))
  mediaType := strings.TrimSpace(c.Query("media_type"))
  if draftID <= 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "draft_version_id is required"})
    return
  }

  limit, offset := parsePagination(c)

  query := `SELECT v.id, v.asset_id, v.version_no, v.file_url, v.file_size, v.width, v.height, v.duration_ms, v.format, v.compress_profile,
    v.created_by, v.created_at, a.draft_version_id, a.module_key, a.media_type, a.file_url
    FROM app_db_media_versions v
    JOIN app_db_media_assets a ON a.id = v.asset_id
    WHERE a.draft_version_id = ?`
  args := []interface{}{draftID}

  if moduleKey != "" {
    query += " AND a.module_key = ?"
    args = append(args, moduleKey)
  }
  if mediaType != "" {
    query += " AND a.media_type = ?"
    args = append(args, mediaType)
  }

  query += " ORDER BY v.id DESC LIMIT ? OFFSET ?"
  args = append(args, limit, offset)

  rows, err := h.db.Query(query, args...)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }
  defer rows.Close()

  var ossService *services.OSSService
  if service, err := services.NewOSSService(h.cfg, h.redis); err == nil {
    ossService = service
  }

  items := make([]gin.H, 0)
  for rows.Next() {
    var (
      id           int64
      assetID      sql.NullInt64
      versionNo    sql.NullInt64
      fileURL      sql.NullString
      fileSize     sql.NullInt64
      width        sql.NullInt64
      height       sql.NullInt64
      duration     sql.NullInt64
      format       sql.NullString
      compress     sql.NullString
      createdBy   sql.NullInt64
      createdAt   sql.NullTime
      draftVersion sql.NullInt64
      moduleKeyV  sql.NullString
      mediaTypeV  sql.NullString
      originURL   sql.NullString
    )

    if err := rows.Scan(
      &id,
      &assetID,
      &versionNo,
      &fileURL,
      &fileSize,
      &width,
      &height,
      &duration,
      &format,
      &compress,
      &createdBy,
      &createdAt,
      &draftVersion,
      &moduleKeyV,
      &mediaTypeV,
      &originURL,
    ); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed"})
      return
    }

    signed := signPath(h.cfg, ossService, nullableString(fileURL), "")
    originSigned := signPath(h.cfg, ossService, nullableString(originURL), "")

    items = append(items, gin.H{
      "id":                id,
      "asset_id":          nullableInt64Pointer(assetID),
      "version_no":        nullableInt64Pointer(versionNo),
      "file_url":          nullableStringValue(fileURL),
      "file_url_signed":   signed,
      "file_size":         nullableInt64Pointer(fileSize),
      "width":             nullableInt64Pointer(width),
      "height":            nullableInt64Pointer(height),
      "duration_ms":       nullableInt64Pointer(duration),
      "format":            nullableStringValue(format),
      "compress_profile":  nullableStringValue(compress),
      "created_by":        nullableInt64Pointer(createdBy),
      "created_at":        nullableTimePointer(createdAt),
      "draft_version_id":  nullableInt64Pointer(draftVersion),
      "module_key":        nullableStringValue(moduleKeyV),
      "media_type":        nullableStringValue(mediaTypeV),
      "origin_url":        nullableStringValue(originURL),
      "origin_url_signed": originSigned,
    })
  }

  c.JSON(http.StatusOK, gin.H{"data": items})
}

func parsePagination(c *gin.Context) (int, int) {
  limit := int(parseInt64Query(c, "limit"))
  offset := int(parseInt64Query(c, "offset"))
  if limit <= 0 {
    limit = 50
  }
  if limit > 200 {
    limit = 200
  }
  if offset < 0 {
    offset = 0
  }
  return limit, offset
}

func decodeJSON(value sql.NullString) interface{} {
  if !value.Valid {
    return nil
  }
  trimmed := strings.TrimSpace(value.String)
  if trimmed == "" {
    return nil
  }
  var payload interface{}
  if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
    return trimmed
  }
  return payload
}
