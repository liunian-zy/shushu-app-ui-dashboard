package handlers

import (
  "database/sql"
  "errors"
  "io"
  "net/http"
  "os"
  "path/filepath"
  "strings"
  "time"

  "github.com/gin-gonic/gin"
  "github.com/redis/go-redis/v9"

  "shushu-app-ui-dashboard/internal/config"
  "shushu-app-ui-dashboard/internal/http/middleware"
  "shushu-app-ui-dashboard/internal/services"
)

type IdentityTemplateHandler struct {
  cfg   *config.Config
  db    *sql.DB
  redis *redis.Client
}

type templateRequest struct {
  Name        string `json:"name"`
  Description string `json:"description"`
  Status      *int   `json:"status"`
}

type templateItemRequest struct {
  Name   string `json:"name"`
  Image  string `json:"image"`
  Sort   *int   `json:"sort"`
  Status *int   `json:"status"`
}

type applyTemplateRequest struct {
  DraftVersionID int64 `json:"draft_version_id"`
  TemplateID     int64 `json:"template_id"`
  Replace        *bool `json:"replace"`
}

// NewIdentityTemplateHandler creates a handler for identity templates.
// Args:
//   db: Database connection.
// Returns:
//   *IdentityTemplateHandler: Initialized handler.
func NewIdentityTemplateHandler(cfg *config.Config, db *sql.DB, redis *redis.Client) *IdentityTemplateHandler {
  return &IdentityTemplateHandler{cfg: cfg, db: db, redis: redis}
}

// ListTemplates returns template list.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *IdentityTemplateHandler) ListTemplates(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  rows, err := h.db.Query(
    "SELECT t.id, t.name, t.description, t.status, t.created_at, t.updated_at, COUNT(i.id) AS item_count FROM app_db_identity_templates t LEFT JOIN app_db_identity_template_items i ON i.template_id = t.id GROUP BY t.id, t.name, t.description, t.status, t.created_at, t.updated_at ORDER BY t.id DESC",
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
      name        sql.NullString
      description sql.NullString
      status      sql.NullInt64
      createdAt   sql.NullTime
      updatedAt   sql.NullTime
      itemCount   sql.NullInt64
    )
    if err := rows.Scan(&id, &name, &description, &status, &createdAt, &updatedAt, &itemCount); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed"})
      return
    }
    items = append(items, gin.H{
      "id":           id,
      "name":         nullableString(name),
      "description":  nullableString(description),
      "status":       nullableInt(status),
      "item_count":   nullableInt(itemCount),
      "created_at":   nullableTimePointer(createdAt),
      "updated_at":   nullableTimePointer(updatedAt),
    })
  }

  c.JSON(http.StatusOK, gin.H{"data": items})
}

// CreateTemplate creates a template (admin only).
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *IdentityTemplateHandler) CreateTemplate(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  claims, ok := middleware.GetAuthClaims(c)
  if !ok {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
    return
  }

  var req templateRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
    return
  }
  name := strings.TrimSpace(req.Name)
  if name == "" {
    c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
    return
  }
  status := 1
  if req.Status != nil {
    status = *req.Status
  }

  result, err := h.db.Exec(
    "INSERT INTO app_db_identity_templates (name, description, status, created_by, updated_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
    name,
    nullIfEmpty(strings.TrimSpace(req.Description)),
    status,
    claims.UserID,
    claims.UserID,
    time.Now(),
    time.Now(),
  )
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "insert failed"})
    return
  }
  id, _ := result.LastInsertId()
  c.JSON(http.StatusOK, gin.H{"id": id})
}

// UpdateTemplate updates a template (admin only).
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *IdentityTemplateHandler) UpdateTemplate(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  claims, ok := middleware.GetAuthClaims(c)
  if !ok {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
    return
  }

  id, err := parseInt64ParamValue(c.Param("id"))
  if err != nil || id <= 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
    return
  }

  var req templateRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
    return
  }

  payload := map[string]interface{}{}
  if strings.TrimSpace(req.Name) != "" {
    payload["name"] = strings.TrimSpace(req.Name)
  }
  if strings.TrimSpace(req.Description) != "" {
    payload["description"] = strings.TrimSpace(req.Description)
  }
  if req.Status != nil {
    payload["status"] = *req.Status
  }
  if len(payload) == 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "empty payload"})
    return
  }
  payload["updated_by"] = claims.UserID
  payload["updated_at"] = time.Now()

  sqlText, args, err := BuildUpdateSQL("app_db_identity_templates", "id", id, payload)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  result, err := h.db.Exec(sqlText, args...)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
    return
  }
  rows, _ := result.RowsAffected()
  if rows == 0 {
    c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
    return
  }
  c.JSON(http.StatusOK, gin.H{"id": id})
}

