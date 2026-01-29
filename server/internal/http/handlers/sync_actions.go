package handlers

import (
  "database/sql"
  "strings"
  "time"
)

func updateDraftSyncStatus(db *sql.DB, draftVersionID int64, status, message string, targetID int64) error {
  _, err := db.Exec(
    "UPDATE app_db_version_names SET sync_status = ?, sync_message = ?, target_app_version_name_id = ? WHERE id = ?",
    status,
    nullIfEmpty(message),
    nullableID(targetID),
    draftVersionID,
  )
  return err
}

func updateDraftSyncStatusTx(tx *sql.Tx, draftVersionID int64, status, message string, targetID int64, now time.Time) error {
  _, err := tx.Exec(
    "UPDATE app_db_version_names SET sync_status = ?, sync_message = ?, target_app_version_name_id = ?, synced_at = ? WHERE id = ?",
    status,
    nullIfEmpty(message),
    nullableID(targetID),
    now,
    draftVersionID,
  )
  return err
}

func insertSyncJob(tx *sql.Tx, draftVersionID, triggerBy int64, status string, now time.Time) (int64, error) {
  result, err := tx.Exec(
    "INSERT INTO app_db_sync_jobs (draft_version_id, trigger_by, status, started_at, created_at) VALUES (?, ?, ?, ?, ?)",
    draftVersionID,
    triggerBy,
    status,
    now,
    now,
  )
  if err != nil {
    return 0, err
  }
  return result.LastInsertId()
}

func insertSyncModuleJob(tx *sql.Tx, draftVersionID, triggerBy int64, moduleKey, status string, now time.Time) (int64, error) {
  result, err := tx.Exec(
    "INSERT INTO app_db_sync_module_jobs (draft_version_id, module_key, trigger_by, status, started_at, created_at) VALUES (?, ?, ?, ?, ?, ?)",
    draftVersionID,
    moduleKey,
    triggerBy,
    status,
    now,
    now,
  )
  if err != nil {
    return 0, err
  }
  return result.LastInsertId()
}

func finishSyncJob(tx *sql.Tx, jobID int64, now time.Time) error {
  _, err := tx.Exec(
    "UPDATE app_db_sync_jobs SET status = ?, finished_at = ? WHERE id = ?",
    "success",
    now,
    jobID,
  )
  return err
}

func finishSyncModuleJob(tx *sql.Tx, jobID int64, now time.Time) error {
  _, err := tx.Exec(
    "UPDATE app_db_sync_module_jobs SET status = ?, finished_at = ? WHERE id = ?",
    "success",
    now,
    jobID,
  )
  return err
}

func failSyncJob(tx *sql.Tx, jobID int64, message string, now time.Time) error {
  _, err := tx.Exec(
    "UPDATE app_db_sync_jobs SET status = ?, error_message = ?, finished_at = ? WHERE id = ?",
    "failed",
    nullIfEmpty(message),
    now,
    jobID,
  )
  return err
}

func failSyncModuleJob(tx *sql.Tx, jobID int64, status, message string, now time.Time) error {
  if status == "" {
    status = "failed"
  }
  _, err := tx.Exec(
    "UPDATE app_db_sync_module_jobs SET status = ?, error_message = ?, finished_at = ? WHERE id = ?",
    status,
    nullIfEmpty(message),
    now,
    jobID,
  )
  return err
}

func failSync(tx *sql.Tx, draftVersionID, targetID, jobID int64, message string, now time.Time) {
  _ = updateDraftSyncStatusTx(tx, draftVersionID, "failed", message, targetID, now)
  _ = failSyncJob(tx, jobID, message, now)
}

func updateAppVersionName(tx *sql.Tx, id int64, appVersionName, locationName string, status int64, feishuFields, aiModal string, now time.Time) error {
  _, err := tx.Exec(
    "UPDATE app_version_names SET app_version_name = ?, location_name = ?, status = ?, updated_at = ?, feishu_field_names = ?, ai_modal = ? WHERE id = ?",
    appVersionName,
    locationName,
    status,
    now,
    nullIfEmpty(feishuFields),
    aiModal,
    id,
  )
  return err
}

func insertAppVersionName(tx *sql.Tx, appVersionName, locationName string, status int64, feishuFields, aiModal string, now time.Time) (int64, error) {
  result, err := tx.Exec(
    "INSERT INTO app_version_names (app_version_name, location_name, status, created_at, updated_at, feishu_field_names, ai_modal) VALUES (?, ?, ?, ?, ?, ?, ?)",
    appVersionName,
    locationName,
    status,
    now,
    now,
    nullIfEmpty(feishuFields),
    aiModal,
  )
  if err != nil {
    return 0, err
  }
  return result.LastInsertId()
}

