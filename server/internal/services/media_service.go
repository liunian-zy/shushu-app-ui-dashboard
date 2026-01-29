package services

import (
  "context"
  "crypto/rand"
  "encoding/hex"
  "encoding/json"
  "errors"
  "fmt"
  "math"
  "os"
  "os/exec"
  "path/filepath"
  "strconv"
  "strings"
  "time"
)

type MediaService struct {
  oss         *OSSService
  ffmpegPath  string
  ffprobePath string
}

type MediaRule struct {
  ID             int64
  ModuleKey      string
  MediaType      string
  MaxSizeKB      int64
  MinWidth       int64
  MaxWidth       int64
  MinHeight      int64
  MaxHeight      int64
  RatioWidth     int64
  RatioHeight    int64
  MinDurationMS  int64
  MaxDurationMS  int64
  AllowFormats   string
  ResizeMode     string
  TargetFormat   string
  CompressQuality int64
}

type MediaMeta struct {
  SizeBytes  int64
  Width      int64
  Height     int64
  DurationMS int64
  Format     string
  FileExt    string
}

type MediaViolation struct {
  Field  string      `json:"field"`
  Rule   interface{} `json:"rule"`
  Actual interface{} `json:"actual"`
}

// NewMediaService creates a media service for validation and transformation.
// Args:
//   ossService: OSS service instance.
// Returns:
//   *MediaService: Initialized media service.
//   error: Error when required tools are missing.
func NewMediaService(ossService *OSSService) (*MediaService, error) {

  ffmpegPath, err := exec.LookPath("ffmpeg")
  if err != nil {
    return nil, errors.New("ffmpeg not found")
  }
  ffprobePath, err := exec.LookPath("ffprobe")
  if err != nil {
    return nil, errors.New("ffprobe not found")
  }

  return &MediaService{
    oss:         ossService,
    ffmpegPath:  ffmpegPath,
    ffprobePath: ffprobePath,
  }, nil
}

// DownloadToTemp downloads an OSS object to a temp file.
// Args:
//   objectPath: OSS object path.
// Returns:
//   string: Local temp file path.
//   func(): Cleanup function.
//   error: Error when download fails.
func (s *MediaService) DownloadToTemp(objectPath string) (string, func(), error) {
  if strings.TrimSpace(objectPath) == "" {
    return "", nil, errors.New("object path is required")
  }

  tempDir := os.TempDir()
  ext := filepath.Ext(objectPath)
  if ext == "" {
    ext = ".bin"
  }
  tempFile, err := os.CreateTemp(tempDir, "media-*")
  if err != nil {
    return "", nil, err
  }
  tempPath := tempFile.Name() + ext
  _ = tempFile.Close()

  if err := s.oss.DownloadToFile(objectPath, tempPath); err != nil {
    _ = os.Remove(tempPath)
    return "", nil, err
  }

  cleanup := func() {
    _ = os.Remove(tempPath)
  }
  return tempPath, cleanup, nil
}

// Probe reads media metadata using ffprobe.
// Args:
//   localPath: Local file path.
// Returns:
//   *MediaMeta: Metadata result.
//   error: Error when probe fails.
func (s *MediaService) Probe(localPath string) (*MediaMeta, error) {
  if strings.TrimSpace(localPath) == "" {
    return nil, errors.New("local path is required")
  }

  ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
  defer cancel()

  cmd := exec.CommandContext(ctx, s.ffprobePath, "-v", "quiet", "-print_format", "json", "-show_streams", "-show_format", localPath)
  output, err := cmd.Output()
  if err != nil {
    return nil, fmt.Errorf("ffprobe failed: %w", err)
  }

  meta := &MediaMeta{}
  if err := parseFFProbe(output, meta); err != nil {
    return nil, err
  }

  fileInfo, err := os.Stat(localPath)
  if err == nil {
    meta.SizeBytes = fileInfo.Size()
  }

  ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(localPath), "."))
  if ext != "" {
    meta.FileExt = ext
    if strings.TrimSpace(meta.Format) == "" {
      meta.Format = ext
    }
  }

  return meta, nil
}