// DeleteTemplate deletes a template and its items (admin only).
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *IdentityTemplateHandler) DeleteTemplate(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  id, err := parseInt64ParamValue(c.Param("id"))
  if err != nil || id <= 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
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

  if _, err := tx.Exec("DELETE FROM app_db_identity_template_items WHERE template_id = ?", id); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
    return
  }

  result, err := tx.Exec("DELETE FROM app_db_identity_templates WHERE id = ?", id)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
    return
  }
  rows, _ := result.RowsAffected()
  if rows == 0 {
    c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
    return
  }

  if err := tx.Commit(); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "commit failed"})
    return
  }

  c.JSON(http.StatusOK, gin.H{"id": id})
}

// ListTemplateItems returns items for a template.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *IdentityTemplateHandler) ListTemplateItems(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  templateID, err := parseInt64ParamValue(c.Param("id"))
  if err != nil || templateID <= 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template id"})
    return
  }

  rows, err := h.db.Query(
    "SELECT id, template_id, name, image, sort, status, created_at, updated_at FROM app_db_identity_template_items WHERE template_id = ? ORDER BY sort ASC, id ASC",
    templateID,
  )
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }
  defer rows.Close()

  ossService, _ := services.NewOSSService(h.cfg, h.redis)

  items := make([]gin.H, 0)
  for rows.Next() {
    var (
      id        int64
      template  sql.NullInt64
      name      sql.NullString
      image     sql.NullString
      sort      sql.NullInt64
      status    sql.NullInt64
      createdAt sql.NullTime
      updatedAt sql.NullTime
    )
    if err := rows.Scan(&id, &template, &name, &image, &sort, &status, &createdAt, &updatedAt); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed"})
      return
    }
    imageURL := signPath(h.cfg, ossService, nullableString(image), "")
    items = append(items, gin.H{
      "id":          id,
      "template_id": nullableInt(template),
      "name":        nullableString(name),
      "image":       nullableString(image),
      "image_url":   imageURL,
      "sort":        nullableInt(sort),
      "status":      nullableInt(status),
      "created_at":  nullableTimePointer(createdAt),
      "updated_at":  nullableTimePointer(updatedAt),
    })
  }

  c.JSON(http.StatusOK, gin.H{"data": items})
}

// CreateTemplateItem adds a template item (admin only).
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *IdentityTemplateHandler) CreateTemplateItem(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  claims, ok := middleware.GetAuthClaims(c)
  if !ok {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
    return
  }

  templateID, err := parseInt64ParamValue(c.Param("id"))
  if err != nil || templateID <= 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template id"})
    return
  }

  var req templateItemRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
    return
  }
  name := strings.TrimSpace(req.Name)
  if name == "" {
    c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
    return
  }
  status := 1
  if req.Status != nil {
    status = *req.Status
  }
  sort := 0
  if req.Sort != nil {
    sort = *req.Sort
  }

  result, err := h.db.Exec(
    "INSERT INTO app_db_identity_template_items (template_id, name, image, sort, status, created_by, updated_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
    templateID,
    name,
    nullIfEmpty(strings.TrimSpace(req.Image)),
    sort,
    status,
    claims.UserID,
    claims.UserID,
    time.Now(),
    time.Now(),
  )
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "insert failed"})
    return
  }
  id, _ := result.LastInsertId()
  c.JSON(http.StatusOK, gin.H{"id": id})
}

// UpdateTemplateItem updates a template item (admin only).
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *IdentityTemplateHandler) UpdateTemplateItem(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  claims, ok := middleware.GetAuthClaims(c)
  if !ok {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
    return
  }

  id, err := parseInt64ParamValue(c.Param("id"))
  if err != nil || id <= 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
    return
  }

  var req templateItemRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
    return
  }

  payload := map[string]interface{}{}
  if strings.TrimSpace(req.Name) != "" {
    payload["name"] = strings.TrimSpace(req.Name)
  }
  if strings.TrimSpace(req.Image) != "" {
    payload["image"] = strings.TrimSpace(req.Image)
  }
  if req.Sort != nil {
    payload["sort"] = *req.Sort
  }
  if req.Status != nil {
    payload["status"] = *req.Status
  }
  if len(payload) == 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "empty payload"})
    return
  }
  payload["updated_by"] = claims.UserID
  payload["updated_at"] = time.Now()

  sqlText, args, err := BuildUpdateSQL("app_db_identity_template_items", "id", id, payload)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  result, err := h.db.Exec(sqlText, args...)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
    return
  }
  rows, _ := result.RowsAffected()
  if rows == 0 {
    c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
    return
  }
  c.JSON(http.StatusOK, gin.H{"id": id})
}

