package handlers

import (
  "database/sql"
  "errors"
  "fmt"
  "net/http"
  "os"
  "strings"
  "time"

  "github.com/gin-gonic/gin"

  "shushu-app-ui-dashboard/internal/http/middleware"
  "shushu-app-ui-dashboard/internal/services"
)

type uploadCacheEntry struct {
  ossPath  string
  localAbs string
}

// CompleteUpload uploads local media to OSS for a task and marks it completed.
// Args:
//   c: Gin context.
// Returns:
//   None.
func (h *TaskHandler) CompleteUpload(c *gin.Context) {
  if h.db == nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
    return
  }

  claims, ok := middleware.GetAuthClaims(c)
  if !ok {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
    return
  }

  taskID, err := parseInt64ParamValue(c.Param("id"))
  if err != nil || taskID <= 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
    return
  }

  tx, err := h.db.Begin()
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "transaction failed"})
    return
  }
  defer func() {
    _ = tx.Rollback()
  }()

  var (
    draftVersionID sql.NullInt64
    moduleKey      sql.NullString
  )
  row := tx.QueryRow("SELECT draft_version_id, module_key FROM app_db_tasks WHERE id = ?", taskID)
  if err := row.Scan(&draftVersionID, &moduleKey); err != nil {
    if errors.Is(err, sql.ErrNoRows) {
      c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
    return
  }
  if !draftVersionID.Valid || draftVersionID.Int64 <= 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "draft_version_id missing"})
    return
  }
  module := strings.TrimSpace(moduleKey.String)
  if module == "" {
    c.JSON(http.StatusBadRequest, gin.H{"error": "module_key missing"})
    return
  }

  ossService, err := services.NewOSSService(h.cfg, h.redis)
  if err != nil {
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "oss not ready"})
    return
  }

  cache := make(map[string]uploadCacheEntry)
  uploadCount, err := h.uploadModuleAssets(tx, ossService, draftVersionID.Int64, module, cache)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }

  now := time.Now()
  if _, err := tx.Exec("UPDATE app_db_tasks SET status = ?, updated_by = ?, updated_at = ? WHERE id = ?", "completed", claims.UserID, now, taskID); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
    return
  }

  _ = insertTaskAction(tx, taskID, "complete_upload", claims.UserID, map[string]interface{}{
    "module_key": module,
    "uploaded":   uploadCount,
  })

  if err := tx.Commit(); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "commit failed"})
    return
  }

  for _, entry := range cache {
    if entry.localAbs != "" {
      _ = os.Remove(entry.localAbs)
    }
  }

  c.JSON(http.StatusOK, gin.H{
    "status":       "ok",
    "uploaded":     uploadCount,
    "task_id":      taskID,
    "module_key":   module,
    "draft_version_id": draftVersionID.Int64,
  })
}

func (h *TaskHandler) uploadModuleAssets(tx *sql.Tx, ossService *services.OSSService, draftVersionID int64, moduleKey string, cache map[string]uploadCacheEntry) (int, error) {
  switch moduleKey {
  case "banners":
    return h.uploadBannerAssets(tx, ossService, draftVersionID, cache)
  case "identities":
    return h.uploadIdentityAssets(tx, ossService, draftVersionID, cache)
  case "scenes":
    return h.uploadSceneAssets(tx, ossService, draftVersionID, cache)
  case "app_ui_fields", "print_wait":
    return h.uploadAppUIAssets(tx, ossService, draftVersionID, cache)
  case "config_extra_steps":
    return h.uploadExtraStepAssets(tx, ossService, draftVersionID, cache)
  case "clothes_categories":
    return h.uploadClothesAssets(tx, ossService, draftVersionID, cache)
  case "photo_hobbies":
    return h.uploadPhotoAssets(tx, ossService, draftVersionID, cache)
  default:
    return 0, errors.New("unsupported module_key")
  }
}

