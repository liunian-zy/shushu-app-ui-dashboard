package handlers

import "strings"

type SyncValidationError struct {
  Module  string `json:"module"`
  RowID   int64  `json:"row_id"`
  Field   string `json:"field"`
  Message string `json:"message"`
}

type SyncVersionDraft struct {
  AppVersionName string `json:"app_version_name"`
  LocationName   string `json:"location_name"`
}

type SyncBannerDraft struct {
  ID    int64  `json:"id"`
  Image string `json:"image"`
}

type SyncIdentityDraft struct {
  ID   int64  `json:"id"`
  Name string `json:"name"`
}

type SyncSceneDraft struct {
  ID   int64  `json:"id"`
  Name string `json:"name"`
}

type SyncNamedDraft struct {
  ID   int64  `json:"id"`
  Name string `json:"name"`
}

type SyncExtraStepDraft struct {
  ID        int64  `json:"id"`
  StepIndex *int   `json:"step_index"`
  FieldName string `json:"field_name"`
  Label     string `json:"label"`
}

type SyncValidationPayload struct {
  Version           SyncVersionDraft     `json:"version"`
  Banners           []SyncBannerDraft    `json:"banners"`
  Identities        []SyncIdentityDraft  `json:"identities"`
  Scenes            []SyncSceneDraft     `json:"scenes"`
  ClothesCategories []SyncNamedDraft     `json:"clothes_categories"`
  PhotoHobbies      []SyncNamedDraft     `json:"photo_hobbies"`
  ExtraSteps        []SyncExtraStepDraft `json:"config_extra_steps"`
}

// ValidateSyncPayload validates required fields before sync.
// Args:
//   payload: Sync data payload built from draft records.
// Returns:
//   []SyncValidationError: Validation errors.
func ValidateSyncPayload(payload SyncValidationPayload, modules []string) []SyncValidationError {
  errors := make([]SyncValidationError, 0)

  versionName := strings.TrimSpace(payload.Version.AppVersionName)
  if versionName == "" {
    errors = append(errors, SyncValidationError{
      Module:  "version_names",
      RowID:   0,
      Field:   "app_version_name",
      Message: "app_version_name is required",
    })
  }

  locationName := strings.TrimSpace(payload.Version.LocationName)
  if locationName == "" {
    errors = append(errors, SyncValidationError{
      Module:  "version_names",
      RowID:   0,
      Field:   "location_name",
      Message: "location_name is required",
    })
  }

  if shouldSyncModule(modules, "banners") {
    for _, item := range payload.Banners {
      if strings.TrimSpace(item.Image) == "" {
        errors = append(errors, SyncValidationError{
          Module:  "banners",
          RowID:   item.ID,
          Field:   "image",
          Message: "image is required",
        })
      }
    }
  }

  if shouldSyncModule(modules, "identities") {
    for _, item := range payload.Identities {
      if strings.TrimSpace(item.Name) == "" {
        errors = append(errors, SyncValidationError{
          Module:  "identities",
          RowID:   item.ID,
          Field:   "name",
          Message: "name is required",
        })
      }
    }
  }

  if shouldSyncModule(modules, "scenes") {
    if len(payload.Scenes) == 0 {
      errors = append(errors, SyncValidationError{
        Module:  "scenes",
        RowID:   0,
        Field:   "name",
        Message: "at least one scene is required",
      })
    }

    for _, item := range payload.Scenes {
      if strings.TrimSpace(item.Name) == "" {
        errors = append(errors, SyncValidationError{
          Module:  "scenes",
          RowID:   item.ID,
          Field:   "name",
          Message: "name is required",
        })
      }
    }
  }

  if shouldSyncModule(modules, "clothes_categories") {
    for _, item := range payload.ClothesCategories {
      if strings.TrimSpace(item.Name) == "" {
        errors = append(errors, SyncValidationError{
          Module:  "clothes_categories",
          RowID:   item.ID,
          Field:   "name",
          Message: "name is required",
        })
      }
    }
  }

  if shouldSyncModule(modules, "photo_hobbies") {
    for _, item := range payload.PhotoHobbies {
      if strings.TrimSpace(item.Name) == "" {
        errors = append(errors, SyncValidationError{
          Module:  "photo_hobbies",
          RowID:   item.ID,
          Field:   "name",
          Message: "name is required",
        })
      }
    }
  }

  if shouldSyncModule(modules, "config_extra_steps") {
    for _, item := range payload.ExtraSteps {
      if item.StepIndex == nil {
        errors = append(errors, SyncValidationError{
          Module:  "config_extra_steps",
          RowID:   item.ID,
          Field:   "step_index",
          Message: "step_index is required",
        })
      }
      if strings.TrimSpace(item.FieldName) == "" {
        errors = append(errors, SyncValidationError{
          Module:  "config_extra_steps",
          RowID:   item.ID,
          Field:   "field_name",
          Message: "field_name is required",
        })
      }
      if strings.TrimSpace(item.Label) == "" {
        errors = append(errors, SyncValidationError{
          Module:  "config_extra_steps",
          RowID:   item.ID,
          Field:   "label",
          Message: "label is required",
        })
      }
    }
  }

  return errors
}

func buildSyncValidationPayload(data draftData, appVersionName, locationName string) SyncValidationPayload {
  payload := SyncValidationPayload{
    Version: SyncVersionDraft{
      AppVersionName: appVersionName,
      LocationName:   locationName,
    },
    Banners:           make([]SyncBannerDraft, 0, len(data.Banners)),
    Identities:        make([]SyncIdentityDraft, 0, len(data.Identities)),
    Scenes:            make([]SyncSceneDraft, 0, len(data.Scenes)),
    ClothesCategories: make([]SyncNamedDraft, 0, len(data.ClothesCategories)),
    PhotoHobbies:      make([]SyncNamedDraft, 0, len(data.PhotoHobbies)),
    ExtraSteps:        make([]SyncExtraStepDraft, 0, len(data.ExtraSteps)),
  }

  for _, item := range data.Banners {
    payload.Banners = append(payload.Banners, SyncBannerDraft{
      ID:    item.ID,
      Image: nullableStringValue(item.Image),
    })
  }

  for _, item := range data.Identities {
    payload.Identities = append(payload.Identities, SyncIdentityDraft{
      ID:   item.ID,
      Name: nullableStringValue(item.Name),
    })
  }

  for _, item := range data.Scenes {
    payload.Scenes = append(payload.Scenes, SyncSceneDraft{
      ID:   item.ID,
      Name: nullableStringValue(item.Name),
    })
  }

  for _, item := range data.ClothesCategories {
    payload.ClothesCategories = append(payload.ClothesCategories, SyncNamedDraft{
      ID:   item.ID,
      Name: nullableStringValue(item.Name),
    })
  }

  for _, item := range data.PhotoHobbies {
    payload.PhotoHobbies = append(payload.PhotoHobbies, SyncNamedDraft{
      ID:   item.ID,
      Name: nullableStringValue(item.Name),
    })
  }

  for _, item := range data.ExtraSteps {
    payload.ExtraSteps = append(payload.ExtraSteps, SyncExtraStepDraft{
      ID:        item.ID,
      StepIndex: nullableInt(item.StepIndex),
      FieldName: nullableStringValue(item.FieldName),
      Label:     nullableStringValue(item.Label),
    })
  }

  return payload
}
