package handlers

import (
	"database/sql"
	"strings"
)

type SyncPushVersion struct {
	AppVersionName   string `json:"app_version_name"`
	LocationName     string `json:"location_name"`
	FeishuFieldNames string `json:"feishu_field_names"`
	AiModal          string `json:"ai_modal"`
	Status           *int64 `json:"status"`
}

type SyncPushBanner struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Image    string `json:"image"`
	Sort     *int64 `json:"sort"`
	IsActive *int64 `json:"is_active"`
	Type     *int64 `json:"type"`
}

type SyncPushIdentity struct {
	ID       int64  `json:"id"`
	TargetID *int64 `json:"target_id"`
	Name     string `json:"name"`
	Image    string `json:"image"`
	Sort     *int64 `json:"sort"`
	Status   *int64 `json:"status"`
}

type SyncPushScene struct {
	ID            int64  `json:"id"`
	TargetID      *int64 `json:"target_id"`
	Name          string `json:"name"`
	Image         string `json:"image"`
	Desc          string `json:"desc"`
	Music         string `json:"music"`
	WatermarkPath string `json:"watermark_path"`
	NeedWatermark *int64 `json:"need_watermark"`
	Sort          *int64 `json:"sort"`
	Status        *int64 `json:"status"`
	OssStyle      string `json:"oss_style"`
}

type SyncPushClothes struct {
	ID        int64  `json:"id"`
	TargetID  *int64 `json:"target_id"`
	Name      string `json:"name"`
	Image     string `json:"image"`
	Sort      *int64 `json:"sort"`
	Status    *int64 `json:"status"`
	Music     string `json:"music"`
	Desc      string `json:"desc"`
	MusicText string `json:"music_text"`
}

type SyncPushPhotoHobby struct {
	ID        int64  `json:"id"`
	TargetID  *int64 `json:"target_id"`
	Name      string `json:"name"`
	Image     string `json:"image"`
	Sort      *int64 `json:"sort"`
	Status    *int64 `json:"status"`
	Music     string `json:"music"`
	MusicText string `json:"music_text"`
	Desc      string `json:"desc"`
}

type SyncPushExtraStep struct {
	ID        int64  `json:"id"`
	TargetID  *int64 `json:"target_id"`
	StepIndex *int64 `json:"step_index"`
	FieldName string `json:"field_name"`
	Label     string `json:"label"`
	Music     string `json:"music"`
	MusicText string `json:"music_text"`
	Status    *int64 `json:"status"`
}

type SyncPushAppUIFields struct {
	ID              int64  `json:"id"`
	TargetID        *int64 `json:"target_id"`
	HomeTitleLeft   string `json:"home_title_left"`
	HomeTitleRight  string `json:"home_title_right"`
	HomeSubtitle    string `json:"home_subtitle"`
	StartExperience string `json:"start_experience"`
	Step1Music      string `json:"step1_music"`
	Step1MusicText  string `json:"step1_music_text"`
	Step1Title      string `json:"step1_title"`
	Step2Music      string `json:"step2_music"`
	Step2MusicText  string `json:"step2_music_text"`
	Step2Title      string `json:"step2_title"`
	Status          *int64 `json:"status"`
	PrintWait       string `json:"print_wait"`
}

type SyncPushRequest struct {
	DraftVersionID    int64                `json:"draft_version_id"`
	TriggerBy         int64                `json:"trigger_by"`
	Confirm           bool                 `json:"confirm"`
	Modules           []string             `json:"modules"`
	Version           SyncPushVersion      `json:"version"`
	AppUIFields       *SyncPushAppUIFields `json:"app_ui_fields"`
	Banners           []SyncPushBanner     `json:"banners"`
	Identities        []SyncPushIdentity   `json:"identities"`
	Scenes            []SyncPushScene      `json:"scenes"`
	ClothesCategories []SyncPushClothes    `json:"clothes_categories"`
	PhotoHobbies      []SyncPushPhotoHobby `json:"photo_hobbies"`
	ExtraSteps        []SyncPushExtraStep  `json:"config_extra_steps"`
}

type SyncIDMapping struct {
	ModuleKey string `json:"module_key"`
	DraftID   int64  `json:"draft_id"`
	TargetID  int64  `json:"target_id"`
}

