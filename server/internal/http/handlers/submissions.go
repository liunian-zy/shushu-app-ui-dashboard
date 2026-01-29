package handlers

import (
  "database/sql"
  "encoding/json"
  "net/http"
  "sort"
  "strings"
  "time"

  "github.com/gin-gonic/gin"
)

type SubmissionHandler struct {
  db *sql.DB
}

type submitRequest struct {
  DraftVersionID int64           `json:"draft_version_id"`
  ModuleKey      string          `json:"module_key"`
  EntityTable    string          `json:"entity_table"`
  EntityID       int64           `json:"entity_id"`
  SubmitBy       int64           `json:"submit_by"`
  Payload        json.RawMessage `json:"payload"`
}

type confirmRequest struct {
  SubmissionID int64 `json:"submission_id"`
  ConfirmedBy  int64 `json:"confirmed_by"`
}

type DiffItem struct {
  Field string      `json:"field"`
  Old   interface{} `json:"old"`
  New   interface{} `json:"new"`
}

// NewSubmissionHandler creates a handler for submission workflows.
// Args:
//   db: Database connection.
// Returns:
//   *SubmissionHandler: Initialized handler.
func NewSubmissionHandler(db *sql.DB) *SubmissionHandler {
  return &SubmissionHandler{db: db}
}

// Submit creates a new submission snapshot.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *SubmissionHandler) Submit(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  var req submitRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
    return
  }

  req.ModuleKey = strings.TrimSpace(req.ModuleKey)
  req.EntityTable = strings.TrimSpace(req.EntityTable)
  if req.DraftVersionID <= 0 || req.ModuleKey == "" || req.EntityTable == "" || req.EntityID <= 0 || req.SubmitBy <= 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "missing required fields"})
    return
  }

  if len(req.Payload) == 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "payload is required"})
    return
  }

  payloadMap, err := decodePayload(req.Payload)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "payload must be valid json"})
    return
  }

  tx, err := h.db.Begin()
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "transaction failed"})
    return
  }
  defer func() {
    _ = tx.Rollback()
  }()

  var (
    prevID       sql.NullInt64
    prevBy       sql.NullInt64
    prevVersion  sql.NullInt64
    prevPayload  sql.NullString
  )

  prevRow := tx.QueryRow(
    "SELECT id, submit_by, submit_version, payload_json FROM app_db_submissions WHERE draft_version_id = ? AND module_key = ? AND entity_table = ? AND entity_id = ? ORDER BY submit_version DESC LIMIT 1",
    req.DraftVersionID,
    req.ModuleKey,
    req.EntityTable,
    req.EntityID,
  )
  if err := prevRow.Scan(&prevID, &prevBy, &prevVersion, &prevPayload); err != nil {
    if err != sql.ErrNoRows {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
      return
    }
  }

  submitVersion := int64(1)
  if prevVersion.Valid {
    submitVersion = prevVersion.Int64 + 1
  }

  diffItems := make([]DiffItem, 0)
  if prevPayload.Valid {
    prevMap, err := decodePayload([]byte(prevPayload.String))
    if err == nil {
      diffItems = BuildPayloadDiff(prevMap, payloadMap)
    }
  }

  needConfirm := false
  if prevBy.Valid && prevBy.Int64 != req.SubmitBy {
    needConfirm = true
  }

  diffJSON := interface{}(nil)
  if len(diffItems) > 0 {
    if raw, err := json.Marshal(diffItems); err == nil {
      diffJSON = string(raw)
    }
  }

  status := SubmissionStatus(needConfirm)

  result, err := tx.Exec(
    "INSERT INTO app_db_submissions (draft_version_id, module_key, entity_table, entity_id, submit_version, submit_by, payload_json, diff_json, need_confirm, status, prev_submission_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
    req.DraftVersionID,
    req.ModuleKey,
    req.EntityTable,
    req.EntityID,
    submitVersion,
    req.SubmitBy,
    string(req.Payload),
    diffJSON,
    needConfirm,
    status,
    nullableInt64Value(prevID),
    time.Now(),
  )
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "insert failed"})
    return
  }

  submissionID, err := result.LastInsertId()
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "insert failed"})
    return
  }

  if len(diffItems) > 0 {
    for _, diff := range diffItems {
      _, err := tx.Exec(
        "INSERT INTO app_db_field_history (draft_version_id, entity_table, entity_id, field_name, old_value, new_value, submit_id, changed_by, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
        req.DraftVersionID,
        req.EntityTable,
        req.EntityID,
        diff.Field,
        jsonValue(diff.Old),
        jsonValue(diff.New),
        submissionID,
        req.SubmitBy,
        time.Now(),
      )
      if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "history insert failed"})
        return
      }
    }
  }

  _, err = tx.Exec(
    "UPDATE app_db_tasks SET status = ?, updated_by = ?, updated_at = ? WHERE draft_version_id = ? AND module_key = ?",
    status,
    req.SubmitBy,
    time.Now(),
    req.DraftVersionID,
    req.ModuleKey,
  )
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "update tasks failed"})
    return
  }

  if err := insertTaskActions(tx, req.DraftVersionID, req.ModuleKey, "submit", req.SubmitBy, submissionID); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "update task actions failed"})
    return
  }

  if _, err := tx.Exec(
    "UPDATE app_db_version_names SET last_submit_by = ?, last_submit_at = ?, submit_version = ?, draft_status = ? WHERE id = ?",
    req.SubmitBy,
    time.Now(),
    submitVersion,
    status,
    req.DraftVersionID,
  ); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "update version failed"})
    return
  }

  auditPayload := map[string]interface{}{
    "module_key":     req.ModuleKey,
    "entity_table":   req.EntityTable,
    "entity_id":      req.EntityID,
    "submit_version": submitVersion,
    "need_confirm":   needConfirm,
  }

  if raw, err := json.Marshal(auditPayload); err == nil {
    _, _ = tx.Exec(
      "INSERT INTO app_db_audit_logs (draft_version_id, entity_table, entity_id, action, actor_id, detail_json, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
      req.DraftVersionID,
      req.EntityTable,
      req.EntityID,
      "submit",
      req.SubmitBy,
      string(raw),
      time.Now(),
    )
  }

  if err := tx.Commit(); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "commit failed"})
    return
  }

  c.JSON(http.StatusOK, gin.H{
    "submission_id": submissionID,
    "need_confirm":  needConfirm,
    "diff":          diffItems,
  })
}

