package handlers

import (
  "database/sql"
  "net/http"
  "strings"
  "time"

  "github.com/gin-gonic/gin"

  "shushu-app-ui-dashboard/internal/config"
  "shushu-app-ui-dashboard/internal/services"
)

type UserHandler struct {
  cfg *config.Config
  db  *sql.DB
}

type createUserRequest struct {
  Username    string `json:"username"`
  DisplayName string `json:"display_name"`
  Role        string `json:"role"`
  Password    string `json:"password"`
}

// NewUserHandler creates a handler for user operations.
// Args:
//   cfg: App config instance.
//   db: Database connection.
// Returns:
//   *UserHandler: Initialized handler.
func NewUserHandler(cfg *config.Config, db *sql.DB) *UserHandler {
  return &UserHandler{cfg: cfg, db: db}
}

// Create creates a new user (admin only).
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *UserHandler) Create(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  var req createUserRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
    return
  }

  username := strings.TrimSpace(req.Username)
  password := strings.TrimSpace(req.Password)
  if username == "" || password == "" {
    c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
    return
  }

  role := strings.ToLower(strings.TrimSpace(req.Role))
  if role == "" {
    role = "user"
  }
  if role != "admin" && role != "user" {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
    return
  }

  var exists int64
  if err := h.db.QueryRow("SELECT COUNT(1) FROM app_db_users WHERE username = ?", username).Scan(&exists); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }
  if exists > 0 {
    c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
    return
  }

  authService, err := services.NewAuthService(h.cfg)
  if err != nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
    return
  }
  hash, err := authService.HashPassword(password)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "hash failed"})
    return
  }

  result, err := h.db.Exec(
    "INSERT INTO app_db_users (username, display_name, role, status, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
    username,
    nullIfEmpty(strings.TrimSpace(req.DisplayName)),
    role,
    1,
    hash,
    time.Now(),
    time.Now(),
  )
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "insert failed"})
    return
  }

  id, _ := result.LastInsertId()
  c.JSON(http.StatusOK, gin.H{
    "id":           id,
    "username":     username,
    "display_name": strings.TrimSpace(req.DisplayName),
    "role":         role,
  })
}

// List returns users (admin only).
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *UserHandler) List(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  rows, err := h.db.Query(
    "SELECT id, username, display_name, role, status, created_at, updated_at, last_login_at FROM app_db_users ORDER BY id DESC",
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
      username    sql.NullString
      displayName sql.NullString
      role        sql.NullString
      status      sql.NullInt64
      createdAt   sql.NullTime
      updatedAt   sql.NullTime
      lastLoginAt sql.NullTime
    )
    if err := rows.Scan(&id, &username, &displayName, &role, &status, &createdAt, &updatedAt, &lastLoginAt); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed"})
      return
    }

    items = append(items, gin.H{
      "id":            id,
      "username":      nullableString(username),
      "display_name":  nullableString(displayName),
      "role":          nullableString(role),
      "status":        nullableInt(status),
      "created_at":    nullableTimePointer(createdAt),
      "updated_at":    nullableTimePointer(updatedAt),
      "last_login_at": nullableTimePointer(lastLoginAt),
    })
  }

  c.JSON(http.StatusOK, gin.H{"data": items})
}
