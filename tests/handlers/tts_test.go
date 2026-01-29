package handlers_test

import (
  "strings"
  "testing"

  "shushu-app-ui-dashboard/internal/http/handlers"
)

func TestValidateTTSText(t *testing.T) {
  if msg := handlers.ValidateTTSText(" "); msg == "" {
    t.Fatalf("expected error for empty text")
  }

  longText := strings.Repeat("a", 5001)
  if msg := handlers.ValidateTTSText(longText); msg == "" {
    t.Fatalf("expected error for long text")
  }

  if msg := handlers.ValidateTTSText("hello"); msg != "" {
    t.Fatalf("expected no error, got %s", msg)
  }
}

func TestBuildTTSAudioPath(t *testing.T) {
  path := handlers.BuildTTSAudioPath("my module", 123)
  if !strings.HasPrefix(path, "tts/my_module/123/tts_") {
    t.Fatalf("unexpected path prefix: %s", path)
  }
  if !strings.HasSuffix(path, ".mp3") {
    t.Fatalf("expected mp3 suffix, got %s", path)
  }
}
