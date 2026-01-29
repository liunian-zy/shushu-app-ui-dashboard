package handlers

import (
  "errors"
  "path"
  "path/filepath"
  "strings"

  "shushu-app-ui-dashboard/internal/config"
)

const localPathPrefix = "local://"

func isLocalPath(value string) bool {
  return strings.HasPrefix(strings.TrimSpace(value), localPathPrefix)
}

func trimLocalPrefix(value string) string {
  return strings.TrimPrefix(strings.TrimSpace(value), localPathPrefix)
}

func buildLocalURL(cfg *config.Config, relativePath string) string {
  trimmed := strings.TrimPrefix(strings.TrimSpace(relativePath), "/")
  base := strings.TrimSpace(cfg.LocalStorageBaseURL)
  if base == "" {
    base = "/api/local-files/"
  }
  if strings.HasPrefix(base, "http://") || strings.HasPrefix(base, "https://") {
    return strings.TrimSuffix(base, "/") + "/" + trimmed
  }
  return path.Join(base, trimmed)
}

func buildLocalFilePath(cfg *config.Config, relativePath string) (string, error) {
  cleaned := filepath.Clean("/" + strings.TrimSpace(relativePath))
  if strings.HasPrefix(cleaned, "/..") {
    return "", errors.New("invalid path")
  }
  relative := strings.TrimPrefix(cleaned, "/")
  return filepath.Join(cfg.LocalStorageRoot, relative), nil
}

func resolveLocalFilePath(cfg *config.Config, storagePath string) (string, error) {
  if !isLocalPath(storagePath) {
    return "", errors.New("path is not local")
  }
  relative := trimLocalPrefix(storagePath)
  if strings.TrimSpace(relative) == "" {
    return "", errors.New("path is empty")
  }
  return buildLocalFilePath(cfg, relative)
}
