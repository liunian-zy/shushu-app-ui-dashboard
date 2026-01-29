package handlers

import "fmt"

// BuildDraftFilterByName builds a SQL WHERE clause for tables that use draft_version_id or app_version_name.
// Args:
//   draftVersionID: The draft version id value.
//   appVersionName: The app version name value.
// Returns:
//   whereClause: SQL WHERE clause without the leading keyword.
//   args: SQL arguments matching the clause.
//   err: Error when neither filter is provided.
func BuildDraftFilterByName(draftVersionID int64, appVersionName string) (string, []any, error) {
  if draftVersionID > 0 {
    return "draft_version_id = ?", []any{draftVersionID}, nil
  }
  if appVersionName != "" {
    return "app_version_name = ?", []any{appVersionName}, nil
  }
  return "", nil, fmt.Errorf("draft_version_id or app_version_name is required")
}

// BuildDraftFilterByNameID builds a SQL WHERE clause for tables that use draft_version_id or app_version_name_id.
// Args:
//   draftVersionID: The draft version id value.
//   appVersionNameID: The app version name id value.
// Returns:
//   whereClause: SQL WHERE clause without the leading keyword.
//   args: SQL arguments matching the clause.
//   err: Error when neither filter is provided.
func BuildDraftFilterByNameID(draftVersionID int64, appVersionNameID int64) (string, []any, error) {
  if draftVersionID > 0 {
    return "draft_version_id = ?", []any{draftVersionID}, nil
  }
  if appVersionNameID > 0 {
    return "app_version_name_id = ?", []any{appVersionNameID}, nil
  }
  return "", nil, fmt.Errorf("draft_version_id or app_version_name_id is required")
}