// Confirm marks a submission as confirmed.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *SubmissionHandler) Confirm(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  var req confirmRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
    return
  }

  if req.SubmissionID <= 0 || req.ConfirmedBy <= 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "missing required fields"})
    return
  }

  tx, err := h.db.Begin()
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "transaction failed"})
    return
  }
  defer func() {
    _ = tx.Rollback()
  }()

  var (
    draftVersionID int64
    moduleKey      string
    entityTable    string
    entityID       int64
  )

  row := tx.QueryRow(
    "SELECT draft_version_id, module_key, entity_table, entity_id FROM app_db_submissions WHERE id = ?",
    req.SubmissionID,
  )
  if err := row.Scan(&draftVersionID, &moduleKey, &entityTable, &entityID); err != nil {
    if err == sql.ErrNoRows {
      c.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }

  result, err := tx.Exec(
    "UPDATE app_db_submissions SET status = ?, confirmed_by = ?, confirmed_at = ? WHERE id = ?",
    "confirmed",
    req.ConfirmedBy,
    time.Now(),
    req.SubmissionID,
  )
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "confirm failed"})
    return
  }

  rows, err := result.RowsAffected()
  if err != nil || rows == 0 {
    c.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
    return
  }

  _, err = tx.Exec(
    "UPDATE app_db_tasks SET status = ?, updated_by = ?, updated_at = ? WHERE draft_version_id = ? AND module_key = ?",
    "confirmed",
    req.ConfirmedBy,
    time.Now(),
    draftVersionID,
    moduleKey,
  )
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "update tasks failed"})
    return
  }

  if err := insertTaskActions(tx, draftVersionID, moduleKey, "confirm", req.ConfirmedBy, req.SubmissionID); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "update task actions failed"})
    return
  }

  pendingCount, err := countPendingConfirm(tx, draftVersionID)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query pending failed"})
    return
  }

  if pendingCount == 0 {
    _, err := tx.Exec(
      "UPDATE app_db_version_names SET draft_status = ?, confirmed_by = ?, confirmed_at = ? WHERE id = ?",
      "confirmed",
      req.ConfirmedBy,
      time.Now(),
      draftVersionID,
    )
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "update version failed"})
      return
    }
  }

  auditPayload := map[string]interface{}{
    "module_key":   moduleKey,
    "entity_table": entityTable,
    "entity_id":    entityID,
  }

  if raw, err := json.Marshal(auditPayload); err == nil {
    _, _ = tx.Exec(
      "INSERT INTO app_db_audit_logs (draft_version_id, entity_table, entity_id, action, actor_id, detail_json, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
      draftVersionID,
      entityTable,
      entityID,
      "confirm",
      req.ConfirmedBy,
      string(raw),
      time.Now(),
    )
  }

  if err := tx.Commit(); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "commit failed"})
    return
  }

  c.JSON(http.StatusOK, gin.H{"status": "confirmed"})
}