func buildSyncPushFromDraft(req syncRequest, version draftVersionRow, data draftData) SyncPushRequest {
	payload := SyncPushRequest{
		DraftVersionID: req.DraftVersionID,
		TriggerBy:      req.TriggerBy,
		Confirm:        req.Confirm,
		Modules:        req.Modules,
		Version: SyncPushVersion{
			AppVersionName:   nullableStringValue(version.AppVersionName),
			LocationName:     nullableStringValue(version.LocationName),
			FeishuFieldNames: nullableStringValue(version.FeishuFieldNames),
			AiModal:          nullableStringValue(version.AiModal),
			Status:           int64Pointer(version.Status),
		},
		Banners:           make([]SyncPushBanner, 0, len(data.Banners)),
		Identities:        make([]SyncPushIdentity, 0, len(data.Identities)),
		Scenes:            make([]SyncPushScene, 0, len(data.Scenes)),
		ClothesCategories: make([]SyncPushClothes, 0, len(data.ClothesCategories)),
		PhotoHobbies:      make([]SyncPushPhotoHobby, 0, len(data.PhotoHobbies)),
		ExtraSteps:        make([]SyncPushExtraStep, 0, len(data.ExtraSteps)),
	}

	if data.AppUIFields != nil {
		payload.AppUIFields = &SyncPushAppUIFields{
			ID:              data.AppUIFields.ID,
			TargetID:        int64Pointer(data.AppUIFields.TargetID),
			HomeTitleLeft:   nullableStringValue(data.AppUIFields.HomeTitleLeft),
			HomeTitleRight:  nullableStringValue(data.AppUIFields.HomeTitleRight),
			HomeSubtitle:    nullableStringValue(data.AppUIFields.HomeSubtitle),
			StartExperience: nullableStringValue(data.AppUIFields.StartExperience),
			Step1Music:      nullableStringValue(data.AppUIFields.Step1Music),
			Step1MusicText:  nullableStringValue(data.AppUIFields.Step1MusicText),
			Step1Title:      nullableStringValue(data.AppUIFields.Step1Title),
			Step2Music:      nullableStringValue(data.AppUIFields.Step2Music),
			Step2MusicText:  nullableStringValue(data.AppUIFields.Step2MusicText),
			Step2Title:      nullableStringValue(data.AppUIFields.Step2Title),
			Status:          int64Pointer(data.AppUIFields.Status),
			PrintWait:       nullableStringValue(data.AppUIFields.PrintWait),
		}
	}

	for _, item := range data.Banners {
		payload.Banners = append(payload.Banners, SyncPushBanner{
			ID:       item.ID,
			Title:    nullableStringValue(item.Title),
			Image:    nullableStringValue(item.Image),
			Sort:     int64Pointer(item.Sort),
			IsActive: int64Pointer(item.IsActive),
			Type:     int64Pointer(item.Type),
		})
	}

	for _, item := range data.Identities {
		payload.Identities = append(payload.Identities, SyncPushIdentity{
			ID:       item.ID,
			TargetID: int64Pointer(item.TargetID),
			Name:     nullableStringValue(item.Name),
			Image:    nullableStringValue(item.Image),
			Sort:     int64Pointer(item.Sort),
			Status:   int64Pointer(item.Status),
		})
	}

	for _, item := range data.Scenes {
		payload.Scenes = append(payload.Scenes, SyncPushScene{
			ID:            item.ID,
			TargetID:      int64Pointer(item.TargetID),
			Name:          nullableStringValue(item.Name),
			Image:         nullableStringValue(item.Image),
			Desc:          nullableStringValue(item.Desc),
			Music:         nullableStringValue(item.Music),
			WatermarkPath: nullableStringValue(item.WatermarkPath),
			NeedWatermark: int64Pointer(item.NeedWatermark),
			Sort:          int64Pointer(item.Sort),
			Status:        int64Pointer(item.Status),
			OssStyle:      nullableStringValue(item.OssStyle),
		})
	}

	for _, item := range data.ClothesCategories {
		payload.ClothesCategories = append(payload.ClothesCategories, SyncPushClothes{
			ID:        item.ID,
			TargetID:  int64Pointer(item.TargetID),
			Name:      nullableStringValue(item.Name),
			Image:     nullableStringValue(item.Image),
			Sort:      int64Pointer(item.Sort),
			Status:    int64Pointer(item.Status),
			Music:     nullableStringValue(item.Music),
			Desc:      nullableStringValue(item.Desc),
			MusicText: nullableStringValue(item.MusicText),
		})
	}

	for _, item := range data.PhotoHobbies {
		payload.PhotoHobbies = append(payload.PhotoHobbies, SyncPushPhotoHobby{
			ID:        item.ID,
			TargetID:  int64Pointer(item.TargetID),
			Name:      nullableStringValue(item.Name),
			Image:     nullableStringValue(item.Image),
			Sort:      int64Pointer(item.Sort),
			Status:    int64Pointer(item.Status),
			Music:     nullableStringValue(item.Music),
			MusicText: nullableStringValue(item.MusicText),
			Desc:      nullableStringValue(item.Desc),
		})
	}

	for _, item := range data.ExtraSteps {
		payload.ExtraSteps = append(payload.ExtraSteps, SyncPushExtraStep{
			ID:        item.ID,
			TargetID:  int64Pointer(item.TargetID),
			StepIndex: int64Pointer(item.StepIndex),
			FieldName: nullableStringValue(item.FieldName),
			Label:     nullableStringValue(item.Label),
			Music:     nullableStringValue(item.Music),
			MusicText: nullableStringValue(item.MusicText),
			Status:    int64Pointer(item.Status),
		})
	}

	return payload
}

