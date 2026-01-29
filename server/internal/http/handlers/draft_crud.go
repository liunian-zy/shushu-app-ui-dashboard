package handlers

import (
  "database/sql"
  "encoding/json"
  "errors"
  "fmt"
  "net/http"
  "sort"
  "strconv"
  "strings"
  "time"

  "github.com/gin-gonic/gin"
)

type DraftCRUDHandler struct {
  db *sql.DB
}

type DraftKeyMode int

const (
  DraftKeyByName DraftKeyMode = iota
  DraftKeyByNameID
)

var (
  versionNameColumns = []string{
    "app_version_name",
    "location_name",
    "feishu_field_names",
    "ai_modal",
    "status",
    "draft_status",
    "submit_version",
    "last_submit_by",
    "last_submit_at",
    "confirmed_by",
    "confirmed_at",
    "sync_status",
    "sync_message",
    "synced_at",
    "target_app_version_name_id",
    "created_by",
    "updated_by",
    "created_at",
    "updated_at",
  }
  bannerColumns = []string{
    "draft_version_id",
    "title",
    "image",
    "sort",
    "is_active",
    "type",
    "app_version_name",
    "created_by",
    "updated_by",
    "created_at",
    "updated_at",
  }
  identityColumns = []string{
    "draft_version_id",
    "name",
    "image",
    "sort",
    "status",
    "app_version_name",
    "created_by",
    "updated_by",
    "created_at",
    "updated_at",
  }
  sceneColumns = []string{
    "draft_version_id",
    "name",
    "image",
    "desc",
    "music",
    "watermark_path",
    "need_watermark",
    "sort",
    "status",
    "app_version_name",
    "oss_style",
    "created_by",
    "updated_by",
    "created_at",
    "updated_at",
  }
  clothesColumns = []string{
    "draft_version_id",
    "name",
    "image",
    "sort",
    "status",
    "music",
    "desc",
    "music_text",
    "app_version_name",
    "created_by",
    "updated_by",
    "created_at",
    "updated_at",
  }
  photoHobbyColumns = []string{
    "draft_version_id",
    "name",
    "image",
    "sort",
    "status",
    "music",
    "music_text",
    "desc",
    "app_version_name",
    "created_by",
    "updated_by",
    "created_at",
    "updated_at",
  }
  extraStepColumns = []string{
    "draft_version_id",
    "app_version_name_id",
    "step_index",
    "field_name",
    "label",
    "music",
    "music_text",
    "status",
    "created_by",
    "updated_by",
    "created_at",
    "updated_at",
  }
  appUIColumns = []string{
    "draft_version_id",
    "app_version_name_id",
    "home_title_left",
    "home_title_right",
    "home_subtitle",
    "start_experience",
    "step1_music",
    "step1_music_text",
    "step1_title",
    "step2_music",
    "step2_music_text",
    "step2_title",
    "status",
    "print_wait",
    "created_by",
    "updated_by",
    "created_at",
    "updated_at",
  }
)

// NewDraftCRUDHandler creates a handler for draft CRUD.
// Args:
//   db: Database connection.
// Returns:
//   *DraftCRUDHandler: Initialized handler.
func NewDraftCRUDHandler(db *sql.DB) *DraftCRUDHandler {
  return &DraftCRUDHandler{db: db}
}

// ListVersionNames returns draft version names.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftCRUDHandler) ListVersionNames(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  rows, err := h.db.Query(
    "SELECT id, app_version_name, location_name, feishu_field_names, ai_modal, status, draft_status, submit_version, last_submit_by, last_submit_at, confirmed_by, confirmed_at FROM app_db_version_names ORDER BY id DESC",
  )
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }
  defer rows.Close()

  items := make([]gin.H, 0)
  for rows.Next() {
    var (
      id             int64
      versionName    sql.NullString
      locationName   sql.NullString
      feishuFields   sql.NullString
      aiModal        sql.NullString
      status         sql.NullInt64
      draftStatus    sql.NullString
      submitVersion  sql.NullInt64
      lastSubmitBy   sql.NullInt64
      lastSubmitAt   sql.NullTime
      confirmedBy    sql.NullInt64
      confirmedAt    sql.NullTime
    )

    if err := rows.Scan(&id, &versionName, &locationName, &feishuFields, &aiModal, &status, &draftStatus, &submitVersion, &lastSubmitBy, &lastSubmitAt, &confirmedBy, &confirmedAt); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed"})
      return
    }

    feishuList := parseFeishuFieldList(feishuFields)

    items = append(items, gin.H{
      "id":               id,
      "app_version_name": nullableString(versionName),
      "location_name":    nullableString(locationName),
      "feishu_field_names": nullableString(feishuFields),
      "feishu_field_list": feishuList,
      "ai_modal":         nullableString(aiModal),
      "status":           nullableInt(status),
      "draft_status":     nullableString(draftStatus),
      "submit_version":   nullableInt(submitVersion),
      "last_submit_by":   nullableInt(lastSubmitBy),
      "last_submit_at":   nullableTimePointer(lastSubmitAt),
      "confirmed_by":     nullableInt(confirmedBy),
      "confirmed_at":     nullableTimePointer(confirmedAt),
    })
  }

  c.JSON(http.StatusOK, gin.H{"data": items})
}

