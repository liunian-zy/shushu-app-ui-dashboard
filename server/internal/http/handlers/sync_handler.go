package handlers

import (
  "bytes"
  "context"
  "database/sql"
  "encoding/json"
  "errors"
  "io"
  "net/http"
  "os"
  "strings"
  "time"

  "github.com/gin-gonic/gin"

  "shushu-app-ui-dashboard/internal/config"
)

type SyncHandler struct {
  cfg    *config.Config
  db     *sql.DB
  client *http.Client
}

type syncRequest struct {
  DraftVersionID int64 `json:"draft_version_id"`
  TriggerBy      int64 `json:"trigger_by"`
  Confirm        bool  `json:"confirm"`
  Modules        []string `json:"modules"`
}

// NewSyncHandler creates a handler for sync flow.
// Args:
//   db: Database connection.
// Returns:
//   *SyncHandler: Initialized handler.
func NewSyncHandler(cfg *config.Config, db *sql.DB) *SyncHandler {
  timeout := time.Duration(cfg.SyncTimeoutSeconds) * time.Second
  if timeout <= 0 {
    timeout = 20 * time.Second
  }
  return &SyncHandler{
    cfg: cfg,
    db:  db,
    client: &http.Client{
      Timeout: timeout,
    },
  }
}

// Sync validates and pushes draft data to the online sync API.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *SyncHandler) Sync(c *gin.Context) {
  if strings.ToLower(strings.TrimSpace(h.cfg.AppMode)) == "online" {
    c.JSON(http.StatusNotFound, gin.H{"error": "not available in online mode"})
    return
  }

  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  var req syncRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
    return
  }

  if req.DraftVersionID <= 0 || req.TriggerBy <= 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "draft_version_id and trigger_by are required"})
    return
  }

  if strings.TrimSpace(h.cfg.SyncTargetURL) == "" {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sync target not configured"})
    return
  }
  if strings.TrimSpace(h.cfg.SyncAPIKey) == "" {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sync api key not configured"})
    return
  }

  draftVersion, err := loadDraftVersion(h.db, req.DraftVersionID)
  if err != nil {
    if err == sql.ErrNoRows {
      c.JSON(http.StatusNotFound, gin.H{"error": "draft version not found"})
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }

  appVersionName := strings.TrimSpace(nullableStringValue(draftVersion.AppVersionName))
  locationName := strings.TrimSpace(nullableStringValue(draftVersion.LocationName))

  invalidModules := findInvalidModules(req.Modules)
  if len(invalidModules) > 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_modules", "modules": invalidModules})
    return
  }
  modules := normalizeModules(req.Modules)

  draftData, err := loadDraftData(h.db, req.DraftVersionID)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }
  payload := buildSyncValidationPayload(draftData, appVersionName, locationName)
  validationErrors := ValidateSyncPayload(payload, modules)
  if len(validationErrors) > 0 {
    _ = updateDraftSyncStatus(h.db, req.DraftVersionID, "failed", "validation_failed", 0)
    c.JSON(http.StatusBadRequest, gin.H{
      "error":   "validation_failed",
      "details": validationErrors,
    })
    return
  }

  now := time.Now()
  jobID, moduleJobs, err := h.startSyncJob(req.DraftVersionID, req.TriggerBy, modules, now)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "sync job failed"})
    return
  }

  _ = updateDraftSyncStatus(h.db, req.DraftVersionID, "running", "", 0)

  uploadCache, err := h.uploadDraftModules(req.DraftVersionID, modules)
  if err != nil {
    _ = updateDraftSyncStatus(h.db, req.DraftVersionID, "failed", err.Error(), 0)
    _ = h.finishSyncJobWithError(jobID, moduleJobs, "failed", err.Error(), now)
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }

  draftData, err = loadDraftData(h.db, req.DraftVersionID)
  if err != nil {
    _ = updateDraftSyncStatus(h.db, req.DraftVersionID, "failed", "query failed", 0)
    _ = h.finishSyncJobWithError(jobID, moduleJobs, "failed", "query failed", now)
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }

  req.Modules = modules
  pushPayload := buildSyncPushFromDraft(req, draftVersion, draftData)
  result, err := h.pushToRemote(c.Request.Context(), pushPayload)
  if err != nil {
    if pushErr := asSyncPushError(err); pushErr != nil {
      if pushErr.NeedConfirm {
        _ = updateDraftSyncStatus(h.db, req.DraftVersionID, "pending_confirm", pushErr.Message, pushErr.TargetID)
        _ = h.finishSyncJobWithError(jobID, moduleJobs, "pending_confirm", pushErr.Message, now)
        c.JSON(http.StatusConflict, gin.H{
          "need_confirm":               true,
          "reason":                     pushErr.Message,
          "target_app_version_name_id": pushErr.TargetID,
        })
        return
      }
      _ = updateDraftSyncStatus(h.db, req.DraftVersionID, "failed", pushErr.Message, pushErr.TargetID)
      _ = h.finishSyncJobWithError(jobID, moduleJobs, "failed", pushErr.Message, now)
      if len(pushErr.Details) > 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": pushErr.Message, "details": pushErr.Details})
      } else {
        c.JSON(http.StatusBadRequest, gin.H{"error": pushErr.Message})
      }
      return
    }
    _ = updateDraftSyncStatus(h.db, req.DraftVersionID, "failed", err.Error(), 0)
    _ = h.finishSyncJobWithError(jobID, moduleJobs, "failed", err.Error(), now)
    c.JSON(http.StatusInternalServerError, gin.H{"error": "sync failed"})
    return
  }

  if err := h.finishSyncSuccess(req, result.TargetID, draftData, jobID, moduleJobs, now); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "sync failed"})
    return
  }

  for _, entry := range uploadCache {
    if entry.localAbs != "" {
      _ = os.Remove(entry.localAbs)
    }
  }

  c.JSON(http.StatusOK, gin.H{
    "status":                     "synced",
    "draft_version_id":           req.DraftVersionID,
    "target_app_version_name_id": result.TargetID,
  })
}

