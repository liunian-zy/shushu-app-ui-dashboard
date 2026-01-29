package services

import (
  "bytes"
  "context"
  "encoding/json"
  "fmt"
  "io"
  "net/http"
  "strings"
  "time"

  "shushu-app-ui-dashboard/internal/config"
)

type TTSService struct {
  baseURL string
  apiKey  string
  client  *http.Client
}

type TTSConvertRequest struct {
  Text         string   `json:"text"`
  Volume       *int     `json:"volume,omitempty"`
  Speed        *float64 `json:"speed,omitempty"`
  Pitch        *int     `json:"pitch,omitempty"`
  Stability    *int     `json:"stability,omitempty"`
  Similarity   *int     `json:"similarity,omitempty"`
  Exaggeration *int     `json:"exaggeration,omitempty"`
  VoiceID      *string  `json:"voice_id,omitempty"`
  EmotionName  *string  `json:"emotion_name,omitempty"`
  Accent       *string  `json:"accent,omitempty"`
  CountryCode  *string  `json:"country_code,omitempty"`
}

type TTSVoiceData struct {
  ID          int64  `json:"id"`
  DisplayName string `json:"displayName"`
  VoiceAvatar string `json:"voiceAvatar"`
}

type TTSConvertResponse struct {
  Success  bool          `json:"success"`
  AudioURL string        `json:"audioUrl"`
  Message  string        `json:"message"`
  Data     *TTSVoiceData `json:"data"`
}

type TTSVoiceDetailRequest struct {
  VoiceID string `json:"voice_id"`
  SlangID int    `json:"slang_id"`
}

type TTSVoiceDetailResponse struct {
  Success bool            `json:"success"`
  Message string          `json:"message"`
  Data    json.RawMessage `json:"data"`
}

// NewTTSService creates a new TTS service instance.
// Args:
//   cfg: App config instance with TTS settings.
// Returns:
//   *TTSService: Initialized service.
//   error: Error when config is incomplete.
func NewTTSService(cfg *config.Config) (*TTSService, error) {
  if cfg == nil || strings.TrimSpace(cfg.TtsBaseURL) == "" {
    return nil, fmt.Errorf("tts base url is required")
  }
  if strings.TrimSpace(cfg.TtsAPIKey) == "" {
    return nil, fmt.Errorf("tts api key is required")
  }
  return &TTSService{
    baseURL: strings.TrimRight(cfg.TtsBaseURL, "/"),
    apiKey:  cfg.TtsAPIKey,
    client: &http.Client{
      Timeout: 60 * time.Second,
    },
  }, nil
}

// Convert calls the upstream TTS service.
// Args:
//   ctx: Request context.
//   req: Convert request payload.
// Returns:
//   *TTSConvertResponse: Response payload.
//   error: Error when request fails.
func (s *TTSService) Convert(ctx context.Context, req *TTSConvertRequest) (*TTSConvertResponse, error) {
  if s == nil || req == nil {
    return nil, fmt.Errorf("invalid tts request")
  }

  payload, err := json.Marshal(req)
  if err != nil {
    return nil, err
  }

  endpoint := s.baseURL + "/api/tts/convert"
  httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
  if err != nil {
    return nil, err
  }
  httpReq.Header.Set("Content-Type", "application/json")
  httpReq.Header.Set("X-API-Key", s.apiKey)

  resp, err := s.client.Do(httpReq)
  if err != nil {
    return nil, err
  }
  defer resp.Body.Close()

  body, err := io.ReadAll(resp.Body)
  if err != nil {
    return nil, err
  }

  if resp.StatusCode < 200 || resp.StatusCode >= 300 {
    return nil, fmt.Errorf("tts convert failed: %s", strings.TrimSpace(string(body)))
  }

  var result TTSConvertResponse
  if err := json.Unmarshal(body, &result); err != nil {
    return nil, err
  }

  if !result.Success {
    msg := strings.TrimSpace(result.Message)
    if msg == "" {
      msg = "tts convert failed"
    }
    return nil, fmt.Errorf(msg)
  }
  return &result, nil
}

// VoiceDetail fetches voice detail metadata from the TTS service.
// Args:
//   ctx: Request context.
//   voiceID: Voice id.
//   slangID: Accent slang id.
// Returns:
//   *TTSVoiceDetailResponse: Response payload.
//   error: Error when request fails.
func (s *TTSService) VoiceDetail(ctx context.Context, voiceID string, slangID int) (*TTSVoiceDetailResponse, error) {
  if s == nil {
    return nil, fmt.Errorf("invalid tts service")
  }
  trimmed := strings.TrimSpace(voiceID)
  if trimmed == "" {
    return nil, fmt.Errorf("voice_id is required")
  }
  if slangID <= 0 {
    slangID = 18
  }

  payload, err := json.Marshal(&TTSVoiceDetailRequest{
    VoiceID: trimmed,
    SlangID: slangID,
  })
  if err != nil {
    return nil, err
  }

  endpoint := s.baseURL + "/api/tts/voice-detail"
  httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
  if err != nil {
    return nil, err
  }
  httpReq.Header.Set("Content-Type", "application/json")
  httpReq.Header.Set("X-API-Key", s.apiKey)

  resp, err := s.client.Do(httpReq)
  if err != nil {
    return nil, err
  }
  defer resp.Body.Close()

  body, err := io.ReadAll(resp.Body)
  if err != nil {
    return nil, err
  }

  if resp.StatusCode < 200 || resp.StatusCode >= 300 {
    return nil, fmt.Errorf("tts voice detail failed: %s", strings.TrimSpace(string(body)))
  }

  var result TTSVoiceDetailResponse
  if err := json.Unmarshal(body, &result); err != nil {
    return nil, err
  }
  if !result.Success {
    msg := strings.TrimSpace(result.Message)
    if msg == "" {
      msg = "tts voice detail failed"
    }
    return nil, fmt.Errorf(msg)
  }
  return &result, nil
}

// DownloadAudio downloads audio from a given URL.
// Args:
//   ctx: Request context.
//   url: Audio URL.
// Returns:
//   []byte: Audio bytes.
//   error: Error when download fails.
func (s *TTSService) DownloadAudio(ctx context.Context, url string) ([]byte, error) {
  if s == nil || strings.TrimSpace(url) == "" {
    return nil, fmt.Errorf("audio url is required")
  }

  httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
  if err != nil {
    return nil, err
  }

  resp, err := s.client.Do(httpReq)
  if err != nil {
    return nil, err
  }
  defer resp.Body.Close()

  body, err := io.ReadAll(resp.Body)
  if err != nil {
    return nil, err
  }
  if resp.StatusCode < 200 || resp.StatusCode >= 300 {
    return nil, fmt.Errorf("audio download failed: %s", strings.TrimSpace(string(body)))
  }
  return body, nil
}
