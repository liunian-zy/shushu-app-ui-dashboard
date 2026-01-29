package store

import (
  "database/sql"
  "time"

  _ "github.com/go-sql-driver/mysql"
)

func NewMySQL(dsn string) (*sql.DB, error) {
  db, err := sql.Open("mysql", dsn)
  if err != nil {
    return nil, err
  }

  db.SetConnMaxLifetime(5 * time.Minute)
  db.SetMaxIdleConns(5)
  db.SetMaxOpenConns(20)

  if err := db.Ping(); err != nil {
    return nil, err
  }

  return db, nil
}
