package main

import (
  "log"
  "net/http"
  "strings"
  "time"

  "shushu-app-ui-dashboard/internal/config"
  apphttp "shushu-app-ui-dashboard/internal/http"
  "shushu-app-ui-dashboard/internal/store"
)

func main() {
  cfg, err := config.Load()
  if err != nil {
    log.Fatalf("load config failed: %v", err)
  }

  if cfg.AppTimezone != "" {
    if loc, err := time.LoadLocation(cfg.AppTimezone); err != nil {
      log.Printf("load timezone failed: %v", err)
    } else {
      time.Local = loc
    }
  }

  var dbErr error
  var redisErr error
  var deps apphttp.Deps

  if cfg.MysqlDSN != "" {
    deps.DB, dbErr = store.NewMySQL(cfg.MysqlDSN)
    if dbErr != nil {
      log.Printf("mysql connect failed: %v", dbErr)
    } else if strings.ToLower(strings.TrimSpace(cfg.AppMode)) != "online" {
      if err := store.ApplyMigrations(deps.DB); err != nil {
        log.Printf("apply migrations failed: %v", err)
      }
    }
  } else {
    log.Print("MYSQL_DSN not set, skip mysql connection")
  }

  deps.Redis, redisErr = store.NewRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
  if redisErr != nil {
    log.Printf("redis connect failed: %v", redisErr)
  }

  router := apphttp.NewRouter(cfg, &deps)
  server := &http.Server{
    Addr:              ":" + cfg.Port,
    Handler:           router,
    ReadHeaderTimeout: 5 * time.Second,
  }

  log.Printf("server listening on :%s", cfg.Port)
  if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
    log.Fatalf("server exited: %v", err)
  }
}
