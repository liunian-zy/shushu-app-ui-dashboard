package handlers_test

import "testing"

import "shushu-app-ui-dashboard/internal/http/handlers"

func TestValidateTTSParams(t *testing.T) {
  if err := handlers.ValidateTTSParams(58, 1.0, 56, 50, 95, 0); err != nil {
    t.Fatalf("expected valid params, got %v", err)
  }

  if err := handlers.ValidateTTSParams(-1, 1.0, 56, 50, 95, 0); err == nil {
    t.Fatalf("expected volume range error")
  }
  if err := handlers.ValidateTTSParams(58, 2.5, 56, 50, 95, 0); err == nil {
    t.Fatalf("expected speed range error")
  }
  if err := handlers.ValidateTTSParams(58, 1.0, 0, 50, 95, 0); err == nil {
    t.Fatalf("expected pitch range error")
  }
  if err := handlers.ValidateTTSParams(58, 1.0, 56, 120, 95, 0); err == nil {
    t.Fatalf("expected stability range error")
  }
  if err := handlers.ValidateTTSParams(58, 1.0, 56, 50, 120, 0); err == nil {
    t.Fatalf("expected similarity range error")
  }
  if err := handlers.ValidateTTSParams(58, 1.0, 56, 50, 95, 120); err == nil {
    t.Fatalf("expected exaggeration range error")
  }
}
