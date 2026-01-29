package handlers

import (
  "database/sql"
  "database/sql/driver"
  "errors"
  "fmt"
  "net/http"
  "os"
  "path/filepath"
  "strings"
  "time"

  "github.com/gin-gonic/gin"
  "github.com/redis/go-redis/v9"

  "shushu-app-ui-dashboard/internal/config"
  "shushu-app-ui-dashboard/internal/services"
)

type MediaHandler struct {
  cfg   *config.Config
  db    *sql.DB
  redis *redis.Client
}

var errRuleNotFound = errors.New("rule not found")

type mediaRuleRequest struct {
  ModuleKey      string `json:"module_key"`
  MediaType      string `json:"media_type"`
  MaxSizeKB      int64  `json:"max_size_kb"`
  MinWidth       int64  `json:"min_width"`
  MaxWidth       int64  `json:"max_width"`
  MinHeight      int64  `json:"min_height"`
  MaxHeight      int64  `json:"max_height"`
  RatioWidth     int64  `json:"ratio_width"`
  RatioHeight    int64  `json:"ratio_height"`
  MinDurationMS  int64  `json:"min_duration_ms"`
  MaxDurationMS  int64  `json:"max_duration_ms"`
  AllowFormats   string `json:"allow_formats"`
  ResizeMode     string `json:"resize_mode"`
  TargetFormat   string `json:"target_format"`
  CompressQuality int64 `json:"compress_quality"`
  Status         int64  `json:"status"`
  CreatedBy      int64  `json:"created_by"`
  UpdatedBy      int64  `json:"updated_by"`
}

type mediaValidateRequest struct {
  ModuleKey   string `json:"module_key"`
  MediaType   string `json:"media_type"`
  Path        string `json:"path"`
  RuleID      int64  `json:"rule_id"`
  Rule        *mediaRuleOverride `json:"rule"`
}

type mediaTransformRequest struct {
  DraftVersionID int64  `json:"draft_version_id"`
  ModuleKey      string `json:"module_key"`
  MediaType      string `json:"media_type"`
  Path           string `json:"path"`
  RuleID         int64  `json:"rule_id"`
  TargetPath     string `json:"target_path"`
  OperatorID     int64  `json:"operator_id"`
  Rule           *mediaRuleOverride `json:"rule"`
}

type mediaRuleOverride struct {
  MaxSizeKB      int64  `json:"max_size_kb"`
  MinWidth       int64  `json:"min_width"`
  MaxWidth       int64  `json:"max_width"`
  MinHeight      int64  `json:"min_height"`
  MaxHeight      int64  `json:"max_height"`
  RatioWidth     int64  `json:"ratio_width"`
  RatioHeight    int64  `json:"ratio_height"`
  MinDurationMS  int64  `json:"min_duration_ms"`
  MaxDurationMS  int64  `json:"max_duration_ms"`
  AllowFormats   string `json:"allow_formats"`
  ResizeMode     string `json:"resize_mode"`
  TargetFormat   string `json:"target_format"`
  CompressQuality int64 `json:"compress_quality"`
}

var mediaRuleColumns = []string{
  "module_key",
  "media_type",
  "max_size_kb",
  "min_width",
  "max_width",
  "min_height",
  "max_height",
  "ratio_width",
  "ratio_height",
  "min_duration_ms",
  "max_duration_ms",
  "allow_formats",
  "resize_mode",
  "target_format",
  "compress_quality",
  "status",
  "created_by",
  "updated_by",
  "created_at",
  "updated_at",
}

// NewMediaHandler creates a handler for media validation and rules.
// Args:
//   cfg: App config instance.
//   db: Database connection.
//   redis: Redis client for OSS cache.
// Returns:
//   *MediaHandler: Initialized handler.
func NewMediaHandler(cfg *config.Config, db *sql.DB, redis *redis.Client) *MediaHandler {
  return &MediaHandler{cfg: cfg, db: db, redis: redis}
}

