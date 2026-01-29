package handlers

import (
  "encoding/json"
  "errors"
  "strings"

  "github.com/mozillazg/go-pinyin"
)

var baseFeishuFields = []string{
  "编号顺序",
  "场景",
  "身份",
  "提示词",
  "人像位置",
  "输入图1",
  "输入图2",
  "输入图3",
  "审核情况",
  "作废",
}

// NormalizeAiModal validates and normalizes ai modal value.
// Args:
//   value: Raw ai modal value.
// Returns:
//   string: Normalized ai modal.
//   error: Error when value is invalid.
func NormalizeAiModal(value string) (string, error) {
  trimmed := strings.ToUpper(strings.TrimSpace(value))
  if trimmed == "" {
    return "SD", nil
  }
  if trimmed != "SD" && trimmed != "NANO" {
    return "", errors.New("invalid ai_modal")
  }
  return trimmed, nil
}

// NormalizeVersionName normalizes a version name string.
// Args:
//   value: Raw version name.
// Returns:
//   string: Normalized version name.
func NormalizeVersionName(value string) string {
  trimmed := strings.TrimSpace(value)
  if trimmed == "" {
    return ""
  }
  upper := strings.ToUpper(trimmed)
  return sanitizeVersionName(upper)
}

// GenerateVersionName builds a version name from location name.
// Args:
//   locationName: Location name.
// Returns:
//   string: Generated version name.
func GenerateVersionName(locationName string) string {
  trimmed := strings.TrimSpace(locationName)
  if trimmed == "" {
    return ""
  }

  args := pinyin.NewArgs()
  args.Style = pinyin.Normal
  parts := pinyin.LazyPinyin(trimmed, args)
  raw := strings.ToUpper(strings.Join(parts, ""))
  return sanitizeVersionName(raw)
}

// BuildDefaultFeishuFields builds the default feishu field list.
// Args:
//   aiModal: Normalized ai modal.
// Returns:
//   []string: Field list.
func BuildDefaultFeishuFields(aiModal string) []string {
  fields := make([]string, 0, len(baseFeishuFields)+1)
  fields = append(fields, baseFeishuFields...)
  if strings.EqualFold(strings.TrimSpace(aiModal), "SD") {
    fields = append(fields, "SD模式")
  }
  return fields
}

// EncodeFeishuFieldNames encodes fields to JSON string.
// Args:
//   fields: Field list.
// Returns:
//   string: JSON string.
//   error: Error when encoding fails.
func EncodeFeishuFieldNames(fields []string) (string, error) {
  raw, err := json.Marshal(fields)
  if err != nil {
    return "", err
  }
  return string(raw), nil
}

func sanitizeVersionName(value string) string {
  if value == "" {
    return ""
  }
  var builder strings.Builder
  prevUnderscore := false
  for _, ch := range value {
    if (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
      builder.WriteRune(ch)
      prevUnderscore = false
      continue
    }
    if !prevUnderscore {
      builder.WriteByte('_')
      prevUnderscore = true
    }
  }
  cleaned := strings.Trim(builder.String(), "_")
  return cleaned
}
