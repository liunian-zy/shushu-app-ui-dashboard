package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type SyncPushHandler struct {
	db *sql.DB
}

// NewSyncPushHandler creates a handler for online sync push.
// Args:
//
//	db: Database connection.
//
// Returns:
//
//	*SyncPushHandler: Initialized handler.
func NewSyncPushHandler(db *sql.DB) *SyncPushHandler {
	return &SyncPushHandler{db: db}
}

// Push receives sync payload and writes to production tables.
// Args:
//
//	c: Gin context.
//
// Returns:
//
//	None.
func (h *SyncPushHandler) Push(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db not ready"})
		return
	}

	var req SyncPushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	appVersionName := strings.TrimSpace(req.Version.AppVersionName)
	locationName := strings.TrimSpace(req.Version.LocationName)
	if appVersionName == "" || locationName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app_version_name and location_name are required"})
		return
	}

	invalidModules := findInvalidModules(req.Modules)
	if len(invalidModules) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_modules", "modules": invalidModules})
		return
	}
	modules := normalizeModules(req.Modules)
	validationPayload := buildValidationPayloadFromPush(req)
	validationErrors := ValidateSyncPayload(validationPayload, modules)
	if len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"details": validationErrors,
		})
		return
	}

	targetID, err := findAppVersionNameID(h.db, appVersionName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	if targetID > 0 && !req.Confirm {
		c.JSON(http.StatusConflict, gin.H{
			"need_confirm":               true,
			"reason":                     "app_version_name_exists",
			"target_app_version_name_id": targetID,
		})
		return
	}
	if len(modules) > 0 && !shouldSyncModule(modules, "version_names") && targetID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "version_not_synced"})
		return
	}

	aiModal := strings.TrimSpace(req.Version.AiModal)
	if aiModal == "" {
		aiModal = "SD"
	}
	status := int64(1)
	if req.Version.Status != nil {
		status = *req.Version.Status
	}
	feishuFields := strings.TrimSpace(req.Version.FeishuFieldNames)

	draftData := buildDraftDataFromPush(req)
	now := time.Now()
	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "transaction failed"})
		return
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if shouldSyncModule(modules, "version_names") {
		if targetID > 0 {
			if err := updateAppVersionName(tx, targetID, appVersionName, locationName, status, feishuFields, aiModal, now); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "sync failed"})
				return
			}
		} else {
			targetID, err = insertAppVersionName(tx, appVersionName, locationName, status, feishuFields, aiModal, now)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "sync failed"})
				return
			}
		}
	}

	if shouldSyncModule(modules, "app_ui_fields") {
		if err := syncAppUIFields(tx, targetID, draftData.AppUIFields, now); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "sync failed"})
			return
		}
	}
	if shouldSyncModule(modules, "banners") {
		if err := syncBanners(tx, appVersionName, draftData.Banners, now); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "sync failed"})
			return
		}
	}
	if shouldSyncModule(modules, "identities") {
		if err := syncIdentities(tx, appVersionName, draftData.Identities, now); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "sync failed"})
			return
		}
	}
	if shouldSyncModule(modules, "scenes") {
		if err := syncScenes(tx, appVersionName, draftData.Scenes, now); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "sync failed"})
			return
		}
	}
	if shouldSyncModule(modules, "clothes_categories") {
		if err := syncClothesCategories(tx, appVersionName, draftData.ClothesCategories, now); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "sync failed"})
			return
		}
	}
	if shouldSyncModule(modules, "photo_hobbies") {
		if err := syncPhotoHobbies(tx, appVersionName, draftData.PhotoHobbies, now); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "sync failed"})
			return
		}
	}
	if shouldSyncModule(modules, "config_extra_steps") {
		if err := syncExtraSteps(tx, targetID, draftData.ExtraSteps, now); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "sync failed"})
			return
		}
	}

	mappings, err := buildSyncMappings(tx, modules, appVersionName, targetID, draftData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sync failed"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sync failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":                     "synced",
		"target_app_version_name_id": targetID,
		"mappings":                   mappings,
	})
}
