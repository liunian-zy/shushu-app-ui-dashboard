package handlers

import (
  "fmt"
  "path"
  "path/filepath"
  "strings"
  "time"

  "github.com/gin-gonic/gin"
  "github.com/redis/go-redis/v9"

  "shushu-app-ui-dashboard/internal/config"
  "shushu-app-ui-dashboard/internal/services"
)

type OSSHandler struct {
  cfg   *config.Config
  redis *redis.Client
}

type preSignRequest struct {
  Filename string `json:"filename"`
  Module   string `json:"module"`
  Path     string `json:"path"`
  Expires  int64  `json:"expires"`
}

type signURLRequest struct {
  Path        string `json:"path"`
  Style       string `json:"style"`
  UseInternal bool   `json:"use_internal"`
}

// NewOSSHandler creates a handler for OSS signing endpoints.
// Args:
//   cfg: App config instance.
//   redis: Redis client for caching.
// Returns:
//   *OSSHandler: Initialized handler.
func NewOSSHandler(cfg *config.Config, redis *redis.Client) *OSSHandler {
  return &OSSHandler{cfg: cfg, redis: redis}
}

// PreSign returns a pre-signed upload URL and storage path.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *OSSHandler) PreSign(c *gin.Context) {
  var req preSignRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(400, gin.H{"error": "invalid request"})
    return
  }

  uploadPath, err := BuildUploadPath(req.Path, req.Module, req.Filename)
  if err != nil {
    c.JSON(400, gin.H{"error": err.Error()})
    return
  }

  service, err := services.NewOSSService(h.cfg, h.redis)
  if err != nil {
    c.JSON(500, gin.H{"error": "oss service init failed"})
    return
  }

  preSignedURL, err := service.GetUploadPreSignedURL(uploadPath, req.Expires)
  if err != nil {
    c.JSON(500, gin.H{"error": "generate pre-signed url failed"})
    return
  }

  c.JSON(200, gin.H{
    "upload_pre_url": preSignedURL,
    "path":           uploadPath,
    "expires":        req.Expires,
  })
}

// SignURL returns a signed URL for an existing OSS path.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *OSSHandler) SignURL(c *gin.Context) {
  var req signURLRequest
  if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Path) == "" {
    c.JSON(400, gin.H{"error": "invalid request"})
    return
  }

  service, err := services.NewOSSService(h.cfg, h.redis)
  if err != nil {
    c.JSON(500, gin.H{"error": "oss service init failed"})
    return
  }

  signedURL, err := service.GetSignedURL(req.Path, req.UseInternal, req.Style)
  if err != nil {
    c.JSON(500, gin.H{"error": "generate signed url failed"})
    return
  }

  c.JSON(200, gin.H{
    "signed_url": signedURL,
  })
}

// BuildUploadPath generates an OSS path by priority order: path, module+filename, or filename.
// Args:
//   rawPath: Explicit path value.
//   module: Module name used as directory prefix.
//   filename: File name for upload.
// Returns:
//   string: Resolved upload path.
//   error: Error when path cannot be resolved.
func BuildUploadPath(rawPath, module, filename string) (string, error) {
  cleaned := strings.TrimSpace(rawPath)
  if cleaned != "" {
    cleaned = strings.TrimPrefix(cleaned, "/")
    dir := path.Dir(cleaned)
    ext := filepath.Ext(cleaned)
    if ext == "" {
      ext = filepath.Ext(filename)
    }
    if ext == "" {
      ext = ".bin"
    }
    timestamp := time.Now().Format("20060102150405")
    randomized := fmt.Sprintf("file_%s_%s%s", timestamp, randomSuffix(8), ext)
    if dir == "." || dir == "/" {
      return randomized, nil
    }
    return path.Join(dir, randomized), nil
  }

  filename = strings.TrimSpace(filename)
  if filename == "" {
    return "", fmt.Errorf("filename is required")
  }
  ext := filepath.Ext(filename)
  if ext == "" {
    ext = ".bin"
  }
  timestamp := time.Now().Format("20060102150405")
  randomized := fmt.Sprintf("file_%s_%s%s", timestamp, randomSuffix(8), ext)

  module = strings.TrimSpace(module)
  if module == "" {
    return randomized, nil
  }

  return path.Join(module, randomized), nil
}
