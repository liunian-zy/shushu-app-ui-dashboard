package handlers

import (
  "database/sql"
  "errors"
  "net/http"
  "strings"
  "time"

  "github.com/gin-gonic/gin"

  "shushu-app-ui-dashboard/internal/http/middleware"
)

type TTSPreset struct {
  ID           int64
  Name         string
  VoiceID      string
  EmotionName  string
  Volume       int
  Speed        float64
  Pitch        int
  Stability    int
  Similarity   int
  Exaggeration int
  Status       int
  IsDefault    int
  CreatedAt    sql.NullTime
  UpdatedAt    sql.NullTime
  CreatedBy    sql.NullInt64
  UpdatedBy    sql.NullInt64
}

type ttsPresetPayload struct {
  Name         *string  `json:"name"`
  VoiceID      *string  `json:"voice_id"`
  EmotionName  *string  `json:"emotion_name"`
  Volume       *int     `json:"volume"`
  Speed        *float64 `json:"speed"`
  Pitch        *int     `json:"pitch"`
  Stability    *int     `json:"stability"`
  Similarity   *int     `json:"similarity"`
  Exaggeration *int     `json:"exaggeration"`
  Status       *int     `json:"status"`
  IsDefault    *int     `json:"is_default"`
}

type TTSPresetHandler struct {
  db *sql.DB
}

const (
  defaultVoiceID      = "70eb6772-4cd1-11f0-9276-00163e0fe4f9"
  defaultEmotionName  = "Happy"
  defaultVolume       = 58
  defaultSpeed        = 1.0
  defaultPitch        = 56
  defaultStability    = 50
  defaultSimilarity   = 95
  defaultExaggeration = 0
)

func defaultTTSPreset() TTSPreset {
  return TTSPreset{
    Name:         "默认预设",
    VoiceID:      defaultVoiceID,
    EmotionName:  defaultEmotionName,
    Volume:       defaultVolume,
    Speed:        defaultSpeed,
    Pitch:        defaultPitch,
    Stability:    defaultStability,
    Similarity:   defaultSimilarity,
    Exaggeration: defaultExaggeration,
    Status:       1,
    IsDefault:    1,
  }
}

// ValidateTTSParams validates parameter ranges.
// Args:
//   volume: 0-100
//   speed: 0.5-2.0
//   pitch: 1-100
//   stability: 0-100
//   similarity: 0-100
//   exaggeration: 0-100
// Returns:
//   error: Error when any value is out of range.
func ValidateTTSParams(volume int, speed float64, pitch int, stability int, similarity int, exaggeration int) error {
  if volume < 0 || volume > 100 {
    return errors.New("volume must be between 0 and 100")
  }
  if speed < 0.5 || speed > 2.0 {
    return errors.New("speed must be between 0.5 and 2.0")
  }
  if pitch < 1 || pitch > 100 {
    return errors.New("pitch must be between 1 and 100")
  }
  if stability < 0 || stability > 100 {
    return errors.New("stability must be between 0 and 100")
  }
  if similarity < 0 || similarity > 100 {
    return errors.New("similarity must be between 0 and 100")
  }
  if exaggeration < 0 || exaggeration > 100 {
    return errors.New("exaggeration must be between 0 and 100")
  }
  return nil
}

// NewTTSPresetHandler creates a handler for TTS presets.
// Args:
//   db: Database connection.
// Returns:
//   *TTSPresetHandler: Handler instance.
func NewTTSPresetHandler(db *sql.DB) *TTSPresetHandler {
  return &TTSPresetHandler{db: db}
}

