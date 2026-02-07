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

type UserHandler struct {
	cfg *config.Config
	db  *sql.DB
}

type createUserRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	Status      *int   `json:"status"`
	Password    string `json:"password"`
}

type updateUserRequest struct {
	Username    *string `json:"username"`
	DisplayName *string `json:"display_name"`
	Role        *string `json:"role"`
	Status      *int    `json:"status"`
	Password    *string `json:"password"`
}

type changeMyPasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// NewUserHandler creates a handler for user operations.
// Args:
//
//	cfg: App config instance.
//	db: Database connection.
//
// Returns:
//
//	*UserHandler: Initialized handler.
func NewUserHandler(cfg *config.Config, db *sql.DB) *UserHandler {
	return &UserHandler{cfg: cfg, db: db}
}

// Create creates a new user (admin only).
// Args:
//
//	c: Gin context.
//
// Returns:
//
//	None.
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

	status := 1
	if req.Status != nil {
		if *req.Status != 0 && *req.Status != 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
			return
		}
		status = *req.Status
	}

	result, err := h.db.Exec(
		"INSERT INTO app_db_users (username, display_name, role, status, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		username,
		nullIfEmpty(strings.TrimSpace(req.DisplayName)),
		role,
		status,
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
//
//	c: Gin context.
//
// Returns:
//
//	None.
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

// Update updates user profile fields (admin only).
// Args:
//
//	c: Gin context.
//
// Returns:
//
//	None.
func (h *UserHandler) Update(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
		return
	}

	id, err := parseInt64ParamValue(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	payload := map[string]interface{}{}
	if req.Username != nil {
		username := strings.TrimSpace(*req.Username)
		if username == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
			return
		}
		var exists int64
		if err := h.db.QueryRow("SELECT COUNT(1) FROM app_db_users WHERE username = ? AND id <> ?", username, id).Scan(&exists); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
			return
		}
		if exists > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
			return
		}
		payload["username"] = username
	}

	if req.DisplayName != nil {
		payload["display_name"] = nullIfEmpty(strings.TrimSpace(*req.DisplayName))
	}

	if req.Role != nil {
		role := strings.ToLower(strings.TrimSpace(*req.Role))
		if role != "admin" && role != "user" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
			return
		}
		payload["role"] = role
	}

	if req.Status != nil {
		if *req.Status != 0 && *req.Status != 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
			return
		}
		payload["status"] = *req.Status
	}

	if req.Password != nil {
		password := strings.TrimSpace(*req.Password)
		if password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "password is required"})
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
		payload["password_hash"] = hash
	}

	if len(payload) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty payload"})
		return
	}
	payload["updated_at"] = time.Now()

	sqlText, args, err := BuildUpdateSQL("app_db_users", "id", id, payload)
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

// ChangeMyPassword updates current user's password.
// Args:
//
//	c: Gin context.
//
// Returns:
//
//	None.
func (h *UserHandler) ChangeMyPassword(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
		return
	}

	claims, ok := middleware.GetAuthClaims(c)
	if !ok || claims.UserID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req changeMyPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	oldPassword := strings.TrimSpace(req.OldPassword)
	newPassword := strings.TrimSpace(req.NewPassword)
	if oldPassword == "" || newPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "old_password and new_password are required"})
		return
	}

	authService, err := services.NewAuthService(h.cfg)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	var (
		status       sql.NullInt64
		passwordHash sql.NullString
	)
	row := h.db.QueryRow("SELECT status, password_hash FROM app_db_users WHERE id = ? LIMIT 1", claims.UserID)
	if err := row.Scan(&status, &passwordHash); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	if status.Valid && status.Int64 == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "user disabled"})
		return
	}
	if !passwordHash.Valid || strings.TrimSpace(passwordHash.String) == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "password not set"})
		return
	}
	if !authService.VerifyPassword(passwordHash.String, oldPassword) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "old password is incorrect"})
		return
	}

	hash, err := authService.HashPassword(newPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "hash failed"})
		return
	}

	if _, err := h.db.Exec("UPDATE app_db_users SET password_hash = ?, updated_at = ? WHERE id = ?", hash, time.Now(), claims.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": claims.UserID})
}
