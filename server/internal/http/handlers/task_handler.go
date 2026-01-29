package handlers

import (
  "database/sql"
  "encoding/json"
  "errors"
  "net/http"
  "strconv"
  "strings"
  "time"

  "github.com/gin-gonic/gin"
  "github.com/redis/go-redis/v9"

  "shushu-app-ui-dashboard/internal/config"
  "shushu-app-ui-dashboard/internal/http/middleware"
)

type TaskHandler struct {
  cfg   *config.Config
  db    *sql.DB
  redis *redis.Client
}

type createTaskRequest struct {
  DraftVersionID int64  `json:"draft_version_id"`
  ModuleKey      string `json:"module_key"`
  Title          string `json:"title"`
  Description    string `json:"description"`
  AssignedTo     *int64 `json:"assigned_to"`
  AllowAssist    *int   `json:"allow_assist"`
  Priority       *int   `json:"priority"`
  Status         string `json:"status"`
}

type updateTaskRequest struct {
  Title       *string `json:"title"`
  Description *string `json:"description"`
  AssignedTo  *int64  `json:"assigned_to"`
  AllowAssist *int    `json:"allow_assist"`
  Priority    *int    `json:"priority"`
  Status      *string `json:"status"`
}

type assistTaskRequest struct {
  Note string `json:"note"`
}

// NewTaskHandler creates a handler for task operations.
// Args:
//   db: Database connection.
// Returns:
//   *TaskHandler: Initialized handler.
func NewTaskHandler(cfg *config.Config, db *sql.DB, redis *redis.Client) *TaskHandler {
  return &TaskHandler{cfg: cfg, db: db, redis: redis}
}

// List returns task list for a draft version.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *TaskHandler) List(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  draftVersionID := parseInt64Query(c, "draft_version_id")
  if draftVersionID <= 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "draft_version_id is required"})
    return
  }

  moduleKey := strings.TrimSpace(c.Query("module_key"))
  status := strings.TrimSpace(c.Query("status"))
  assignedTo := parseInt64Query(c, "assigned_to")

  where := []string{"draft_version_id = ?"}
  args := []any{draftVersionID}
  if moduleKey != "" {
    where = append(where, "module_key = ?")
    args = append(args, moduleKey)
  }
  if status != "" {
    where = append(where, "status = ?")
    args = append(args, strings.ToLower(status))
  }
  if assignedTo > 0 {
    where = append(where, "assigned_to = ?")
    args = append(args, assignedTo)
  }

  query := "SELECT id, draft_version_id, module_key, title, description, status, assigned_to, allow_assist, priority, created_by, updated_by, created_at, updated_at FROM app_db_tasks WHERE " + strings.Join(where, " AND ") + " ORDER BY priority DESC, id DESC"
  rows, err := h.db.Query(query, args...)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }
  defer rows.Close()

  items := make([]gin.H, 0)
  for rows.Next() {
    var (
      id             int64
      draftID        int64
      moduleField    sql.NullString
      title          sql.NullString
      desc           sql.NullString
      taskStatus     sql.NullString
      assigned       sql.NullInt64
      allowAssist    sql.NullInt64
      priority       sql.NullInt64
      createdBy      sql.NullInt64
      updatedBy      sql.NullInt64
      createdAt      sql.NullTime
      updatedAt      sql.NullTime
    )

    if err := rows.Scan(&id, &draftID, &moduleField, &title, &desc, &taskStatus, &assigned, &allowAssist, &priority, &createdBy, &updatedBy, &createdAt, &updatedAt); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed"})
      return
    }

    items = append(items, gin.H{
      "id":              id,
      "draft_version_id": draftID,
      "module_key":      nullableString(moduleField),
      "title":           nullableString(title),
      "description":     nullableString(desc),
      "status":          nullableString(taskStatus),
      "assigned_to":     nullableInt(assigned),
      "allow_assist":    nullableInt(allowAssist),
      "priority":        nullableInt(priority),
      "created_by":      nullableInt(createdBy),
      "updated_by":      nullableInt(updatedBy),
      "created_at":      nullableTimePointer(createdAt),
      "updated_at":      nullableTimePointer(updatedAt),
    })
  }

  c.JSON(http.StatusOK, gin.H{"data": items})
}

