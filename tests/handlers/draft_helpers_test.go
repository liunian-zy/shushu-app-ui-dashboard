package handlers_test

import (
  "testing"

  "shushu-app-ui-dashboard/internal/http/handlers"
)

// TestBuildDraftFilterByName verifies draft filter selection.
func TestBuildDraftFilterByName(t *testing.T) {
  where, args, err := handlers.BuildDraftFilterByName(12, "")
  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if where != "draft_version_id = ?" || len(args) != 1 || args[0] != int64(12) {
    t.Fatalf("unexpected filter: %s %#v", where, args)
  }

  where, args, err = handlers.BuildDraftFilterByName(0, "STANDARD")
  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if where != "app_version_name = ?" || len(args) != 1 || args[0] != "STANDARD" {
    t.Fatalf("unexpected filter: %s %#v", where, args)
  }

  if _, _, err = handlers.BuildDraftFilterByName(0, ""); err == nil {
    t.Fatalf("expected error for empty filters")
  }
}

// TestBuildDraftFilterByNameID verifies draft filter selection for name id fields.
func TestBuildDraftFilterByNameID(t *testing.T) {
  where, args, err := handlers.BuildDraftFilterByNameID(5, 0)
  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if where != "draft_version_id = ?" || len(args) != 1 || args[0] != int64(5) {
    t.Fatalf("unexpected filter: %s %#v", where, args)
  }

  where, args, err = handlers.BuildDraftFilterByNameID(0, 7)
  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if where != "app_version_name_id = ?" || len(args) != 1 || args[0] != int64(7) {
    t.Fatalf("unexpected filter: %s %#v", where, args)
  }

  if _, _, err = handlers.BuildDraftFilterByNameID(0, 0); err == nil {
    t.Fatalf("expected error for empty filters")
  }
}
