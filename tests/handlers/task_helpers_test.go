package handlers_test

import (
  "testing"

  "shushu-app-ui-dashboard/internal/http/handlers"
)

func TestNormalizeTaskStatus(t *testing.T) {
  status, err := handlers.NormalizeTaskStatus("")
  if err != nil || status != "open" {
    t.Fatalf("expected open, got %s, err=%v", status, err)
  }

  status, err = handlers.NormalizeTaskStatus("IN_PROGRESS")
  if err != nil || status != "in_progress" {
    t.Fatalf("expected in_progress, got %s, err=%v", status, err)
  }

  if _, err := handlers.NormalizeTaskStatus("unknown"); err == nil {
    t.Fatalf("expected error for invalid status")
  }
}

func TestNormalizeAssignedTo(t *testing.T) {
  if got := handlers.NormalizeAssignedTo(nil); got != 0 {
    t.Fatalf("expected 0, got %d", got)
  }

  zero := int64(0)
  if got := handlers.NormalizeAssignedTo(&zero); got != 0 {
    t.Fatalf("expected 0, got %d", got)
  }

  value := int64(9)
  if got := handlers.NormalizeAssignedTo(&value); got != 9 {
    t.Fatalf("expected 9, got %d", got)
  }
}
