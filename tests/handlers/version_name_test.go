package handlers_test

import (
  "testing"

  "shushu-app-ui-dashboard/internal/http/handlers"
)

func TestNormalizeAiModal(t *testing.T) {
  if got, err := handlers.NormalizeAiModal(""); err != nil || got != "SD" {
    t.Fatalf("expected SD, got %s, err=%v", got, err)
  }

  if got, err := handlers.NormalizeAiModal("nano"); err != nil || got != "NANO" {
    t.Fatalf("expected NANO, got %s, err=%v", got, err)
  }

  if _, err := handlers.NormalizeAiModal("bad"); err == nil {
    t.Fatalf("expected error for invalid ai_modal")
  }
}

func TestGenerateVersionName(t *testing.T) {
  got := handlers.GenerateVersionName("博物馆")
  if got != "BOWUGUAN" {
    t.Fatalf("expected BOWUGUAN, got %s", got)
  }
}

func TestBuildDefaultFeishuFields(t *testing.T) {
  fields := handlers.BuildDefaultFeishuFields("SD")
  if len(fields) == 0 {
    t.Fatalf("expected default fields")
  }
  if fields[len(fields)-1] != "SD模式" {
    t.Fatalf("expected SD模式 appended")
  }
}
