package middleware

import (
  "crypto/rand"
  "encoding/hex"

  "github.com/gin-gonic/gin"
)

func RequestID() gin.HandlerFunc {
  return func(c *gin.Context) {
    if c.GetHeader("X-Request-Id") == "" {
      c.Header("X-Request-Id", newID())
    }
    c.Next()
  }
}

func newID() string {
  buf := make([]byte, 8)
  if _, err := rand.Read(buf); err != nil {
    return "fallback"
  }
  return hex.EncodeToString(buf)
}
