package handlers

import (
  "fmt"
  "io"
  "net/http"
  "os"
  "path/filepath"
  "strings"
  "time"

  "github.com/gin-gonic/gin"

  "shushu-app-ui-dashboard/internal/config"
)

type LocalFileHandler struct {
  cfg *config.Config
}

// NewLocalFileHandler creates a handler for local file operations.
// Args:
//   cfg: App config instance.
// Returns:
//   *LocalFileHandler: Initialized handler.
func NewLocalFileHandler(cfg *config.Config) *LocalFileHandler {
  return &LocalFileHandler{cfg: cfg}
}

// Upload stores a file to local storage.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *LocalFileHandler) Upload(c *gin.Context) {
  file, header, err := c.Request.FormFile("file")
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
    return
  }
  defer func() {
    _ = file.Close()
  }()

  moduleKey := strings.TrimSpace(c.PostForm("module_key"))
  draftVersionID := parseInt64Query(c, "draft_version_id")
  if draftVersionID == 0 {
    draftVersionID = parseInt64Value(c.PostForm("draft_version_id"))
  }

  relativePath, err := buildLocalUploadPath(moduleKey, draftVersionID, header.Filename)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  absPath, err := buildLocalFilePath(h.cfg, relativePath)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "mkdir failed"})
    return
  }

  output, err := os.Create(absPath)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "save failed"})
    return
  }
  defer func() {
    _ = output.Close()
  }()

  if _, err := io.Copy(output, file); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "write failed"})
    return
  }

  url := buildLocalURL(h.cfg, relativePath)
  storagePath := localPathPrefix + relativePath

  c.JSON(http.StatusOK, gin.H{
    "path":      storagePath,
    "url":       url,
    "file_name": header.Filename,
  })
}

// Serve returns a local file content.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *LocalFileHandler) Serve(c *gin.Context) {
  raw := strings.TrimPrefix(c.Param("path"), "/")
  absPath, err := buildLocalFilePath(h.cfg, raw)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path"})
    return
  }
  if _, err := os.Stat(absPath); err != nil {
    c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
    return
  }
  c.File(absPath)
}

func buildLocalUploadPath(moduleKey string, draftVersionID int64, filename string) (string, error) {
  cleanedModule := sanitizePathSegment(moduleKey)
  if cleanedModule == "" {
    cleanedModule = "misc"
  }
  versionPart := "0"
  if draftVersionID > 0 {
    versionPart = formatInt64(draftVersionID)
  }
  base := filepath.Base(strings.TrimSpace(filename))
  if base == "" {
    return "", fmt.Errorf("filename is required")
  }
  ext := filepath.Ext(base)
  if ext == "" {
    ext = ".bin"
  }
  timestamp := time.Now().Format("20060102150405")
  newName := fmt.Sprintf("file_%s_%s%s", timestamp, randomSuffix(8), ext)
  return filepath.ToSlash(filepath.Join("drafts", versionPart, cleanedModule, newName)), nil
}

func parseInt64Value(raw string) int64 {
  trimmed := strings.TrimSpace(raw)
  if trimmed == "" {
    return 0
  }
  value, err := parseInt64(trimmed)
  if err != nil {
    return 0
  }
  return value
}
