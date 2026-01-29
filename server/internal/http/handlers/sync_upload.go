package handlers

import (
  "errors"

  "shushu-app-ui-dashboard/internal/services"
)

var uploadableSyncModules = map[string]struct{}{
  "banners":            {},
  "identities":         {},
  "scenes":             {},
  "app_ui_fields":      {},
  "config_extra_steps": {},
  "clothes_categories": {},
  "photo_hobbies":      {},
}

func (h *SyncHandler) uploadDraftModules(draftVersionID int64, modules []string) (map[string]uploadCacheEntry, error) {
  resolved := resolveSyncModules(modules)
  pending := make([]string, 0, len(resolved))
  for _, module := range resolved {
    if _, ok := uploadableSyncModules[module]; ok {
      pending = append(pending, module)
    }
  }
  if len(pending) == 0 {
    return map[string]uploadCacheEntry{}, nil
  }

  ossService, err := services.NewOSSService(h.cfg, nil)
  if err != nil {
    return nil, err
  }

  tx, err := h.db.Begin()
  if err != nil {
    return nil, err
  }
  defer func() {
    _ = tx.Rollback()
  }()

  cache := make(map[string]uploadCacheEntry)
  taskHandler := NewTaskHandler(h.cfg, h.db, nil)
  for _, module := range pending {
    if _, err := taskHandler.uploadModuleAssets(tx, ossService, draftVersionID, module, cache); err != nil {
      return nil, err
    }
  }

  if err := tx.Commit(); err != nil {
    return nil, err
  }
  if len(cache) == 0 {
    return map[string]uploadCacheEntry{}, nil
  }
  return cache, nil
}

func ensureSyncModules(modules []string) ([]string, error) {
  resolved := resolveSyncModules(modules)
  if len(resolved) == 0 {
    return nil, errors.New("no sync modules")
  }
  return resolved, nil
}