// ListRules returns media rules.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *MediaHandler) ListRules(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  moduleKey := strings.TrimSpace(c.Query("module_key"))
  mediaType := strings.TrimSpace(c.Query("media_type"))

  query := "SELECT id, module_key, media_type, max_size_kb, min_width, max_width, min_height, max_height, ratio_width, ratio_height, min_duration_ms, max_duration_ms, allow_formats, resize_mode, target_format, compress_quality, status FROM app_db_media_rules WHERE 1=1"
  args := make([]any, 0)

  if moduleKey != "" {
    query += " AND module_key = ?"
    args = append(args, moduleKey)
  }
  if mediaType != "" {
    query += " AND media_type = ?"
    args = append(args, mediaType)
  }

  query += " ORDER BY id DESC"

  rows, err := h.db.Query(query, args...)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }
  defer rows.Close()

  items := make([]gin.H, 0)
  for rows.Next() {
    rule := services.MediaRule{}
    var (
      allowFormats sql.NullString
      resizeMode   sql.NullString
      targetFormat sql.NullString
      status       sql.NullInt64
    )

    if err := rows.Scan(
      &rule.ID,
      &rule.ModuleKey,
      &rule.MediaType,
      &rule.MaxSizeKB,
      &rule.MinWidth,
      &rule.MaxWidth,
      &rule.MinHeight,
      &rule.MaxHeight,
      &rule.RatioWidth,
      &rule.RatioHeight,
      &rule.MinDurationMS,
      &rule.MaxDurationMS,
      &allowFormats,
      &resizeMode,
      &targetFormat,
      &rule.CompressQuality,
      &status,
    ); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed"})
      return
    }

    rule.AllowFormats = nullableStringValue(allowFormats)
    rule.ResizeMode = nullableStringValue(resizeMode)
    rule.TargetFormat = nullableStringValue(targetFormat)

    items = append(items, gin.H{
      "id":               rule.ID,
      "module_key":       rule.ModuleKey,
      "media_type":       rule.MediaType,
      "max_size_kb":      rule.MaxSizeKB,
      "min_width":        rule.MinWidth,
      "max_width":        rule.MaxWidth,
      "min_height":       rule.MinHeight,
      "max_height":       rule.MaxHeight,
      "ratio_width":      rule.RatioWidth,
      "ratio_height":     rule.RatioHeight,
      "min_duration_ms":  rule.MinDurationMS,
      "max_duration_ms":  rule.MaxDurationMS,
      "allow_formats":    rule.AllowFormats,
      "resize_mode":      rule.ResizeMode,
      "target_format":    rule.TargetFormat,
      "compress_quality": rule.CompressQuality,
      "status":           nullableInt(status),
    })
  }

  c.JSON(http.StatusOK, gin.H{"data": items})
}

// CreateRule creates a media rule.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *MediaHandler) CreateRule(c *gin.Context) {
  h.createOrUpdateRule(c, 0)
}

// UpdateRule updates a media rule.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *MediaHandler) UpdateRule(c *gin.Context) {
  id := parseInt64Param(c, "id")
  if id <= 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
    return
  }
  h.createOrUpdateRule(c, id)
}

// DeleteRule deletes a media rule.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *MediaHandler) DeleteRule(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  id := parseInt64Param(c, "id")
  if id <= 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
    return
  }

  result, err := h.db.Exec("DELETE FROM app_db_media_rules WHERE id = ?", id)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
    return
  }

  rows, err := result.RowsAffected()
  if err != nil || rows == 0 {
    c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
    return
  }

  c.JSON(http.StatusOK, gin.H{"id": id})
}