// CreateVersionName creates a new draft version name.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftCRUDHandler) CreateVersionName(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  payload, err := readPayload(c)
  if err != nil {
    return
  }

  filtered := FilterPayload(payload, versionNameColumns)

  locationName := parseStringValue(filtered["location_name"])
  appVersionName := parseStringValue(filtered["app_version_name"])
  aiModalRaw := parseStringValue(filtered["ai_modal"])
  aiModal, err := NormalizeAiModal(aiModalRaw)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ai_modal"})
    return
  }
  filtered["ai_modal"] = aiModal

  if appVersionName == "" && locationName != "" {
    appVersionName = GenerateVersionName(locationName)
  }
  if appVersionName == "" {
    c.JSON(http.StatusBadRequest, gin.H{"error": "app_version_name is required"})
    return
  }
  filtered["app_version_name"] = NormalizeVersionName(appVersionName)

  normalizedFeishu, hasFeishu := normalizeFeishuFields(filtered["feishu_field_names"])
  if hasFeishu {
    filtered["feishu_field_names"] = normalizedFeishu
  } else {
    defaults := BuildDefaultFeishuFields(aiModal)
    if raw, err := EncodeFeishuFieldNames(defaults); err == nil {
      filtered["feishu_field_names"] = raw
    }
  }

  applyTimestamps(filtered, true)

  sqlText, args, err := BuildInsertSQL("app_db_version_names", filtered)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  result, err := h.db.Exec(sqlText, args...)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "insert failed"})
    return
  }

  id, _ := result.LastInsertId()
  c.JSON(http.StatusOK, gin.H{"id": id})
}

// UpdateVersionName updates a draft version name.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftCRUDHandler) UpdateVersionName(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  id := parseInt64Param(c, "id")
  if id <= 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
    return
  }

  payload, err := readPayload(c)
  if err != nil {
    return
  }

  filtered := FilterPayload(payload, versionNameColumns)
  if len(filtered) == 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "empty payload"})
    return
  }

  if raw, ok := filtered["ai_modal"]; ok {
    aiModal := parseStringValue(raw)
    normalized, err := NormalizeAiModal(aiModal)
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ai_modal"})
      return
    }
    filtered["ai_modal"] = normalized
  }

  if raw, ok := filtered["app_version_name"]; ok {
    name := parseStringValue(raw)
    if name != "" {
      filtered["app_version_name"] = NormalizeVersionName(name)
    }
  }

  if raw, ok := filtered["feishu_field_names"]; ok {
    normalized, hasValue := normalizeFeishuFields(raw)
    if hasValue {
      filtered["feishu_field_names"] = normalized
    } else {
      delete(filtered, "feishu_field_names")
    }
  }

  applyTimestamps(filtered, false)

  sqlText, args, err := BuildUpdateSQL("app_db_version_names", "id", id, filtered)
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

// DeleteVersionName deletes a draft version name.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftCRUDHandler) DeleteVersionName(c *gin.Context) {
  h.deleteEntity(c, "app_db_version_names", "id")
}

// CreateBanner creates a new banner.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftCRUDHandler) CreateBanner(c *gin.Context) {
  h.createEntity(c, "app_db_banners", bannerColumns, DraftKeyByName)
}

// UpdateBanner updates a banner by id.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftCRUDHandler) UpdateBanner(c *gin.Context) {
  h.updateEntity(c, "app_db_banners", bannerColumns, "id")
}

// DeleteBanner deletes a banner by id.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftCRUDHandler) DeleteBanner(c *gin.Context) {
  h.deleteEntity(c, "app_db_banners", "id")
}

// CreateIdentity creates a new identity.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftCRUDHandler) CreateIdentity(c *gin.Context) {
  h.createEntity(c, "app_db_identities", identityColumns, DraftKeyByName)
}

