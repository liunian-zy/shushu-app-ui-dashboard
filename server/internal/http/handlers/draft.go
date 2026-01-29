package handlers

import (
  "database/sql"
  "net/http"
  "os"
  "strconv"
  "strings"

  "github.com/gin-gonic/gin"
  "github.com/redis/go-redis/v9"

  "shushu-app-ui-dashboard/internal/config"
  "shushu-app-ui-dashboard/internal/services"
)

type DraftHandler struct {
  cfg   *config.Config
  db    *sql.DB
  redis *redis.Client
}

type submissionSummary struct {
  status    sql.NullString
  createdAt sql.NullTime
}

// NewDraftHandler creates a draft handler for listing draft data.
// Args:
//   cfg: App config instance.
//   db: Database connection.
//   redis: Redis client for OSS cache.
// Returns:
//   *DraftHandler: Initialized handler.
func NewDraftHandler(cfg *config.Config, db *sql.DB, redis *redis.Client) *DraftHandler {
  return &DraftHandler{cfg: cfg, db: db, redis: redis}
}

// ListBanners returns draft banners with signed URLs.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftHandler) ListBanners(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  draftID, appVersionName := parseDraftFilters(c)
  where, args, err := BuildDraftFilterByName(draftID, appVersionName)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  rows, err := h.db.Query(
    "SELECT id, title, image, sort, is_active, type, app_version_name FROM app_db_banners WHERE "+where+" ORDER BY sort ASC, id ASC",
    args...,
  )
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }
  defer rows.Close()

  ossService := h.newOSSService()
  submissionMap, err := h.loadLatestSubmissions(draftID, "banners", "app_db_banners")
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }

  items := make([]gin.H, 0)
  for rows.Next() {
    var (
      id              uint64
      title           sql.NullString
      image           sql.NullString
      sort            sql.NullInt64
      isActive        sql.NullInt64
      bannerType      sql.NullInt64
      appVersionField sql.NullString
    )

    if err := rows.Scan(&id, &title, &image, &sort, &isActive, &bannerType, &appVersionField); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed"})
      return
    }

    imageURL := signPath(h.cfg, ossService, nullableString(image), "")
    summary := submissionMap[int64(id)]

    items = append(items, gin.H{
      "id":               id,
      "title":            nullableString(title),
      "image":            nullableString(image),
      "image_url":        imageURL,
      "sort":             nullableInt(sort),
      "is_active":        nullableInt(isActive),
      "type":             nullableInt(bannerType),
      "app_version_name": nullableString(appVersionField),
      "submit_status":    nullableStringValue(summary.status),
      "last_submit_at":   nullableTimePointer(summary.createdAt),
    })
  }

  c.JSON(http.StatusOK, gin.H{"data": items})
}

// ListIdentities returns draft identities with signed URLs.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftHandler) ListIdentities(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  draftID, appVersionName := parseDraftFilters(c)
  where, args, err := BuildDraftFilterByName(draftID, appVersionName)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  rows, err := h.db.Query(
    "SELECT id, name, image, sort, status, app_version_name FROM app_db_identities WHERE "+where+" ORDER BY sort ASC, id ASC",
    args...,
  )
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }
  defer rows.Close()

  ossService := h.newOSSService()
  submissionMap, err := h.loadLatestSubmissions(draftID, "identities", "app_db_identities")
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }

  items := make([]gin.H, 0)
  for rows.Next() {
    var (
      id              int
      name            sql.NullString
      image           sql.NullString
      sort            sql.NullInt64
      status          sql.NullInt64
      appVersionField sql.NullString
    )

    if err := rows.Scan(&id, &name, &image, &sort, &status, &appVersionField); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed"})
      return
    }

    imageURL := signPath(h.cfg, ossService, nullableString(image), "")
    summary := submissionMap[int64(id)]

    items = append(items, gin.H{
      "id":               id,
      "name":             nullableString(name),
      "image":            nullableString(image),
      "image_url":        imageURL,
      "sort":             nullableInt(sort),
      "status":           nullableInt(status),
      "app_version_name": nullableString(appVersionField),
      "submit_status":    nullableStringValue(summary.status),
      "last_submit_at":   nullableTimePointer(summary.createdAt),
    })
  }

  c.JSON(http.StatusOK, gin.H{"data": items})
}

