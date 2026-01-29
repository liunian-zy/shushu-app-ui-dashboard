package handlers_test

import (
  "reflect"
  "testing"

  "shushu-app-ui-dashboard/internal/http/handlers"
)

// TestBuildPayloadDiff verifies payload diff computation.
func TestBuildPayloadDiff(t *testing.T) {
  prev := map[string]interface{}{
    "title": "old",
    "count": 1.0,
    "remove": "gone",
  }
  curr := map[string]interface{}{
    "title": "new",
    "count": 1.0,
    "add":   true,
  }

  diff := handlers.BuildPayloadDiff(prev, curr)
  if len(diff) != 3 {
    t.Fatalf("expected 3 diff items, got %d", len(diff))
  }

  fields := []string{diff[0].Field, diff[1].Field, diff[2].Field}
  expected := []string{"add", "remove", "title"}
  if !reflect.DeepEqual(fields, expected) {
    t.Fatalf("unexpected fields order: %v", fields)
  }

  if diff[2].Old != "old" || diff[2].New != "new" {
    t.Fatalf("unexpected title diff: %#v", diff[2])
  }
}

func TestSubmissionStatus(t *testing.T) {
  if got := handlers.SubmissionStatus(true); got != "pending_confirm" {
    t.Fatalf("expected pending_confirm, got %s", got)
  }
  if got := handlers.SubmissionStatus(false); got != "submitted" {
    t.Fatalf("expected submitted, got %s", got)
  }
}