// UpdateIdentity updates an identity by id.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftCRUDHandler) UpdateIdentity(c *gin.Context) {
  h.updateEntity(c, "app_db_identities", identityColumns, "id")
}

// DeleteIdentity deletes an identity by id.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftCRUDHandler) DeleteIdentity(c *gin.Context) {
  h.deleteEntity(c, "app_db_identities", "id")
}

// CreateScene creates a new scene.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftCRUDHandler) CreateScene(c *gin.Context) {
  h.createEntity(c, "app_db_scenes", sceneColumns, DraftKeyByName)
}

// UpdateScene updates a scene by id.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftCRUDHandler) UpdateScene(c *gin.Context) {
  h.updateEntity(c, "app_db_scenes", sceneColumns, "id")
}

// DeleteScene deletes a scene by id.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftCRUDHandler) DeleteScene(c *gin.Context) {
  h.deleteEntity(c, "app_db_scenes", "id")
}

// CreateClothesCategory creates a new clothes category.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftCRUDHandler) CreateClothesCategory(c *gin.Context) {
  h.createEntity(c, "app_db_clothes_categories", clothesColumns, DraftKeyByName)
}

// UpdateClothesCategory updates a clothes category by id.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftCRUDHandler) UpdateClothesCategory(c *gin.Context) {
  h.updateEntity(c, "app_db_clothes_categories", clothesColumns, "id")
}

// DeleteClothesCategory deletes a clothes category by id.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftCRUDHandler) DeleteClothesCategory(c *gin.Context) {
  h.deleteEntity(c, "app_db_clothes_categories", "id")
}

// CreatePhotoHobby creates a new photo hobby.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftCRUDHandler) CreatePhotoHobby(c *gin.Context) {
  h.createEntity(c, "app_db_photo_hobbies", photoHobbyColumns, DraftKeyByName)
}

// UpdatePhotoHobby updates a photo hobby by id.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftCRUDHandler) UpdatePhotoHobby(c *gin.Context) {
  h.updateEntity(c, "app_db_photo_hobbies", photoHobbyColumns, "id")
}

// DeletePhotoHobby deletes a photo hobby by id.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftCRUDHandler) DeletePhotoHobby(c *gin.Context) {
  h.deleteEntity(c, "app_db_photo_hobbies", "id")
}

// CreateConfigExtraStep creates a new config extra step.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftCRUDHandler) CreateConfigExtraStep(c *gin.Context) {
  h.createEntity(c, "app_db_config_extra_steps", extraStepColumns, DraftKeyByNameID)
}

// UpdateConfigExtraStep updates a config extra step by id.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftCRUDHandler) UpdateConfigExtraStep(c *gin.Context) {
  h.updateEntity(c, "app_db_config_extra_steps", extraStepColumns, "id")
}

// DeleteConfigExtraStep deletes a config extra step by id.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftCRUDHandler) DeleteConfigExtraStep(c *gin.Context) {
  h.deleteEntity(c, "app_db_config_extra_steps", "id")
}

// UpsertAppUIFields creates or updates app ui fields by draft key.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *DraftCRUDHandler) UpsertAppUIFields(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  payload, err := readPayload(c)
  if err != nil {
    return
  }

  filtered := FilterPayload(payload, appUIColumns)
  if err := ValidateDraftKey(filtered, DraftKeyByNameID); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  applyTimestamps(filtered, true)

  draftVersionID := parseID(filtered["draft_version_id"])
  appVersionNameID := parseID(filtered["app_version_name_id"])

  row := h.db.QueryRow(
    "SELECT id FROM app_db_app_ui_fields WHERE draft_version_id = ? OR app_version_name_id = ? LIMIT 1",
    draftVersionID,
    appVersionNameID,
  )
  var existingID int64
  if err := row.Scan(&existingID); err != nil {
    if err != sql.ErrNoRows {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
      return
    }
  }

  if existingID > 0 {
    filtered["updated_at"] = time.Now()
    sqlText, args, err := BuildUpdateSQL("app_db_app_ui_fields", "id", existingID, filtered)
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }
    if _, err := h.db.Exec(sqlText, args...); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
      return
    }
    c.JSON(http.StatusOK, gin.H{"id": existingID})
    return
  }

  sqlText, args, err := BuildInsertSQL("app_db_app_ui_fields", filtered)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  result, err := h.db.Exec(sqlText, args...)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "insert failed"})
    return
  }
  id, _ := result.LastInsertId()
  c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *DraftCRUDHandler) createEntity(c *gin.Context, table string, allowed []string, mode DraftKeyMode) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  payload, err := readPayload(c)
  if err != nil {
    return
  }

  filtered := FilterPayload(payload, allowed)
  if err := ValidateDraftKey(filtered, mode); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  applyTimestamps(filtered, true)

  sqlText, args, err := BuildInsertSQL(table, filtered)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  result, err := h.db.Exec(sqlText, args...)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "insert failed"})
    return
  }

  id, _ := result.LastInsertId()
  c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *DraftCRUDHandler) updateEntity(c *gin.Context, table string, allowed []string, idColumn string) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  id := parseInt64Param(c, "id")
  if id <= 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
    return
  }

  payload, err := readPayload(c)
  if err != nil {
    return
  }

  filtered := FilterPayload(payload, allowed)
  if len(filtered) == 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "empty payload"})
    return
  }

  applyTimestamps(filtered, false)

  sqlText, args, err := BuildUpdateSQL(table, idColumn, id, filtered)
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