func (h *TaskHandler) uploadBannerAssets(tx *sql.Tx, ossService *services.OSSService, draftVersionID int64, cache map[string]uploadCacheEntry) (int, error) {
  rows, err := tx.Query("SELECT id, image FROM app_db_banners WHERE draft_version_id = ?", draftVersionID)
  if err != nil {
    return 0, err
  }
  defer rows.Close()

  type bannerAsset struct {
    id    int64
    image sql.NullString
  }
  items := make([]bannerAsset, 0)
  for rows.Next() {
    var item bannerAsset
    if err := rows.Scan(&item.id, &item.image); err != nil {
      return 0, err
    }
    items = append(items, item)
  }
  if err := rows.Err(); err != nil {
    return 0, err
  }

  uploaded := 0
  for _, item := range items {
    if !item.image.Valid || strings.TrimSpace(item.image.String) == "" {
      continue
    }
    newPath, changed, err := h.uploadIfLocal(ossService, item.image.String, cache)
    if err != nil {
      return uploaded, err
    }
    if changed {
      if _, err := tx.Exec("UPDATE app_db_banners SET image = ?, updated_at = ? WHERE id = ?", newPath, time.Now(), item.id); err != nil {
        return uploaded, err
      }
      uploaded++
    }
  }
  return uploaded, nil
}

func (h *TaskHandler) uploadIdentityAssets(tx *sql.Tx, ossService *services.OSSService, draftVersionID int64, cache map[string]uploadCacheEntry) (int, error) {
  rows, err := tx.Query("SELECT id, image FROM app_db_identities WHERE draft_version_id = ?", draftVersionID)
  if err != nil {
    return 0, err
  }
  defer rows.Close()

  type identityAsset struct {
    id    int64
    image sql.NullString
  }
  items := make([]identityAsset, 0)
  for rows.Next() {
    var item identityAsset
    if err := rows.Scan(&item.id, &item.image); err != nil {
      return 0, err
    }
    items = append(items, item)
  }
  if err := rows.Err(); err != nil {
    return 0, err
  }

  uploaded := 0
  for _, item := range items {
    if !item.image.Valid || strings.TrimSpace(item.image.String) == "" {
      continue
    }
    newPath, changed, err := h.uploadIfLocal(ossService, item.image.String, cache)
    if err != nil {
      return uploaded, err
    }
    if changed {
      if _, err := tx.Exec("UPDATE app_db_identities SET image = ?, updated_at = ? WHERE id = ?", newPath, time.Now(), item.id); err != nil {
        return uploaded, err
      }
      uploaded++
    }
  }
  return uploaded, nil
}

func (h *TaskHandler) uploadSceneAssets(tx *sql.Tx, ossService *services.OSSService, draftVersionID int64, cache map[string]uploadCacheEntry) (int, error) {
  rows, err := tx.Query("SELECT id, image, music, watermark_path FROM app_db_scenes WHERE draft_version_id = ?", draftVersionID)
  if err != nil {
    return 0, err
  }
  defer rows.Close()

  type sceneAsset struct {
    id        int64
    image     sql.NullString
    music     sql.NullString
    watermark sql.NullString
  }
  items := make([]sceneAsset, 0)
  for rows.Next() {
    var item sceneAsset
    if err := rows.Scan(&item.id, &item.image, &item.music, &item.watermark); err != nil {
      return 0, err
    }
    items = append(items, item)
  }
  if err := rows.Err(); err != nil {
    return 0, err
  }

  uploaded := 0
  for _, item := range items {
    updated := false
    if item.image.Valid && strings.TrimSpace(item.image.String) != "" {
      newPath, changed, err := h.uploadIfLocal(ossService, item.image.String, cache)
      if err != nil {
        return uploaded, err
      }
      if changed {
        item.image.String = newPath
        updated = true
        uploaded++
      }
    }
    if item.music.Valid && strings.TrimSpace(item.music.String) != "" {
      newPath, changed, err := h.uploadIfLocal(ossService, item.music.String, cache)
      if err != nil {
        return uploaded, err
      }
      if changed {
        item.music.String = newPath
        updated = true
        uploaded++
      }
    }
    if item.watermark.Valid && strings.TrimSpace(item.watermark.String) != "" {
      newPath, changed, err := h.uploadIfLocal(ossService, item.watermark.String, cache)
      if err != nil {
        return uploaded, err
      }
      if changed {
        item.watermark.String = newPath
        updated = true
        uploaded++
      }
    }
    if updated {
      if _, err := tx.Exec(
        "UPDATE app_db_scenes SET image = ?, music = ?, watermark_path = ?, updated_at = ? WHERE id = ?",
        nullableStringValue(item.image),
        nullableStringValue(item.music),
        nullableStringValue(item.watermark),
        time.Now(),
        item.id,
      ); err != nil {
        return uploaded, err
      }
    }
  }
  return uploaded, nil
}