// List returns submission history.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *SubmissionHandler) List(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  draftID := parseInt64Query(c, "draft_version_id")
  moduleKey := strings.TrimSpace(c.Query("module_key"))
  entityTable := strings.TrimSpace(c.Query("entity_table"))
  entityID := parseInt64Query(c, "entity_id")

  if draftID <= 0 || moduleKey == "" || entityTable == "" {
    c.JSON(http.StatusBadRequest, gin.H{"error": "missing required fields"})
    return
  }

  query := `SELECT s.id, s.module_key, s.entity_table, s.entity_id, s.submit_version, s.submit_by, s.need_confirm, s.status,
    s.prev_submission_id, s.confirmed_by, s.confirmed_at, s.created_at, s.diff_json,
    su.display_name, su.username, cu.display_name, cu.username
    FROM app_db_submissions s
    LEFT JOIN app_db_users su ON su.id = s.submit_by
    LEFT JOIN app_db_users cu ON cu.id = s.confirmed_by
    WHERE s.draft_version_id = ? AND s.module_key = ? AND s.entity_table = ?`
  args := []interface{}{draftID, moduleKey, entityTable}

  if entityID > 0 {
    query += " AND entity_id = ?"
    args = append(args, entityID)
  }

  query += " ORDER BY s.submit_version DESC"

  rows, err := h.db.Query(query, args...)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }
  defer rows.Close()

  items := make([]gin.H, 0)
  for rows.Next() {
    var (
      id              int64
      moduleKeyValue  string
      entityTableValue string
      entityIDValue   int64
      submitVersion   int64
      submitBy        int64
      needConfirm     bool
      status          string
      prevSubmission  sql.NullInt64
      confirmedBy     sql.NullInt64
      confirmedAt     sql.NullTime
      createdAt       time.Time
      diffJSON        sql.NullString
      submitName      sql.NullString
      submitUsername  sql.NullString
      confirmName     sql.NullString
      confirmUsername sql.NullString
    )

    if err := rows.Scan(
      &id,
      &moduleKeyValue,
      &entityTableValue,
      &entityIDValue,
      &submitVersion,
      &submitBy,
      &needConfirm,
      &status,
      &prevSubmission,
      &confirmedBy,
      &confirmedAt,
      &createdAt,
      &diffJSON,
      &submitName,
      &submitUsername,
      &confirmName,
      &confirmUsername,
    ); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed"})
      return
    }

    items = append(items, gin.H{
      "id":                 id,
      "module_key":         moduleKeyValue,
      "entity_table":       entityTableValue,
      "entity_id":          entityIDValue,
      "submit_version":     submitVersion,
      "submit_by":          submitBy,
      "submit_name":        nullableStringValue(submitName),
      "submit_username":    nullableStringValue(submitUsername),
      "need_confirm":       needConfirm,
      "status":             status,
      "prev_submission_id": nullableInt64Pointer(prevSubmission),
      "confirmed_by":       nullableInt64Pointer(confirmedBy),
      "confirmed_name":     nullableStringValue(confirmName),
      "confirmed_username": nullableStringValue(confirmUsername),
      "confirmed_at":       nullableTimePointer(confirmedAt),
      "created_at":         createdAt,
      "diff":               decodeDiff(diffJSON),
    })
  }

  c.JSON(http.StatusOK, gin.H{"data": items})
}

