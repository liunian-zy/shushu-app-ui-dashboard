package services_test

import (
  "strings"
  "testing"

  "shushu-app-ui-dashboard/internal/services"
)

func TestComputeTargetPath(t *testing.T) {
  result := services.ComputeTargetPath("banners/hero.jpg", "png")
  if !strings.HasPrefix(result, "banners/file_") {
    t.Fatalf("unexpected prefix: %s", result)
  }
  if !strings.HasSuffix(result, ".png") {
    t.Fatalf("unexpected suffix: %s", result)
  }
}

func TestValidateMediaRule(t *testing.T) {
  rule := &services.MediaRule{
    MaxSizeKB:    100,
    MaxWidth:     800,
    MaxHeight:    600,
    AllowFormats: "jpg,png",
  }
  meta := &services.MediaMeta{
    SizeBytes: 200 * 1024,
    Width:     1200,
    Height:    600,
    Format:    "jpg",
  }

  violations := services.ValidateMediaRule(rule, meta)
  if len(violations) == 0 {
    t.Fatalf("expected violations")
  }
}
