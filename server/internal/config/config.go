package config

import (
  "os"
  "strconv"
  "strings"

  "github.com/joho/godotenv"
)

type Config struct {
  Env           string
  Port          string
  AppMode       string
  AppTimezone   string
  MysqlDSN      string
  RedisAddr     string
  RedisPassword string
  RedisDB       int
  LocalStorageRoot string
  LocalStorageBaseURL string
  OssEndpoint   string
  OssAccessKey  string
  OssSecret     string
  OssBucket     string
  OssInternal   string
  OssSignTTL    int64
  TtsBaseURL    string
  TtsAPIKey     string
  JwtSecret     string
  JwtIssuer     string
  JwtExpireHours int
  SyncTargetURL string
  SyncAPIKey    string
  SyncTimeoutSeconds int
}

func Load() (*Config, error) {
  _ = godotenv.Load("../.env", ".env")

  redisDB := 0
  if raw := os.Getenv("REDIS_DB"); raw != "" {
    parsed, err := strconv.Atoi(raw)
    if err != nil {
      return nil, err
    }
    redisDB = parsed
  }

  cfg := &Config{
    Env:           envOrDefault("APP_ENV", "dev"),
    Port:          envOrDefault("APP_PORT", "8080"),
    AppMode:       envOrDefault("APP_MODE", "internal"),
    AppTimezone:   envOrDefault("APP_TIMEZONE", "Asia/Shanghai"),
    MysqlDSN:      normalizeMySQLDSN(os.Getenv("MYSQL_DSN")),
    RedisAddr:     envOrDefault("REDIS_ADDR", "127.0.0.1:6379"),
    RedisPassword: os.Getenv("REDIS_PASSWORD"),
    RedisDB:       redisDB,
    LocalStorageRoot: envOrDefault("LOCAL_STORAGE_ROOT", "/data/shushu-app-ui/uploads"),
    LocalStorageBaseURL: envOrDefault("LOCAL_STORAGE_BASE_URL", "/api/local-files/"),
    OssEndpoint:   envOrDefault("ALI_URL", envOrDefault("ALI_ENDPOINT", "")),
    OssAccessKey:  os.Getenv("ALI_ACCESS_KEY_ID"),
    OssSecret:     os.Getenv("ALI_SECRET_ACCESS_KEY"),
    OssBucket:     os.Getenv("ALI_BUCKET"),
    OssInternal:   os.Getenv("ALI_INTERNAL_ENDPOINT"),
    OssSignTTL:    envInt64("OSS_SIGN_TTL", 3600),
    TtsBaseURL:    envOrDefault("TTS_BASE_URL", "http://127.0.0.1:3001"),
    TtsAPIKey:     os.Getenv("TTS_API_KEY"),
    JwtSecret:     envOrDefault("JWT_SECRET", "dev-secret"),
    JwtIssuer:     envOrDefault("JWT_ISSUER", "shushu-app-ui-dashboard"),
    JwtExpireHours: envInt("JWT_EXPIRE_HOURS", 24),
    SyncTargetURL: strings.TrimSpace(os.Getenv("SYNC_TARGET_URL")),
    SyncAPIKey:    strings.TrimSpace(os.Getenv("SYNC_API_KEY")),
    SyncTimeoutSeconds: envInt("SYNC_TIMEOUT_SECONDS", 20),
  }

  return cfg, nil
}

func envOrDefault(key, value string) string {
  if v := os.Getenv(key); v != "" {
    return v
  }
  return value
}

func envInt64(key string, value int64) int64 {
  if v := os.Getenv(key); v != "" {
    if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
      return parsed
    }
  }
  return value
}

func envInt(key string, value int) int {
  if v := os.Getenv(key); v != "" {
    if parsed, err := strconv.Atoi(v); err == nil {
      return parsed
    }
  }
  return value
}

func normalizeMySQLDSN(dsn string) string {
  if strings.TrimSpace(dsn) == "" {
    return dsn
  }
  dsn = ensureDSNParam(dsn, "loc", "Asia%2FShanghai")
  dsn = ensureDSNParam(dsn, "time_zone", "%27%2B08:00%27")
  return dsn
}

func ensureDSNParam(dsn, key, value string) string {
  if strings.Contains(dsn, key+"=") {
    return dsn
  }
  sep := "?"
  if strings.Contains(dsn, "?") {
    sep = "&"
  }
  return dsn + sep + key + "=" + value
}
