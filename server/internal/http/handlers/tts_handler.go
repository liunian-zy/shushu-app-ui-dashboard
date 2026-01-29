package handlers

import (
  "database/sql"
  "encoding/json"
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

type TTSHandler struct {
  db    *sql.DB
  cfg   *config.Config
  redis *redis.Client
}

type ttsRequest struct {
  Text           string   `json:"text"`
  PresetID       *int64   `json:"preset_id"`
  Volume         *int     `json:"volume"`
  Speed          *float64 `json:"speed"`
  Pitch          *int     `json:"pitch"`
  Stability      *int     `json:"stability"`
  Similarity     *int     `json:"similarity"`
  Exaggeration   *int     `json:"exaggeration"`
  VoiceID        *string  `json:"voice_id"`
  EmotionName    *string  `json:"emotion_name"`
  Accent         *string  `json:"accent"`
  CountryCode    *string  `json:"country_code"`
  ModuleKey      string   `json:"module_key"`
  DraftVersionID int64    `json:"draft_version_id"`
}

type ttsVoiceDetailRequest struct {
  VoiceID string `json:"voice_id"`
  SlangID int    `json:"slang_id"`
}

// NewTTSHandler creates a handler for TTS requests.
// Args:
//   cfg: App config instance.
//   db: Database connection.
//   redis: Redis client for OSS cache.
// Returns:
//   *TTSHandler: Initialized handler.
func NewTTSHandler(cfg *config.Config, db *sql.DB, redis *redis.Client) *TTSHandler {
  return &TTSHandler{cfg: cfg, db: db, redis: redis}
}

// Convert generates audio from text and uploads it to OSS.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *TTSHandler) Convert(c *gin.Context) {
  var req ttsRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
    return
  }

  if msg := ValidateTTSText(req.Text); msg != "" {
    c.JSON(http.StatusBadRequest, gin.H{"error": msg})
    return
  }

  ttsService, err := services.NewTTSService(h.cfg)
  if err != nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
    return
  }

  preset := defaultTTSPreset()
  if req.PresetID != nil && *req.PresetID > 0 && h.db != nil {
    if selected, err := fetchTTSPresetByID(h.db, *req.PresetID); err == nil {
      preset = *selected
    }
  } else if h.db != nil {
    if selected, err := fetchDefaultTTSPreset(h.db); err == nil {
      preset = *selected
    }
  }

  preset, err = applyTTSRequestOverrides(preset, &req)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  if err := ValidateTTSParams(preset.Volume, preset.Speed, preset.Pitch, preset.Stability, preset.Similarity, preset.Exaggeration); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  convertReq := buildTTSConvertRequest(&req, &preset)
  result, err := ttsService.Convert(c.Request.Context(), convertReq)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }

  audioURL := strings.TrimSpace(result.AudioURL)
  if audioURL == "" {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "tts audio url missing"})
    return
  }

  audioBytes, err := ttsService.DownloadAudio(c.Request.Context(), audioURL)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }

  objectPath := BuildTTSAudioPath(req.ModuleKey, req.DraftVersionID)
  localPath, err := buildLocalFilePath(h.cfg, objectPath)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "local path failed"})
    return
  }
  if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "mkdir failed"})
    return
  }
  if err := os.WriteFile(localPath, audioBytes, 0644); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "write failed"})
    return
  }

  localURL := buildLocalURL(h.cfg, objectPath)
  storagePath := localPathPrefix + objectPath

  c.JSON(http.StatusOK, gin.H{
    "success":     true,
    "audio_path":  storagePath,
    "audio_url":   localURL,
    "source_url":  audioURL,
    "size_bytes":  len(audioBytes),
    "voice":       result.Data,
    "created_at":  time.Now().Format(time.RFC3339),
    "module_key":  strings.TrimSpace(req.ModuleKey),
    "file_name":   filepath.Base(objectPath),
  })
}

