package handlers

import (
  "database/sql"
  "net/http"
  "strings"
  "time"

  "github.com/gin-gonic/gin"

  "shushu-app-ui-dashboard/internal/config"
  "shushu-app-ui-dashboard/internal/http/middleware"
  "shushu-app-ui-dashboard/internal/services"
)

type AuthHandler struct {
  cfg *config.Config
  db  *sql.DB
}

type loginRequest struct {
  Username string `json:"username"`
  Password string `json:"password"`
}

type bootstrapRequest struct {
  Username    string `json:"username"`
  DisplayName string `json:"display_name"`
  Password    string `json:"password"`
}

// NewAuthHandler creates a handler for auth operations.
// Args:
//   cfg: App config instance.
//   db: Database connection.
// Returns:
//   *AuthHandler: Initialized handler.
func NewAuthHandler(cfg *config.Config, db *sql.DB) *AuthHandler {
  return &AuthHandler{cfg: cfg, db: db}
}

// Login authenticates a user and returns a token.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *AuthHandler) Login(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  var req loginRequest
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

  authService, err := services.NewAuthService(h.cfg)
  if err != nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
    return
  }

  var (
    id           int64
    dbUsername   string
    displayName  sql.NullString
    role         sql.NullString
    status       sql.NullInt64
    passwordHash sql.NullString
  )

  row := h.db.QueryRow(
    "SELECT id, username, display_name, role, status, password_hash FROM app_db_users WHERE username = ? ORDER BY id DESC LIMIT 1",
    username,
  )
  if err := row.Scan(&id, &dbUsername, &displayName, &role, &status, &passwordHash); err != nil {
    if err == sql.ErrNoRows {
      c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }

  if status.Valid && status.Int64 == 0 {
    c.JSON(http.StatusForbidden, gin.H{"error": "user disabled"})
    return
  }
  if !passwordHash.Valid || passwordHash.String == "" {
    c.JSON(http.StatusForbidden, gin.H{"error": "password not set"})
    return
  }

  if !authService.VerifyPassword(passwordHash.String, password) {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
    return
  }

  user := &services.AuthUser{
    ID:          id,
    Username:    dbUsername,
    DisplayName: nullableStringValue(displayName),
    Role:        normalizeRole(role),
  }

  token, expiresAt, err := authService.IssueToken(user)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "token failed"})
    return
  }

  _, _ = h.db.Exec("UPDATE app_db_users SET last_login_at = ?, updated_at = ? WHERE id = ?", time.Now(), time.Now(), id)

  c.JSON(http.StatusOK, gin.H{
    "token":      token,
    "expires_at": expiresAt.Format(time.RFC3339),
    "user": gin.H{
      "id":           user.ID,
      "username":     user.Username,
      "display_name": user.DisplayName,
      "role":         user.Role,
    },
  })
}

// Bootstrap creates the first admin user when no users exist.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *AuthHandler) Bootstrap(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  var count int64
  if err := h.db.QueryRow("SELECT COUNT(1) FROM app_db_users").Scan(&count); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }
  if count > 0 {
    c.JSON(http.StatusForbidden, gin.H{"error": "bootstrap not allowed"})
    return
  }

  var req bootstrapRequest
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
    "admin",
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
    "id":       id,
    "username": username,
    "role":     "admin",
  })
}

// Me returns current user info.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *AuthHandler) Me(c *gin.Context) {
  claims, ok := middleware.GetAuthClaims(c)
  if !ok {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
    return
  }

  c.JSON(http.StatusOK, gin.H{
    "user": gin.H{
      "id":           claims.UserID,
      "username":     claims.Username,
      "display_name": claims.DisplayName,
      "role":         claims.Role,
    },
  })
}

func normalizeRole(value sql.NullString) string {
  if value.Valid {
    raw := strings.TrimSpace(strings.ToLower(value.String))
    if raw != "" {
      return raw
    }
  }
  return "user"
}
