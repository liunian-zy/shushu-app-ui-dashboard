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
  ID     int64
  Name   sql.NullString
  Image  sql.NullString
  Sort   sql.NullInt64
  Status sql.NullInt64
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
}

type draftExtraStepRow struct {
  ID        int64
  StepIndex sql.NullInt64
  FieldName sql.NullString
  Label     sql.NullString
  Music     sql.NullString
  MusicText sql.NullString
  Status    sql.NullInt64
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
    "SELECT id, home_title_left, home_title_right, home_subtitle, start_experience, step1_music, step1_music_text, step1_title, step2_music, step2_music_text, step2_title, status, print_wait FROM app_db_app_ui_fields WHERE draft_version_id = ?",
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
    "SELECT id, name, image, sort, status FROM app_db_identities WHERE draft_version_id = ?",
    draftVersionID,
  ); err == nil {
    defer rows.Close()
    for rows.Next() {
      var row draftIdentityRow
      if err := rows.Scan(&row.ID, &row.Name, &row.Image, &row.Sort, &row.Status); err != nil {
        return data, err
      }
      data.Identities = append(data.Identities, row)
    }
  } else {
    return data, err
  }

  if rows, err := db.Query(
    "SELECT id, name, image, `desc`, music, watermark_path, need_watermark, sort, status, oss_style FROM app_db_scenes WHERE draft_version_id = ?",
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
      ); err != nil {
        return data, err
      }
      data.Scenes = append(data.Scenes, row)
    }
  } else {
    return data, err
  }

  if rows, err := db.Query(
    "SELECT id, name, image, sort, status, music, `desc`, music_text FROM app_db_clothes_categories WHERE draft_version_id = ?",
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
      ); err != nil {
        return data, err
      }
      data.ClothesCategories = append(data.ClothesCategories, row)
    }
  } else {
    return data, err
  }

  if rows, err := db.Query(
    "SELECT id, name, image, sort, status, music, music_text, `desc` FROM app_db_photo_hobbies WHERE draft_version_id = ?",
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
      ); err != nil {
        return data, err
      }
      data.PhotoHobbies = append(data.PhotoHobbies, row)
    }
  } else {
    return data, err
  }

  if rows, err := db.Query(
    "SELECT id, step_index, field_name, label, music, music_text, status FROM app_db_config_extra_steps WHERE draft_version_id = ?",
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