func syncAppUIFields(tx *sql.Tx, appVersionNameID int64, appUI *draftAppUIRow, now time.Time) error {
  if _, err := tx.Exec("DELETE FROM app_ui_fields WHERE app_version_name_id = ?", appVersionNameID); err != nil {
    return err
  }
  if appUI == nil {
    return nil
  }
  status := int64OrDefault(appUI.Status, 1)
  _, err := tx.Exec(
    "INSERT INTO app_ui_fields (app_version_name_id, home_title_left, home_title_right, home_subtitle, start_experience, step1_music, step1_music_text, step1_title, step2_music, step2_music_text, step2_title, status, created_at, updated_at, print_wait) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
    appVersionNameID,
    nullIfEmpty(nullableStringValue(appUI.HomeTitleLeft)),
    nullIfEmpty(nullableStringValue(appUI.HomeTitleRight)),
    nullIfEmpty(nullableStringValue(appUI.HomeSubtitle)),
    nullIfEmpty(nullableStringValue(appUI.StartExperience)),
    nullIfEmpty(nullableStringValue(appUI.Step1Music)),
    nullIfEmpty(nullableStringValue(appUI.Step1MusicText)),
    nullIfEmpty(nullableStringValue(appUI.Step1Title)),
    nullIfEmpty(nullableStringValue(appUI.Step2Music)),
    nullIfEmpty(nullableStringValue(appUI.Step2MusicText)),
    nullIfEmpty(nullableStringValue(appUI.Step2Title)),
    status,
    now,
    now,
    nullIfEmpty(nullableStringValue(appUI.PrintWait)),
  )
  return err
}

func syncBanners(tx *sql.Tx, appVersionName string, banners []draftBannerRow, now time.Time) error {
  if _, err := tx.Exec("DELETE FROM banners WHERE app_version_name = ?", appVersionName); err != nil {
    return err
  }
  for _, banner := range banners {
    sort := int64OrDefault(banner.Sort, 0)
    isActive := int64OrDefault(banner.IsActive, 1)
    bannerType := mapBannerTypeForSync(int64OrDefault(banner.Type, 0))
    _, err := tx.Exec(
      "INSERT INTO banners (title, image, sort, is_active, created_at, updated_at, type, app_version_name) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
      nullIfEmpty(nullableStringValue(banner.Title)),
      nullIfEmpty(nullableStringValue(banner.Image)),
      sort,
      isActive,
      now,
      now,
      bannerType,
      appVersionName,
    )
    if err != nil {
      return err
    }
  }
  return nil
}

func mapBannerTypeForSync(value int64) int64 {
  if value == 3 {
    return 0
  }
  return value
}

func syncIdentities(tx *sql.Tx, appVersionName string, identities []draftIdentityRow, now time.Time) error {
  if _, err := tx.Exec("DELETE FROM identities WHERE app_version_name = ?", appVersionName); err != nil {
    return err
  }
  for _, item := range identities {
    sort := int64OrDefault(item.Sort, 0)
    status := int64OrDefault(item.Status, 1)
    _, err := tx.Exec(
      "INSERT INTO identities (name, image, sort, status, created_at, updated_at, app_version_name) VALUES (?, ?, ?, ?, ?, ?, ?)",
      nullIfEmpty(nullableStringValue(item.Name)),
      nullIfEmpty(nullableStringValue(item.Image)),
      sort,
      status,
      now,
      now,
      appVersionName,
    )
    if err != nil {
      return err
    }
  }
  return nil
}

func syncScenes(tx *sql.Tx, appVersionName string, scenes []draftSceneRow, now time.Time) error {
  if _, err := tx.Exec("DELETE FROM scenes WHERE app_version_name = ?", appVersionName); err != nil {
    return err
  }
  for _, item := range scenes {
    sort := int64OrDefault(item.Sort, 0)
    status := int64OrDefault(item.Status, 1)
    needWatermark := int64OrDefault(item.NeedWatermark, 1)
    _, err := tx.Exec(
      "INSERT INTO scenes (name, image, `desc`, music, watermark_path, need_watermark, sort, status, created_at, updated_at, app_version_name, oss_style) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
      nullIfEmpty(nullableStringValue(item.Name)),
      nullIfEmpty(nullableStringValue(item.Image)),
      nullIfEmpty(nullableStringValue(item.Desc)),
      nullIfEmpty(nullableStringValue(item.Music)),
      nullIfEmpty(nullableStringValue(item.WatermarkPath)),
      needWatermark,
      sort,
      status,
      now,
      now,
      appVersionName,
      nullIfEmpty(nullableStringValue(item.OssStyle)),
    )
    if err != nil {
      return err
    }
  }
  return nil
}