// VoiceDetail returns emotion/accent info for a voice id.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *TTSHandler) VoiceDetail(c *gin.Context) {
  var req ttsVoiceDetailRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
    return
  }
  voiceID := strings.TrimSpace(req.VoiceID)
  if voiceID == "" {
    c.JSON(http.StatusBadRequest, gin.H{"error": "voice_id is required"})
    return
  }
  if req.SlangID <= 0 {
    req.SlangID = 18
  }

  ttsService, err := services.NewTTSService(h.cfg)
  if err != nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
    return
  }

  detail, err := ttsService.VoiceDetail(c.Request.Context(), voiceID, req.SlangID)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }

  if len(detail.Data) == 0 {
    c.JSON(http.StatusOK, gin.H{"data": gin.H{}})
    return
  }

  var payload map[string]interface{}
  if err := json.Unmarshal(detail.Data, &payload); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "decode failed"})
    return
  }
  c.JSON(http.StatusOK, gin.H{"data": payload})
}

func buildTTSConvertRequest(req *ttsRequest, preset *TTSPreset) *services.TTSConvertRequest {
  voice := ""
  volume := defaultVolume
  speed := defaultSpeed
  pitch := defaultPitch
  stability := defaultStability
  similarity := defaultSimilarity
  exaggeration := defaultExaggeration
  presetEmotion := ""
  if preset != nil {
    voice = preset.VoiceID
    volume = preset.Volume
    speed = preset.Speed
    pitch = preset.Pitch
    stability = preset.Stability
    similarity = preset.Similarity
    exaggeration = preset.Exaggeration
    presetEmotion = preset.EmotionName
  }
  cleanedVoice := trimStringPointer(&voice)
  cleanedEmotion := trimStringPointer(req.EmotionName)
  if cleanedEmotion == nil {
    cleanedEmotion = trimStringPointer(&presetEmotion)
  }
  cleanedAccent := trimStringPointer(req.Accent)
  cleanedCountry := trimStringPointer(req.CountryCode)

  return &services.TTSConvertRequest{
    Text:         strings.TrimSpace(req.Text),
    Volume:       &volume,
    Speed:        &speed,
    Pitch:        &pitch,
    Stability:    &stability,
    Similarity:   &similarity,
    Exaggeration: &exaggeration,
    VoiceID:      cleanedVoice,
    EmotionName:  cleanedEmotion,
    Accent:       cleanedAccent,
    CountryCode:  cleanedCountry,
  }
}

func trimStringPointer(value *string) *string {
  if value == nil {
    return nil
  }
  trimmed := strings.TrimSpace(*value)
  if trimmed == "" {
    return nil
  }
  return &trimmed
}

// newOSSService builds an OSS service instance for uploads.
// Args:
//   None.
// Returns:
//   *services.OSSService: OSS service or nil if init fails.
func (h *TTSHandler) newOSSService() *services.OSSService {
  service, err := services.NewOSSService(h.cfg, h.redis)
  if err != nil {
    return nil
  }
  return service
}

// ValidateTTSText validates text input for TTS.
// Args:
//   text: Input text.
// Returns:
//   string: Error message or empty when valid.
func ValidateTTSText(text string) string {
  trimmed := strings.TrimSpace(text)
  if trimmed == "" {
    return "text is required"
  }
  if len(trimmed) > 5000 {
    return "text length exceeds 5000"
  }
  return ""
}

// BuildTTSAudioPath creates an OSS path for TTS outputs.
// Args:
//   moduleKey: Module key for grouping.
//   draftVersionID: Draft version id.
// Returns:
//   string: Object path.
func BuildTTSAudioPath(moduleKey string, draftVersionID int64) string {
  cleanedModule := sanitizePathSegment(moduleKey)
  if cleanedModule == "" {
    cleanedModule = "default"
  }
  versionPart := "0"
  if draftVersionID > 0 {
    versionPart = strings.TrimSpace(formatInt64(draftVersionID))
  }
  timestamp := time.Now().Format("20060102150405")
  return "tts/" + cleanedModule + "/" + versionPart + "/tts_" + timestamp + "_" + randomSuffix(6) + ".mp3"
}