func buildDraftDataFromPush(req SyncPushRequest) draftData {
	data := draftData{
		Banners:           make([]draftBannerRow, 0, len(req.Banners)),
		Identities:        make([]draftIdentityRow, 0, len(req.Identities)),
		Scenes:            make([]draftSceneRow, 0, len(req.Scenes)),
		ClothesCategories: make([]draftClothesRow, 0, len(req.ClothesCategories)),
		PhotoHobbies:      make([]draftPhotoHobbyRow, 0, len(req.PhotoHobbies)),
		ExtraSteps:        make([]draftExtraStepRow, 0, len(req.ExtraSteps)),
	}

	if req.AppUIFields != nil {
		data.AppUIFields = &draftAppUIRow{
			ID:              req.AppUIFields.ID,
			HomeTitleLeft:   toNullString(req.AppUIFields.HomeTitleLeft),
			HomeTitleRight:  toNullString(req.AppUIFields.HomeTitleRight),
			HomeSubtitle:    toNullString(req.AppUIFields.HomeSubtitle),
			StartExperience: toNullString(req.AppUIFields.StartExperience),
			Step1Music:      toNullString(req.AppUIFields.Step1Music),
			Step1MusicText:  toNullString(req.AppUIFields.Step1MusicText),
			Step1Title:      toNullString(req.AppUIFields.Step1Title),
			Step2Music:      toNullString(req.AppUIFields.Step2Music),
			Step2MusicText:  toNullString(req.AppUIFields.Step2MusicText),
			Step2Title:      toNullString(req.AppUIFields.Step2Title),
			Status:          toNullInt64(req.AppUIFields.Status),
			PrintWait:       toNullString(req.AppUIFields.PrintWait),
			TargetID:        toNullInt64(req.AppUIFields.TargetID),
		}
	}

	for _, item := range req.Banners {
		data.Banners = append(data.Banners, draftBannerRow{
			ID:       item.ID,
			Title:    toNullString(item.Title),
			Image:    toNullString(item.Image),
			Sort:     toNullInt64(item.Sort),
			IsActive: toNullInt64(item.IsActive),
			Type:     toNullInt64(item.Type),
		})
	}

	for _, item := range req.Identities {
		data.Identities = append(data.Identities, draftIdentityRow{
			ID:       item.ID,
			Name:     toNullString(item.Name),
			Image:    toNullString(item.Image),
			Sort:     toNullInt64(item.Sort),
			Status:   toNullInt64(item.Status),
			TargetID: toNullInt64(item.TargetID),
		})
	}

	for _, item := range req.Scenes {
		data.Scenes = append(data.Scenes, draftSceneRow{
			ID:            item.ID,
			Name:          toNullString(item.Name),
			Image:         toNullString(item.Image),
			Desc:          toNullString(item.Desc),
			Music:         toNullString(item.Music),
			WatermarkPath: toNullString(item.WatermarkPath),
			NeedWatermark: toNullInt64(item.NeedWatermark),
			Sort:          toNullInt64(item.Sort),
			Status:        toNullInt64(item.Status),
			OssStyle:      toNullString(item.OssStyle),
			TargetID:      toNullInt64(item.TargetID),
		})
	}

	for _, item := range req.ClothesCategories {
		data.ClothesCategories = append(data.ClothesCategories, draftClothesRow{
			ID:        item.ID,
			Name:      toNullString(item.Name),
			Image:     toNullString(item.Image),
			Sort:      toNullInt64(item.Sort),
			Status:    toNullInt64(item.Status),
			Music:     toNullString(item.Music),
			Desc:      toNullString(item.Desc),
			MusicText: toNullString(item.MusicText),
			TargetID:  toNullInt64(item.TargetID),
		})
	}

	for _, item := range req.PhotoHobbies {
		data.PhotoHobbies = append(data.PhotoHobbies, draftPhotoHobbyRow{
			ID:        item.ID,
			Name:      toNullString(item.Name),
			Image:     toNullString(item.Image),
			Sort:      toNullInt64(item.Sort),
			Status:    toNullInt64(item.Status),
			Music:     toNullString(item.Music),
			MusicText: toNullString(item.MusicText),
			Desc:      toNullString(item.Desc),
			TargetID:  toNullInt64(item.TargetID),
		})
	}

	for _, item := range req.ExtraSteps {
		data.ExtraSteps = append(data.ExtraSteps, draftExtraStepRow{
			ID:        item.ID,
			StepIndex: toNullInt64(item.StepIndex),
			FieldName: toNullString(item.FieldName),
			Label:     toNullString(item.Label),
			Music:     toNullString(item.Music),
			MusicText: toNullString(item.MusicText),
			Status:    toNullInt64(item.Status),
			TargetID:  toNullInt64(item.TargetID),
		})
	}

	return data
}