// Validate validates media against a rule.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *MediaHandler) Validate(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  var req mediaValidateRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
    return
  }

  req.MediaType = strings.TrimSpace(req.MediaType)
  if req.MediaType == "" {
    c.JSON(http.StatusBadRequest, gin.H{"error": "media_type is required"})
    return
  }

  req.Path = strings.TrimSpace(req.Path)
  if req.Path == "" {
    c.JSON(http.StatusBadRequest, gin.H{"error": "path is required"})
    return
  }

  var (
    rule        *services.MediaRule
    err         error
    ruleMissing bool
  )
  if req.Rule != nil {
    rule = buildRuleOverride(req.Rule, req.ModuleKey, req.MediaType)
  } else {
    rule, err = h.fetchRule(req.RuleID, req.ModuleKey, req.MediaType)
    if err != nil {
      if errors.Is(err, errRuleNotFound) {
        ruleMissing = true
      } else {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
      }
    }
  }

  localPath, _, isLocal, err := resolveMediaLocalPath(h.cfg, req.Path)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  var ossService *services.OSSService
  if !isLocal {
    ossService, err = services.NewOSSService(h.cfg, h.redis)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "oss init failed"})
      return
    }
  }

  mediaService, err := services.NewMediaService(ossService)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }

  var cleanup func()
  if !isLocal {
    localPath, cleanup, err = mediaService.DownloadToTemp(req.Path)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "download failed"})
      return
    }
    defer cleanup()
  }

  meta, err := mediaService.Probe(localPath)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "probe failed"})
    return
  }

  violations := services.ValidateMediaRule(rule, meta)
  warning := ""
  if ruleMissing {
    warning = "no_rule"
  }

  c.JSON(http.StatusOK, gin.H{
    "path":         req.Path,
    "rule":         formatMediaRule(rule),
    "meta":         meta,
    "valid":        len(violations) == 0,
    "violations":   violations,
    "warning":      warning,
    "rule_missing": ruleMissing,
  })
}

// Transform validates and transforms media, then uploads to OSS.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *MediaHandler) Transform(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  var req mediaTransformRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
    return
  }

  if req.DraftVersionID <= 0 || strings.TrimSpace(req.ModuleKey) == "" {
    c.JSON(http.StatusBadRequest, gin.H{"error": "draft_version_id and module_key are required"})
    return
  }

  req.MediaType = strings.TrimSpace(req.MediaType)
  if req.MediaType == "" {
    c.JSON(http.StatusBadRequest, gin.H{"error": "media_type is required"})
    return
  }

  req.Path = strings.TrimSpace(req.Path)
  if req.Path == "" {
    c.JSON(http.StatusBadRequest, gin.H{"error": "path is required"})
    return
  }

  var rule *services.MediaRule
  var err error
  if req.Rule != nil {
    rule = buildRuleOverride(req.Rule, req.ModuleKey, req.MediaType)
  } else {
    rule, err = h.fetchRule(req.RuleID, req.ModuleKey, req.MediaType)
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }
  }

  localPath, relativePath, isLocal, err := resolveMediaLocalPath(h.cfg, req.Path)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  var ossService *services.OSSService
  if !isLocal {
    ossService, err = services.NewOSSService(h.cfg, h.redis)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "oss init failed"})
      return
    }
  }

  mediaService, err := services.NewMediaService(ossService)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }

  var cleanup func()
  if !isLocal {
    localPath, cleanup, err = mediaService.DownloadToTemp(req.Path)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "download failed"})
      return
    }
    defer cleanup()
  }

  outputPath := strings.TrimSpace(req.TargetPath)
  if outputPath == "" {
    basePath := req.Path
    if isLocal {
      basePath = relativePath
    }
    outputPath = services.ComputeTargetPath(basePath, rule.TargetFormat)
    if isLocal {
      outputPath = localPathPrefix + outputPath
    }
  }

  var storedOutputPath string
  var tempOutput string
  if isLocal {
    storedOutputPath = outputPath
    outputRelative := outputPath
    if isLocalPath(outputRelative) {
      outputRelative = trimLocalPrefix(outputRelative)
    }
    localOutput, err := buildLocalFilePath(h.cfg, outputRelative)
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }
    if err := os.MkdirAll(filepath.Dir(localOutput), 0755); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "mkdir failed"})
      return
    }
    tempOutput = localOutput
  } else {
    storedOutputPath = outputPath
    tempOutput = filepath.Join(os.TempDir(), fmt.Sprintf("media-out-%d%s", time.Now().UnixNano(), filepath.Ext(outputPath)))
    defer func() { _ = os.Remove(tempOutput) }()
  }

  if err := mediaService.Transform(localPath, tempOutput, req.MediaType, rule); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "transform failed"})
    return
  }

  meta, err := mediaService.Probe(tempOutput)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "probe failed"})
    return
  }

  violations := services.ValidateMediaRule(rule, meta)
  if len(violations) > 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "rule violated after transform", "violations": violations})
    return
  }

  if !isLocal {
    if err := ossService.UploadFileFromPath(storedOutputPath, tempOutput); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
      return
    }
  }

  assetID, err := h.ensureAsset(req.DraftVersionID, req.ModuleKey, req.MediaType, req.Path, meta, req.OperatorID)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "asset save failed", "detail": err.Error()})
    return
  }

  versionID, err := h.insertMediaVersion(assetID, storedOutputPath, meta, rule)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "version save failed"})
    return
  }

  previewURL := ""
  if isLocal {
    previewURL = buildLocalURL(h.cfg, trimLocalPrefix(storedOutputPath))
  } else if ossService != nil {
    if signed, err := ossService.GetSignedURL(storedOutputPath, false, ""); err == nil {
      previewURL = signed
    }
  }

  c.JSON(http.StatusOK, gin.H{
    "asset_id":   assetID,
    "version_id": versionID,
    "path":       storedOutputPath,
    "url":        previewURL,
    "meta":       meta,
  })
}

