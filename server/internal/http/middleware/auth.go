package middleware

import (
  "net/http"
  "strings"

  "github.com/gin-gonic/gin"

  "shushu-app-ui-dashboard/internal/config"
  "shushu-app-ui-dashboard/internal/services"
)

const AuthContextKey = "auth_claims"

// AuthRequired validates JWT token and stores claims in context.
// Args:
//   cfg: App config instance.
// Returns:
//   gin.HandlerFunc: Middleware handler.
func AuthRequired(cfg *config.Config) gin.HandlerFunc {
  service, err := services.NewAuthService(cfg)

  return func(c *gin.Context) {
    if err != nil {
      c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
      c.Abort()
      return
    }

    token := extractToken(c.GetHeader("Authorization"))
    if token == "" {
      c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
      c.Abort()
      return
    }

    claims, err := service.ParseToken(token)
    if err != nil {
      c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
      c.Abort()
      return
    }

    c.Set(AuthContextKey, claims)
    c.Next()
  }
}

// RequireAdmin ensures the user role is admin.
// Returns:
//   gin.HandlerFunc: Middleware handler.
func RequireAdmin() gin.HandlerFunc {
  return func(c *gin.Context) {
    claims, ok := GetAuthClaims(c)
    if !ok || !strings.EqualFold(claims.Role, "admin") {
      c.JSON(http.StatusForbidden, gin.H{"error": "admin required"})
      c.Abort()
      return
    }
    c.Next()
  }
}

// GetAuthClaims returns auth claims from context.
// Args:
//   c: Gin context.
// Returns:
//   *services.AuthClaims: Claims data.
//   bool: True when exists.
func GetAuthClaims(c *gin.Context) (*services.AuthClaims, bool) {
  raw, ok := c.Get(AuthContextKey)
  if !ok {
    return nil, false
  }
  claims, ok := raw.(*services.AuthClaims)
  return claims, ok
}

func extractToken(authHeader string) string {
  trimmed := strings.TrimSpace(authHeader)
  if trimmed == "" {
    return ""
  }
  if strings.HasPrefix(strings.ToLower(trimmed), "bearer ") {
    return strings.TrimSpace(trimmed[7:])
  }
  return trimmed
}