// BuildPayloadDiff computes differences between two payloads.
// Args:
//   prev: Previous payload.
//   curr: Current payload.
// Returns:
//   []DiffItem: Field-level differences.
func BuildPayloadDiff(prev, curr map[string]interface{}) []DiffItem {
  if prev == nil {
    prev = map[string]interface{}{}
  }
  if curr == nil {
    curr = map[string]interface{}{}
  }

  keys := make(map[string]struct{})
  for key := range prev {
    keys[key] = struct{}{}
  }
  for key := range curr {
    keys[key] = struct{}{}
  }

  ordered := make([]string, 0, len(keys))
  for key := range keys {
    ordered = append(ordered, key)
  }
  sort.Strings(ordered)

  diffs := make([]DiffItem, 0)
  for _, key := range ordered {
    oldValue, oldOK := prev[key]
    newValue, newOK := curr[key]
    if oldOK && newOK {
      if deepEqual(oldValue, newValue) {
        continue
      }
    } else if !oldOK && !newOK {
      continue
    }

    diffs = append(diffs, DiffItem{
      Field: key,
      Old:   oldValue,
      New:   newValue,
    })
  }

  return diffs
}

// SubmissionStatus returns the submit status by confirm requirement.
// Args:
//   needConfirm: True when a second confirmation is needed.
// Returns:
//   string: Submission status.
func SubmissionStatus(needConfirm bool) string {
  if needConfirm {
    return "pending_confirm"
  }
  return "submitted"
}

// insertTaskActions records task actions for the given module.
// Args:
//   tx: Active transaction.
//   draftVersionID: Draft version id.
//   moduleKey: Module key for task group.
//   action: Action type (submit/confirm).
//   actorID: Actor user id.
//   submissionID: Related submission id.
// Returns:
//   error: Error when insert fails.
func insertTaskActions(tx *sql.Tx, draftVersionID int64, moduleKey, action string, actorID int64, submissionID int64) error {
  rows, err := tx.Query(
    "SELECT id FROM app_db_tasks WHERE draft_version_id = ? AND module_key = ?",
    draftVersionID,
    moduleKey,
  )
  if err != nil {
    return err
  }
  defer rows.Close()

  for rows.Next() {
    var taskID int64
    if err := rows.Scan(&taskID); err != nil {
      return err
    }

    payload := map[string]interface{}{
      "submission_id": submissionID,
      "module_key":    moduleKey,
    }

    raw, _ := json.Marshal(payload)
    _, err := tx.Exec(
      "INSERT INTO app_db_task_actions (task_id, action, actor_id, detail_json, created_at) VALUES (?, ?, ?, ?, ?)",
      taskID,
      action,
      actorID,
      string(raw),
      time.Now(),
    )
    if err != nil {
      return err
    }
  }

  return nil
}

// countPendingConfirm counts pending confirmations for a draft version.
// Args:
//   tx: Active transaction.
//   draftVersionID: Draft version id.
// Returns:
//   int64: Pending count.
//   error: Error when query fails.
func countPendingConfirm(tx *sql.Tx, draftVersionID int64) (int64, error) {
  row := tx.QueryRow(
    "SELECT COUNT(1) FROM app_db_submissions WHERE draft_version_id = ? AND status = ?",
    draftVersionID,
    "pending_confirm",
  )
  var count int64
  if err := row.Scan(&count); err != nil {
    return 0, err
  }
  return count, nil
}

func decodePayload(raw []byte) (map[string]interface{}, error) {
  var payload map[string]interface{}
  if err := json.Unmarshal(raw, &payload); err != nil {
    return nil, err
  }
  return payload, nil
}

func jsonValue(value interface{}) interface{} {
  if value == nil {
    return nil
  }
  raw, err := json.Marshal(value)
  if err != nil {
    return nil
  }
  return string(raw)
}

func nullableInt64Value(value sql.NullInt64) interface{} {
  if value.Valid {
    return value.Int64
  }
  return nil
}

func nullableInt64Pointer(value sql.NullInt64) *int64 {
  if value.Valid {
    v := value.Int64
    return &v
  }
  return nil
}

func nullableTimePointer(value sql.NullTime) *time.Time {
  if value.Valid {
    v := value.Time
    return &v
  }
  return nil
}

func decodeDiff(raw sql.NullString) []DiffItem {
  if !raw.Valid || strings.TrimSpace(raw.String) == "" {
    return nil
  }
  var diff []DiffItem
  if err := json.Unmarshal([]byte(raw.String), &diff); err != nil {
    return nil
  }
  return diff
}

func deepEqual(a, b interface{}) bool {
  aBytes, aErr := json.Marshal(a)
  bBytes, bErr := json.Marshal(b)
  if aErr != nil || bErr != nil {
    return a == b
  }
  return string(aBytes) == string(bBytes)
}
