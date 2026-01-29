package services_test

import (
  "testing"

  "shushu-app-ui-dashboard/internal/config"
  "shushu-app-ui-dashboard/internal/services"
)

func TestAuthServiceHashVerify(t *testing.T) {
  cfg := &config.Config{JwtSecret: "test-secret", JwtIssuer: "test", JwtExpireHours: 1}
  service, err := services.NewAuthService(cfg)
  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }

  hash, err := service.HashPassword("pass123")
  if err != nil {
    t.Fatalf("hash failed: %v", err)
  }

  if !service.VerifyPassword(hash, "pass123") {
    t.Fatalf("expected password match")
  }
  if service.VerifyPassword(hash, "wrong") {
    t.Fatalf("expected password mismatch")
  }
}

func TestAuthServiceToken(t *testing.T) {
  cfg := &config.Config{JwtSecret: "test-secret", JwtIssuer: "test", JwtExpireHours: 1}
  service, err := services.NewAuthService(cfg)
  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }

  token, _, err := service.IssueToken(&services.AuthUser{
    ID:          10,
    Username:    "demo",
    DisplayName: "Demo",
    Role:        "admin",
  })
  if err != nil {
    t.Fatalf("issue token failed: %v", err)
  }

  claims, err := service.ParseToken(token)
  if err != nil {
    t.Fatalf("parse token failed: %v", err)
  }
  if claims.UserID != 10 || claims.Username != "demo" || claims.Role != "admin" {
    t.Fatalf("unexpected claims: %#v", claims)
  }
}
