package service

import (
	"testing"
	"time"

	"omnilogs-api/dto"
	"omnilogs-api/models"
)

func TestArchiveNeedsRearchiveAfterRestore(t *testing.T) {
	restored := "RESTORED"
	archive := models.LogArchive{RestoreStatus: &restored, DocumentCount: 1}
	if !archiveNeedsRearchive(archive) {
		t.Fatal("restored archive must produce a new GZIP generation")
	}
	rearchived := "REARCHIVED"
	archive.RestoreStatus = &rearchived
	if archiveNeedsRearchive(archive) {
		t.Fatal("rearchived archive must be reusable")
	}
}

func TestArchiveReusableDoesNotHideLateLogs(t *testing.T) {
	verified := "VERIFIED"
	deletedAt := time.Now().UTC()
	archive := models.LogArchive{Status: &verified, DeletedFromActiveAt: &deletedAt}
	if archiveReusable(archive, 1) {
		t.Fatal("a verified archive must not hide late active logs")
	}
	if !archiveReusable(archive, 0) {
		t.Fatal("a verified archive with no active logs should be reused")
	}
}

func TestArchiveDayQueryFromHonorsNewLogsOnlyBoundary(t *testing.T) {
	day := time.Date(2026, time.August, 19, 0, 0, 0, 0, time.UTC)
	effective := day.Add(13*time.Hour + 45*time.Minute)
	policy := &models.LogRetentionPolicy{EffectiveFrom: &effective}
	if got := archiveDayQueryFrom(policy, day, day.AddDate(0, 0, 1), false); !got.Equal(effective) {
		t.Fatalf("query started at %s, want policy boundary %s", got, effective)
	}
	if got := archiveDayQueryFrom(policy, day, day.AddDate(0, 0, 1), true); !got.Equal(day) {
		t.Fatalf("explicit backfill started at %s, want %s", got, day)
	}
}

func TestEmptyRestoredArchiveDoesNotNeedRearchive(t *testing.T) {
	restored := "RESTORED"
	archive := models.LogArchive{RestoreStatus: &restored, DocumentCount: 0}
	if archiveNeedsRearchive(archive) {
		t.Fatal("empty restored archive must be treated as a successful no-op")
	}
}

func TestBackupIdentityKeepsManualAndPolicyCoverageEquivalent(t *testing.T) {
	location, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		t.Fatal(err)
	}
	day := time.Date(2026, 8, 16, 0, 0, 0, 0, location)
	manual := archiveCoverageKey(7, 9, nil, nil, nil, day)
	policy := archiveCoverageKey(7, 9, nil, nil, nil, day)
	if manual != policy {
		t.Fatalf("coverage must not depend on backup type: %q != %q", manual, policy)
	}
	if got := buildBackupTag("MANUAL", 12, day, day.AddDate(0, 0, 8)); got != "manual:p12:2026-08-16_2026-08-23" {
		t.Fatalf("unexpected backup tag %q", got)
	}
}

func TestNormalizeBackupTypeDefaultsAPIRequestsToManual(t *testing.T) {
	got, err := normalizeBackupType("")
	if err != nil || got != "MANUAL" {
		t.Fatalf("normalize backup type = %q, %v", got, err)
	}
	if _, err := normalizeBackupType("other"); err == nil {
		t.Fatal("expected invalid backup type error")
	}
}

func TestHasRestoredArchiveRequiresMatchingLogScope(t *testing.T) {
	restored := "RESTORED"
	featureOne, featureTwo := 1, 2
	archives := []models.LogArchive{{RestoreStatus: &restored, FeatureID: &featureOne, DocumentCount: 1}}
	if !hasRestoredArchive(archives, dto.CreateArchiveRequest{FeatureID: &featureOne}) {
		t.Fatal("matching restored scope was not detected")
	}
	if hasRestoredArchive(archives, dto.CreateArchiveRequest{FeatureID: &featureTwo}) {
		t.Fatal("different feature scope must not be rearchived together")
	}
}
