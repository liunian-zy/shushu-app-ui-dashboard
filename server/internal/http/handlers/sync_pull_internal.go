package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"shushu-app-ui-dashboard/internal/http/middleware"
)

type syncImportRequest struct {
	TargetID       int64  `json:"target_app_version_name_id"`
	AppVersionName string `json:"app_version_name"`
	DraftVersionID int64  `json:"draft_version_id"`
}

// PullVersions loads online version list for import entry.
// Args:
//
//	c: Gin context.
//
// Returns:
//
//	None.
func (h *SyncHandler) PullVersions(c *gin.Context) {
	if strings.ToLower(strings.TrimSpace(h.cfg.AppMode)) == "online" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not available in online mode"})
		return
	}
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
		return
	}

	if strings.TrimSpace(h.cfg.SyncTargetURL) == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sync target not configured"})
		return
	}
	if strings.TrimSpace(h.cfg.SyncAPIKey) == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sync api key not configured"})
		return
	}

	data, err := h.fetchRemoteVersions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// ImportFromOnline imports one online version snapshot into draft tables.
// Args:
//
//	c: Gin context.
//
// Returns:
//
//	None.
func (h *SyncHandler) ImportFromOnline(c *gin.Context) {
	if strings.ToLower(strings.TrimSpace(h.cfg.AppMode)) == "online" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not available in online mode"})
		return
	}
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
		return
	}

	if strings.TrimSpace(h.cfg.SyncTargetURL) == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sync target not configured"})
		return
	}
	if strings.TrimSpace(h.cfg.SyncAPIKey) == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sync api key not configured"})
		return
	}

	var req syncImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if req.TargetID <= 0 && strings.TrimSpace(req.AppVersionName) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_app_version_name_id or app_version_name is required"})
		return
	}

	snapshot, err := h.fetchRemoteSnapshot(c.Request.Context(), req.TargetID, req.AppVersionName)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	claims, _ := middleware.GetAuthClaims(c)
	operatorID := int64(0)
	if claims != nil {
		operatorID = claims.UserID
	}

	draftVersionID := req.DraftVersionID
	now := time.Now()
	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "transaction failed"})
		return
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if draftVersionID > 0 {
		exists, err := existsDraftVersionTx(tx, draftVersionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "draft version not found"})
			return
		}
		if err := purgeDraftVersionTx(tx, draftVersionID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "clear draft failed"})
			return
		}
		if err := updateDraftVersionMetaTx(tx, draftVersionID, snapshot.Version, operatorID, now); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "update draft failed"})
			return
		}
	} else {
		draftVersionID, err = insertDraftVersionFromSnapshotTx(tx, snapshot.Version, operatorID, now)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create draft failed"})
			return
		}
	}

	if err := importSnapshotModulesTx(tx, draftVersionID, snapshot, operatorID, now); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "import failed"})
		return
	}

	if err := recordImportAuditTx(tx, draftVersionID, snapshot.Version.TargetID, operatorID, req.DraftVersionID > 0, now); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "audit failed"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "import failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":                     "imported",
		"draft_version_id":           draftVersionID,
		"target_app_version_name_id": snapshot.Version.TargetID,
		"app_version_name":           snapshot.Version.AppVersionName,
		"location_name":              snapshot.Version.LocationName,
	})
}

func (h *SyncHandler) fetchRemoteVersions(ctx context.Context) ([]SyncRemoteVersion, error) {
	endpoint := buildRemoteSyncURL(h.cfg.SyncTargetURL, "/sync/versions")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", strings.TrimSpace(h.cfg.SyncAPIKey))

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= http.StatusMultipleChoices {
		return nil, parseRemoteError(body, "pull versions failed")
	}

	var parsed struct {
		Data []SyncRemoteVersion `json:"data"`
	}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &parsed); err != nil {
			return nil, fmt.Errorf("parse versions response failed")
		}
	}
	return parsed.Data, nil
}

func (h *SyncHandler) fetchRemoteSnapshot(ctx context.Context, targetID int64, appVersionName string) (*SyncPullSnapshot, error) {
	endpoint := buildRemoteSyncURL(h.cfg.SyncTargetURL, "/sync/snapshot")
	query := url.Values{}
	if targetID > 0 {
		query.Set("target_app_version_name_id", fmt.Sprintf("%d", targetID))
	} else {
		query.Set("app_version_name", strings.TrimSpace(appVersionName))
	}
	if strings.Contains(endpoint, "?") {
		endpoint = endpoint + "&" + query.Encode()
	} else {
		endpoint = endpoint + "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", strings.TrimSpace(h.cfg.SyncAPIKey))

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= http.StatusMultipleChoices {
		return nil, parseRemoteError(body, "pull snapshot failed")
	}

	var parsed struct {
		Data SyncPullSnapshot `json:"data"`
	}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &parsed); err != nil {
			return nil, fmt.Errorf("parse snapshot response failed")
		}
	}
	return &parsed.Data, nil
}

