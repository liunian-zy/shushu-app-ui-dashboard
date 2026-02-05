package handlers

import (
	"database/sql"
	"strconv"
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
	if appUI == nil {
		_, err := tx.Exec("DELETE FROM app_ui_fields WHERE app_version_name_id = ?", appVersionNameID)
		return err
	}
	status := int64OrDefault(appUI.Status, 1)
	targetID := int64OrDefault(appUI.TargetID, 0)
	if targetID > 0 {
		result, err := tx.Exec(
			"UPDATE app_ui_fields SET home_title_left = ?, home_title_right = ?, home_subtitle = ?, start_experience = ?, step1_music = ?, step1_music_text = ?, step1_title = ?, step2_music = ?, step2_music_text = ?, step2_title = ?, status = ?, updated_at = ?, print_wait = ? WHERE id = ?",
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
			nullIfEmpty(nullableStringValue(appUI.PrintWait)),
			targetID,
		)
		if err != nil {
			return err
		}
		if affected, _ := result.RowsAffected(); affected > 0 {
			return nil
		}
	}
	var existingID int64
	row := tx.QueryRow("SELECT id FROM app_ui_fields WHERE app_version_name_id = ? ORDER BY id DESC LIMIT 1", appVersionNameID)
	if err := row.Scan(&existingID); err != nil {
		if err != sql.ErrNoRows {
			return err
		}
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
	_, err := tx.Exec(
		"UPDATE app_ui_fields SET home_title_left = ?, home_title_right = ?, home_subtitle = ?, start_experience = ?, step1_music = ?, step1_music_text = ?, step1_title = ?, step2_music = ?, step2_music_text = ?, step2_title = ?, status = ?, updated_at = ?, print_wait = ? WHERE id = ?",
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
		nullIfEmpty(nullableStringValue(appUI.PrintWait)),
		existingID,
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
	nameToID, existingIDs, err := loadExistingNameIDSet(tx, "SELECT id, name FROM identities WHERE app_version_name = ?", appVersionName)
	if err != nil {
		return err
	}
	seenIDs := make(map[int64]struct{}, len(identities))
	for _, item := range identities {
		nameKey := strings.TrimSpace(nullableStringValue(item.Name))
		if nameKey == "" {
			continue
		}
		sort := int64OrDefault(item.Sort, 0)
		status := int64OrDefault(item.Status, 1)
		usedID := int64(0)
		if item.TargetID.Valid {
			targetID := item.TargetID.Int64
			result, err := tx.Exec(
				"UPDATE identities SET name = ?, image = ?, sort = ?, status = ?, updated_at = ? WHERE id = ?",
				nullIfEmpty(nullableStringValue(item.Name)),
				nullIfEmpty(nullableStringValue(item.Image)),
				sort,
				status,
				now,
				targetID,
			)
			if err != nil {
				return err
			}
			if affected, _ := result.RowsAffected(); affected > 0 {
				usedID = targetID
			}
		}
		if usedID == 0 {
			if existingID, ok := nameToID[nameKey]; ok {
				if _, err := tx.Exec(
					"UPDATE identities SET name = ?, image = ?, sort = ?, status = ?, updated_at = ? WHERE id = ?",
					nullIfEmpty(nullableStringValue(item.Name)),
					nullIfEmpty(nullableStringValue(item.Image)),
					sort,
					status,
					now,
					existingID,
				); err != nil {
					return err
				}
				usedID = existingID
			} else {
				result, err := tx.Exec(
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
				insertedID, _ := result.LastInsertId()
				usedID = insertedID
			}
		}
		if usedID > 0 {
			seenIDs[usedID] = struct{}{}
		}
	}
	for existingID := range existingIDs {
		if _, ok := seenIDs[existingID]; ok {
			continue
		}
		if _, err := tx.Exec("DELETE FROM identities WHERE id = ?", existingID); err != nil {
			return err
		}
	}
	return nil
}

func syncScenes(tx *sql.Tx, appVersionName string, scenes []draftSceneRow, now time.Time) error {
	nameToID, existingIDs, err := loadExistingNameIDSet(tx, "SELECT id, name FROM scenes WHERE app_version_name = ?", appVersionName)
	if err != nil {
		return err
	}
	seenIDs := make(map[int64]struct{}, len(scenes))
	for _, item := range scenes {
		nameKey := strings.TrimSpace(nullableStringValue(item.Name))
		if nameKey == "" {
			continue
		}
		sort := int64OrDefault(item.Sort, 0)
		status := int64OrDefault(item.Status, 1)
		needWatermark := int64OrDefault(item.NeedWatermark, 1)
		usedID := int64(0)
		if item.TargetID.Valid {
			targetID := item.TargetID.Int64
			result, err := tx.Exec(
				"UPDATE scenes SET name = ?, image = ?, `desc` = ?, music = ?, watermark_path = ?, need_watermark = ?, sort = ?, status = ?, updated_at = ?, oss_style = ? WHERE id = ?",
				nullIfEmpty(nullableStringValue(item.Name)),
				nullIfEmpty(nullableStringValue(item.Image)),
				nullIfEmpty(nullableStringValue(item.Desc)),
				nullIfEmpty(nullableStringValue(item.Music)),
				nullIfEmpty(nullableStringValue(item.WatermarkPath)),
				needWatermark,
				sort,
				status,
				now,
				nullIfEmpty(nullableStringValue(item.OssStyle)),
				targetID,
			)
			if err != nil {
				return err
			}
			if affected, _ := result.RowsAffected(); affected > 0 {
				usedID = targetID
			}
		}
		if usedID == 0 {
			if existingID, ok := nameToID[nameKey]; ok {
				if _, err := tx.Exec(
					"UPDATE scenes SET name = ?, image = ?, `desc` = ?, music = ?, watermark_path = ?, need_watermark = ?, sort = ?, status = ?, updated_at = ?, oss_style = ? WHERE id = ?",
					nullIfEmpty(nullableStringValue(item.Name)),
					nullIfEmpty(nullableStringValue(item.Image)),
					nullIfEmpty(nullableStringValue(item.Desc)),
					nullIfEmpty(nullableStringValue(item.Music)),
					nullIfEmpty(nullableStringValue(item.WatermarkPath)),
					needWatermark,
					sort,
					status,
					now,
					nullIfEmpty(nullableStringValue(item.OssStyle)),
					existingID,
				); err != nil {
					return err
				}
				usedID = existingID
			} else {
				result, err := tx.Exec(
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
				insertedID, _ := result.LastInsertId()
				usedID = insertedID
			}
		}
		if usedID > 0 {
			seenIDs[usedID] = struct{}{}
		}
	}
	for existingID := range existingIDs {
		if _, ok := seenIDs[existingID]; ok {
			continue
		}
		if _, err := tx.Exec("DELETE FROM scenes WHERE id = ?", existingID); err != nil {
			return err
		}
	}
	return nil
}

func syncClothesCategories(tx *sql.Tx, appVersionName string, items []draftClothesRow, now time.Time) error {
	nameToID, existingIDs, err := loadExistingNameIDSet(tx, "SELECT id, name FROM clothes_categories WHERE app_version_name = ?", appVersionName)
	if err != nil {
		return err
	}
	seenIDs := make(map[int64]struct{}, len(items))
	for _, item := range items {
		nameKey := strings.TrimSpace(nullableStringValue(item.Name))
		if nameKey == "" {
			continue
		}
		sort := int64OrDefault(item.Sort, 0)
		status := int64OrDefault(item.Status, 1)
		usedID := int64(0)
		if item.TargetID.Valid {
			targetID := item.TargetID.Int64
			result, err := tx.Exec(
				"UPDATE clothes_categories SET name = ?, image = ?, sort = ?, status = ?, updated_at = ?, music = ?, `desc` = ?, music_text = ? WHERE id = ?",
				nullIfEmpty(nullableStringValue(item.Name)),
				nullIfEmpty(nullableStringValue(item.Image)),
				sort,
				status,
				now,
				nullIfEmpty(nullableStringValue(item.Music)),
				nullIfEmpty(nullableStringValue(item.Desc)),
				nullIfEmpty(nullableStringValue(item.MusicText)),
				targetID,
			)
			if err != nil {
				return err
			}
			if affected, _ := result.RowsAffected(); affected > 0 {
				usedID = targetID
			}
		}
		if usedID == 0 {
			if existingID, ok := nameToID[nameKey]; ok {
				if _, err := tx.Exec(
					"UPDATE clothes_categories SET name = ?, image = ?, sort = ?, status = ?, updated_at = ?, music = ?, `desc` = ?, music_text = ? WHERE id = ?",
					nullIfEmpty(nullableStringValue(item.Name)),
					nullIfEmpty(nullableStringValue(item.Image)),
					sort,
					status,
					now,
					nullIfEmpty(nullableStringValue(item.Music)),
					nullIfEmpty(nullableStringValue(item.Desc)),
					nullIfEmpty(nullableStringValue(item.MusicText)),
					existingID,
				); err != nil {
					return err
				}
				usedID = existingID
			} else {
				result, err := tx.Exec(
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
				insertedID, _ := result.LastInsertId()
				usedID = insertedID
			}
		}
		if usedID > 0 {
			seenIDs[usedID] = struct{}{}
		}
	}
	for existingID := range existingIDs {
		if _, ok := seenIDs[existingID]; ok {
			continue
		}
		if _, err := tx.Exec("DELETE FROM clothes_categories WHERE id = ?", existingID); err != nil {
			return err
		}
	}
	return nil
}

func syncPhotoHobbies(tx *sql.Tx, appVersionName string, items []draftPhotoHobbyRow, now time.Time) error {
	nameToID, existingIDs, err := loadExistingNameIDSet(tx, "SELECT id, name FROM photo_hobbies WHERE app_version_name = ?", appVersionName)
	if err != nil {
		return err
	}
	seenIDs := make(map[int64]struct{}, len(items))
	for _, item := range items {
		nameKey := strings.TrimSpace(nullableStringValue(item.Name))
		if nameKey == "" {
			continue
		}
		sort := int64OrDefault(item.Sort, 0)
		status := int64OrDefault(item.Status, 1)
		usedID := int64(0)
		if item.TargetID.Valid {
			targetID := item.TargetID.Int64
			result, err := tx.Exec(
				"UPDATE photo_hobbies SET name = ?, image = ?, sort = ?, status = ?, updated_at = ?, music = ?, music_text = ?, `desc` = ? WHERE id = ?",
				nullIfEmpty(nullableStringValue(item.Name)),
				nullIfEmpty(nullableStringValue(item.Image)),
				sort,
				status,
				now,
				nullIfEmpty(nullableStringValue(item.Music)),
				nullIfEmpty(nullableStringValue(item.MusicText)),
				nullIfEmpty(nullableStringValue(item.Desc)),
				targetID,
			)
			if err != nil {
				return err
			}
			if affected, _ := result.RowsAffected(); affected > 0 {
				usedID = targetID
			}
		}
		if usedID == 0 {
			if existingID, ok := nameToID[nameKey]; ok {
				if _, err := tx.Exec(
					"UPDATE photo_hobbies SET name = ?, image = ?, sort = ?, status = ?, updated_at = ?, music = ?, music_text = ?, `desc` = ? WHERE id = ?",
					nullIfEmpty(nullableStringValue(item.Name)),
					nullIfEmpty(nullableStringValue(item.Image)),
					sort,
					status,
					now,
					nullIfEmpty(nullableStringValue(item.Music)),
					nullIfEmpty(nullableStringValue(item.MusicText)),
					nullIfEmpty(nullableStringValue(item.Desc)),
					existingID,
				); err != nil {
					return err
				}
				usedID = existingID
			} else {
				result, err := tx.Exec(
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
				insertedID, _ := result.LastInsertId()
				usedID = insertedID
			}
		}
		if usedID > 0 {
			seenIDs[usedID] = struct{}{}
		}
	}
	for existingID := range existingIDs {
		if _, ok := seenIDs[existingID]; ok {
			continue
		}
		if _, err := tx.Exec("DELETE FROM photo_hobbies WHERE id = ?", existingID); err != nil {
			return err
		}
	}
	return nil
}

func syncExtraSteps(tx *sql.Tx, appVersionNameID int64, items []draftExtraStepRow, now time.Time) error {
	keyToID, existingIDs, err := loadExistingExtraStepIDSet(tx, appVersionNameID)
	if err != nil {
		return err
	}
	seenIDs := make(map[int64]struct{}, len(items))
	for _, item := range items {
		stepIndex := int64OrDefault(item.StepIndex, 0)
		fieldName := strings.TrimSpace(nullableStringValue(item.FieldName))
		if fieldName == "" {
			continue
		}
		key := buildExtraStepKey(stepIndex, fieldName)
		status := int64OrDefault(item.Status, 1)
		usedID := int64(0)
		if item.TargetID.Valid {
			targetID := item.TargetID.Int64
			result, err := tx.Exec(
				"UPDATE config_extra_steps SET step_index = ?, field_name = ?, label = ?, music = ?, music_text = ?, status = ?, updated_at = ? WHERE id = ?",
				stepIndex,
				nullIfEmpty(nullableStringValue(item.FieldName)),
				nullIfEmpty(nullableStringValue(item.Label)),
				nullIfEmpty(nullableStringValue(item.Music)),
				nullIfEmpty(nullableStringValue(item.MusicText)),
				status,
				now,
				targetID,
			)
			if err != nil {
				return err
			}
			if affected, _ := result.RowsAffected(); affected > 0 {
				usedID = targetID
			}
		}
		if usedID == 0 {
			if existingID, ok := keyToID[key]; ok {
				if _, err := tx.Exec(
					"UPDATE config_extra_steps SET step_index = ?, field_name = ?, label = ?, music = ?, music_text = ?, status = ?, updated_at = ? WHERE id = ?",
					stepIndex,
					nullIfEmpty(nullableStringValue(item.FieldName)),
					nullIfEmpty(nullableStringValue(item.Label)),
					nullIfEmpty(nullableStringValue(item.Music)),
					nullIfEmpty(nullableStringValue(item.MusicText)),
					status,
					now,
					existingID,
				); err != nil {
					return err
				}
				usedID = existingID
			} else {
				result, err := tx.Exec(
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
				insertedID, _ := result.LastInsertId()
				usedID = insertedID
			}
		}
		if usedID > 0 {
			seenIDs[usedID] = struct{}{}
		}
	}
	for existingID := range existingIDs {
		if _, ok := seenIDs[existingID]; ok {
			continue
		}
		if _, err := tx.Exec("DELETE FROM config_extra_steps WHERE id = ?", existingID); err != nil {
			return err
		}
	}
	return nil
}

func buildSyncMappings(tx *sql.Tx, modules []string, appVersionName string, appVersionNameID int64, data draftData) ([]SyncIDMapping, error) {
	mappings := make([]SyncIDMapping, 0)
	if shouldSyncModule(modules, "app_ui_fields") && data.AppUIFields != nil {
		targetID := int64OrDefault(data.AppUIFields.TargetID, 0)
		if targetID == 0 {
			row := tx.QueryRow("SELECT id FROM app_ui_fields WHERE app_version_name_id = ? ORDER BY id DESC LIMIT 1", appVersionNameID)
			_ = row.Scan(&targetID)
		}
		if data.AppUIFields.ID > 0 && targetID > 0 {
			mappings = append(mappings, SyncIDMapping{
				ModuleKey: "app_ui_fields",
				DraftID:   data.AppUIFields.ID,
				TargetID:  targetID,
			})
		}
	}

	if shouldSyncModule(modules, "identities") {
		nameToID, err := loadNameToIDMap(tx, "SELECT id, name FROM identities WHERE app_version_name = ?", appVersionName)
		if err != nil {
			return nil, err
		}
		for _, item := range data.Identities {
			nameKey := strings.TrimSpace(nullableStringValue(item.Name))
			if item.ID == 0 || nameKey == "" {
				continue
			}
			targetID := int64OrDefault(item.TargetID, 0)
			if targetID == 0 {
				targetID = nameToID[nameKey]
			}
			if targetID > 0 {
				mappings = append(mappings, SyncIDMapping{
					ModuleKey: "identities",
					DraftID:   item.ID,
					TargetID:  targetID,
				})
			}
		}
	}

	if shouldSyncModule(modules, "scenes") {
		nameToID, err := loadNameToIDMap(tx, "SELECT id, name FROM scenes WHERE app_version_name = ?", appVersionName)
		if err != nil {
			return nil, err
		}
		for _, item := range data.Scenes {
			nameKey := strings.TrimSpace(nullableStringValue(item.Name))
			if item.ID == 0 || nameKey == "" {
				continue
			}
			targetID := int64OrDefault(item.TargetID, 0)
			if targetID == 0 {
				targetID = nameToID[nameKey]
			}
			if targetID > 0 {
				mappings = append(mappings, SyncIDMapping{
					ModuleKey: "scenes",
					DraftID:   item.ID,
					TargetID:  targetID,
				})
			}
		}
	}

	if shouldSyncModule(modules, "clothes_categories") {
		nameToID, err := loadNameToIDMap(tx, "SELECT id, name FROM clothes_categories WHERE app_version_name = ?", appVersionName)
		if err != nil {
			return nil, err
		}
		for _, item := range data.ClothesCategories {
			nameKey := strings.TrimSpace(nullableStringValue(item.Name))
			if item.ID == 0 || nameKey == "" {
				continue
			}
			targetID := int64OrDefault(item.TargetID, 0)
			if targetID == 0 {
				targetID = nameToID[nameKey]
			}
			if targetID > 0 {
				mappings = append(mappings, SyncIDMapping{
					ModuleKey: "clothes_categories",
					DraftID:   item.ID,
					TargetID:  targetID,
				})
			}
		}
	}

	if shouldSyncModule(modules, "photo_hobbies") {
		nameToID, err := loadNameToIDMap(tx, "SELECT id, name FROM photo_hobbies WHERE app_version_name = ?", appVersionName)
		if err != nil {
			return nil, err
		}
		for _, item := range data.PhotoHobbies {
			nameKey := strings.TrimSpace(nullableStringValue(item.Name))
			if item.ID == 0 || nameKey == "" {
				continue
			}
			targetID := int64OrDefault(item.TargetID, 0)
			if targetID == 0 {
				targetID = nameToID[nameKey]
			}
			if targetID > 0 {
				mappings = append(mappings, SyncIDMapping{
					ModuleKey: "photo_hobbies",
					DraftID:   item.ID,
					TargetID:  targetID,
				})
			}
		}
	}

	if shouldSyncModule(modules, "config_extra_steps") {
		keyToID, err := loadExtraStepKeyToIDMap(tx, appVersionNameID)
		if err != nil {
			return nil, err
		}
		for _, item := range data.ExtraSteps {
			fieldName := strings.TrimSpace(nullableStringValue(item.FieldName))
			if item.ID == 0 || fieldName == "" {
				continue
			}
			key := buildExtraStepKey(int64OrDefault(item.StepIndex, 0), fieldName)
			targetID := int64OrDefault(item.TargetID, 0)
			if targetID == 0 {
				targetID = keyToID[key]
			}
			if targetID > 0 {
				mappings = append(mappings, SyncIDMapping{
					ModuleKey: "config_extra_steps",
					DraftID:   item.ID,
					TargetID:  targetID,
				})
			}
		}
	}

	return mappings, nil
}

func upsertSyncIDMappings(tx *sql.Tx, draftVersionID int64, mappings []SyncIDMapping, now time.Time) error {
	if len(mappings) == 0 {
		return nil
	}
	stmt, err := tx.Prepare(
		"INSERT INTO app_db_sync_id_map (draft_version_id, module_key, draft_row_id, target_row_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE target_row_id = VALUES(target_row_id), updated_at = VALUES(updated_at)",
	)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, mapping := range mappings {
		if mapping.ModuleKey == "" || mapping.DraftID <= 0 || mapping.TargetID <= 0 {
			continue
		}
		if _, err := stmt.Exec(draftVersionID, mapping.ModuleKey, mapping.DraftID, mapping.TargetID, now, now); err != nil {
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

func loadExistingNameIDSet(tx *sql.Tx, query string, args ...interface{}) (map[string]int64, map[int64]struct{}, error) {
	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	nameToID := make(map[string]int64)
	ids := make(map[int64]struct{})
	for rows.Next() {
		var id int64
		var name sql.NullString
		if err := rows.Scan(&id, &name); err != nil {
			return nil, nil, err
		}
		ids[id] = struct{}{}
		key := strings.TrimSpace(name.String)
		if key == "" {
			continue
		}
		nameToID[key] = id
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return nameToID, ids, nil
}

func loadNameToIDMap(tx *sql.Tx, query string, args ...interface{}) (map[string]int64, error) {
	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]int64)
	for rows.Next() {
		var id int64
		var name sql.NullString
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		key := strings.TrimSpace(name.String)
		if key == "" {
			continue
		}
		result[key] = id
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func loadExistingExtraStepIDSet(tx *sql.Tx, appVersionNameID int64) (map[string]int64, map[int64]struct{}, error) {
	rows, err := tx.Query(
		"SELECT id, step_index, field_name FROM config_extra_steps WHERE app_version_name_id = ?",
		appVersionNameID,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	keyToID := make(map[string]int64)
	ids := make(map[int64]struct{})
	for rows.Next() {
		var id int64
		var stepIndex sql.NullInt64
		var fieldName sql.NullString
		if err := rows.Scan(&id, &stepIndex, &fieldName); err != nil {
			return nil, nil, err
		}
		ids[id] = struct{}{}
		key := buildExtraStepKey(int64OrDefault(stepIndex, 0), strings.TrimSpace(fieldName.String))
		if key == "" {
			continue
		}
		keyToID[key] = id
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return keyToID, ids, nil
}

func loadExtraStepKeyToIDMap(tx *sql.Tx, appVersionNameID int64) (map[string]int64, error) {
	rows, err := tx.Query(
		"SELECT id, step_index, field_name FROM config_extra_steps WHERE app_version_name_id = ?",
		appVersionNameID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]int64)
	for rows.Next() {
		var id int64
		var stepIndex sql.NullInt64
		var fieldName sql.NullString
		if err := rows.Scan(&id, &stepIndex, &fieldName); err != nil {
			return nil, err
		}
		key := buildExtraStepKey(int64OrDefault(stepIndex, 0), strings.TrimSpace(fieldName.String))
		if key == "" {
			continue
		}
		result[key] = id
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func buildExtraStepKey(stepIndex int64, fieldName string) string {
	trimmed := strings.TrimSpace(fieldName)
	if trimmed == "" {
		return ""
	}
	return strconv.FormatInt(stepIndex, 10) + ":" + trimmed
}
