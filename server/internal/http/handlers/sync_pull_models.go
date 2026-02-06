package handlers

import "time"

type SyncRemoteVersion struct {
	TargetID         int64      `json:"target_app_version_name_id"`
	AppVersionName   string     `json:"app_version_name"`
	LocationName     string     `json:"location_name"`
	FeishuFieldNames string     `json:"feishu_field_names"`
	AiModal          string     `json:"ai_modal"`
	Status           *int64     `json:"status"`
	UpdatedAt        *time.Time `json:"updated_at,omitempty"`
}

type SyncPullSnapshot struct {
	Version           SyncRemoteVersion     `json:"version"`
	AppUIFields       *SyncPushAppUIFields  `json:"app_ui_fields"`
	Banners           []SyncPushBanner      `json:"banners"`
	Identities        []SyncPushIdentity    `json:"identities"`
	Scenes            []SyncPushScene       `json:"scenes"`
	ClothesCategories []SyncPushClothes     `json:"clothes_categories"`
	PhotoHobbies      []SyncPushPhotoHobby  `json:"photo_hobbies"`
	ExtraSteps        []SyncPushExtraStep   `json:"config_extra_steps"`
}
