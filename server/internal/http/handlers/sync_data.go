package handlers

import (
	"database/sql"
	"strings"
)

type draftVersionRow struct {
	AppVersionName   sql.NullString
	LocationName     sql.NullString
	FeishuFieldNames sql.NullString
	AiModal          sql.NullString
	Status           sql.NullInt64
}

type draftBannerRow struct {
	ID       int64
	Title    sql.NullString
	Image    sql.NullString
	Sort     sql.NullInt64
	IsActive sql.NullInt64
	Type     sql.NullInt64
}

type draftIdentityRow struct {
	ID       int64
	Name     sql.NullString
	Image    sql.NullString
	Sort     sql.NullInt64
	Status   sql.NullInt64
	TargetID sql.NullInt64
}

type draftSceneRow struct {
	ID            int64
	Name          sql.NullString
	Image         sql.NullString
	Desc          sql.NullString
	Music         sql.NullString
	WatermarkPath sql.NullString
	NeedWatermark sql.NullInt64
	Sort          sql.NullInt64
	Status        sql.NullInt64
	OssStyle      sql.NullString
	TargetID      sql.NullInt64
}

type draftClothesRow struct {
	ID        int64
	Name      sql.NullString
	Image     sql.NullString
	Sort      sql.NullInt64
	Status    sql.NullInt64
	Music     sql.NullString
	Desc      sql.NullString
	MusicText sql.NullString
	TargetID  sql.NullInt64
}

type draftPhotoHobbyRow struct {
	ID        int64
	Name      sql.NullString
	Image     sql.NullString
	Sort      sql.NullInt64
	Status    sql.NullInt64
	Music     sql.NullString
	MusicText sql.NullString
	Desc      sql.NullString
	TargetID  sql.NullInt64
}

type draftExtraStepRow struct {
	ID        int64
	StepIndex sql.NullInt64
	FieldName sql.NullString
	Label     sql.NullString
	Music     sql.NullString
	MusicText sql.NullString
	Status    sql.NullInt64
	TargetID  sql.NullInt64
}

type draftAppUIRow struct {
	ID              int64
	HomeTitleLeft   sql.NullString
	HomeTitleRight  sql.NullString
	HomeSubtitle    sql.NullString
	StartExperience sql.NullString
	Step1Music      sql.NullString
	Step1MusicText  sql.NullString
	Step1Title      sql.NullString
	Step2Music      sql.NullString
	Step2MusicText  sql.NullString
	Step2Title      sql.NullString
	Status          sql.NullInt64
	PrintWait       sql.NullString
	TargetID        sql.NullInt64
}

type draftData struct {
	AppUIFields       *draftAppUIRow
	Banners           []draftBannerRow
	Identities        []draftIdentityRow
	Scenes            []draftSceneRow
	ClothesCategories []draftClothesRow
	PhotoHobbies      []draftPhotoHobbyRow
	ExtraSteps        []draftExtraStepRow
}

func loadDraftVersion(db *sql.DB, draftVersionID int64) (draftVersionRow, error) {
	row := db.QueryRow(
		"SELECT app_version_name, location_name, feishu_field_names, ai_modal, status FROM app_db_version_names WHERE id = ?",
		draftVersionID,
	)
	var data draftVersionRow
	if err := row.Scan(&data.AppVersionName, &data.LocationName, &data.FeishuFieldNames, &data.AiModal, &data.Status); err != nil {
		return data, err
	}
	return data, nil
}

func findAppVersionNameID(db *sql.DB, appVersionName string) (int64, error) {
	if strings.TrimSpace(appVersionName) == "" {
		return 0, nil
	}
	row := db.QueryRow("SELECT id FROM app_version_names WHERE app_version_name = ? ORDER BY id DESC LIMIT 1", appVersionName)
	var id int64
	if err := row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return id, nil
}