// ListScenes returns draft scenes with signed URLs.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftHandler) ListScenes(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  draftID, appVersionName := parseDraftFilters(c)
  where, args, err := BuildDraftFilterByName(draftID, appVersionName)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  rows, err := h.db.Query(
    "SELECT id, name, image, `desc`, music, sort, status, app_version_name FROM app_db_scenes WHERE "+where+" ORDER BY sort ASC, id ASC",
    args...,
  )
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }
  defer rows.Close()

  ossService := h.newOSSService()
  submissionMap, err := h.loadLatestSubmissions(draftID, "scenes", "app_db_scenes")
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }

  items := make([]gin.H, 0)
  for rows.Next() {
    var (
      id              int
      name            sql.NullString
      image           sql.NullString
      desc            sql.NullString
      music           sql.NullString
      sort            sql.NullInt64
      status          sql.NullInt64
      appVersionField sql.NullString
    )

    if err := rows.Scan(&id, &name, &image, &desc, &music, &sort, &status, &appVersionField); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed"})
      return
    }

    imageURL := signPath(h.cfg, ossService, nullableString(image), "")
    musicURL := signPath(h.cfg, ossService, nullableString(music), "")
    summary := submissionMap[int64(id)]

    items = append(items, gin.H{
      "id":               id,
      "name":             nullableString(name),
      "image":            nullableString(image),
      "image_url":        imageURL,
      "desc":             nullableString(desc),
      "music":            nullableString(music),
      "music_url":        musicURL,
      "sort":             nullableInt(sort),
      "status":           nullableInt(status),
      "app_version_name": nullableString(appVersionField),
      "submit_status":    nullableStringValue(summary.status),
      "last_submit_at":   nullableTimePointer(summary.createdAt),
    })
  }

  c.JSON(http.StatusOK, gin.H{"data": items})
}

// ListClothesCategories returns draft clothes categories with signed URLs.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftHandler) ListClothesCategories(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  draftID, appVersionName := parseDraftFilters(c)
  where, args, err := BuildDraftFilterByName(draftID, appVersionName)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  rows, err := h.db.Query(
    "SELECT id, name, image, sort, status, music, `desc`, music_text, app_version_name FROM app_db_clothes_categories WHERE "+where+" ORDER BY sort ASC, id ASC",
    args...,
  )
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }
  defer rows.Close()

  ossService := h.newOSSService()
  submissionMap, err := h.loadLatestSubmissions(draftID, "clothes_categories", "app_db_clothes_categories")
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }

  items := make([]gin.H, 0)
  for rows.Next() {
    var (
      id              int
      name            sql.NullString
      image           sql.NullString
      sort            sql.NullInt64
      status          sql.NullInt64
      music           sql.NullString
      desc            sql.NullString
      musicText       sql.NullString
      appVersionField sql.NullString
    )

    if err := rows.Scan(&id, &name, &image, &sort, &status, &music, &desc, &musicText, &appVersionField); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed"})
      return
    }

    imageURL := signPath(h.cfg, ossService, nullableString(image), "")
    musicURL := signPath(h.cfg, ossService, nullableString(music), "")
    summary := submissionMap[int64(id)]

    items = append(items, gin.H{
      "id":               id,
      "name":             nullableString(name),
      "image":            nullableString(image),
      "image_url":        imageURL,
      "sort":             nullableInt(sort),
      "status":           nullableInt(status),
      "music":            nullableString(music),
      "music_url":        musicURL,
      "desc":             nullableString(desc),
      "music_text":       nullableString(musicText),
      "app_version_name": nullableString(appVersionField),
      "submit_status":    nullableStringValue(summary.status),
      "last_submit_at":   nullableTimePointer(summary.createdAt),
    })
  }

  c.JSON(http.StatusOK, gin.H{"data": items})
}

// ListPhotoHobbies returns draft photo hobbies with signed URLs.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftHandler) ListPhotoHobbies(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  draftID, appVersionName := parseDraftFilters(c)
  where, args, err := BuildDraftFilterByName(draftID, appVersionName)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  rows, err := h.db.Query(
    "SELECT id, name, image, sort, status, music, music_text, `desc`, app_version_name FROM app_db_photo_hobbies WHERE "+where+" ORDER BY sort ASC, id ASC",
    args...,
  )
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }
  defer rows.Close()

  ossService := h.newOSSService()
  submissionMap, err := h.loadLatestSubmissions(draftID, "photo_hobbies", "app_db_photo_hobbies")
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }

  items := make([]gin.H, 0)
  for rows.Next() {
    var (
      id              int
      name            sql.NullString
      image           sql.NullString
      sort            sql.NullInt64
      status          sql.NullInt64
      music           sql.NullString
      musicText       sql.NullString
      desc            sql.NullString
      appVersionField sql.NullString
    )

    if err := rows.Scan(&id, &name, &image, &sort, &status, &music, &musicText, &desc, &appVersionField); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed"})
      return
    }

    imageURL := signPath(h.cfg, ossService, nullableString(image), "")
    musicURL := signPath(h.cfg, ossService, nullableString(music), "")
    summary := submissionMap[int64(id)]

    items = append(items, gin.H{
      "id":               id,
      "name":             nullableString(name),
      "image":            nullableString(image),
      "image_url":        imageURL,
      "sort":             nullableInt(sort),
      "status":           nullableInt(status),
      "music":            nullableString(music),
      "music_url":        musicURL,
      "music_text":       nullableString(musicText),
      "desc":             nullableString(desc),
      "app_version_name": nullableString(appVersionField),
      "submit_status":    nullableStringValue(summary.status),
      "last_submit_at":   nullableTimePointer(summary.createdAt),
    })
  }

  c.JSON(http.StatusOK, gin.H{"data": items})
}

