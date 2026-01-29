package handlers_test

import (
  "testing"

  "shushu-app-ui-dashboard/internal/http/handlers"
)

func TestValidateSyncPayloadMissingRequired(t *testing.T) {
  payload := handlers.SyncValidationPayload{
    Version: handlers.SyncVersionDraft{},
    Banners: []handlers.SyncBannerDraft{
      {ID: 10, Image: ""},
    },
    Identities: []handlers.SyncIdentityDraft{
      {ID: 20, Name: ""},
    },
    ClothesCategories: []handlers.SyncNamedDraft{
      {ID: 30, Name: ""},
    },
    PhotoHobbies: []handlers.SyncNamedDraft{
      {ID: 40, Name: ""},
    },
    ExtraSteps: []handlers.SyncExtraStepDraft{
      {ID: 50, StepIndex: nil, FieldName: "", Label: ""},
    },
  }

  errs := handlers.ValidateSyncPayload(payload, nil)
  if len(errs) != 10 {
    t.Fatalf("expected 10 errors, got %d", len(errs))
  }
}

func TestValidateSyncPayloadOK(t *testing.T) {
  stepIndex := 1
  payload := handlers.SyncValidationPayload{
    Version: handlers.SyncVersionDraft{
      AppVersionName: "MUSEUM",
      LocationName:   "Museum",
    },
    Banners: []handlers.SyncBannerDraft{
      {ID: 10, Image: "banner/a.png"},
    },
    Identities: []handlers.SyncIdentityDraft{
      {ID: 20, Name: "Male"},
    },
    Scenes: []handlers.SyncSceneDraft{
      {ID: 30, Name: "Entrance"},
    },
    ClothesCategories: []handlers.SyncNamedDraft{
      {ID: 40, Name: "default"},
    },
    PhotoHobbies: []handlers.SyncNamedDraft{
      {ID: 50, Name: "default"},
    },
    ExtraSteps: []handlers.SyncExtraStepDraft{
      {ID: 60, StepIndex: &stepIndex, FieldName: "clothes_prefer", Label: "Clothes Preference"},
    },
  }

  errs := handlers.ValidateSyncPayload(payload, nil)
  if len(errs) != 0 {
    t.Fatalf("expected no errors, got %d", len(errs))
  }
}

func TestValidateSyncPayloadModules(t *testing.T) {
  payload := handlers.SyncValidationPayload{
    Version: handlers.SyncVersionDraft{
      AppVersionName: "MUSEUM",
      LocationName:   "Museum",
    },
    Banners: []handlers.SyncBannerDraft{
      {ID: 10, Image: ""},
    },
    Identities: []handlers.SyncIdentityDraft{
      {ID: 20, Name: ""},
    },
  }

  errs := handlers.ValidateSyncPayload(payload, []string{"banners"})
  if len(errs) != 1 {
    t.Fatalf("expected 1 error, got %d", len(errs))
  }
  if errs[0].Module != "banners" {
    t.Fatalf("expected banners error, got %s", errs[0].Module)
  }
}
