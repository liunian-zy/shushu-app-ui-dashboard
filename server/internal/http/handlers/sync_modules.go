package handlers

import "strings"

var syncModuleSet = map[string]struct{}{
  "version_names":      {},
  "app_ui_fields":      {},
  "banners":            {},
  "identities":         {},
  "scenes":             {},
  "clothes_categories": {},
  "photo_hobbies":      {},
  "config_extra_steps": {},
}

var syncModuleOrder = []string{
  "version_names",
  "banners",
  "identities",
  "scenes",
  "app_ui_fields",
  "config_extra_steps",
  "clothes_categories",
  "photo_hobbies",
}

func normalizeModules(modules []string) []string {
  if len(modules) == 0 {
    return nil
  }
  normalized := make([]string, 0, len(modules))
  seen := make(map[string]struct{}, len(modules))
  for _, module := range modules {
    key := strings.ToLower(strings.TrimSpace(module))
    if key == "" {
      continue
    }
    if _, ok := syncModuleSet[key]; !ok {
      continue
    }
    if _, ok := seen[key]; ok {
      continue
    }
    seen[key] = struct{}{}
    normalized = append(normalized, key)
  }
  return normalized
}

func resolveSyncModules(modules []string) []string {
  normalized := normalizeModules(modules)
  if len(normalized) > 0 {
    return normalized
  }
  return syncModuleOrder
}

func findInvalidModules(modules []string) []string {
  if len(modules) == 0 {
    return nil
  }
  invalid := make([]string, 0)
  seen := make(map[string]struct{}, len(modules))
  for _, module := range modules {
    key := strings.ToLower(strings.TrimSpace(module))
    if key == "" {
      continue
    }
    if _, ok := syncModuleSet[key]; ok {
      continue
    }
    if _, ok := seen[key]; ok {
      continue
    }
    seen[key] = struct{}{}
    invalid = append(invalid, key)
  }
  return invalid
}

func shouldSyncModule(modules []string, module string) bool {
  if len(modules) == 0 {
    return true
  }
  key := strings.ToLower(strings.TrimSpace(module))
  if key == "" {
    return false
  }
  for _, item := range modules {
    if item == key {
      return true
    }
  }
  return false
}