func (h *DraftCRUDHandler) deleteEntity(c *gin.Context, table string, idColumn string) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  id := parseInt64Param(c, "id")
  if id <= 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
    return
  }

  result, err := h.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE %s = ?", table, idColumn), id)
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

// FilterPayload keeps only allowed fields from payload.
// Args:
//   payload: Incoming request payload.
//   allowed: Allowed field names.
// Returns:
//   map[string]interface{}: Filtered payload.
func FilterPayload(payload map[string]interface{}, allowed []string) map[string]interface{} {
  filtered := make(map[string]interface{})
  allowedSet := make(map[string]struct{}, len(allowed))
  for _, key := range allowed {
    allowedSet[key] = struct{}{}
  }
  for key, value := range payload {
    if _, ok := allowedSet[key]; ok {
      filtered[key] = value
    }
  }
  return filtered
}

// ValidateDraftKey ensures draft linkage fields exist.
// Args:
//   payload: Filtered payload.
//   mode: Draft key mode (name or name_id).
// Returns:
//   error: Error when no valid key found.
func ValidateDraftKey(payload map[string]interface{}, mode DraftKeyMode) error {
  if mode == DraftKeyByName {
    if hasID(payload["draft_version_id"]) || hasString(payload["app_version_name"]) {
      return nil
    }
    return errors.New("draft_version_id or app_version_name is required")
  }

  if hasID(payload["draft_version_id"]) || hasID(payload["app_version_name_id"]) {
    return nil
  }
  return errors.New("draft_version_id or app_version_name_id is required")
}

// BuildInsertSQL builds an INSERT statement from payload.
// Args:
//   table: Table name.
//   payload: Filtered payload.
// Returns:
//   string: SQL string.
//   []any: SQL args.
//   error: Error when payload is empty.
func BuildInsertSQL(table string, payload map[string]interface{}) (string, []any, error) {
  if len(payload) == 0 {
    return "", nil, errors.New("empty payload")
  }

  keys := make([]string, 0, len(payload))
  for key := range payload {
    keys = append(keys, key)
  }
  sort.Strings(keys)

  quotedKeys := make([]string, 0, len(keys))
  placeholders := make([]string, 0, len(keys))
  args := make([]any, 0, len(keys))
  for _, key := range keys {
    quotedKeys = append(quotedKeys, quoteSQLIdent(key))
    placeholders = append(placeholders, "?")
    args = append(args, payload[key])
  }

  sqlText := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", quoteSQLIdent(table), strings.Join(quotedKeys, ","), strings.Join(placeholders, ","))
  return sqlText, args, nil
}

// BuildUpdateSQL builds an UPDATE statement from payload.
// Args:
//   table: Table name.
//   idColumn: ID column name.
//   id: ID value.
//   payload: Filtered payload.
// Returns:
//   string: SQL string.
//   []any: SQL args.
//   error: Error when payload is empty.
func BuildUpdateSQL(table, idColumn string, id int64, payload map[string]interface{}) (string, []any, error) {
  if len(payload) == 0 {
    return "", nil, errors.New("empty payload")
  }

  keys := make([]string, 0, len(payload))
  for key := range payload {
    keys = append(keys, key)
  }
  sort.Strings(keys)

  setParts := make([]string, 0, len(keys))
  args := make([]any, 0, len(keys)+1)
  for _, key := range keys {
    setParts = append(setParts, quoteSQLIdent(key)+" = ?")
    args = append(args, payload[key])
  }
  args = append(args, id)

  sqlText := fmt.Sprintf("UPDATE %s SET %s WHERE %s = ?", quoteSQLIdent(table), strings.Join(setParts, ","), quoteSQLIdent(idColumn))
  return sqlText, args, nil
}