func parseRemoteError(body []byte, fallback string) error {
	if len(body) == 0 {
		return fmt.Errorf(fallback)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err == nil {
		if raw, ok := payload["error"].(string); ok && strings.TrimSpace(raw) != "" {
			return fmt.Errorf(strings.TrimSpace(raw))
		}
	}
	return fmt.Errorf(fallback)
}

func buildRemoteSyncURL(raw, suffix string) string {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimSuffix(trimmed, "/")
	trimmed = strings.TrimSuffix(trimmed, "/sync/push")
	return trimmed + suffix
}

func recordImportAuditTx(tx *sql.Tx, draftVersionID, targetID, operatorID int64, overwrite bool, now time.Time) error {
	actionMode := "create"
	if overwrite {
		actionMode = "overwrite"
	}
	payload := map[string]interface{}{
		"source":                     "online",
		"target_app_version_name_id": targetID,
		"mode":                       actionMode,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = tx.Exec(
		"INSERT INTO app_db_audit_logs (draft_version_id, entity_table, entity_id, action, actor_id, detail_json, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		draftVersionID,
		"sync_import",
		targetID,
		"import_from_online",
		nullableID(operatorID),
		string(raw),
		now,
	)
	return err
}

func existsDraftVersionTx(tx *sql.Tx, id int64) (bool, error) {
	row := tx.QueryRow("SELECT id FROM app_db_version_names WHERE id = ? LIMIT 1", id)
	var draftID int64
	if err := row.Scan(&draftID); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return draftID > 0, nil
}

func purgeDraftVersionTx(tx *sql.Tx, draftVersionID int64) error {
	queries := []string{
		"DELETE FROM app_db_sync_id_map WHERE draft_version_id = ?",
		"DELETE FROM app_db_app_ui_fields WHERE draft_version_id = ?",
		"DELETE FROM app_db_banners WHERE draft_version_id = ?",
		"DELETE FROM app_db_identities WHERE draft_version_id = ?",
		"DELETE FROM app_db_scenes WHERE draft_version_id = ?",
		"DELETE FROM app_db_clothes_categories WHERE draft_version_id = ?",
		"DELETE FROM app_db_photo_hobbies WHERE draft_version_id = ?",
		"DELETE FROM app_db_config_extra_steps WHERE draft_version_id = ?",
	}
	for _, query := range queries {
		if _, err := tx.Exec(query, draftVersionID); err != nil {
			return err
		}
	}
	return nil
}

func updateDraftVersionMetaTx(tx *sql.Tx, draftVersionID int64, version SyncRemoteVersion, operatorID int64, now time.Time) error {
	aiModal, err := NormalizeAiModal(version.AiModal)
	if err != nil {
		aiModal = "SD"
	}
	status := int64(1)
	if version.Status != nil {
		status = *version.Status
	}
	_, err = tx.Exec(
		"UPDATE app_db_version_names SET app_version_name = ?, location_name = ?, feishu_field_names = ?, ai_modal = ?, status = ?, draft_status = ?, submit_version = ?, last_submit_by = NULL, last_submit_at = NULL, confirmed_by = NULL, confirmed_at = NULL, sync_status = ?, sync_message = NULL, synced_at = ?, target_app_version_name_id = ?, updated_by = ?, updated_at = ? WHERE id = ?",
		nullIfEmpty(version.AppVersionName),
		nullIfEmpty(version.LocationName),
		nullIfEmpty(version.FeishuFieldNames),
		aiModal,
		status,
		"draft",
		0,
		"imported",
		now,
		nullableID(version.TargetID),
		nullableID(operatorID),
		now,
		draftVersionID,
	)
	return err
}

func insertDraftVersionFromSnapshotTx(tx *sql.Tx, version SyncRemoteVersion, operatorID int64, now time.Time) (int64, error) {
	aiModal, err := NormalizeAiModal(version.AiModal)
	if err != nil {
		aiModal = "SD"
	}
	status := int64(1)
	if version.Status != nil {
		status = *version.Status
	}
	result, err := tx.Exec(
		"INSERT INTO app_db_version_names (app_version_name, location_name, feishu_field_names, ai_modal, status, draft_status, submit_version, sync_status, synced_at, target_app_version_name_id, created_by, updated_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		nullIfEmpty(version.AppVersionName),
		nullIfEmpty(version.LocationName),
		nullIfEmpty(version.FeishuFieldNames),
		aiModal,
		status,
		"draft",
		0,
		"imported",
		now,
		nullableID(version.TargetID),
		nullableID(operatorID),
		nullableID(operatorID),
		now,
		now,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func importSnapshotModulesTx(tx *sql.Tx, draftVersionID int64, snapshot *SyncPullSnapshot, operatorID int64, now time.Time) error {
	targetVersionID := snapshot.Version.TargetID
	appVersionName := snapshot.Version.AppVersionName

	if snapshot.AppUIFields != nil {
		appUIStatus := intValue(snapshot.AppUIFields.Status, 1)
		result, err := tx.Exec(
			"INSERT INTO app_db_app_ui_fields (draft_version_id, app_version_name_id, home_title_left, home_title_right, home_subtitle, start_experience, step1_music, step1_music_text, step1_title, step2_music, step2_music_text, step2_title, status, print_wait, created_by, updated_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			draftVersionID,
			draftVersionID,
			nullIfEmpty(snapshot.AppUIFields.HomeTitleLeft),
			nullIfEmpty(snapshot.AppUIFields.HomeTitleRight),
			nullIfEmpty(snapshot.AppUIFields.HomeSubtitle),
			nullIfEmpty(snapshot.AppUIFields.StartExperience),
			nullIfEmpty(snapshot.AppUIFields.Step1Music),
			nullIfEmpty(snapshot.AppUIFields.Step1MusicText),
			nullIfEmpty(snapshot.AppUIFields.Step1Title),
			nullIfEmpty(snapshot.AppUIFields.Step2Music),
			nullIfEmpty(snapshot.AppUIFields.Step2MusicText),
			nullIfEmpty(snapshot.AppUIFields.Step2Title),
			appUIStatus,
			nullIfEmpty(snapshot.AppUIFields.PrintWait),
			nullableID(operatorID),
			nullableID(operatorID),
			now,
			now,
		)
		if err != nil {
			return err
		}
		draftID, _ := result.LastInsertId()
		if snapshot.AppUIFields.ID > 0 {
			if err := upsertSyncIDMappings(tx, draftVersionID, []SyncIDMapping{{
				ModuleKey: "app_ui_fields",
				DraftID:   draftID,
				TargetID:  snapshot.AppUIFields.ID,
			}}, now); err != nil {
				return err
			}
		}
	}

	bannerMappings := make([]SyncIDMapping, 0, len(snapshot.Banners))
	for _, item := range snapshot.Banners {
		result, err := tx.Exec(
			"INSERT INTO app_db_banners (draft_version_id, title, image, sort, is_active, type, app_version_name, created_by, updated_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			draftVersionID,
			nullIfEmpty(item.Title),
			nullIfEmpty(item.Image),
			intValue(item.Sort, 0),
			intValue(item.IsActive, 1),
			intValue(item.Type, 0),
			nullIfEmpty(appVersionName),
			nullableID(operatorID),
			nullableID(operatorID),
			now,
			now,
		)
		if err != nil {
			return err
		}
		draftID, _ := result.LastInsertId()
		if item.ID > 0 {
			bannerMappings = append(bannerMappings, SyncIDMapping{ModuleKey: "banners", DraftID: draftID, TargetID: item.ID})
		}
	}

	identityMappings := make([]SyncIDMapping, 0, len(snapshot.Identities))
	for _, item := range snapshot.Identities {
		result, err := tx.Exec(
			"INSERT INTO app_db_identities (draft_version_id, name, image, sort, status, app_version_name, created_by, updated_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			draftVersionID,
			nullIfEmpty(item.Name),
			nullIfEmpty(item.Image),
			intValue(item.Sort, 0),
			intValue(item.Status, 1),
			nullIfEmpty(appVersionName),
			nullableID(operatorID),
			nullableID(operatorID),
			now,
			now,
		)
		if err != nil {
			return err
		}
		draftID, _ := result.LastInsertId()
		if item.ID > 0 {
			identityMappings = append(identityMappings, SyncIDMapping{ModuleKey: "identities", DraftID: draftID, TargetID: item.ID})
		}
	}

	sceneMappings := make([]SyncIDMapping, 0, len(snapshot.Scenes))
	for _, item := range snapshot.Scenes {
		result, err := tx.Exec(
			"INSERT INTO app_db_scenes (draft_version_id, name, image, `desc`, music, watermark_path, need_watermark, sort, status, app_version_name, oss_style, created_by, updated_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			draftVersionID,
			nullIfEmpty(item.Name),
			nullIfEmpty(item.Image),
			nullIfEmpty(item.Desc),
			nullIfEmpty(item.Music),
			nullIfEmpty(item.WatermarkPath),
			intValue(item.NeedWatermark, 1),
			intValue(item.Sort, 0),
			intValue(item.Status, 1),
			nullIfEmpty(appVersionName),
			nullIfEmpty(item.OssStyle),
			nullableID(operatorID),
			nullableID(operatorID),
			now,
			now,
		)
		if err != nil {
			return err
		}
		draftID, _ := result.LastInsertId()
		if item.ID > 0 {
			sceneMappings = append(sceneMappings, SyncIDMapping{ModuleKey: "scenes", DraftID: draftID, TargetID: item.ID})
		}
	}

	clothesMappings := make([]SyncIDMapping, 0, len(snapshot.ClothesCategories))
	for _, item := range snapshot.ClothesCategories {
		result, err := tx.Exec(
			"INSERT INTO app_db_clothes_categories (draft_version_id, name, image, sort, status, music, `desc`, music_text, app_version_name, created_by, updated_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			draftVersionID,
			nullIfEmpty(item.Name),
			nullIfEmpty(item.Image),
			intValue(item.Sort, 0),
			intValue(item.Status, 1),
			nullIfEmpty(item.Music),
			nullIfEmpty(item.Desc),
			nullIfEmpty(item.MusicText),
			nullIfEmpty(appVersionName),
			nullableID(operatorID),
			nullableID(operatorID),
			now,
			now,
		)
		if err != nil {
			return err
		}
		draftID, _ := result.LastInsertId()
		if item.ID > 0 {
			clothesMappings = append(clothesMappings, SyncIDMapping{ModuleKey: "clothes_categories", DraftID: draftID, TargetID: item.ID})
		}
	}

	photoMappings := make([]SyncIDMapping, 0, len(snapshot.PhotoHobbies))
	for _, item := range snapshot.PhotoHobbies {
		result, err := tx.Exec(
			"INSERT INTO app_db_photo_hobbies (draft_version_id, name, image, sort, status, music, music_text, `desc`, app_version_name, created_by, updated_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			draftVersionID,
			nullIfEmpty(item.Name),
			nullIfEmpty(item.Image),
			intValue(item.Sort, 0),
			intValue(item.Status, 1),
			nullIfEmpty(item.Music),
			nullIfEmpty(item.MusicText),
			nullIfEmpty(item.Desc),
			nullIfEmpty(appVersionName),
			nullableID(operatorID),
			nullableID(operatorID),
			now,
			now,
		)
		if err != nil {
			return err
		}
		draftID, _ := result.LastInsertId()
		if item.ID > 0 {
			photoMappings = append(photoMappings, SyncIDMapping{ModuleKey: "photo_hobbies", DraftID: draftID, TargetID: item.ID})
		}
	}

	extraMappings := make([]SyncIDMapping, 0, len(snapshot.ExtraSteps))
	for _, item := range snapshot.ExtraSteps {
		result, err := tx.Exec(
			"INSERT INTO app_db_config_extra_steps (draft_version_id, app_version_name_id, step_index, field_name, label, music, music_text, status, created_by, updated_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			draftVersionID,
			draftVersionID,
			intValue(item.StepIndex, 0),
			nullIfEmpty(item.FieldName),
			nullIfEmpty(item.Label),
			nullIfEmpty(item.Music),
			nullIfEmpty(item.MusicText),
			intValue(item.Status, 1),
			nullableID(operatorID),
			nullableID(operatorID),
			now,
			now,
		)
		if err != nil {
			return err
		}
		draftID, _ := result.LastInsertId()
		if item.ID > 0 {
			extraMappings = append(extraMappings, SyncIDMapping{ModuleKey: "config_extra_steps", DraftID: draftID, TargetID: item.ID})
		}
	}

	allMappings := make([]SyncIDMapping, 0, len(bannerMappings)+len(identityMappings)+len(sceneMappings)+len(clothesMappings)+len(photoMappings)+len(extraMappings)+1)
	allMappings = append(allMappings, bannerMappings...)
	allMappings = append(allMappings, identityMappings...)
	allMappings = append(allMappings, sceneMappings...)
	allMappings = append(allMappings, clothesMappings...)
	allMappings = append(allMappings, photoMappings...)
	allMappings = append(allMappings, extraMappings...)
	if len(allMappings) > 0 {
		if err := upsertSyncIDMappings(tx, draftVersionID, allMappings, now); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(
		"UPDATE app_db_version_names SET target_app_version_name_id = ?, sync_status = ?, sync_message = NULL, synced_at = ? WHERE id = ?",
		nullableID(targetVersionID),
		"imported",
		now,
		draftVersionID,
	); err != nil {
		return err
	}

	return nil
}

func intValue(value *int64, fallback int64) int64 {
	if value != nil {
		return *value
	}
	return fallback
}
