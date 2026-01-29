package handlers_test

import (
  "testing"

  "shushu-app-ui-dashboard/internal/http/handlers"
)

func TestIsPendingTaskStatus(t *testing.T) {
  pending := []string{"open", "in_progress", "submitted", "pending_confirm"}
  for _, status := range pending {
    if !handlers.IsPendingTaskStatus(status) {
      t.Fatalf("expected pending for %s", status)
    }
  }

  notPending := []string{"confirmed", "completed", "closed", "unknown", ""}
  for _, status := range notPending {
    if handlers.IsPendingTaskStatus(status) {
      t.Fatalf("expected non-pending for %s", status)
    }
  }
}