func (h *TaskHandler) uploadAppUIAssets(tx *sql.Tx, ossService *services.OSSService, draftVersionID int64, cache map[string]uploadCacheEntry) (int, error) {
  row := tx.QueryRow(
    "SELECT id, step1_music, step2_music, print_wait FROM app_db_app_ui_fields WHERE draft_version_id = ?",
    draftVersionID,
  )
  var (
    id        int64
    step1     sql.NullString
    step2     sql.NullString
    printWait sql.NullString
  )
  if err := row.Scan(&id, &step1, &step2, &printWait); err != nil {
    if errors.Is(err, sql.ErrNoRows) {
      return 0, nil
    }
    return 0, err
  }

  uploaded := 0
  updated := false
  if step1.Valid && strings.TrimSpace(step1.String) != "" {
    newPath, changed, err := h.uploadIfLocal(ossService, step1.String, cache)
    if err != nil {
      return uploaded, err
    }
    if changed {
      step1.String = newPath
      updated = true
      uploaded++
    }
  }
  if step2.Valid && strings.TrimSpace(step2.String) != "" {
    newPath, changed, err := h.uploadIfLocal(ossService, step2.String, cache)
    if err != nil {
      return uploaded, err
    }
    if changed {
      step2.String = newPath
      updated = true
      uploaded++
    }
  }
  if printWait.Valid && strings.TrimSpace(printWait.String) != "" {
    newPath, changed, err := h.uploadIfLocal(ossService, printWait.String, cache)
    if err != nil {
      return uploaded, err
    }
    if changed {
      printWait.String = newPath
      updated = true
      uploaded++
    }
  }

  if updated {
    if _, err := tx.Exec(
      "UPDATE app_db_app_ui_fields SET step1_music = ?, step2_music = ?, print_wait = ?, updated_at = ? WHERE id = ?",
      nullableStringValue(step1),
      nullableStringValue(step2),
      nullableStringValue(printWait),
      time.Now(),
      id,
    ); err != nil {
      return uploaded, err
    }
  }
  return uploaded, nil
}

func (h *TaskHandler) uploadExtraStepAssets(tx *sql.Tx, ossService *services.OSSService, draftVersionID int64, cache map[string]uploadCacheEntry) (int, error) {
  rows, err := tx.Query("SELECT id, music FROM app_db_config_extra_steps WHERE draft_version_id = ?", draftVersionID)
  if err != nil {
    return 0, err
  }
  defer rows.Close()

  type extraStepAsset struct {
    id    int64
    music sql.NullString
  }
  items := make([]extraStepAsset, 0)
  for rows.Next() {
    var item extraStepAsset
    if err := rows.Scan(&item.id, &item.music); err != nil {
      return 0, err
    }
    items = append(items, item)
  }
  if err := rows.Err(); err != nil {
    return 0, err
  }

  uploaded := 0
  for _, item := range items {
    if !item.music.Valid || strings.TrimSpace(item.music.String) == "" {
      continue
    }
    newPath, changed, err := h.uploadIfLocal(ossService, item.music.String, cache)
    if err != nil {
      return uploaded, err
    }
    if changed {
      if _, err := tx.Exec("UPDATE app_db_config_extra_steps SET music = ?, updated_at = ? WHERE id = ?", newPath, time.Now(), item.id); err != nil {
        return uploaded, err
      }
      uploaded++
    }
  }
  return uploaded, nil
}