// DeleteTemplateItem deletes a template item (admin only).
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *IdentityTemplateHandler) DeleteTemplateItem(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  id, err := parseInt64ParamValue(c.Param("id"))
  if err != nil || id <= 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
    return
  }

  result, err := h.db.Exec("DELETE FROM app_db_identity_template_items WHERE id = ?", id)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
    return
  }
  rows, _ := result.RowsAffected()
  if rows == 0 {
    c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
    return
  }
  c.JSON(http.StatusOK, gin.H{"id": id})
}

// ApplyTemplate copies template items to identities.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *IdentityTemplateHandler) ApplyTemplate(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  claims, ok := middleware.GetAuthClaims(c)
  if !ok {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
    return
  }

  var req applyTemplateRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
    return
  }
  if req.DraftVersionID <= 0 || req.TemplateID <= 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "draft_version_id and template_id are required"})
    return
  }
  if err := h.db.Ping(); err != nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }
  replace := true
  if req.Replace != nil {
    replace = *req.Replace
  }

  row := h.db.QueryRow("SELECT app_version_name FROM app_db_version_names WHERE id = ?", req.DraftVersionID)
  var appVersionName sql.NullString
  if err := row.Scan(&appVersionName); err != nil {
    if errors.Is(err, sql.ErrNoRows) {
      c.JSON(http.StatusBadRequest, gin.H{"error": "draft version not found"})
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
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

  if replace {
    if _, err := tx.Exec("DELETE FROM app_db_identities WHERE draft_version_id = ?", req.DraftVersionID); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
      return
    }
  }

  rows, err := tx.Query(
    "SELECT name, image, sort, status FROM app_db_identity_template_items WHERE template_id = ? ORDER BY sort ASC, id ASC",
    req.TemplateID,
  )
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }

  type templateItem struct {
    name   sql.NullString
    image  sql.NullString
    sort   sql.NullInt64
    status sql.NullInt64
  }
  templateItems := make([]templateItem, 0)
  for rows.Next() {
    var item templateItem
    if err := rows.Scan(&item.name, &item.image, &item.sort, &item.status); err != nil {
      _ = rows.Close()
      c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed"})
      return
    }
    templateItems = append(templateItems, item)
  }
  _ = rows.Close()
  if err := rows.Err(); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }

  inserted := 0
  now := time.Now()
  for _, item := range templateItems {
    if strings.TrimSpace(nullableStringValue(item.name)) == "" {
      continue
    }
    imagePath := nullableStringValue(item.image)
    if imagePath != "" {
      copied, err := h.copyTemplateImage(req.DraftVersionID, imagePath)
      if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "copy template image failed"})
        return
      }
      imagePath = copied
    }
    if _, err := tx.Exec(
      "INSERT INTO app_db_identities (draft_version_id, name, image, sort, status, app_version_name, created_by, updated_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
      req.DraftVersionID,
      nullableStringValue(item.name),
      imagePath,
      nullableIntValue(item.sort),
      nullableIntValue(item.status),
      nullableStringValue(appVersionName),
      claims.UserID,
      claims.UserID,
      now,
      now,
    ); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "insert failed", "detail": err.Error()})
      return
    }
    inserted++
  }

  if err := tx.Commit(); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "commit failed"})
    return
  }

  c.JSON(http.StatusOK, gin.H{
    "inserted": inserted,
    "replaced": replace,
  })
}

func nullableIntValue(value sql.NullInt64) int64 {
  if value.Valid {
    return value.Int64
  }
  return 0
}

func (h *IdentityTemplateHandler) copyTemplateImage(draftVersionID int64, imagePath string) (string, error) {
  if !isLocalPath(imagePath) {
    return imagePath, nil
  }
  relative := trimLocalPrefix(imagePath)
  if strings.TrimSpace(relative) == "" {
    return imagePath, nil
  }
  sourcePath, err := buildLocalFilePath(h.cfg, relative)
  if err != nil {
    return "", err
  }
  filename := filepath.Base(relative)
  targetRelative, err := buildLocalUploadPath("identities", draftVersionID, filename)
  if err != nil {
    return "", err
  }
  targetPath, err := buildLocalFilePath(h.cfg, targetRelative)
  if err != nil {
    return "", err
  }
  if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
    return "", err
  }
  if err := copyLocalFile(sourcePath, targetPath); err != nil {
    return "", err
  }
  return localPathPrefix + targetRelative, nil
}

func copyLocalFile(sourcePath, targetPath string) error {
  source, err := os.Open(sourcePath)
  if err != nil {
    return err
  }
  defer func() {
    _ = source.Close()
  }()

  target, err := os.Create(targetPath)
  if err != nil {
    return err
  }
  defer func() {
    _ = target.Close()
  }()

  if _, err := io.Copy(target, source); err != nil {
    return err
  }
  return target.Sync()
}