func syncClothesCategories(tx *sql.Tx, appVersionName string, items []draftClothesRow, now time.Time) error {
  if _, err := tx.Exec("DELETE FROM clothes_categories WHERE app_version_name = ?", appVersionName); err != nil {
    return err
  }
  for _, item := range items {
    sort := int64OrDefault(item.Sort, 0)
    status := int64OrDefault(item.Status, 1)
    _, err := tx.Exec(
      "INSERT INTO clothes_categories (name, image, sort, status, created_at, updated_at, music, `desc`, music_text, app_version_name) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
      nullIfEmpty(nullableStringValue(item.Name)),
      nullIfEmpty(nullableStringValue(item.Image)),
      sort,
      status,
      now,
      now,
      nullIfEmpty(nullableStringValue(item.Music)),
      nullIfEmpty(nullableStringValue(item.Desc)),
      nullIfEmpty(nullableStringValue(item.MusicText)),
      appVersionName,
    )
    if err != nil {
      return err
    }
  }
  return nil
}

func syncPhotoHobbies(tx *sql.Tx, appVersionName string, items []draftPhotoHobbyRow, now time.Time) error {
  if _, err := tx.Exec("DELETE FROM photo_hobbies WHERE app_version_name = ?", appVersionName); err != nil {
    return err
  }
  for _, item := range items {
    sort := int64OrDefault(item.Sort, 0)
    status := int64OrDefault(item.Status, 1)
    _, err := tx.Exec(
      "INSERT INTO photo_hobbies (name, image, sort, status, created_at, updated_at, music, music_text, `desc`, app_version_name) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
      nullIfEmpty(nullableStringValue(item.Name)),
      nullIfEmpty(nullableStringValue(item.Image)),
      sort,
      status,
      now,
      now,
      nullIfEmpty(nullableStringValue(item.Music)),
      nullIfEmpty(nullableStringValue(item.MusicText)),
      nullIfEmpty(nullableStringValue(item.Desc)),
      appVersionName,
    )
    if err != nil {
      return err
    }
  }
  return nil
}

func syncExtraSteps(tx *sql.Tx, appVersionNameID int64, items []draftExtraStepRow, now time.Time) error {
  if _, err := tx.Exec("DELETE FROM config_extra_steps WHERE app_version_name_id = ?", appVersionNameID); err != nil {
    return err
  }
  for _, item := range items {
    stepIndex := int64OrDefault(item.StepIndex, 0)
    status := int64OrDefault(item.Status, 1)
    _, err := tx.Exec(
      "INSERT INTO config_extra_steps (app_version_name_id, step_index, field_name, label, music, music_text, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
      appVersionNameID,
      stepIndex,
      nullIfEmpty(nullableStringValue(item.FieldName)),
      nullIfEmpty(nullableStringValue(item.Label)),
      nullIfEmpty(nullableStringValue(item.Music)),
      nullIfEmpty(nullableStringValue(item.MusicText)),
      status,
      now,
      now,
    )
    if err != nil {
      return err
    }
  }
  return nil
}

func buildSyncAuditPayload(data draftData) map[string]interface{} {
  payload := map[string]interface{}{
    "banners":            len(data.Banners),
    "identities":         len(data.Identities),
    "scenes":             len(data.Scenes),
    "clothes_categories": len(data.ClothesCategories),
    "photo_hobbies":      len(data.PhotoHobbies),
    "config_extra_steps": len(data.ExtraSteps),
  }
  if data.AppUIFields != nil {
    payload["app_ui_fields"] = 1
  } else {
    payload["app_ui_fields"] = 0
  }
  return payload
}

func int64OrDefault(value sql.NullInt64, fallback int64) int64 {
  if value.Valid {
    return value.Int64
  }
  return fallback
}

func nullIfEmpty(value string) interface{} {
  trimmed := strings.TrimSpace(value)
  if trimmed == "" {
    return nil
  }
  return trimmed
}

func nullableID(value int64) interface{} {
  if value == 0 {
    return nil
  }
  return value
}