func (h *TTSPresetHandler) List(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }
  showAll := strings.TrimSpace(c.Query("all")) == "1"
  if showAll {
    if claims, ok := middleware.GetAuthClaims(c); !ok || !strings.EqualFold(claims.Role, "admin") {
      showAll = false
    }
  }

  query := "SELECT id, name, voice_id, emotion_name, volume, speed, pitch, stability, similarity, exaggeration, status, is_default, created_at, updated_at, created_by, updated_by FROM app_db_tts_presets"
  if !showAll {
    query += " WHERE status = 1"
  }
  query += " ORDER BY is_default DESC, id DESC"

  rows, err := h.db.Query(query)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }
  defer rows.Close()

  items := make([]gin.H, 0)
  for rows.Next() {
    var (
      id           int64
      name         sql.NullString
      voiceID      sql.NullString
      emotionName  sql.NullString
      volume       sql.NullInt64
      speed        sql.NullFloat64
      pitch        sql.NullInt64
      stability    sql.NullInt64
      similarity   sql.NullInt64
      exaggeration sql.NullInt64
      status       sql.NullInt64
      isDefault    sql.NullInt64
      createdAt    sql.NullTime
      updatedAt    sql.NullTime
      createdBy    sql.NullInt64
      updatedBy    sql.NullInt64
    )
    if err := rows.Scan(
      &id,
      &name,
      &voiceID,
      &emotionName,
      &volume,
      &speed,
      &pitch,
      &stability,
      &similarity,
      &exaggeration,
      &status,
      &isDefault,
      &createdAt,
      &updatedAt,
      &createdBy,
      &updatedBy,
    ); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed"})
      return
    }
    items = append(items, gin.H{
      "id":           id,
      "name":         nullableString(name),
      "voice_id":     nullableString(voiceID),
      "emotion_name": nullableString(emotionName),
      "volume":       nullableInt64Pointer(volume),
      "speed":        nullableFloatPointer(speed),
      "pitch":        nullableInt64Pointer(pitch),
      "stability":    nullableInt64Pointer(stability),
      "similarity":   nullableInt64Pointer(similarity),
      "exaggeration": nullableInt64Pointer(exaggeration),
      "status":       nullableInt64Pointer(status),
      "is_default":   nullableInt64Pointer(isDefault),
      "created_at":   nullableTimePointer(createdAt),
      "updated_at":   nullableTimePointer(updatedAt),
      "created_by":   nullableInt64Pointer(createdBy),
      "updated_by":   nullableInt64Pointer(updatedBy),
    })
  }

  c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *TTSPresetHandler) Create(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }
  var req ttsPresetPayload
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
    return
  }

  preset := defaultTTSPreset()
  name := ""
  if req.Name != nil {
    name = strings.TrimSpace(*req.Name)
  }
  if name == "" {
    c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
    return
  }
  preset.Name = name

  preset, err := applyPresetOverrides(preset, &req)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  if err := ValidateTTSParams(preset.Volume, preset.Speed, preset.Pitch, preset.Stability, preset.Similarity, preset.Exaggeration); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  claims, _ := middleware.GetAuthClaims(c)
  now := time.Now()
  payload := map[string]interface{}{
    "name":         preset.Name,
    "voice_id":     preset.VoiceID,
    "emotion_name": normalizeOptionalString(preset.EmotionName),
    "volume":       preset.Volume,
    "speed":        preset.Speed,
    "pitch":        preset.Pitch,
    "stability":    preset.Stability,
    "similarity":   preset.Similarity,
    "exaggeration": preset.Exaggeration,
    "status":       preset.Status,
    "is_default":   preset.IsDefault,
    "created_at":   now,
    "updated_at":   now,
  }
  if claims != nil {
    payload["created_by"] = claims.UserID
    payload["updated_by"] = claims.UserID
  }

  tx, err := h.db.Begin()
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "transaction failed"})
    return
  }
  defer func() { _ = tx.Rollback() }()

  if preset.IsDefault == 1 {
    if _, err := tx.Exec("UPDATE app_db_tts_presets SET is_default = 0 WHERE is_default = 1"); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "reset default failed"})
      return
    }
  }

  sqlText, args, err := BuildInsertSQL("app_db_tts_presets", payload)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  if _, err := tx.Exec(sqlText, args...); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "insert failed"})
    return
  }

  if err := tx.Commit(); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "commit failed"})
    return
  }

  c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *TTSPresetHandler) Update(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }
  id := parseInt64Param(c, "id")
  if id <= 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
    return
  }

  var req ttsPresetPayload
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
    return
  }

  preset, err := h.fetchPresetByID(id)
  if err != nil {
    c.JSON(http.StatusNotFound, gin.H{"error": "preset not found"})
    return
  }

  updated, err := applyPresetOverrides(*preset, &req)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  preset = &updated

  if err := ValidateTTSParams(preset.Volume, preset.Speed, preset.Pitch, preset.Stability, preset.Similarity, preset.Exaggeration); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  claims, _ := middleware.GetAuthClaims(c)
  now := time.Now()
  payload := map[string]interface{}{
    "name":         preset.Name,
    "voice_id":     preset.VoiceID,
    "emotion_name": normalizeOptionalString(preset.EmotionName),
    "volume":       preset.Volume,
    "speed":        preset.Speed,
    "pitch":        preset.Pitch,
    "stability":    preset.Stability,
    "similarity":   preset.Similarity,
    "exaggeration": preset.Exaggeration,
    "status":       preset.Status,
    "is_default":   preset.IsDefault,
    "updated_at":   now,
  }
  if claims != nil {
    payload["updated_by"] = claims.UserID
  }

  tx, err := h.db.Begin()
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "transaction failed"})
    return
  }
  defer func() { _ = tx.Rollback() }()

  if preset.IsDefault == 1 {
    if _, err := tx.Exec("UPDATE app_db_tts_presets SET is_default = 0 WHERE is_default = 1"); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "reset default failed"})
      return
    }
  }

  sqlText, args, err := BuildUpdateSQL("app_db_tts_presets", "id", id, payload)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  if _, err := tx.Exec(sqlText, args...); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
    return
  }

  if err := tx.Commit(); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "commit failed"})
    return
  }

  c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *TTSPresetHandler) Delete(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }
  id := parseInt64Param(c, "id")
  if id <= 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
    return
  }

  preset, err := h.fetchPresetByID(id)
  if err != nil {
    c.JSON(http.StatusNotFound, gin.H{"error": "preset not found"})
    return
  }
  if preset.IsDefault == 1 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "default preset cannot be deleted"})
    return
  }

  if _, err := h.db.Exec("DELETE FROM app_db_tts_presets WHERE id = ?", id); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
    return
  }
  c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *TTSPresetHandler) fetchPresetByID(id int64) (*TTSPreset, error) {
  return fetchTTSPresetByID(h.db, id)
}