// Transform transforms a media file to match the rule.
// Args:
//   localInput: Local input path.
//   localOutput: Local output path.
//   mediaType: Media type (image/video/audio).
//   rule: Media rule for transformation.
// Returns:
//   error: Error when transform fails.
func (s *MediaService) Transform(localInput, localOutput, mediaType string, rule *MediaRule) error {
  if strings.TrimSpace(localInput) == "" || strings.TrimSpace(localOutput) == "" {
    return errors.New("input and output paths are required")
  }

  switch strings.ToLower(mediaType) {
  case "image":
    return s.transformImage(localInput, localOutput, rule)
  case "video":
    return s.transformVideo(localInput, localOutput, rule)
  case "audio":
    return s.transformAudio(localInput, localOutput, rule)
  default:
    return errors.New("unsupported media type")
  }
}

// ValidateMediaRule validates metadata against a media rule.
// Args:
//   rule: Media rule to validate.
//   meta: Media metadata.
// Returns:
//   []MediaViolation: List of violations.
func ValidateMediaRule(rule *MediaRule, meta *MediaMeta) []MediaViolation {
  if rule == nil || meta == nil {
    return nil
  }

  violations := make([]MediaViolation, 0)

  mediaType := strings.ToLower(strings.TrimSpace(rule.MediaType))
  if mediaType == "" {
    mediaType = "image"
  }

  if mediaType != "audio" && rule.MaxSizeKB > 0 {
    maxBytes := rule.MaxSizeKB * 1024
    if meta.SizeBytes > maxBytes {
      violations = append(violations, MediaViolation{Field: "size", Rule: maxBytes, Actual: meta.SizeBytes})
    }
  }

  if mediaType == "image" {
    if rule.MinWidth > 0 && meta.Width > 0 && meta.Width < rule.MinWidth {
      violations = append(violations, MediaViolation{Field: "width_min", Rule: rule.MinWidth, Actual: meta.Width})
    }
    if rule.MaxWidth > 0 && meta.Width > 0 && meta.Width > rule.MaxWidth {
      violations = append(violations, MediaViolation{Field: "width_max", Rule: rule.MaxWidth, Actual: meta.Width})
    }
    if rule.MinHeight > 0 && meta.Height > 0 && meta.Height < rule.MinHeight {
      violations = append(violations, MediaViolation{Field: "height_min", Rule: rule.MinHeight, Actual: meta.Height})
    }
    if rule.MaxHeight > 0 && meta.Height > 0 && meta.Height > rule.MaxHeight {
      violations = append(violations, MediaViolation{Field: "height_max", Rule: rule.MaxHeight, Actual: meta.Height})
    }
    if rule.RatioWidth > 0 && rule.RatioHeight > 0 && meta.Width > 0 && meta.Height > 0 {
      if meta.Width*rule.RatioHeight != meta.Height*rule.RatioWidth {
        violations = append(violations, MediaViolation{
          Field:  "ratio",
          Rule:   fmt.Sprintf("%d:%d", rule.RatioWidth, rule.RatioHeight),
          Actual: fmt.Sprintf("%d:%d", meta.Width, meta.Height),
        })
      }
    }
  }

  if mediaType != "audio" && strings.TrimSpace(rule.AllowFormats) != "" {
    allowed := parseFormatList(rule.AllowFormats)
    if len(allowed) > 0 {
      actual := parseFormatList(meta.FileExt)
      if len(actual) == 0 {
        actual = parseFormatList(meta.Format)
      }
      if !containsAny(allowed, actual) {
        violations = append(violations, MediaViolation{Field: "format", Rule: allowed, Actual: actual})
      }
    }
  }

  return violations
}

// ComputeTargetPath builds a new target path with optional format change.
// Args:
//   sourcePath: Original path.
//   targetFormat: Target format extension (without dot).
// Returns:
//   string: New path with suffix.
func ComputeTargetPath(sourcePath, targetFormat string) string {
  cleaned := strings.TrimSpace(sourcePath)
  if cleaned == "" {
    return ""
  }
  ext := filepath.Ext(cleaned)
  dir := filepath.Dir(cleaned)
  if targetFormat != "" {
    ext = "." + strings.TrimPrefix(strings.ToLower(targetFormat), ".")
  }
  if ext == "" {
    ext = ".bin"
  }
  suffix := time.Now().Format("20060102150405")
  filename := fmt.Sprintf("file_%s_%s%s", suffix, randomSuffix(8), ext)
  if dir == "." || dir == "/" {
    return filename
  }
  return filepath.ToSlash(filepath.Join(dir, filename))
}