func quoteSQLIdent(value string) string {
  trimmed := strings.TrimSpace(value)
  if trimmed == "" {
    return value
  }
  escaped := strings.ReplaceAll(trimmed, "`", "``")
  return "`" + escaped + "`"
}

func readPayload(c *gin.Context) (map[string]interface{}, error) {
  payload := make(map[string]interface{})
  if err := c.ShouldBindJSON(&payload); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
    return nil, err
  }
  return payload, nil
}

func applyTimestamps(payload map[string]interface{}, isCreate bool) {
  now := time.Now()
  if isCreate {
    if _, ok := payload["created_at"]; !ok {
      payload["created_at"] = now
    }
  }
  payload["updated_at"] = now
}

func parseInt64Param(c *gin.Context, name string) int64 {
  raw := strings.TrimSpace(c.Param(name))
  if raw == "" {
    return 0
  }
  parsed, err := strconv.ParseInt(raw, 10, 64)
  if err != nil {
    return 0
  }
  return parsed
}

func parseID(value interface{}) int64 {
  switch raw := value.(type) {
  case int:
    return int64(raw)
  case int64:
    return raw
  case float64:
    return int64(raw)
  case json.Number:
    parsed, err := raw.Int64()
    if err != nil {
      return 0
    }
    return parsed
  case string:
    parsed, err := parseStringID(raw)
    if err != nil {
      return 0
    }
    return parsed
  default:
    return 0
  }
}

func parseStringValue(value interface{}) string {
  switch raw := value.(type) {
  case string:
    return strings.TrimSpace(raw)
  case []byte:
    return strings.TrimSpace(string(raw))
  default:
    return ""
  }
}

func normalizeFeishuFields(value interface{}) (string, bool) {
  switch raw := value.(type) {
  case string:
    trimmed := strings.TrimSpace(raw)
    if trimmed == "" {
      return "", false
    }
    if strings.HasPrefix(trimmed, "[") {
      var fields []string
      if err := json.Unmarshal([]byte(trimmed), &fields); err == nil {
        if encoded, err := EncodeFeishuFieldNames(fields); err == nil {
          return encoded, true
        }
      }
    }
    return trimmed, true
  case []string:
    if len(raw) == 0 {
      return "", false
    }
    if encoded, err := EncodeFeishuFieldNames(raw); err == nil {
      return encoded, true
    }
    return "", false
  case []interface{}:
    fields := make([]string, 0, len(raw))
    for _, item := range raw {
      if s, ok := item.(string); ok {
        trimmed := strings.TrimSpace(s)
        if trimmed != "" {
          fields = append(fields, trimmed)
        }
      }
    }
    if len(fields) == 0 {
      return "", false
    }
    if encoded, err := EncodeFeishuFieldNames(fields); err == nil {
      return encoded, true
    }
    return "", false
  default:
    return "", false
  }
}

func parseFeishuFieldList(value sql.NullString) []string {
  if !value.Valid {
    return nil
  }
  trimmed := strings.TrimSpace(value.String)
  if trimmed == "" {
    return nil
  }
  var fields []string
  if err := json.Unmarshal([]byte(trimmed), &fields); err == nil {
    return fields
  }
  return nil
}

func hasString(value interface{}) bool {
  raw, ok := value.(string)
  return ok && strings.TrimSpace(raw) != ""
}

func hasID(value interface{}) bool {
  switch raw := value.(type) {
  case int:
    return raw > 0
  case int64:
    return raw > 0
  case float64:
    return raw > 0
  case json.Number:
    parsed, _ := raw.Int64()
    return parsed > 0
  case string:
    parsed, err := parseStringID(raw)
    return err == nil && parsed > 0
  default:
    return false
  }
}

func parseStringID(raw string) (int64, error) {
  trimmed := strings.TrimSpace(raw)
  if trimmed == "" {
    return 0, errors.New("empty")
  }
  return parseStringInt64(trimmed)
}

func parseStringInt64(raw string) (int64, error) {
  return parseInt64(raw)
}

func parseInt64(raw string) (int64, error) {
  parsed, err := strconv.ParseInt(raw, 10, 64)
  if err != nil {
    return 0, err
  }
  return parsed, nil
}
