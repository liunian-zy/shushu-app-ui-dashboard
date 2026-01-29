package handlers

import "strings"

var pendingTaskStatuses = []string{"open", "in_progress", "submitted", "pending_confirm"}

// IsPendingTaskStatus checks whether a task status should be counted as pending.
// Args:
//   status: Task status string.
// Returns:
//   bool: True when status is considered pending.
func IsPendingTaskStatus(status string) bool {
  trimmed := strings.TrimSpace(strings.ToLower(status))
  if trimmed == "" {
    return false
  }
  for _, value := range pendingTaskStatuses {
    if trimmed == value {
      return true
    }
  }
  return false
}