type syncPushResult struct {
  TargetID int64
}

type syncPushError struct {
  StatusCode  int
  Message     string
  NeedConfirm bool
  TargetID    int64
  Details     []SyncValidationError
}

func (e *syncPushError) Error() string {
  return e.Message
}

func asSyncPushError(err error) *syncPushError {
  var target *syncPushError
  if errors.As(err, &target) {
    return target
  }
  return nil
}

func (h *SyncHandler) pushToRemote(ctx context.Context, payload SyncPushRequest) (*syncPushResult, error) {
  body, err := json.Marshal(payload)
  if err != nil {
    return nil, err
  }

  req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimSpace(h.cfg.SyncTargetURL), bytes.NewReader(body))
  if err != nil {
    return nil, err
  }
  req.Header.Set("Content-Type", "application/json")
  if apiKey := strings.TrimSpace(h.cfg.SyncAPIKey); apiKey != "" {
    req.Header.Set("X-API-Key", apiKey)
  }

  resp, err := h.client.Do(req)
  if err != nil {
    return nil, err
  }
  defer func() {
    _ = resp.Body.Close()
  }()

  payloadBytes, _ := io.ReadAll(resp.Body)
  var parsed struct {
    Status    string                `json:"status"`
    TargetID  int64                 `json:"target_app_version_name_id"`
    NeedConfirm bool                `json:"need_confirm"`
    Reason    string                `json:"reason"`
    Error     string                `json:"error"`
    Details   []SyncValidationError `json:"details"`
  }
  if len(payloadBytes) > 0 {
    _ = json.Unmarshal(payloadBytes, &parsed)
  }

  if resp.StatusCode >= http.StatusMultipleChoices {
    message := strings.TrimSpace(parsed.Error)
    if message == "" {
      message = strings.TrimSpace(parsed.Reason)
    }
    if message == "" {
      message = "sync failed"
    }
    return nil, &syncPushError{
      StatusCode:  resp.StatusCode,
      Message:     message,
      NeedConfirm: parsed.NeedConfirm,
      TargetID:    parsed.TargetID,
      Details:     parsed.Details,
    }
  }

  return &syncPushResult{TargetID: parsed.TargetID}, nil
}

func (h *SyncHandler) startSyncJob(draftVersionID, triggerBy int64, modules []string, now time.Time) (int64, map[string]int64, error) {
  tx, err := h.db.Begin()
  if err != nil {
    return 0, nil, err
  }
  defer func() {
    _ = tx.Rollback()
  }()
  jobID, err := insertSyncJob(tx, draftVersionID, triggerBy, "running", now)
  if err != nil {
    return 0, nil, err
  }
  moduleJobs := make(map[string]int64)
  for _, moduleKey := range resolveSyncModules(modules) {
    moduleID, err := insertSyncModuleJob(tx, draftVersionID, triggerBy, moduleKey, "running", now)
    if err != nil {
      return 0, nil, err
    }
    moduleJobs[moduleKey] = moduleID
  }
  if err := tx.Commit(); err != nil {
    return 0, nil, err
  }
  return jobID, moduleJobs, nil
}

func (h *SyncHandler) finishSyncJobWithError(jobID int64, moduleJobs map[string]int64, status, message string, now time.Time) error {
  tx, err := h.db.Begin()
  if err != nil {
    return err
  }
  defer func() {
    _ = tx.Rollback()
  }()
  if err := failSyncJob(tx, jobID, message, now); err != nil {
    return err
  }
  for _, moduleID := range moduleJobs {
    if err := failSyncModuleJob(tx, moduleID, status, message, now); err != nil {
      return err
    }
  }
  return tx.Commit()
}

func (h *SyncHandler) finishSyncSuccess(req syncRequest, targetID int64, data draftData, jobID int64, moduleJobs map[string]int64, now time.Time) error {
  tx, err := h.db.Begin()
  if err != nil {
    return err
  }
  defer func() {
    _ = tx.Rollback()
  }()

  if err := updateDraftSyncStatusTx(tx, req.DraftVersionID, "synced", "", targetID, now); err != nil {
    return err
  }

  auditPayload := buildSyncAuditPayload(data)
  if raw, err := json.Marshal(auditPayload); err == nil {
    _, _ = tx.Exec(
      "INSERT INTO app_db_audit_logs (draft_version_id, entity_table, entity_id, action, actor_id, detail_json, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
      req.DraftVersionID,
      "sync",
      targetID,
      "sync",
      req.TriggerBy,
      string(raw),
      now,
    )
  }

  if err := finishSyncJob(tx, jobID, now); err != nil {
    return err
  }
  for _, moduleID := range moduleJobs {
    if err := finishSyncModuleJob(tx, moduleID, now); err != nil {
      return err
    }
  }

  return tx.Commit()
}