func loadDraftData(db *sql.DB, draftVersionID int64) (draftData, error) {
	data := draftData{
		Banners:           make([]draftBannerRow, 0),
		Identities:        make([]draftIdentityRow, 0),
		Scenes:            make([]draftSceneRow, 0),
		ClothesCategories: make([]draftClothesRow, 0),
		PhotoHobbies:      make([]draftPhotoHobbyRow, 0),
		ExtraSteps:        make([]draftExtraStepRow, 0),
	}

	appUIRow := db.QueryRow(
		"SELECT f.id, f.home_title_left, f.home_title_right, f.home_subtitle, f.start_experience, f.step1_music, f.step1_music_text, f.step1_title, f.step2_music, f.step2_music_text, f.step2_title, f.status, f.print_wait, m.target_row_id FROM app_db_app_ui_fields f LEFT JOIN app_db_sync_id_map m ON m.draft_version_id = f.draft_version_id AND m.module_key = 'app_ui_fields' AND m.draft_row_id = f.id WHERE f.draft_version_id = ?",
		draftVersionID,
	)
	var appUI draftAppUIRow
	if err := appUIRow.Scan(
		&appUI.ID,
		&appUI.HomeTitleLeft,
		&appUI.HomeTitleRight,
		&appUI.HomeSubtitle,
		&appUI.StartExperience,
		&appUI.Step1Music,
		&appUI.Step1MusicText,
		&appUI.Step1Title,
		&appUI.Step2Music,
		&appUI.Step2MusicText,
		&appUI.Step2Title,
		&appUI.Status,
		&appUI.PrintWait,
		&appUI.TargetID,
	); err == nil {
		data.AppUIFields = &appUI
	} else if err != sql.ErrNoRows {
		return data, err
	}

	if rows, err := db.Query(
		"SELECT id, title, image, sort, is_active, type FROM app_db_banners WHERE draft_version_id = ?",
		draftVersionID,
	); err == nil {
		defer rows.Close()
		for rows.Next() {
			var row draftBannerRow
			if err := rows.Scan(&row.ID, &row.Title, &row.Image, &row.Sort, &row.IsActive, &row.Type); err != nil {
				return data, err
			}
			data.Banners = append(data.Banners, row)
		}
	} else {
		return data, err
	}

	if rows, err := db.Query(
		"SELECT i.id, i.name, i.image, i.sort, i.status, m.target_row_id FROM app_db_identities i LEFT JOIN app_db_sync_id_map m ON m.draft_version_id = i.draft_version_id AND m.module_key = 'identities' AND m.draft_row_id = i.id WHERE i.draft_version_id = ?",
		draftVersionID,
	); err == nil {
		defer rows.Close()
		for rows.Next() {
			var row draftIdentityRow
			if err := rows.Scan(&row.ID, &row.Name, &row.Image, &row.Sort, &row.Status, &row.TargetID); err != nil {
				return data, err
			}
			data.Identities = append(data.Identities, row)
		}
	} else {
		return data, err
	}

	if rows, err := db.Query(
		"SELECT s.id, s.name, s.image, s.`desc`, s.music, s.watermark_path, s.need_watermark, s.sort, s.status, s.oss_style, m.target_row_id FROM app_db_scenes s LEFT JOIN app_db_sync_id_map m ON m.draft_version_id = s.draft_version_id AND m.module_key = 'scenes' AND m.draft_row_id = s.id WHERE s.draft_version_id = ?",
		draftVersionID,
	); err == nil {
		defer rows.Close()
		for rows.Next() {
			var row draftSceneRow
			if err := rows.Scan(
				&row.ID,
				&row.Name,
				&row.Image,
				&row.Desc,
				&row.Music,
				&row.WatermarkPath,
				&row.NeedWatermark,
				&row.Sort,
				&row.Status,
				&row.OssStyle,
				&row.TargetID,
			); err != nil {
				return data, err
			}
			data.Scenes = append(data.Scenes, row)
		}
	} else {
		return data, err
	}

	if rows, err := db.Query(
		"SELECT c.id, c.name, c.image, c.sort, c.status, c.music, c.`desc`, c.music_text, m.target_row_id FROM app_db_clothes_categories c LEFT JOIN app_db_sync_id_map m ON m.draft_version_id = c.draft_version_id AND m.module_key = 'clothes_categories' AND m.draft_row_id = c.id WHERE c.draft_version_id = ?",
		draftVersionID,
	); err == nil {
		defer rows.Close()
		for rows.Next() {
			var row draftClothesRow
			if err := rows.Scan(
				&row.ID,
				&row.Name,
				&row.Image,
				&row.Sort,
				&row.Status,
				&row.Music,
				&row.Desc,
				&row.MusicText,
				&row.TargetID,
			); err != nil {
				return data, err
			}
			data.ClothesCategories = append(data.ClothesCategories, row)
		}
	} else {
		return data, err
	}

	if rows, err := db.Query(
		"SELECT p.id, p.name, p.image, p.sort, p.status, p.music, p.music_text, p.`desc`, m.target_row_id FROM app_db_photo_hobbies p LEFT JOIN app_db_sync_id_map m ON m.draft_version_id = p.draft_version_id AND m.module_key = 'photo_hobbies' AND m.draft_row_id = p.id WHERE p.draft_version_id = ?",
		draftVersionID,
	); err == nil {
		defer rows.Close()
		for rows.Next() {
			var row draftPhotoHobbyRow
			if err := rows.Scan(
				&row.ID,
				&row.Name,
				&row.Image,
				&row.Sort,
				&row.Status,
				&row.Music,
				&row.MusicText,
				&row.Desc,
				&row.TargetID,
			); err != nil {
				return data, err
			}
			data.PhotoHobbies = append(data.PhotoHobbies, row)
		}
	} else {
		return data, err
	}

	if rows, err := db.Query(
		"SELECT e.id, e.step_index, e.field_name, e.label, e.music, e.music_text, e.status, m.target_row_id FROM app_db_config_extra_steps e LEFT JOIN app_db_sync_id_map m ON m.draft_version_id = e.draft_version_id AND m.module_key = 'config_extra_steps' AND m.draft_row_id = e.id WHERE e.draft_version_id = ?",
		draftVersionID,
	); err == nil {
		defer rows.Close()
		for rows.Next() {
			var row draftExtraStepRow
			if err := rows.Scan(
				&row.ID,
				&row.StepIndex,
				&row.FieldName,
				&row.Label,
				&row.Music,
				&row.MusicText,
				&row.Status,
				&row.TargetID,
			); err != nil {
				return data, err
			}
			data.ExtraSteps = append(data.ExtraSteps, row)
		}
	} else {
		return data, err
	}

	return data, nil
}
