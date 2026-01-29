package handlers_test

import (
  "reflect"
  "testing"

  "shushu-app-ui-dashboard/internal/http/handlers"
)

func TestFilterPayload(t *testing.T) {
  payload := map[string]interface{}{
    "name": "博物馆",
    "status": 1,
    "extra": "ignored",
  }
  filtered := handlers.FilterPayload(payload, []string{"name", "status"})
  if len(filtered) != 2 {
    t.Fatalf("expected 2 fields, got %d", len(filtered))
  }
  if filtered["name"] != "博物馆" || filtered["status"] != 1 {
    t.Fatalf("unexpected filtered payload: %#v", filtered)
  }
}

func TestValidateDraftKey(t *testing.T) {
  payload := map[string]interface{}{ "app_version_name": "STANDARD" }
  if err := handlers.ValidateDraftKey(payload, handlers.DraftKeyByName); err != nil {
    t.Fatalf("unexpected error: %v", err)
  }

  payload = map[string]interface{}{ "draft_version_id": 12 }
  if err := handlers.ValidateDraftKey(payload, handlers.DraftKeyByName); err != nil {
    t.Fatalf("unexpected error: %v", err)
  }

  payload = map[string]interface{}{ "app_version_name_id": 9 }
  if err := handlers.ValidateDraftKey(payload, handlers.DraftKeyByNameID); err != nil {
    t.Fatalf("unexpected error: %v", err)
  }

  if err := handlers.ValidateDraftKey(map[string]interface{}{}, handlers.DraftKeyByName); err == nil {
    t.Fatalf("expected error for missing key")
  }
}

func TestBuildInsertSQL(t *testing.T) {
  payload := map[string]interface{}{ "name": "a", "status": 1 }
  sqlText, args, err := handlers.BuildInsertSQL("app_db_demo", payload)
  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if sqlText != "INSERT INTO `app_db_demo` (`name`,`status`) VALUES (?,?)" {
    t.Fatalf("unexpected sql: %s", sqlText)
  }
  expected := []any{"a", 1}
  if !reflect.DeepEqual(args, expected) {
    t.Fatalf("unexpected args: %#v", args)
  }
}

func TestBuildUpdateSQL(t *testing.T) {
  payload := map[string]interface{}{ "name": "a", "status": 1 }
  sqlText, args, err := handlers.BuildUpdateSQL("app_db_demo", "id", 5, payload)
  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if sqlText != "UPDATE `app_db_demo` SET `name` = ?,`status` = ? WHERE `id` = ?" {
    t.Fatalf("unexpected sql: %s", sqlText)
  }
  expected := []any{"a", 1, int64(5)}
  if !reflect.DeepEqual(args, expected) {
    t.Fatalf("unexpected args: %#v", args)
  }
}