func (h *TaskHandler) uploadClothesAssets(tx *sql.Tx, ossService *services.OSSService, draftVersionID int64, cache map[string]uploadCacheEntry) (int, error) {
  rows, err := tx.Query("SELECT id, image, music FROM app_db_clothes_categories WHERE draft_version_id = ?", draftVersionID)
  if err != nil {
    return 0, err
  }
  defer rows.Close()

  type clothesAsset struct {
    id    int64
    image sql.NullString
    music sql.NullString
  }
  items := make([]clothesAsset, 0)
  for rows.Next() {
    var item clothesAsset
    if err := rows.Scan(&item.id, &item.image, &item.music); err != nil {
      return 0, err
    }
    items = append(items, item)
  }
  if err := rows.Err(); err != nil {
    return 0, err
  }

  uploaded := 0
  for _, item := range items {
    updated := false
    if item.image.Valid && strings.TrimSpace(item.image.String) != "" {
      newPath, changed, err := h.uploadIfLocal(ossService, item.image.String, cache)
      if err != nil {
        return uploaded, err
      }
      if changed {
        item.image.String = newPath
        updated = true
        uploaded++
      }
    }
    if item.music.Valid && strings.TrimSpace(item.music.String) != "" {
      newPath, changed, err := h.uploadIfLocal(ossService, item.music.String, cache)
      if err != nil {
        return uploaded, err
      }
      if changed {
        item.music.String = newPath
        updated = true
        uploaded++
      }
    }
    if updated {
      if _, err := tx.Exec(
        "UPDATE app_db_clothes_categories SET image = ?, music = ?, updated_at = ? WHERE id = ?",
        nullableStringValue(item.image),
        nullableStringValue(item.music),
        time.Now(),
        item.id,
      ); err != nil {
        return uploaded, err
      }
    }
  }
  return uploaded, nil
}

func (h *TaskHandler) uploadPhotoAssets(tx *sql.Tx, ossService *services.OSSService, draftVersionID int64, cache map[string]uploadCacheEntry) (int, error) {
  rows, err := tx.Query("SELECT id, image, music FROM app_db_photo_hobbies WHERE draft_version_id = ?", draftVersionID)
  if err != nil {
    return 0, err
  }
  defer rows.Close()

  type photoAsset struct {
    id    int64
    image sql.NullString
    music sql.NullString
  }
  items := make([]photoAsset, 0)
  for rows.Next() {
    var item photoAsset
    if err := rows.Scan(&item.id, &item.image, &item.music); err != nil {
      return 0, err
    }
    items = append(items, item)
  }
  if err := rows.Err(); err != nil {
    return 0, err
  }

  uploaded := 0
  for _, item := range items {
    updated := false
    if item.image.Valid && strings.TrimSpace(item.image.String) != "" {
      newPath, changed, err := h.uploadIfLocal(ossService, item.image.String, cache)
      if err != nil {
        return uploaded, err
      }
      if changed {
        item.image.String = newPath
        updated = true
        uploaded++
      }
    }
    if item.music.Valid && strings.TrimSpace(item.music.String) != "" {
      newPath, changed, err := h.uploadIfLocal(ossService, item.music.String, cache)
      if err != nil {
        return uploaded, err
      }
      if changed {
        item.music.String = newPath
        updated = true
        uploaded++
      }
    }
    if updated {
      if _, err := tx.Exec(
        "UPDATE app_db_photo_hobbies SET image = ?, music = ?, updated_at = ? WHERE id = ?",
        nullableStringValue(item.image),
        nullableStringValue(item.music),
        time.Now(),
        item.id,
      ); err != nil {
        return uploaded, err
      }
    }
  }
  return uploaded, nil
}

func (h *TaskHandler) uploadIfLocal(ossService *services.OSSService, storagePath string, cache map[string]uploadCacheEntry) (string, bool, error) {
  trimmed := strings.TrimSpace(storagePath)
  if trimmed == "" {
    return storagePath, false, nil
  }
  if !isLocalPath(trimmed) {
    return trimmed, false, nil
  }
  if entry, ok := cache[trimmed]; ok {
    return entry.ossPath, true, nil
  }

  relative := trimLocalPrefix(trimmed)
  localAbs, err := buildLocalFilePath(h.cfg, relative)
  if err != nil {
    return "", false, err
  }
  if _, err := os.Stat(localAbs); err != nil {
    return "", false, fmt.Errorf("local file missing: %s", relative)
  }
  if ossService == nil {
    return "", false, errors.New("oss not ready")
  }
  if err := ossService.UploadFileFromPath(relative, localAbs); err != nil {
    return "", false, err
  }
  cache[trimmed] = uploadCacheEntry{
    ossPath:  relative,
    localAbs: localAbs,
  }
  return relative, true, nil
}

func parseInt64ParamValue(raw string) (int64, error) {
  trimmed := strings.TrimSpace(raw)
  if trimmed == "" {
    return 0, errors.New("empty id")
  }
  return parseInt64(trimmed)
}
