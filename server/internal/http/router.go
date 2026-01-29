package http

import (
  "database/sql"
  "strings"

  "shushu-app-ui-dashboard/internal/config"
  "shushu-app-ui-dashboard/internal/http/handlers"
  "shushu-app-ui-dashboard/internal/http/middleware"

  "github.com/gin-gonic/gin"
  "github.com/redis/go-redis/v9"
)

type Deps struct {
  DB    *sql.DB
  Redis *redis.Client
}

func NewRouter(cfg *config.Config, deps *Deps) *gin.Engine {
  router := gin.New()
  router.Use(gin.Logger(), gin.Recovery(), middleware.RequestID())

  router.GET("/healthz", handlers.Health)

  api := router.Group("/api")
  api.GET("/ping", handlers.Ping)

  if strings.ToLower(strings.TrimSpace(cfg.AppMode)) == "online" {
    syncPushHandler := handlers.NewSyncPushHandler(deps.DB)
    api.POST("/sync/push", middleware.RequireAPIKey(cfg), syncPushHandler.Push)
    return router
  }

  authHandler := handlers.NewAuthHandler(cfg, deps.DB)
  api.POST("/auth/login", authHandler.Login)
  api.POST("/auth/bootstrap", authHandler.Bootstrap)

  localFileHandler := handlers.NewLocalFileHandler(cfg)
  api.GET("/local-files/*path", localFileHandler.Serve)

  secured := api.Group("")
  secured.Use(middleware.AuthRequired(cfg))
  secured.GET("/auth/me", authHandler.Me)

  userHandler := handlers.NewUserHandler(cfg, deps.DB)
  secured.GET("/users", middleware.RequireAdmin(), userHandler.List)
  secured.POST("/users", middleware.RequireAdmin(), userHandler.Create)

  taskHandler := handlers.NewTaskHandler(cfg, deps.DB, deps.Redis)
  secured.GET("/tasks", taskHandler.List)
  secured.POST("/tasks", middleware.RequireAdmin(), taskHandler.Create)
  secured.PUT("/tasks/:id", middleware.RequireAdmin(), taskHandler.Update)
  secured.POST("/tasks/:id/assist", taskHandler.Assist)
  secured.POST("/tasks/:id/complete", taskHandler.CompleteUpload)
  secured.GET("/tasks/:id/actions", taskHandler.Actions)

  ossHandler := handlers.NewOSSHandler(cfg, deps.Redis)
  secured.POST("/oss/pre-sign", ossHandler.PreSign)
  secured.POST("/oss/sign-url", ossHandler.SignURL)

  secured.POST("/local-files/upload", localFileHandler.Upload)

  historyHandler := handlers.NewHistoryHandler(cfg, deps.DB, deps.Redis)
  mediaHandler := handlers.NewMediaHandler(cfg, deps.DB, deps.Redis)
  media := secured.Group("/media")
  media.GET("/rules", mediaHandler.ListRules)
  media.POST("/rules", mediaHandler.CreateRule)
  media.PUT("/rules/:id", mediaHandler.UpdateRule)
  media.DELETE("/rules/:id", mediaHandler.DeleteRule)
  media.POST("/validate", mediaHandler.Validate)
  media.POST("/transform", mediaHandler.Transform)
  media.GET("/versions", historyHandler.ListMediaVersions)

  draftHandler := handlers.NewDraftHandler(cfg, deps.DB, deps.Redis)
  draft := secured.Group("/draft")
  draft.GET("/banners", draftHandler.ListBanners)
  draft.GET("/identities", draftHandler.ListIdentities)
  draft.GET("/scenes", draftHandler.ListScenes)
  draft.GET("/clothes-categories", draftHandler.ListClothesCategories)
  draft.GET("/photo-hobbies", draftHandler.ListPhotoHobbies)
  draft.GET("/app-ui-fields", draftHandler.GetAppUIFields)
  draft.GET("/config-extra-steps", draftHandler.ListConfigExtraSteps)
  crudHandler := handlers.NewDraftCRUDHandler(deps.DB)
  draft.GET("/version-names", crudHandler.ListVersionNames)
  draft.POST("/version-names", crudHandler.CreateVersionName)
  draft.PUT("/version-names/:id", crudHandler.UpdateVersionName)
  draft.DELETE("/version-names/:id", crudHandler.DeleteVersionName)
  draft.POST("/banners", crudHandler.CreateBanner)
  draft.PUT("/banners/:id", crudHandler.UpdateBanner)
  draft.DELETE("/banners/:id", crudHandler.DeleteBanner)
  draft.POST("/identities", crudHandler.CreateIdentity)
  draft.PUT("/identities/:id", crudHandler.UpdateIdentity)
  draft.DELETE("/identities/:id", crudHandler.DeleteIdentity)
  templateHandler := handlers.NewIdentityTemplateHandler(cfg, deps.DB, deps.Redis)
  draft.POST("/identities/apply-template", templateHandler.ApplyTemplate)
  draft.POST("/scenes", crudHandler.CreateScene)
  draft.PUT("/scenes/:id", crudHandler.UpdateScene)
  draft.DELETE("/scenes/:id", crudHandler.DeleteScene)
  draft.POST("/clothes-categories", crudHandler.CreateClothesCategory)
  draft.PUT("/clothes-categories/:id", crudHandler.UpdateClothesCategory)
  draft.DELETE("/clothes-categories/:id", crudHandler.DeleteClothesCategory)
  draft.POST("/photo-hobbies", crudHandler.CreatePhotoHobby)
  draft.PUT("/photo-hobbies/:id", crudHandler.UpdatePhotoHobby)
  draft.DELETE("/photo-hobbies/:id", crudHandler.DeletePhotoHobby)
  draft.POST("/config-extra-steps", crudHandler.CreateConfigExtraStep)
  draft.PUT("/config-extra-steps/:id", crudHandler.UpdateConfigExtraStep)
  draft.DELETE("/config-extra-steps/:id", crudHandler.DeleteConfigExtraStep)
  draft.POST("/app-ui-fields", crudHandler.UpsertAppUIFields)
  submissionHandler := handlers.NewSubmissionHandler(deps.DB)
  draft.POST("/submit", submissionHandler.Submit)
  draft.POST("/confirm", submissionHandler.Confirm)
  draft.GET("/submissions", submissionHandler.List)

  secured.GET("/identity-templates", templateHandler.ListTemplates)
  secured.POST("/identity-templates", middleware.RequireAdmin(), templateHandler.CreateTemplate)
  secured.PUT("/identity-templates/:id", middleware.RequireAdmin(), templateHandler.UpdateTemplate)
  secured.DELETE("/identity-templates/:id", middleware.RequireAdmin(), templateHandler.DeleteTemplate)
  secured.GET("/identity-templates/:id/items", templateHandler.ListTemplateItems)
  secured.POST("/identity-templates/:id/items", middleware.RequireAdmin(), templateHandler.CreateTemplateItem)
  secured.PUT("/identity-template-items/:id", middleware.RequireAdmin(), templateHandler.UpdateTemplateItem)
  secured.DELETE("/identity-template-items/:id", middleware.RequireAdmin(), templateHandler.DeleteTemplateItem)

  secured.GET("/audit/logs", historyHandler.ListAuditLogs)
  secured.GET("/field-history", historyHandler.ListFieldHistory)

  syncHandler := handlers.NewSyncHandler(cfg, deps.DB)
  secured.POST("/sync", syncHandler.Sync)
  secured.GET("/sync/jobs", syncHandler.ListModuleJobs)

  dashboardHandler := handlers.NewDashboardHandler(deps.DB)
  secured.GET("/dashboard/summary", dashboardHandler.Summary)

  presetHandler := handlers.NewTTSPresetHandler(deps.DB)
  secured.GET("/tts/presets", presetHandler.List)
  secured.POST("/tts/presets", middleware.RequireAdmin(), presetHandler.Create)
  secured.PUT("/tts/presets/:id", middleware.RequireAdmin(), presetHandler.Update)
  secured.DELETE("/tts/presets/:id", middleware.RequireAdmin(), presetHandler.Delete)

  ttsHandler := handlers.NewTTSHandler(cfg, deps.DB, deps.Redis)
  secured.POST("/tts/convert", ttsHandler.Convert)
  secured.POST("/tts/voice-detail", ttsHandler.VoiceDetail)

  return router
}