func fetchTTSPresetByID(db *sql.DB, id int64) (*TTSPreset, error) {
  if db == nil {
    return nil, errors.New("db not ready")
  }
  row := db.QueryRow("SELECT id, name, voice_id, emotion_name, volume, speed, pitch, stability, similarity, exaggeration, status, is_default, created_at, updated_at, created_by, updated_by FROM app_db_tts_presets WHERE id = ? LIMIT 1", id)
  var preset TTSPreset
  var (
    name         sql.NullString
    voiceID      sql.NullString
    emotionName  sql.NullString
    volume       sql.NullInt64
    speed        sql.NullFloat64
    pitch        sql.NullInt64
    stability    sql.NullInt64
    similarity   sql.NullInt64
    exaggeration sql.NullInt64
    status       sql.NullInt64
    isDefault    sql.NullInt64
  )
  if err := row.Scan(
    &preset.ID,
    &name,
    &voiceID,
    &emotionName,
    &volume,
    &speed,
    &pitch,
    &stability,
    &similarity,
    &exaggeration,
    &status,
    &isDefault,
    &preset.CreatedAt,
    &preset.UpdatedAt,
    &preset.CreatedBy,
    &preset.UpdatedBy,
  ); err != nil {
    return nil, err
  }

  preset.Name = nullableStringValue(name)
  preset.VoiceID = nullableStringValue(voiceID)
  preset.EmotionName = nullableStringValue(emotionName)
  preset.Volume = int(nullableInt64ValueSafe(volume))
  preset.Speed = nullableFloatValue(speed)
  preset.Pitch = int(nullableInt64ValueSafe(pitch))
  preset.Stability = int(nullableInt64ValueSafe(stability))
  preset.Similarity = int(nullableInt64ValueSafe(similarity))
  preset.Exaggeration = int(nullableInt64ValueSafe(exaggeration))
  preset.Status = int(nullableInt64ValueSafe(status))
  preset.IsDefault = int(nullableInt64ValueSafe(isDefault))

  return &preset, nil
}

