package store

import (
  "database/sql"
  "fmt"
  "os"
  "path/filepath"
  "sort"
  "strings"
)

func ApplyMigrations(db *sql.DB) error {
  if db == nil {
    return fmt.Errorf("db is nil")
  }
  dir, err := discoverMigrationsDir()
  if err != nil {
    return err
  }
  entries, err := os.ReadDir(dir)
  if err != nil {
    return err
  }

  files := make([]string, 0)
  for _, entry := range entries {
    if entry.IsDir() {
      continue
    }
    name := entry.Name()
    if strings.HasSuffix(name, ".sql") {
      files = append(files, name)
    }
  }
  sort.Strings(files)

  for _, name := range files {
    path := filepath.Join(dir, name)
    raw, err := os.ReadFile(path)
    if err != nil {
      return fmt.Errorf("read migration %s failed: %w", name, err)
    }
    statements := splitSQLStatements(string(raw))
    for _, stmt := range statements {
      if strings.TrimSpace(stmt) == "" {
        continue
      }
      if _, err := db.Exec(stmt); err != nil {
        return fmt.Errorf("apply migration %s failed: %w", name, err)
      }
    }
  }
  return nil
}

func discoverMigrationsDir() (string, error) {
  candidates := []string{
    "migrations",
    filepath.Join("server", "migrations"),
  }
  for _, dir := range candidates {
    info, err := os.Stat(dir)
    if err != nil || !info.IsDir() {
      continue
    }
    return dir, nil
  }
  return "", fmt.Errorf("migrations directory not found")
}

func splitSQLStatements(content string) []string {
  cleaned := strings.ReplaceAll(content, "\r\n", "\n")
  parts := strings.Split(cleaned, ";")
  return parts
}