func randomSuffix(length int) string {
  if length <= 0 {
    return ""
  }
  raw := make([]byte, (length+1)/2)
  if _, err := rand.Read(raw); err != nil {
    return strings.Repeat("0", length)
  }
  encoded := hex.EncodeToString(raw)
  if len(encoded) > length {
    return encoded[:length]
  }
  if len(encoded) < length {
    return encoded + strings.Repeat("0", length-len(encoded))
  }
  return encoded
}

// transformImage applies image transformations using ffmpeg.
// Args:
//   localInput: Local input path.
//   localOutput: Local output path.
//   rule: Media rule for transformation.
// Returns:
//   error: Error when transform fails.
func (s *MediaService) transformImage(localInput, localOutput string, rule *MediaRule) error {
  args := []string{"-y", "-i", localInput}

  if scale := buildScaleFilter(rule); scale != "" {
    args = append(args, "-vf", scale)
  }

  if rule != nil && rule.CompressQuality > 0 {
    quality := normalizeImageQuality(rule.CompressQuality)
    if quality > 0 {
      args = append(args, "-q:v", strconv.FormatInt(quality, 10))
    }
  }

  args = append(args, localOutput)
  return runCommand(s.ffmpegPath, args...)
}

// transformVideo applies video transformations using ffmpeg.
// Args:
//   localInput: Local input path.
//   localOutput: Local output path.
//   rule: Media rule for transformation.
// Returns:
//   error: Error when transform fails.
func (s *MediaService) transformVideo(localInput, localOutput string, rule *MediaRule) error {
  args := []string{"-y", "-i", localInput}

  if scale := buildScaleFilter(rule); scale != "" {
    args = append(args, "-vf", scale)
  }

  quality := int64(28)
  if rule != nil && rule.CompressQuality > 0 {
    quality = rule.CompressQuality
  }

  args = append(args,
    "-c:v", "libx264",
    "-crf", strconv.FormatInt(quality, 10),
    "-preset", "slow",
    "-c:a", "aac",
    "-b:a", "128k",
    localOutput,
  )

  return runCommand(s.ffmpegPath, args...)
}

// transformAudio applies audio transformations using ffmpeg.
// Args:
//   localInput: Local input path.
//   localOutput: Local output path.
//   rule: Media rule for transformation.
// Returns:
//   error: Error when transform fails.
func (s *MediaService) transformAudio(localInput, localOutput string, rule *MediaRule) error {
  if rule != nil && strings.EqualFold(strings.TrimSpace(rule.ResizeMode), "lossless") {
    args := []string{
      "-y",
      "-i", localInput,
      "-c:a", "copy",
      localOutput,
    }
    return runCommand(s.ffmpegPath, args...)
  }

  args := []string{
    "-y",
    "-i", localInput,
    "-c:a", "aac",
    "-b:a", "128k",
    localOutput,
  }
  return runCommand(s.ffmpegPath, args...)
}

func normalizeImageQuality(value int64) int64 {
  if value <= 0 {
    return 0
  }
  if value <= 31 {
    return value
  }
  percent := value
  if percent < 1 {
    percent = 1
  }
  if percent > 100 {
    percent = 100
  }
  mapped := 31 - int64(math.Round(float64(percent-1)*29.0/99.0))
  if mapped < 2 {
    mapped = 2
  }
  if mapped > 31 {
    mapped = 31
  }
  return mapped
}

