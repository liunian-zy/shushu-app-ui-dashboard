package handlers

import (
  "net/http"
  "time"

  "github.com/gin-gonic/gin"
)

func Health(c *gin.Context) {
  c.JSON(http.StatusOK, gin.H{
    "status": "ok",
    "time":   time.Now().Format(time.RFC3339),
  })
}

func Ping(c *gin.Context) {
  c.JSON(http.StatusOK, gin.H{
    "message": "pong",
  })
}