// GetAppUIFields returns draft UI fields with signed URLs.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftHandler) GetAppUIFields(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  draftID := parseInt64Query(c, "draft_version_id")
  appVersionNameID := parseInt64Query(c, "app_version_name_id")
  where, args, err := BuildDraftFilterByNameID(draftID, appVersionNameID)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  row := h.db.QueryRow(
    "SELECT id, app_version_name_id, home_title_left, home_title_right, home_subtitle, start_experience, step1_music, step1_music_text, step1_title, step2_music, step2_music_text, step2_title, status, print_wait FROM app_db_app_ui_fields WHERE "+where+" LIMIT 1",
    args...,
  )

  var (
    id               uint64
    appVersionNameIDValue sql.NullInt64
    homeTitleLeft    sql.NullString
    homeTitleRight   sql.NullString
    homeSubtitle     sql.NullString
    startExperience  sql.NullString
    step1Music       sql.NullString
    step1MusicText   sql.NullString
    step1Title       sql.NullString
    step2Music       sql.NullString
    step2MusicText   sql.NullString
    step2Title       sql.NullString
    status           sql.NullInt64
    printWait        sql.NullString
  )

  if err := row.Scan(
    &id,
    &appVersionNameIDValue,
    &homeTitleLeft,
    &homeTitleRight,
    &homeSubtitle,
    &startExperience,
    &step1Music,
    &step1MusicText,
    &step1Title,
    &step2Music,
    &step2MusicText,
    &step2Title,
    &status,
    &printWait,
  ); err != nil {
    if err == sql.ErrNoRows {
      c.JSON(http.StatusOK, gin.H{"data": nil})
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }

  ossService := h.newOSSService()
  submissionMap, err := h.loadLatestSubmissions(draftID, "app_ui_fields", "app_db_app_ui_fields")
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }
  summary := submissionMap[int64(id)]

  data := gin.H{
    "id":                  id,
    "app_version_name_id": nullableInt(appVersionNameIDValue),
    "home_title_left":     nullableString(homeTitleLeft),
    "home_title_right":    nullableString(homeTitleRight),
    "home_subtitle":       nullableString(homeSubtitle),
    "start_experience":    nullableString(startExperience),
    "step1_music":         nullableString(step1Music),
    "step1_music_url":     signPath(h.cfg, ossService, nullableString(step1Music), ""),
    "step1_music_text":    nullableString(step1MusicText),
    "step1_title":         nullableString(step1Title),
    "step2_music":         nullableString(step2Music),
    "step2_music_url":     signPath(h.cfg, ossService, nullableString(step2Music), ""),
    "step2_music_text":    nullableString(step2MusicText),
    "step2_title":         nullableString(step2Title),
    "status":              nullableInt(status),
    "print_wait":          nullableString(printWait),
    "print_wait_url":      signPath(h.cfg, ossService, nullableString(printWait), ""),
    "submit_status":       nullableStringValue(summary.status),
    "last_submit_at":      nullableTimePointer(summary.createdAt),
  }

  c.JSON(http.StatusOK, gin.H{"data": data})
}

