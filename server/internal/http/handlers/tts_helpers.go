package handlers

import (
  "crypto/rand"
  "encoding/hex"
  "strconv"
  "strings"
)

func sanitizePathSegment(value string) string {
  trimmed := strings.TrimSpace(value)
  if trimmed == "" {
    return ""
  }
  var builder strings.Builder
  for _, ch := range trimmed {
    switch {
    case ch >= 'a' && ch <= 'z':
      builder.WriteRune(ch)
    case ch >= 'A' && ch <= 'Z':
      builder.WriteRune(ch)
    case ch >= '0' && ch <= '9':
      builder.WriteRune(ch)
    case ch == '-' || ch == '_':
      builder.WriteRune(ch)
    default:
      builder.WriteByte('_')
    }
  }
  return strings.Trim(builder.String(), "_-")
}

func formatInt64(value int64) string {
  return strconv.FormatInt(value, 10)
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
