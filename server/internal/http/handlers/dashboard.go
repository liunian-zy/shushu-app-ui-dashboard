package handlers

import (
  "database/sql"
  "fmt"
  "net/http"
  "strings"
  "time"

  "github.com/gin-gonic/gin"
)

type DashboardHandler struct {
  db *sql.DB
}

type dashboardTaskSummary struct {
  Total        int64            `json:"total"`
  Pending      int64            `json:"pending"`
  AssistActive int64            `json:"assist_active"`
  StatusCounts map[string]int64 `json:"status_counts"`
}

type dashboardMediaSummary struct {
  TodayChecked   int64 `json:"today_checked"`
  TodayCompliant int64 `json:"today_compliant"`
  TodayPending   int64 `json:"today_pending"`
}

type dashboardSyncSummary struct {
  PendingVersions int64      `json:"pending_versions"`
  LastSyncAt      *time.Time `json:"last_sync_at"`
}

// NewDashboardHandler creates a handler for dashboard summary.
// Args:
//   db: Database connection.
// Returns:
//   *DashboardHandler: Handler instance.
func NewDashboardHandler(db *sql.DB) *DashboardHandler {
  return &DashboardHandler{db: db}
}

// Summary returns dashboard summary data by draft version.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DashboardHandler) Summary(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  draftID := parseInt64Query(c, "draft_version_id")
  if draftID <= 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "draft_version_id is required"})
    return
  }

  tasks, err := h.loadTaskSummary(draftID)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "task summary failed"})
    return
  }

  media, err := h.loadMediaSummary(draftID)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "media summary failed"})
    return
  }

  syncSummary, err := h.loadSyncSummary(draftID)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "sync summary failed"})
    return
  }

  c.JSON(http.StatusOK, gin.H{
    "tasks": tasks,
    "media": media,
    "sync":  syncSummary,
  })
}

func (h *DashboardHandler) loadTaskSummary(draftID int64) (dashboardTaskSummary, error) {
  rows, err := h.db.Query(
    "SELECT status, COUNT(*) FROM app_db_tasks WHERE draft_version_id = ? GROUP BY status",
    draftID,
  )
  if err != nil {
    return dashboardTaskSummary{}, err
  }
  defer rows.Close()

  summary := dashboardTaskSummary{
    StatusCounts: make(map[string]int64),
  }

  for rows.Next() {
    var status sql.NullString
    var count int64
    if err := rows.Scan(&status, &count); err != nil {
      return dashboardTaskSummary{}, err
    }
    statusValue := strings.TrimSpace(nullableStringValue(status))
    if statusValue == "" {
      statusValue = "unknown"
    }
    summary.StatusCounts[statusValue] = count
    summary.Total += count
    if IsPendingTaskStatus(statusValue) {
      summary.Pending += count
    }
  }

  placeholders := strings.TrimRight(strings.Repeat("?,", len(pendingTaskStatuses)), ",")
  args := make([]interface{}, 0, len(pendingTaskStatuses)+1)
  args = append(args, draftID)
  for _, status := range pendingTaskStatuses {
    args = append(args, status)
  }
  query := fmt.Sprintf("SELECT COUNT(*) FROM app_db_tasks WHERE draft_version_id = ? AND allow_assist = 1 AND status IN (%s)", placeholders)
  if err := h.db.QueryRow(query, args...).Scan(&summary.AssistActive); err != nil {
    return dashboardTaskSummary{}, err
  }

  return summary, rows.Err()
}

func (h *DashboardHandler) loadMediaSummary(draftID int64) (dashboardMediaSummary, error) {
  summary := dashboardMediaSummary{}

  err := h.db.QueryRow(
    `SELECT COUNT(*)
     FROM app_db_media_versions v
     JOIN app_db_media_assets a ON a.id = v.asset_id
     WHERE a.draft_version_id = ? AND DATE(v.created_at) = CURDATE()`,
    draftID,
  ).Scan(&summary.TodayChecked)
  if err != nil {
    return dashboardMediaSummary{}, err
  }

  err = h.db.QueryRow(
    `SELECT COUNT(*)
     FROM app_db_media_assets
     WHERE draft_version_id = ? AND DATE(created_at) = CURDATE()
       AND (status IS NULL OR status != 'active')`,
    draftID,
  ).Scan(&summary.TodayPending)
  if err != nil {
    return dashboardMediaSummary{}, err
  }

  summary.TodayCompliant = summary.TodayChecked
  return summary, nil
}

func (h *DashboardHandler) loadSyncSummary(draftID int64) (dashboardSyncSummary, error) {
  summary := dashboardSyncSummary{}

  var syncStatus sql.NullString
  var syncedAt sql.NullTime
  row := h.db.QueryRow(
    "SELECT sync_status, synced_at FROM app_db_version_names WHERE id = ? LIMIT 1",
    draftID,
  )
  if err := row.Scan(&syncStatus, &syncedAt); err != nil {
    if err == sql.ErrNoRows {
      return dashboardSyncSummary{}, err
    }
    return dashboardSyncSummary{}, err
  }

  status := strings.ToLower(strings.TrimSpace(nullableStringValue(syncStatus)))
  if status != "success" {
    summary.PendingVersions = 1
  }

  var lastSync sql.NullTime
  err := h.db.QueryRow(
    "SELECT MAX(COALESCE(finished_at, started_at, created_at)) FROM app_db_sync_jobs WHERE draft_version_id = ?",
    draftID,
  ).Scan(&lastSync)
  if err != nil {
    return dashboardSyncSummary{}, err
  }

  if lastSync.Valid {
    summary.LastSyncAt = nullableTimePointer(lastSync)
  } else if syncedAt.Valid {
    summary.LastSyncAt = nullableTimePointer(syncedAt)
  }

  return summary, nil
}