// Create creates a task (admin only).
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *TaskHandler) Create(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  claims, ok := middleware.GetAuthClaims(c)
  if !ok {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
    return
  }

  var req createTaskRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
    return
  }

  req.ModuleKey = strings.TrimSpace(req.ModuleKey)
  req.Title = strings.TrimSpace(req.Title)
  if req.DraftVersionID <= 0 || req.ModuleKey == "" || req.Title == "" {
    c.JSON(http.StatusBadRequest, gin.H{"error": "draft_version_id, module_key, title are required"})
    return
  }

  status, err := NormalizeTaskStatus(req.Status)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
    return
  }

  allowAssist := int64(1)
  if req.AllowAssist != nil {
    allowAssist = int64(*req.AllowAssist)
  }
  priority := int64(0)
  if req.Priority != nil {
    priority = int64(*req.Priority)
  }
  assignedTo := NormalizeAssignedTo(req.AssignedTo)

  tx, err := h.db.Begin()
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "transaction failed"})
    return
  }
  defer func() {
    _ = tx.Rollback()
  }()

  result, err := tx.Exec(
    "INSERT INTO app_db_tasks (draft_version_id, module_key, title, description, status, assigned_to, allow_assist, priority, created_by, updated_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
    req.DraftVersionID,
    req.ModuleKey,
    req.Title,
    nullIfEmpty(strings.TrimSpace(req.Description)),
    status,
    nullableID(assignedTo),
    allowAssist,
    priority,
    claims.UserID,
    claims.UserID,
    time.Now(),
    time.Now(),
  )
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "insert failed"})
    return
  }

  taskID, err := result.LastInsertId()
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "insert failed"})
    return
  }

  _ = insertTaskAction(tx, taskID, "create", claims.UserID, map[string]interface{}{
    "status":     status,
    "module_key": req.ModuleKey,
  })

  if assignedTo > 0 {
    _ = insertTaskAction(tx, taskID, "assign", claims.UserID, map[string]interface{}{
      "assigned_to": assignedTo,
    })
  }

  if err := tx.Commit(); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "insert failed"})
    return
  }

  c.JSON(http.StatusOK, gin.H{
    "id": taskID,
  })
}