// ListConfigExtraSteps returns draft extra steps with signed URLs.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftHandler) ListConfigExtraSteps(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  draftID := parseInt64Query(c, "draft_version_id")
  appVersionNameID := parseInt64Query(c, "app_version_name_id")
  where, args, err := BuildDraftFilterByNameID(draftID, appVersionNameID)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  rows, err := h.db.Query(
    "SELECT id, app_version_name_id, step_index, field_name, label, music, music_text, status FROM app_db_config_extra_steps WHERE "+where+" ORDER BY step_index ASC, id ASC",
    args...,
  )
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }
  defer rows.Close()

  ossService := h.newOSSService()
  submissionMap, err := h.loadLatestSubmissions(draftID, "config_extra_steps", "app_db_config_extra_steps")
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }

  items := make([]gin.H, 0)
  for rows.Next() {
    var (
      id              int
      appVersionNameIDValue sql.NullInt64
      stepIndex       sql.NullInt64
      fieldName       sql.NullString
      label           sql.NullString
      music           sql.NullString
      musicText       sql.NullString
      status          sql.NullInt64
    )

    if err := rows.Scan(&id, &appVersionNameIDValue, &stepIndex, &fieldName, &label, &music, &musicText, &status); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed"})
      return
    }
    summary := submissionMap[int64(id)]

    items = append(items, gin.H{
      "id":                  id,
      "app_version_name_id": nullableInt(appVersionNameIDValue),
      "step_index":          nullableInt(stepIndex),
      "field_name":          nullableString(fieldName),
      "label":               nullableString(label),
      "music":               nullableString(music),
      "music_url":           signPath(h.cfg, ossService, nullableString(music), ""),
      "music_text":          nullableString(musicText),
      "status":              nullableInt(status),
      "submit_status":       nullableStringValue(summary.status),
      "last_submit_at":      nullableTimePointer(summary.createdAt),
    })
  }

  c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *DraftHandler) loadLatestSubmissions(draftVersionID int64, moduleKey, entityTable string) (map[int64]submissionSummary, error) {
  if h.db == nil || draftVersionID <= 0 {
    return nil, nil
  }

  rows, err := h.db.Query(
    "SELECT entity_id, status, created_at FROM app_db_submissions WHERE draft_version_id = ? AND module_key = ? AND entity_table = ? ORDER BY created_at DESC, id DESC",
    draftVersionID,
    moduleKey,
    entityTable,
  )
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  summary := make(map[int64]submissionSummary)
  for rows.Next() {
    var (
      entityID int64
      status   sql.NullString
      created  sql.NullTime
    )
    if err := rows.Scan(&entityID, &status, &created); err != nil {
      return nil, err
    }
    if _, ok := summary[entityID]; ok {
      continue
    }
    summary[entityID] = submissionSummary{
      status:    status,
      createdAt: created,
    }
  }

  return summary, rows.Err()
}

// newOSSService builds an OSS service instance for signing URLs.
// Args:
//   None.
// Returns:
//   *services.OSSService: OSS service or nil if init fails.
func (h *DraftHandler) newOSSService() *services.OSSService {
  service, err := services.NewOSSService(h.cfg, h.redis)
  if err != nil {
    return nil
  }
  return service
}

// parseDraftFilters extracts draft_version_id and app_version_name from the query.
// Args:
//   c: Gin context.
// Returns:
//   int64: Draft version id value.
//   string: App version name value.
func parseDraftFilters(c *gin.Context) (int64, string) {
  return parseInt64Query(c, "draft_version_id"), strings.TrimSpace(c.Query("app_version_name"))
}

// parseInt64Query parses an int64 query parameter.
// Args:
//   c: Gin context.
//   name: Query parameter name.
// Returns:
//   int64: Parsed value or 0 if missing/invalid.
func parseInt64Query(c *gin.Context, name string) int64 {
  raw := strings.TrimSpace(c.Query(name))
  if raw == "" {
    return 0
  }
  value, err := strconv.ParseInt(raw, 10, 64)
  if err != nil {
    return 0
  }
  return value
}

// nullableString converts sql.NullString to pointer string.
// Args:
//   value: sql.NullString value.
// Returns:
//   *string: Pointer or nil when invalid.
func nullableString(value sql.NullString) *string {
  if value.Valid {
    return &value.String
  }
  return nil
}

// nullableStringValue converts sql.NullString to string.
// Args:
//   value: sql.NullString value.
// Returns:
//   string: Value or empty string.
func nullableStringValue(value sql.NullString) string {
  if value.Valid {
    return value.String
  }
  return ""
}

// nullableInt converts sql.NullInt64 to pointer int.
// Args:
//   value: sql.NullInt64 value.
// Returns:
//   *int: Pointer or nil when invalid.
func nullableInt(value sql.NullInt64) *int {
  if value.Valid {
    v := int(value.Int64)
    return &v
  }
  return nil
}

// signPath generates a signed or local URL for a storage path.
// Args:
//   cfg: App config instance.
//   service: OSS service instance.
//   pathValue: Storage path value.
//   style: Optional OSS style process string.
// Returns:
//   *string: URL or nil if not available.
func signPath(cfg *config.Config, service *services.OSSService, pathValue *string, style string) *string {
  if pathValue == nil {
    return nil
  }
  trimmed := strings.TrimSpace(*pathValue)
  if trimmed == "" {
    return nil
  }
  if isLocalPath(trimmed) {
    url := buildLocalURL(cfg, trimLocalPrefix(trimmed))
    return &url
  }
  if !strings.Contains(trimmed, "://") {
    if abs, err := buildLocalFilePath(cfg, trimmed); err == nil {
      if _, err := os.Stat(abs); err == nil {
        url := buildLocalURL(cfg, trimmed)
        return &url
      }
    }
  }
  if service == nil {
    return nil
  }
  signed, err := service.GetSignedURL(trimmed, false, style)
  if err != nil {
    return nil
  }
  return &signed
}
