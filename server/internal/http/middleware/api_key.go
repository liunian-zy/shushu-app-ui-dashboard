package middleware

import (
  "net/http"
  "strings"

  "github.com/gin-gonic/gin"

  "shushu-app-ui-dashboard/internal/config"
)

// RequireAPIKey validates the sync API key for online mode.
// Args:
//   cfg: App config instance.
// Returns:
//   gin.HandlerFunc: Middleware handler.
func RequireAPIKey(cfg *config.Config) gin.HandlerFunc {
  return func(c *gin.Context) {
    expected := strings.TrimSpace(cfg.SyncAPIKey)
    if expected == "" {
      c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sync api key not configured"})
      c.Abort()
      return
    }

    apiKey := strings.TrimSpace(c.GetHeader("X-API-Key"))
    if apiKey == "" {
      apiKey = strings.TrimSpace(c.Query("api_key"))
    }
    if apiKey == "" || apiKey != expected {
      c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
      c.Abort()
      return
    }

    c.Next()
  }
}
