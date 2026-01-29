package handlers

import (
  "fmt"
  "strings"
)

var taskStatusAllowlist = map[string]struct{}{
  "open":            {},
  "in_progress":     {},
  "submitted":       {},
  "pending_confirm": {},
  "confirmed":       {},
  "completed":       {},
  "closed":          {},
}

// NormalizeTaskStatus normalizes and validates a task status.
// Args:
//   value: Raw status value.
// Returns:
//   string: Normalized status.
//   error: Error when status is invalid.
func NormalizeTaskStatus(value string) (string, error) {
  trimmed := strings.TrimSpace(strings.ToLower(value))
  if trimmed == "" {
    return "open", nil
  }
  if _, ok := taskStatusAllowlist[trimmed]; !ok {
    return "", fmt.Errorf("invalid status")
  }
  return trimmed, nil
}

// NormalizeAssignedTo converts a pointer to an optional id.
// Args:
//   value: Optional id.
// Returns:
//   int64: Normalized id (0 when empty).
func NormalizeAssignedTo(value *int64) int64 {
  if value == nil {
    return 0
  }
  if *value <= 0 {
    return 0
  }
  return *value
}
