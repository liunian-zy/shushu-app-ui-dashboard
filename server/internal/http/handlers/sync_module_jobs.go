package handlers

import (
  "database/sql"
  "net/http"
  "strings"

  "github.com/gin-gonic/gin"
)

// ListModuleJobs returns sync job history by module.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *SyncHandler) ListModuleJobs(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  draftID := parseInt64Query(c, "draft_version_id")
  if draftID <= 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "draft_version_id is required"})
    return
  }

  moduleKey := strings.TrimSpace(c.Query("module_key"))

  query := `SELECT j.id, j.module_key, j.status, j.error_message, j.started_at, j.finished_at, j.created_at,
    j.trigger_by, u.display_name, u.username
    FROM app_db_sync_module_jobs j
    LEFT JOIN app_db_users u ON u.id = j.trigger_by
    WHERE j.draft_version_id = ?`
  args := []any{draftID}
  if moduleKey != "" {
    query += " AND j.module_key = ?"
    args = append(args, moduleKey)
  }
  query += " ORDER BY j.created_at DESC, j.id DESC"

  rows, err := h.db.Query(query, args...)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }
  defer rows.Close()

  items := make([]gin.H, 0)
  for rows.Next() {
    var (
      id           int64
      module       sql.NullString
      status       sql.NullString
      errorMessage sql.NullString
      startedAt    sql.NullTime
      finishedAt   sql.NullTime
      createdAt    sql.NullTime
      triggerBy    sql.NullInt64
      displayName  sql.NullString
      username     sql.NullString
    )
    if err := rows.Scan(&id, &module, &status, &errorMessage, &startedAt, &finishedAt, &createdAt, &triggerBy, &displayName, &username); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed"})
      return
    }
    items = append(items, gin.H{
      "id":             id,
      "module_key":     nullableString(module),
      "status":         nullableString(status),
      "error_message":  nullableString(errorMessage),
      "started_at":     nullableTimePointer(startedAt),
      "finished_at":    nullableTimePointer(finishedAt),
      "created_at":     nullableTimePointer(createdAt),
      "trigger_by":     nullableInt64Pointer(triggerBy),
      "trigger_name":   nullableString(displayName),
      "trigger_username": nullableString(username),
    })
  }

  c.JSON(http.StatusOK, gin.H{"data": items})
}
