package handlers_test

import (
  "strings"
  "testing"

  "shushu-app-ui-dashboard/internal/http/handlers"
)

// TestBuildUploadPath verifies upload path resolution rules.
func TestBuildUploadPath(t *testing.T) {
  path, err := handlers.BuildUploadPath("", "banners", "hero.png")
  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if !strings.HasPrefix(path, "banners/file_") {
    t.Fatalf("unexpected path: %s", path)
  }
  if !strings.HasSuffix(path, ".png") {
    t.Fatalf("unexpected path: %s", path)
  }

  path, err = handlers.BuildUploadPath("custom/path.png", "banners", "hero.png")
  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if !strings.HasPrefix(path, "custom/file_") {
    t.Fatalf("unexpected path: %s", path)
  }
  if !strings.HasSuffix(path, ".png") {
    t.Fatalf("unexpected path: %s", path)
  }

  path, err = handlers.BuildUploadPath("/custom/path.png", "", "hero.png")
  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if !strings.HasPrefix(path, "custom/file_") {
    t.Fatalf("unexpected path: %s", path)
  }
  if !strings.HasSuffix(path, ".png") {
    t.Fatalf("unexpected path: %s", path)
  }

  if _, err = handlers.BuildUploadPath("", "", ""); err == nil {
    t.Fatalf("expected error for empty inputs")
  }
}