// buildScaleFilter builds an ffmpeg scale filter based on rule settings.
// Args:
//   rule: Media rule for resizing.
// Returns:
//   string: Scale filter expression or empty string.
func buildScaleFilter(rule *MediaRule) string {
  if rule == nil || (rule.MaxWidth <= 0 && rule.MaxHeight <= 0) {
    return ""
  }

  mode := strings.ToLower(strings.TrimSpace(rule.ResizeMode))
  if mode == "fill" && rule.MaxWidth > 0 && rule.MaxHeight > 0 {
    return fmt.Sprintf("scale=%d:%d", rule.MaxWidth, rule.MaxHeight)
  }
  if mode == "cover" && rule.MaxWidth > 0 && rule.MaxHeight > 0 {
    return fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d", rule.MaxWidth, rule.MaxHeight, rule.MaxWidth, rule.MaxHeight)
  }

  widthExpr := "iw"
  heightExpr := "ih"
  if rule.MaxWidth > 0 {
    widthExpr = fmt.Sprintf("min(iw,%d)", rule.MaxWidth)
  }
  if rule.MaxHeight > 0 {
    heightExpr = fmt.Sprintf("min(ih,%d)", rule.MaxHeight)
  }

  return fmt.Sprintf("scale='%s':'%s':force_original_aspect_ratio=decrease", widthExpr, heightExpr)
}

// parseFFProbe parses ffprobe JSON output into metadata.
// Args:
//   raw: ffprobe JSON output.
//   meta: Metadata output target.
// Returns:
//   error: Error when parsing fails.
func parseFFProbe(raw []byte, meta *MediaMeta) error {
  var payload struct {
    Streams []struct {
      CodecType string `json:"codec_type"`
      Width     int64  `json:"width"`
      Height    int64  `json:"height"`
      Duration  string `json:"duration"`
    } `json:"streams"`
    Format struct {
      Duration   string `json:"duration"`
      Size       string `json:"size"`
      FormatName string `json:"format_name"`
    } `json:"format"`
  }

  if err := json.Unmarshal(raw, &payload); err != nil {
    return err
  }

  if payload.Format.Size != "" {
    if size, err := strconv.ParseInt(payload.Format.Size, 10, 64); err == nil {
      meta.SizeBytes = size
    }
  }

  if payload.Format.Duration != "" {
    if seconds, err := strconv.ParseFloat(payload.Format.Duration, 64); err == nil {
      meta.DurationMS = int64(seconds * 1000)
    }
  }

  meta.Format = payload.Format.FormatName

  for _, stream := range payload.Streams {
    if stream.Width > 0 || stream.Height > 0 {
      meta.Width = stream.Width
      meta.Height = stream.Height
      if meta.DurationMS == 0 && stream.Duration != "" {
        if seconds, err := strconv.ParseFloat(stream.Duration, 64); err == nil {
          meta.DurationMS = int64(seconds * 1000)
        }
      }
      break
    }
  }

  return nil
}

// parseFormatList normalizes a comma separated format list.
// Args:
//   raw: Raw format list string.
// Returns:
//   []string: Normalized format list.
func parseFormatList(raw string) []string {
  normalized := strings.ToLower(strings.TrimSpace(raw))
  if normalized == "" {
    return nil
  }
  parts := strings.Split(normalized, ",")
  out := make([]string, 0, len(parts))
  for _, part := range parts {
    trimmed := strings.TrimPrefix(strings.TrimSpace(part), ".")
    if trimmed != "" {
      out = append(out, trimmed)
    }
  }
  return out
}

// containsAny checks whether any actual format is allowed.
// Args:
//   allowed: Allowed formats.
//   actual: Actual formats.
// Returns:
//   bool: True when any match exists.
func containsAny(allowed, actual []string) bool {
  if len(allowed) == 0 || len(actual) == 0 {
    return false
  }
  set := make(map[string]struct{}, len(actual))
  for _, item := range actual {
    set[item] = struct{}{}
  }
  for _, item := range allowed {
    if _, ok := set[item]; ok {
      return true
    }
  }
  return false
}

// runCommand executes a command with timeout.
// Args:
//   path: Command path.
//   args: Command arguments.
// Returns:
//   error: Error when execution fails.
func runCommand(path string, args ...string) error {
  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
  defer cancel()
  cmd := exec.CommandContext(ctx, path, args...)
  cmd.Stdout = nil
  cmd.Stderr = nil
  if err := cmd.Run(); err != nil {
    return err
  }
  return nil
}
