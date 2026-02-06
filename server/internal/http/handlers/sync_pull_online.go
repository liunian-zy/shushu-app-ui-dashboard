package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ListVersions returns version list from online tables.
// Args:
//
//	c: Gin context.
//
// Returns:
//
//	None.
func (h *SyncPushHandler) ListVersions(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
		return
	}

	rows, err := h.db.Query(
		"SELECT id, app_version_name, location_name, feishu_field_names, ai_modal, status, updated_at FROM app_version_names ORDER BY id DESC",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	defer rows.Close()

	items := make([]SyncRemoteVersion, 0)
	for rows.Next() {
		var (
			id           int64
			name         sql.NullString
			locationName sql.NullString
			feishuFields sql.NullString
			aiModal      sql.NullString
			status       sql.NullInt64
			updatedAt    sql.NullTime
		)
		if err := rows.Scan(&id, &name, &locationName, &feishuFields, &aiModal, &status, &updatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed"})
			return
		}
		items = append(items, SyncRemoteVersion{
			TargetID:         id,
			AppVersionName:   nullableStringValue(name),
			LocationName:     nullableStringValue(locationName),
			FeishuFieldNames: nullableStringValue(feishuFields),
			AiModal:          nullableStringValue(aiModal),
			Status:           int64Pointer(status),
			UpdatedAt:        nullableTimePointer(updatedAt),
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

// Snapshot returns full module snapshot for a target online version.
// Args:
//
//	c: Gin context.
//
// Returns:
//
//	None.
func (h *SyncPushHandler) Snapshot(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
		return
	}

	targetID := parseInt64Query(c, "target_app_version_name_id")
	appVersionName := strings.TrimSpace(c.Query("app_version_name"))
	if targetID <= 0 && appVersionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_app_version_name_id or app_version_name is required"})
		return
	}

	if targetID <= 0 {
		id, err := findAppVersionNameID(h.db, appVersionName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
			return
		}
		targetID = id
	}
	if targetID <= 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
		return
	}

	version, err := h.loadRemoteVersion(targetID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	appUI, err := h.loadRemoteAppUIFields(targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	banners, err := h.loadRemoteBanners(version.AppVersionName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	identities, err := h.loadRemoteIdentities(version.AppVersionName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	scenes, err := h.loadRemoteScenes(version.AppVersionName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	clothes, err := h.loadRemoteClothesCategories(version.AppVersionName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	photoHobbies, err := h.loadRemotePhotoHobbies(version.AppVersionName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	extraSteps, err := h.loadRemoteExtraSteps(targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": SyncPullSnapshot{
			Version:           version,
			AppUIFields:       appUI,
			Banners:           banners,
			Identities:        identities,
			Scenes:            scenes,
			ClothesCategories: clothes,
			PhotoHobbies:      photoHobbies,
			ExtraSteps:        extraSteps,
		},
	})
}

func (h *SyncPushHandler) loadRemoteVersion(targetID int64) (SyncRemoteVersion, error) {
	row := h.db.QueryRow(
		"SELECT id, app_version_name, location_name, feishu_field_names, ai_modal, status, updated_at FROM app_version_names WHERE id = ? LIMIT 1",
		targetID,
	)
	var (
		id           int64
		name         sql.NullString
		locationName sql.NullString
		feishuFields sql.NullString
		aiModal      sql.NullString
		status       sql.NullInt64
		updatedAt    sql.NullTime
	)
	if err := row.Scan(&id, &name, &locationName, &feishuFields, &aiModal, &status, &updatedAt); err != nil {
		return SyncRemoteVersion{}, err
	}
	return SyncRemoteVersion{
		TargetID:         id,
		AppVersionName:   nullableStringValue(name),
		LocationName:     nullableStringValue(locationName),
		FeishuFieldNames: nullableStringValue(feishuFields),
		AiModal:          nullableStringValue(aiModal),
		Status:           int64Pointer(status),
		UpdatedAt:        nullableTimePointer(updatedAt),
	}, nil
}

func (h *SyncPushHandler) loadRemoteAppUIFields(targetID int64) (*SyncPushAppUIFields, error) {
	row := h.db.QueryRow(
		"SELECT id, home_title_left, home_title_right, home_subtitle, start_experience, step1_music, step1_music_text, step1_title, step2_music, step2_music_text, step2_title, status, print_wait FROM app_ui_fields WHERE app_version_name_id = ? ORDER BY id DESC LIMIT 1",
		targetID,
	)
	var (
		id              int64
		homeTitleLeft   sql.NullString
		homeTitleRight  sql.NullString
		homeSubtitle    sql.NullString
		startExperience sql.NullString
		step1Music      sql.NullString
		step1MusicText  sql.NullString
		step1Title      sql.NullString
		step2Music      sql.NullString
		step2MusicText  sql.NullString
		step2Title      sql.NullString
		status          sql.NullInt64
		printWait       sql.NullString
	)
	if err := row.Scan(
		&id,
		&homeTitleLeft,
		&homeTitleRight,
		&homeSubtitle,
		&startExperience,
		&step1Music,
		&step1MusicText,
		&step1Title,
		&step2Music,
		&step2MusicText,
		&step2Title,
		&status,
		&printWait,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &SyncPushAppUIFields{
		ID:              id,
		HomeTitleLeft:   nullableStringValue(homeTitleLeft),
		HomeTitleRight:  nullableStringValue(homeTitleRight),
		HomeSubtitle:    nullableStringValue(homeSubtitle),
		StartExperience: nullableStringValue(startExperience),
		Step1Music:      nullableStringValue(step1Music),
		Step1MusicText:  nullableStringValue(step1MusicText),
		Step1Title:      nullableStringValue(step1Title),
		Step2Music:      nullableStringValue(step2Music),
		Step2MusicText:  nullableStringValue(step2MusicText),
		Step2Title:      nullableStringValue(step2Title),
		Status:          int64Pointer(status),
		PrintWait:       nullableStringValue(printWait),
	}, nil
}

func (h *SyncPushHandler) loadRemoteBanners(appVersionName string) ([]SyncPushBanner, error) {
	rows, err := h.db.Query(
		"SELECT id, title, image, sort, is_active, type FROM banners WHERE app_version_name = ? ORDER BY sort ASC, id ASC",
		appVersionName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]SyncPushBanner, 0)
	for rows.Next() {
		var (
			id       int64
			title    sql.NullString
			image    sql.NullString
			sort     sql.NullInt64
			isActive sql.NullInt64
			typeVal  sql.NullInt64
		)
		if err := rows.Scan(&id, &title, &image, &sort, &isActive, &typeVal); err != nil {
			return nil, err
		}
		items = append(items, SyncPushBanner{
			ID:       id,
			Title:    nullableStringValue(title),
			Image:    nullableStringValue(image),
			Sort:     int64Pointer(sort),
			IsActive: int64Pointer(isActive),
			Type:     int64Pointer(typeVal),
		})
	}
	return items, nil
}

func (h *SyncPushHandler) loadRemoteIdentities(appVersionName string) ([]SyncPushIdentity, error) {
	rows, err := h.db.Query(
		"SELECT id, name, image, sort, status FROM identities WHERE app_version_name = ? ORDER BY sort ASC, id ASC",
		appVersionName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]SyncPushIdentity, 0)
	for rows.Next() {
		var (
			id     int64
			name   sql.NullString
			image  sql.NullString
			sort   sql.NullInt64
			status sql.NullInt64
		)
		if err := rows.Scan(&id, &name, &image, &sort, &status); err != nil {
			return nil, err
		}
		items = append(items, SyncPushIdentity{
			ID:     id,
			Name:   nullableStringValue(name),
			Image:  nullableStringValue(image),
			Sort:   int64Pointer(sort),
			Status: int64Pointer(status),
		})
	}
	return items, nil
}

func (h *SyncPushHandler) loadRemoteScenes(appVersionName string) ([]SyncPushScene, error) {
	rows, err := h.db.Query(
		"SELECT id, name, image, `desc`, music, watermark_path, need_watermark, sort, status, oss_style FROM scenes WHERE app_version_name = ? ORDER BY sort ASC, id ASC",
		appVersionName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]SyncPushScene, 0)
	for rows.Next() {
		var (
			id            int64
			name          sql.NullString
			image         sql.NullString
			desc          sql.NullString
			music         sql.NullString
			watermarkPath sql.NullString
			needWatermark sql.NullInt64
			sort          sql.NullInt64
			status        sql.NullInt64
			ossStyle      sql.NullString
		)
		if err := rows.Scan(
			&id,
			&name,
			&image,
			&desc,
			&music,
			&watermarkPath,
			&needWatermark,
			&sort,
			&status,
			&ossStyle,
		); err != nil {
			return nil, err
		}
		items = append(items, SyncPushScene{
			ID:            id,
			Name:          nullableStringValue(name),
			Image:         nullableStringValue(image),
			Desc:          nullableStringValue(desc),
			Music:         nullableStringValue(music),
			WatermarkPath: nullableStringValue(watermarkPath),
			NeedWatermark: int64Pointer(needWatermark),
			Sort:          int64Pointer(sort),
			Status:        int64Pointer(status),
			OssStyle:      nullableStringValue(ossStyle),
		})
	}
	return items, nil
}

func (h *SyncPushHandler) loadRemoteClothesCategories(appVersionName string) ([]SyncPushClothes, error) {
	rows, err := h.db.Query(
		"SELECT id, name, image, sort, status, music, `desc`, music_text FROM clothes_categories WHERE app_version_name = ? ORDER BY sort ASC, id ASC",
		appVersionName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]SyncPushClothes, 0)
	for rows.Next() {
		var (
			id        int64
			name      sql.NullString
			image     sql.NullString
			sort      sql.NullInt64
			status    sql.NullInt64
			music     sql.NullString
			desc      sql.NullString
			musicText sql.NullString
		)
		if err := rows.Scan(&id, &name, &image, &sort, &status, &music, &desc, &musicText); err != nil {
			return nil, err
		}
		items = append(items, SyncPushClothes{
			ID:        id,
			Name:      nullableStringValue(name),
			Image:     nullableStringValue(image),
			Sort:      int64Pointer(sort),
			Status:    int64Pointer(status),
			Music:     nullableStringValue(music),
			Desc:      nullableStringValue(desc),
			MusicText: nullableStringValue(musicText),
		})
	}
	return items, nil
}

func (h *SyncPushHandler) loadRemotePhotoHobbies(appVersionName string) ([]SyncPushPhotoHobby, error) {
	rows, err := h.db.Query(
		"SELECT id, name, image, sort, status, music, music_text, `desc` FROM photo_hobbies WHERE app_version_name = ? ORDER BY sort ASC, id ASC",
		appVersionName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]SyncPushPhotoHobby, 0)
	for rows.Next() {
		var (
			id        int64
			name      sql.NullString
			image     sql.NullString
			sort      sql.NullInt64
			status    sql.NullInt64
			music     sql.NullString
			musicText sql.NullString
			desc      sql.NullString
		)
		if err := rows.Scan(&id, &name, &image, &sort, &status, &music, &musicText, &desc); err != nil {
			return nil, err
		}
		items = append(items, SyncPushPhotoHobby{
			ID:        id,
			Name:      nullableStringValue(name),
			Image:     nullableStringValue(image),
			Sort:      int64Pointer(sort),
			Status:    int64Pointer(status),
			Music:     nullableStringValue(music),
			MusicText: nullableStringValue(musicText),
			Desc:      nullableStringValue(desc),
		})
	}
	return items, nil
}

func (h *SyncPushHandler) loadRemoteExtraSteps(targetID int64) ([]SyncPushExtraStep, error) {
	rows, err := h.db.Query(
		"SELECT id, step_index, field_name, label, music, music_text, status FROM config_extra_steps WHERE app_version_name_id = ? ORDER BY step_index ASC, id ASC",
		targetID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]SyncPushExtraStep, 0)
	for rows.Next() {
		var (
			id        int64
			stepIndex sql.NullInt64
			fieldName sql.NullString
			label     sql.NullString
			music     sql.NullString
			musicText sql.NullString
			status    sql.NullInt64
		)
		if err := rows.Scan(&id, &stepIndex, &fieldName, &label, &music, &musicText, &status); err != nil {
			return nil, err
		}
		items = append(items, SyncPushExtraStep{
			ID:        id,
			StepIndex: int64Pointer(stepIndex),
			FieldName: nullableStringValue(fieldName),
			Label:     nullableStringValue(label),
			Music:     nullableStringValue(music),
			MusicText: nullableStringValue(musicText),
			Status:    int64Pointer(status),
		})
	}
	return items, nil
}