// createOrUpdateRule creates or updates a media rule by id.
// Args:
//   c: Gin context.
//   id: Rule id (0 for create).
// Returns:
//   None.
func (h *MediaHandler) createOrUpdateRule(c *gin.Context, id int64) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  payload, err := readPayload(c)
  if err != nil {
    return
  }

  filtered := FilterPayload(payload, mediaRuleColumns)
  if id == 0 {
    if !hasString(filtered["module_key"]) || !hasString(filtered["media_type"]) {
      c.JSON(http.StatusBadRequest, gin.H{"error": "module_key and media_type are required"})
      return
    }
  }

  applyTimestamps(filtered, id == 0)

  if id == 0 {
    sqlText, args, err := BuildInsertSQL("app_db_media_rules", filtered)
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }
    result, err := h.db.Exec(sqlText, args...)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "insert failed"})
      return
    }
    newID, _ := result.LastInsertId()
    c.JSON(http.StatusOK, gin.H{"id": newID})
    return
  }

  sqlText, args, err := BuildUpdateSQL("app_db_media_rules", "id", id, filtered)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  result, err := h.db.Exec(sqlText, args...)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
    return
  }
  rows, err := result.RowsAffected()
  if err != nil || rows == 0 {
    c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
    return
  }
  c.JSON(http.StatusOK, gin.H{"id": id})
}