func buildValidationPayloadFromPush(req SyncPushRequest) SyncValidationPayload {
	payload := SyncValidationPayload{
		Version: SyncVersionDraft{
			AppVersionName: strings.TrimSpace(req.Version.AppVersionName),
			LocationName:   strings.TrimSpace(req.Version.LocationName),
		},
		Banners:           make([]SyncBannerDraft, 0, len(req.Banners)),
		Identities:        make([]SyncIdentityDraft, 0, len(req.Identities)),
		Scenes:            make([]SyncSceneDraft, 0, len(req.Scenes)),
		ClothesCategories: make([]SyncNamedDraft, 0, len(req.ClothesCategories)),
		PhotoHobbies:      make([]SyncNamedDraft, 0, len(req.PhotoHobbies)),
		ExtraSteps:        make([]SyncExtraStepDraft, 0, len(req.ExtraSteps)),
	}

	for _, item := range req.Banners {
		payload.Banners = append(payload.Banners, SyncBannerDraft{
			ID:    item.ID,
			Image: item.Image,
		})
	}

	for _, item := range req.Identities {
		payload.Identities = append(payload.Identities, SyncIdentityDraft{
			ID:   item.ID,
			Name: item.Name,
		})
	}

	for _, item := range req.Scenes {
		payload.Scenes = append(payload.Scenes, SyncSceneDraft{
			ID:   item.ID,
			Name: item.Name,
		})
	}

	for _, item := range req.ClothesCategories {
		payload.ClothesCategories = append(payload.ClothesCategories, SyncNamedDraft{
			ID:   item.ID,
			Name: item.Name,
		})
	}

	for _, item := range req.PhotoHobbies {
		payload.PhotoHobbies = append(payload.PhotoHobbies, SyncNamedDraft{
			ID:   item.ID,
			Name: item.Name,
		})
	}

	for _, item := range req.ExtraSteps {
		payload.ExtraSteps = append(payload.ExtraSteps, SyncExtraStepDraft{
			ID:        item.ID,
			StepIndex: intPointer(item.StepIndex),
			FieldName: item.FieldName,
			Label:     item.Label,
		})
	}

	return payload
}

func int64Pointer(value sql.NullInt64) *int64 {
	if value.Valid {
		v := value.Int64
		return &v
	}
	return nil
}

func toNullString(value string) sql.NullString {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: trimmed, Valid: true}
}

func toNullInt64(value *int64) sql.NullInt64 {
	if value == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *value, Valid: true}
}

func intPointer(value *int64) *int {
	if value == nil {
		return nil
	}
	v := int(*value)
	return &v
}