// Update updates a task (admin only).
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *TaskHandler) Update(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  claims, ok := middleware.GetAuthClaims(c)
  if !ok {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
    return
  }

  taskID, err := strconv.ParseInt(c.Param("id"), 10, 64)
  if err != nil || taskID <= 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
    return
  }

  var req updateTaskRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
    return
  }

  if req.Title == nil && req.Description == nil && req.AssignedTo == nil && req.AllowAssist == nil && req.Priority == nil && req.Status == nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
    return
  }

  tx, err := h.db.Begin()
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "transaction failed"})
    return
  }
  defer func() {
    _ = tx.Rollback()
  }()

  var (
    currentAssigned sql.NullInt64
    currentStatus   sql.NullString
    currentAllow    sql.NullInt64
  )
  row := tx.QueryRow("SELECT assigned_to, status, allow_assist FROM app_db_tasks WHERE id = ?", taskID)
  if err := row.Scan(&currentAssigned, &currentStatus, &currentAllow); err != nil {
    if errors.Is(err, sql.ErrNoRows) {
      c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }

  setParts := make([]string, 0)
  args := make([]any, 0)
  actions := make([]map[string]interface{}, 0)

  if req.Title != nil {
    title := strings.TrimSpace(*req.Title)
    setParts = append(setParts, "title = ?")
    args = append(args, nullIfEmpty(title))
  }
  if req.Description != nil {
    desc := strings.TrimSpace(*req.Description)
    setParts = append(setParts, "description = ?")
    args = append(args, nullIfEmpty(desc))
  }
  if req.AssignedTo != nil {
    assignedTo := NormalizeAssignedTo(req.AssignedTo)
    setParts = append(setParts, "assigned_to = ?")
    args = append(args, nullableID(assignedTo))

    if currentAssigned.Valid && currentAssigned.Int64 != assignedTo {
      actions = append(actions, map[string]interface{}{
        "action":       "assign",
        "assigned_to":  assignedTo,
        "previous_assigned_to": currentAssigned.Int64,
      })
    }
    if !currentAssigned.Valid && assignedTo > 0 {
      actions = append(actions, map[string]interface{}{
        "action":      "assign",
        "assigned_to": assignedTo,
      })
    }
    if currentAssigned.Valid && assignedTo == 0 {
      actions = append(actions, map[string]interface{}{
        "action":      "assign",
        "assigned_to": nil,
      })
    }
  }
  if req.AllowAssist != nil {
    setParts = append(setParts, "allow_assist = ?")
    args = append(args, *req.AllowAssist)
    if currentAllow.Valid && currentAllow.Int64 != int64(*req.AllowAssist) {
      actions = append(actions, map[string]interface{}{
        "action":       "allow_assist",
        "allow_assist": *req.AllowAssist,
      })
    }
  }
  if req.Priority != nil {
    setParts = append(setParts, "priority = ?")
    args = append(args, *req.Priority)
  }
  if req.Status != nil {
    status, err := NormalizeTaskStatus(*req.Status)
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
      return
    }
    setParts = append(setParts, "status = ?")
    args = append(args, status)
    if currentStatus.Valid && currentStatus.String != status {
      actions = append(actions, map[string]interface{}{
        "action":   "status_change",
        "status":   status,
        "previous": currentStatus.String,
      })
    }
    if !currentStatus.Valid {
      actions = append(actions, map[string]interface{}{
        "action": "status_change",
        "status": status,
      })
    }
  }

  if len(setParts) == 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
    return
  }

  setParts = append(setParts, "updated_by = ?", "updated_at = ?")
  args = append(args, claims.UserID, time.Now(), taskID)

  query := "UPDATE app_db_tasks SET " + strings.Join(setParts, ", ") + " WHERE id = ?"
  if _, err := tx.Exec(query, args...); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
    return
  }

  for _, detail := range actions {
    action := detail["action"].(string)
    delete(detail, "action")
    _ = insertTaskAction(tx, taskID, action, claims.UserID, detail)
  }

  if err := tx.Commit(); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
    return
  }

  c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Assist records assistance on a task.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *TaskHandler) Assist(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  claims, ok := middleware.GetAuthClaims(c)
  if !ok {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
    return
  }

  taskID, err := strconv.ParseInt(c.Param("id"), 10, 64)
  if err != nil || taskID <= 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
    return
  }

  var req assistTaskRequest
  _ = c.ShouldBindJSON(&req)

  var allowAssist sql.NullInt64
  row := h.db.QueryRow("SELECT allow_assist FROM app_db_tasks WHERE id = ?", taskID)
  if err := row.Scan(&allowAssist); err != nil {
    if errors.Is(err, sql.ErrNoRows) {
      c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }

  if allowAssist.Valid && allowAssist.Int64 == 0 {
    c.JSON(http.StatusForbidden, gin.H{"error": "assist disabled"})
    return
  }

  detail := map[string]interface{}{}
  note := strings.TrimSpace(req.Note)
  if note != "" {
    detail["note"] = note
  }

  _, err = h.db.Exec("UPDATE app_db_tasks SET updated_by = ?, updated_at = ? WHERE id = ?", claims.UserID, time.Now(), taskID)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
    return
  }

  if err := insertTaskAction(h.db, taskID, "assist", claims.UserID, detail); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "action failed"})
    return
  }

  c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Actions returns task action history.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *TaskHandler) Actions(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  taskID, err := strconv.ParseInt(c.Param("id"), 10, 64)
  if err != nil || taskID <= 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
    return
  }

  rows, err := h.db.Query(
    "SELECT a.id, a.action, a.actor_id, a.detail_json, a.created_at, u.display_name, u.username FROM app_db_task_actions a LEFT JOIN app_db_users u ON u.id = a.actor_id WHERE a.task_id = ? ORDER BY a.id ASC",
    taskID,
  )
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }
  defer rows.Close()

  items := make([]gin.H, 0)
  for rows.Next() {
    var (
      id          int64
      action      sql.NullString
      actorID     sql.NullInt64
      detailRaw   sql.NullString
      createdAt   sql.NullTime
      displayName sql.NullString
      username    sql.NullString
    )
    if err := rows.Scan(&id, &action, &actorID, &detailRaw, &createdAt, &displayName, &username); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed"})
      return
    }
    detail := interface{}(nil)
    if detailRaw.Valid && strings.TrimSpace(detailRaw.String) != "" {
      var parsed interface{}
      if err := json.Unmarshal([]byte(detailRaw.String), &parsed); err == nil {
        detail = parsed
      }
    }
    items = append(items, gin.H{
      "id":         id,
      "action":     nullableString(action),
      "actor_id":   nullableInt(actorID),
      "actor_name": nullableStringValue(displayName),
      "actor_username": nullableStringValue(username),
      "detail":     detail,
      "created_at": nullableTimePointer(createdAt),
    })
  }

  c.JSON(http.StatusOK, gin.H{"data": items})
}

func insertTaskAction(exec sqlExecutor, taskID int64, action string, actorID int64, detail map[string]interface{}) error {
  raw := interface{}(nil)
  if len(detail) > 0 {
    if encoded, err := json.Marshal(detail); err == nil {
      raw = string(encoded)
    }
  }

  _, err := exec.Exec(
    "INSERT INTO app_db_task_actions (task_id, action, actor_id, detail_json, created_at) VALUES (?, ?, ?, ?, ?)",
    taskID,
    action,
    actorID,
    raw,
    time.Now(),
  )
  return err
}

type sqlExecutor interface {
  Exec(query string, args ...interface{}) (sql.Result, error)
}