// fetchRule loads a media rule by id or module/media type.
// Args:
//   ruleID: Rule id to load.
//   moduleKey: Module key filter.
//   mediaType: Media type filter.
// Returns:
//   *services.MediaRule: Matched rule.
//   error: Error when not found or query fails.
func (h *MediaHandler) fetchRule(ruleID int64, moduleKey, mediaType string) (*services.MediaRule, error) {
  if ruleID <= 0 && strings.TrimSpace(moduleKey) == "" {
    return nil, fmt.Errorf("module_key or rule_id is required")
  }

  query := "SELECT id, module_key, media_type, max_size_kb, min_width, max_width, min_height, max_height, ratio_width, ratio_height, min_duration_ms, max_duration_ms, allow_formats, resize_mode, target_format, compress_quality, status FROM app_db_media_rules WHERE "
  args := make([]any, 0)
  if ruleID > 0 {
    query += "id = ?"
    args = append(args, ruleID)
  } else {
    query += "module_key = ? AND media_type = ? AND status = 1"
    args = append(args, moduleKey, mediaType)
  }

  row := h.db.QueryRow(query+" ORDER BY id DESC LIMIT 1", args...)
  rule := &services.MediaRule{}
  var (
    allowFormats sql.NullString
    resizeMode   sql.NullString
    targetFormat sql.NullString
    status       sql.NullInt64
  )

  if err := row.Scan(
    &rule.ID,
    &rule.ModuleKey,
    &rule.MediaType,
    &rule.MaxSizeKB,
    &rule.MinWidth,
    &rule.MaxWidth,
    &rule.MinHeight,
    &rule.MaxHeight,
    &rule.RatioWidth,
    &rule.RatioHeight,
    &rule.MinDurationMS,
    &rule.MaxDurationMS,
    &allowFormats,
    &resizeMode,
    &targetFormat,
    &rule.CompressQuality,
    &status,
  ); err != nil {
    if err == sql.ErrNoRows {
      return nil, errRuleNotFound
    }
    return nil, err
  }

  rule.AllowFormats = nullableStringValue(allowFormats)
  rule.ResizeMode = nullableStringValue(resizeMode)
  rule.TargetFormat = nullableStringValue(targetFormat)

  return rule, nil
}

func buildRuleOverride(override *mediaRuleOverride, moduleKey, mediaType string) *services.MediaRule {
  if override == nil {
    return nil
  }
  return &services.MediaRule{
    ModuleKey:       strings.TrimSpace(moduleKey),
    MediaType:       strings.TrimSpace(mediaType),
    MaxSizeKB:       override.MaxSizeKB,
    MinWidth:        override.MinWidth,
    MaxWidth:        override.MaxWidth,
    MinHeight:       override.MinHeight,
    MaxHeight:       override.MaxHeight,
    RatioWidth:      override.RatioWidth,
    RatioHeight:     override.RatioHeight,
    MinDurationMS:   override.MinDurationMS,
    MaxDurationMS:   override.MaxDurationMS,
    AllowFormats:    strings.TrimSpace(override.AllowFormats),
    ResizeMode:      strings.TrimSpace(override.ResizeMode),
    TargetFormat:    strings.TrimSpace(override.TargetFormat),
    CompressQuality: override.CompressQuality,
  }
}

func formatMediaRule(rule *services.MediaRule) gin.H {
  if rule == nil {
    return nil
  }
  return gin.H{
    "id":               rule.ID,
    "module_key":       rule.ModuleKey,
    "media_type":       rule.MediaType,
    "max_size_kb":      rule.MaxSizeKB,
    "min_width":        rule.MinWidth,
    "max_width":        rule.MaxWidth,
    "min_height":       rule.MinHeight,
    "max_height":       rule.MaxHeight,
    "ratio_width":      rule.RatioWidth,
    "ratio_height":     rule.RatioHeight,
    "min_duration_ms":  rule.MinDurationMS,
    "max_duration_ms":  rule.MaxDurationMS,
    "allow_formats":    rule.AllowFormats,
    "resize_mode":      rule.ResizeMode,
    "target_format":    rule.TargetFormat,
    "compress_quality": rule.CompressQuality,
  }
}

func resolveMediaLocalPath(cfg *config.Config, storagePath string) (string, string, bool, error) {
  trimmed := strings.TrimSpace(storagePath)
  if trimmed == "" {
    return "", "", false, errors.New("path is empty")
  }

  if isLocalPath(trimmed) {
    relative := trimLocalPrefix(trimmed)
    abs, err := buildLocalFilePath(cfg, relative)
    if err != nil {
      return "", "", false, err
    }
    return abs, relative, true, nil
  }

  if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") || strings.Contains(trimmed, "://") {
    return "", "", false, nil
  }

  abs, err := buildLocalFilePath(cfg, trimmed)
  if err != nil {
    return "", "", false, err
  }
  if _, err := os.Stat(abs); err != nil {
    if os.IsNotExist(err) {
      return "", "", false, nil
    }
    return "", "", false, err
  }
  return abs, trimmed, true, nil
}