func fetchDefaultTTSPreset(db *sql.DB) (*TTSPreset, error) {
  if db == nil {
    return nil, errors.New("db not ready")
  }
  row := db.QueryRow("SELECT id, name, voice_id, emotion_name, volume, speed, pitch, stability, similarity, exaggeration, status, is_default, created_at, updated_at, created_by, updated_by FROM app_db_tts_presets WHERE is_default = 1 LIMIT 1")
  var preset TTSPreset
  var (
    name         sql.NullString
    voiceID      sql.NullString
    emotionName  sql.NullString
    volume       sql.NullInt64
    speed        sql.NullFloat64
    pitch        sql.NullInt64
    stability    sql.NullInt64
    similarity   sql.NullInt64
    exaggeration sql.NullInt64
    status       sql.NullInt64
    isDefault    sql.NullInt64
  )
  if err := row.Scan(
    &preset.ID,
    &name,
    &voiceID,
    &emotionName,
    &volume,
    &speed,
    &pitch,
    &stability,
    &similarity,
    &exaggeration,
    &status,
    &isDefault,
    &preset.CreatedAt,
    &preset.UpdatedAt,
    &preset.CreatedBy,
    &preset.UpdatedBy,
  ); err != nil {
    return nil, err
  }

  preset.Name = nullableStringValue(name)
  preset.VoiceID = nullableStringValue(voiceID)
  preset.EmotionName = nullableStringValue(emotionName)
  preset.Volume = int(nullableInt64ValueSafe(volume))
  preset.Speed = nullableFloatValue(speed)
  preset.Pitch = int(nullableInt64ValueSafe(pitch))
  preset.Stability = int(nullableInt64ValueSafe(stability))
  preset.Similarity = int(nullableInt64ValueSafe(similarity))
  preset.Exaggeration = int(nullableInt64ValueSafe(exaggeration))
  preset.Status = int(nullableInt64ValueSafe(status))
  preset.IsDefault = int(nullableInt64ValueSafe(isDefault))

  return &preset, nil
}

func applyPresetOverrides(base TTSPreset, req *ttsPresetPayload) (TTSPreset, error) {
  if req == nil {
    return base, nil
  }
  if req.Name != nil {
    name := strings.TrimSpace(*req.Name)
    if name == "" {
      return base, errors.New("name is required")
    }
    base.Name = name
  }
  if req.VoiceID != nil {
    voice := strings.TrimSpace(*req.VoiceID)
    if voice == "" {
      return base, errors.New("voice_id is required")
    }
    base.VoiceID = voice
  }
  if req.EmotionName != nil {
    base.EmotionName = strings.TrimSpace(*req.EmotionName)
  }
  if req.EmotionName != nil {
    base.EmotionName = strings.TrimSpace(*req.EmotionName)
  }
  if req.Volume != nil {
    base.Volume = *req.Volume
  }
  if req.Speed != nil {
    base.Speed = *req.Speed
  }
  if req.Pitch != nil {
    base.Pitch = *req.Pitch
  }
  if req.Stability != nil {
    base.Stability = *req.Stability
  }
  if req.Similarity != nil {
    base.Similarity = *req.Similarity
  }
  if req.Exaggeration != nil {
    base.Exaggeration = *req.Exaggeration
  }
  if req.Status != nil {
    if *req.Status != 0 && *req.Status != 1 {
      return base, errors.New("status must be 0 or 1")
    }
    base.Status = *req.Status
  }
  if req.IsDefault != nil {
    if *req.IsDefault != 0 && *req.IsDefault != 1 {
      return base, errors.New("is_default must be 0 or 1")
    }
    base.IsDefault = *req.IsDefault
  }
  if base.IsDefault == 1 {
    base.Status = 1
  }
  return base, nil
}

func applyTTSRequestOverrides(base TTSPreset, req *ttsRequest) (TTSPreset, error) {
  if req == nil {
    return base, nil
  }
  if req.VoiceID != nil {
    voice := strings.TrimSpace(*req.VoiceID)
    if voice == "" {
      return base, errors.New("voice_id is required")
    }
    base.VoiceID = voice
  }
  if req.Volume != nil {
    base.Volume = *req.Volume
  }
  if req.Speed != nil {
    base.Speed = *req.Speed
  }
  if req.Pitch != nil {
    base.Pitch = *req.Pitch
  }
  if req.Stability != nil {
    base.Stability = *req.Stability
  }
  if req.Similarity != nil {
    base.Similarity = *req.Similarity
  }
  if req.Exaggeration != nil {
    base.Exaggeration = *req.Exaggeration
  }
  return base, nil
}

func nullableFloatPointer(value sql.NullFloat64) *float64 {
  if value.Valid {
    v := value.Float64
    return &v
  }
  return nil
}

func nullableInt64ValueSafe(value sql.NullInt64) int64 {
  if value.Valid {
    return value.Int64
  }
  return 0
}

func nullableFloatValue(value sql.NullFloat64) float64 {
  if value.Valid {
    return value.Float64
  }
  return 0
}

func normalizeOptionalString(value string) interface{} {
  trimmed := strings.TrimSpace(value)
  if trimmed == "" {
    return nil
  }
  return trimmed
}