// ensureAsset creates a media asset if not exists.
// Args:
//   draftVersionID: Draft version id.
//   moduleKey: Module key.
//   mediaType: Media type.
//   path: Media path.
//   meta: Media metadata.
//   operatorID: Operator user id.
// Returns:
//   int64: Asset id.
//   error: Error when database operations fail.
func (h *MediaHandler) ensureAsset(draftVersionID int64, moduleKey, mediaType, path string, meta *services.MediaMeta, operatorID int64) (int64, error) {
  var lastErr error
  for attempt := 0; attempt < 2; attempt++ {
    row := h.db.QueryRow(
      "SELECT id FROM app_db_media_assets WHERE draft_version_id = ? AND module_key = ? AND media_type = ? AND file_url = ? LIMIT 1",
      draftVersionID,
      moduleKey,
      mediaType,
      path,
    )

    var assetID int64
    if err := row.Scan(&assetID); err == nil {
      return assetID, nil
    } else if !errors.Is(err, sql.ErrNoRows) {
      if errors.Is(err, driver.ErrBadConn) {
        lastErr = err
        continue
      }
      return 0, err
    }

  normalizedFormat := normalizeMediaFormat(meta)
  result, err := h.db.Exec(
    "INSERT INTO app_db_media_assets (draft_version_id, module_key, media_type, file_url, file_name, file_size, width, height, duration_ms, format, status, created_by, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
    draftVersionID,
    moduleKey,
    mediaType,
    path,
    filepath.Base(path),
    meta.SizeBytes,
    meta.Width,
    meta.Height,
    meta.DurationMS,
    normalizedFormat,
    "active",
    operatorID,
    time.Now(),
  )
    if err != nil {
      if errors.Is(err, driver.ErrBadConn) {
        lastErr = err
        continue
      }
      return 0, err
    }

    newID, err := result.LastInsertId()
    if err != nil {
      if errors.Is(err, driver.ErrBadConn) {
        lastErr = err
        continue
      }
      return 0, err
    }

    return newID, nil
  }

  if lastErr != nil {
    return 0, lastErr
  }
  return 0, errors.New("asset save failed")
}

// insertMediaVersion stores a new media version record.
// Args:
//   assetID: Asset id.
//   path: Output path.
//   meta: Media metadata.
//   rule: Media rule used for transform.
// Returns:
//   int64: Version id.
//   error: Error when database operations fail.
func (h *MediaHandler) insertMediaVersion(assetID int64, path string, meta *services.MediaMeta, rule *services.MediaRule) (int64, error) {
  row := h.db.QueryRow("SELECT COALESCE(MAX(version_no), 0) FROM app_db_media_versions WHERE asset_id = ?", assetID)
  var current int64
  if err := row.Scan(&current); err != nil {
    return 0, err
  }

  profile := "manual"
  if rule != nil && rule.ID > 0 {
    profile = fmt.Sprintf("rule_%d", rule.ID)
  }

  normalizedFormat := normalizeMediaFormat(meta)
  result, err := h.db.Exec(
    "INSERT INTO app_db_media_versions (asset_id, version_no, file_url, file_size, width, height, duration_ms, format, compress_profile, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
    assetID,
    current+1,
    path,
    meta.SizeBytes,
    meta.Width,
    meta.Height,
    meta.DurationMS,
    normalizedFormat,
    profile,
    time.Now(),
  )
  if err != nil {
    return 0, err
  }
  newID, err := result.LastInsertId()
  if err != nil {
    return 0, err
  }
  return newID, nil
}

func normalizeMediaFormat(meta *services.MediaMeta) string {
  if meta == nil {
    return ""
  }
  format := strings.TrimSpace(meta.FileExt)
  if format == "" {
    format = strings.TrimSpace(meta.Format)
  }
  if len(format) > 20 {
    return format[:20]
  }
  return format
}
