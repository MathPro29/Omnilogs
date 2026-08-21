package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"omnilogs-api/dto"
	globalauth "omnilogs-api/internal/global_auth/usecase"
	"omnilogs-api/internal/retention_policy/helper"
	"omnilogs-api/internal/retention_policy/repository"
	retentionschedule "omnilogs-api/internal/retention_policy/schedule"
	"omnilogs-api/models"
	"omnilogs-api/responses"
)

type Actor struct {
	UserID        int
	PlatformAdmin bool
}

type actorContextKey struct{}

func WithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, actorContextKey{}, actor)
}

func actorFromContext(ctx context.Context) (Actor, bool) {
	value, ok := ctx.Value(actorContextKey{}).(Actor)
	return value, ok
}

type Usecase interface {
	Create(ctx context.Context, req dto.CreateRetentionPolicyRequest) (*dto.RetentionPolicyResponse, error)
	Update(ctx context.Context, req dto.UpdateRetentionPolicyRequest) (*dto.RetentionPolicyResponse, error)
	Delete(ctx context.Context, policyID int, productID int) error
	GetByID(ctx context.Context, policyID int, productID int) (*dto.RetentionPolicyResponse, error)
	List(ctx context.Context, productID int, environmentID *int) ([]dto.RetentionPolicyResponse, error)
	ToggleActive(ctx context.Context, policyID int, productID int, isActive bool) error
	TriggerNow(ctx context.Context, policyID int, productID int) (*dto.RetentionPolicyResponse, error)
	Simulate(ctx context.Context, req dto.SimulateRetentionRequest) (*dto.SimulationResultDTO, error)
	GetStats(ctx context.Context, productID int, environmentID *int) (*dto.PolicyStatsDTO, error)
	Historical(ctx context.Context, req dto.RetentionHistoricalRequest) (*dto.RetentionHistoricalResponse, error)
	Preview(ctx context.Context, req dto.RetentionPreviewRequest) (*dto.RetentionPreviewResponse, error)
}

type usecase struct {
	repo  repository.Repository
	authz globalauth.Usecase
}

func NewUsecase(repo repository.Repository, authz ...globalauth.Usecase) Usecase {
	var checker globalauth.Usecase
	if len(authz) > 0 {
		checker = authz[0]
	}
	return &usecase{repo: repo, authz: checker}
}

func (u *usecase) Create(ctx context.Context, req dto.CreateRetentionPolicyRequest) (*dto.RetentionPolicyResponse, error) {
	if req.EnvironmentID == nil || *req.EnvironmentID <= 0 {
		return nil, errors.New("environment_id is required")
	}
	if err := u.authorize(ctx, req.ProductID, *req.EnvironmentID, "CREATE"); err != nil {
		return nil, err
	}
	if err := u.validateEnvironment(ctx, req.ProductID, *req.EnvironmentID); err != nil {
		return nil, err
	}
	activeValue, activeUnit, err := normalizeActiveDuration(req.ActiveRetentionValue, req.ActiveRetentionUnit, req.RetentionValue, req.RetentionUnit)
	if err != nil {
		return nil, err
	}
	archiveAfterValue := req.ArchiveAfterValue
	archiveAfterUnit := normalizeOptionalUnit(req.ArchiveAfterUnit)
	if req.ArchiveEnabled && archiveAfterValue == nil {
		archiveAfterValue = intPointer(activeValue)
	}
	if req.ArchiveEnabled && archiveAfterUnit == "" {
		archiveAfterUnit = activeUnit
	}
	archiveRetentionUnit := normalizeOptionalUnit(req.ArchiveRetentionUnit)
	if err := validateArchiveFields(req.ArchiveEnabled, archiveAfterValue, archiveAfterUnit, req.ArchiveRetentionValue, archiveRetentionUnit, req.ArchiveNeverDelete); err != nil {
		return nil, err
	}
	legacyMode := strings.ToUpper(strings.TrimSpace(req.RetentionMode))
	if legacyMode == "" {
		legacyMode = "CUSTOM"
	}
	legacyUnit := req.RetentionUnit
	if strings.TrimSpace(legacyUnit) == "" {
		legacyUnit = activeUnit
	}
	legacyValue := req.RetentionValue
	if legacyValue <= 0 {
		legacyValue = activeValue
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = fmt.Sprintf("Retention %d/%d", req.ProductID, *req.EnvironmentID)
	}
	totalDays := calendarDays(activeValue, activeUnit)
	folderStructure := req.FolderStructure
	if folderStructure == "" {
		folderStructure = defaultFolderStructure(req.RetentionMode)
	}
	autoPurgeAction := req.AutoPurgeAction
	if autoPurgeAction == "" {
		autoPurgeAction = "DELETE"
	}
	storageProvider := req.StorageProvider
	if storageProvider == "" {
		storageProvider = "LOCAL"
	}
	cronSchedule := strings.TrimSpace(req.CronSchedule)
	if cronSchedule == "" {
		cronSchedule = "0 1 * * *"
	}
	scheduleTimezone := strings.TrimSpace(req.ScheduleTimezone)
	if scheduleTimezone == "" {
		scheduleTimezone = retentionschedule.DefaultTimezone
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	now := time.Now().UTC()
	nextPurge, err := retentionschedule.Next(cronSchedule, scheduleTimezone, now)
	if err != nil {
		return nil, err
	}
	var effectiveFrom *time.Time
	if !req.ApplyToExistingLogs {
		effectiveFrom = &now
	}

	policy := &models.LogRetentionPolicy{
		ProductID:                req.ProductID,
		EnvironmentID:            req.EnvironmentID,
		ProjectID:                req.ProjectID,
		Name:                     name,
		Description:              req.Description,
		RetentionMode:            legacyMode,
		RetentionUnit:            legacyUnit,
		RetentionValue:           legacyValue,
		TotalDays:                totalDays,
		FolderStructure:          folderStructure,
		AutoPurgeAction:          autoPurgeAction,
		StorageProvider:          storageProvider,
		BucketName:               req.BucketName,
		CronSchedule:             cronSchedule,
		ScheduleTimezone:         scheduleTimezone,
		IsActive:                 isActive,
		NextPurgeAt:              &nextPurge,
		ActiveRetentionValue:     activeValue,
		ActiveRetentionUnit:      activeUnit,
		ArchiveEnabled:           req.ArchiveEnabled,
		ArchiveAfterValue:        archiveAfterValue,
		ArchiveAfterUnit:         archiveAfterUnit,
		DeleteActiveAfterArchive: req.ArchiveEnabled && req.DeleteActiveAfterArchive,
		ArchiveRetentionValue:    req.ArchiveRetentionValue,
		ArchiveRetentionUnit:     archiveRetentionUnit,
		ArchiveNeverDelete:       req.ArchiveNeverDelete,
		ApplyToExistingLogs:      req.ApplyToExistingLogs,
		EffectiveFrom:            effectiveFrom,
	}

	if err := u.repo.Create(ctx, policy); err != nil {
		return nil, err
	}

	return toResponse(policy), nil
}

func (u *usecase) Update(ctx context.Context, req dto.UpdateRetentionPolicyRequest) (*dto.RetentionPolicyResponse, error) {
	policy, err := u.repo.GetByID(ctx, req.PolicyID, req.ProductID)
	if err != nil {
		return nil, err
	}
	environmentID := policy.EnvironmentID
	if req.EnvironmentID != nil {
		environmentID = req.EnvironmentID
	}
	if environmentID == nil || *environmentID <= 0 {
		return nil, errors.New("environment_id is required")
	}
	if err := u.authorize(ctx, req.ProductID, *environmentID, "UPDATE"); err != nil {
		return nil, err
	}
	if err := u.validateEnvironment(ctx, req.ProductID, *environmentID); err != nil {
		return nil, err
	}
	policy.EnvironmentID = environmentID

	if req.Name != nil {
		policy.Name = *req.Name
	}
	if req.Description != nil {
		policy.Description = *req.Description
	}
	if req.RetentionMode != nil {
		policy.RetentionMode = *req.RetentionMode
	}
	if req.RetentionUnit != nil {
		policy.RetentionUnit = *req.RetentionUnit
		policy.ActiveRetentionUnit = *req.RetentionUnit
	}
	if req.RetentionValue != nil {
		policy.RetentionValue = *req.RetentionValue
		policy.ActiveRetentionValue = *req.RetentionValue
	}
	if req.FolderStructure != nil {
		policy.FolderStructure = *req.FolderStructure
	}
	if req.AutoPurgeAction != nil {
		policy.AutoPurgeAction = *req.AutoPurgeAction
	}
	if req.StorageProvider != nil {
		policy.StorageProvider = *req.StorageProvider
	}
	if req.BucketName != nil {
		policy.BucketName = req.BucketName
	}
	if req.CronSchedule != nil {
		policy.CronSchedule = strings.TrimSpace(*req.CronSchedule)
	}
	if req.ScheduleTimezone != nil {
		policy.ScheduleTimezone = strings.TrimSpace(*req.ScheduleTimezone)
	}
	if req.IsActive != nil {
		policy.IsActive = *req.IsActive
	}
	if req.ActiveRetentionValue != nil {
		policy.ActiveRetentionValue = *req.ActiveRetentionValue
	}
	if req.ActiveRetentionUnit != nil {
		policy.ActiveRetentionUnit = *req.ActiveRetentionUnit
	}
	if req.ArchiveEnabled != nil {
		policy.ArchiveEnabled = *req.ArchiveEnabled
	}
	if req.ArchiveAfterValue != nil {
		policy.ArchiveAfterValue = req.ArchiveAfterValue
	}
	if req.ArchiveAfterUnit != nil {
		policy.ArchiveAfterUnit = normalizeOptionalUnit(*req.ArchiveAfterUnit)
	}
	if req.DeleteActiveAfterArchive != nil {
		policy.DeleteActiveAfterArchive = *req.DeleteActiveAfterArchive
	}
	if req.ArchiveRetentionValue != nil {
		policy.ArchiveRetentionValue = req.ArchiveRetentionValue
	}
	if req.ArchiveRetentionUnit != nil {
		policy.ArchiveRetentionUnit = normalizeOptionalUnit(*req.ArchiveRetentionUnit)
	}
	if req.ArchiveNeverDelete != nil {
		policy.ArchiveNeverDelete = *req.ArchiveNeverDelete
	}
	if req.ApplyToExistingLogs != nil {
		wasApplyingToExisting := policy.ApplyToExistingLogs
		policy.ApplyToExistingLogs = *req.ApplyToExistingLogs
		if *req.ApplyToExistingLogs {
			policy.EffectiveFrom = nil
		} else if wasApplyingToExisting || policy.EffectiveFrom == nil {
			now := time.Now().UTC()
			policy.EffectiveFrom = &now
		}
	}
	if policy.ActiveRetentionValue <= 0 {
		policy.ActiveRetentionValue = policy.RetentionValue
	}
	if strings.TrimSpace(policy.ActiveRetentionUnit) == "" {
		policy.ActiveRetentionUnit = policy.RetentionUnit
	}
	if policy.ArchiveEnabled {
		if policy.ArchiveAfterValue == nil {
			policy.ArchiveAfterValue = intPointer(policy.ActiveRetentionValue)
		}
		if strings.TrimSpace(policy.ArchiveAfterUnit) == "" {
			policy.ArchiveAfterUnit = normalizeOptionalUnit(policy.ActiveRetentionUnit)
		}
	} else {
		policy.DeleteActiveAfterArchive = false
	}
	if err := validateCalendarPolicy(policy); err != nil {
		return nil, err
	}

	policy.TotalDays = calendarDays(policy.ActiveRetentionValue, policy.ActiveRetentionUnit)
	now := time.Now().UTC()
	nextPurge, err := retentionschedule.Next(policy.CronSchedule, policy.ScheduleTimezone, now)
	if err != nil {
		return nil, err
	}
	policy.NextPurgeAt = &nextPurge

	if err := u.repo.Update(ctx, policy); err != nil {
		return nil, err
	}

	return toResponse(policy), nil
}

func (u *usecase) Delete(ctx context.Context, policyID int, productID int) error {
	policy, err := u.repo.GetByID(ctx, policyID, productID)
	if err != nil {
		return err
	}
	if policy.EnvironmentID == nil {
		return errors.New("policy environment_id is required")
	}
	if err := u.authorize(ctx, productID, *policy.EnvironmentID, "DELETE"); err != nil {
		return err
	}
	return u.repo.Delete(ctx, policyID, productID)
}

func (u *usecase) GetByID(ctx context.Context, policyID int, productID int) (*dto.RetentionPolicyResponse, error) {
	policy, err := u.repo.GetByID(ctx, policyID, productID)
	if err != nil {
		return nil, err
	}
	if policy.EnvironmentID != nil {
		if err := u.authorize(ctx, productID, *policy.EnvironmentID, "READ"); err != nil {
			return nil, err
		}
	}
	return toResponse(policy), nil
}

func (u *usecase) List(ctx context.Context, productID int, environmentID *int) ([]dto.RetentionPolicyResponse, error) {
	if environmentID == nil || *environmentID <= 0 {
		return nil, errors.New("environment_id is required")
	}
	if environmentID != nil {
		if err := u.authorize(ctx, productID, *environmentID, "READ"); err != nil {
			return nil, err
		}
	}
	policies, err := u.repo.List(ctx, productID, environmentID)
	if err != nil {
		return nil, err
	}

	res := make([]dto.RetentionPolicyResponse, len(policies))
	for i, p := range policies {
		res[i] = *toResponse(&p)
	}
	return res, nil
}

func (u *usecase) ToggleActive(ctx context.Context, policyID int, productID int, isActive bool) error {
	policy, err := u.repo.GetByID(ctx, policyID, productID)
	if err != nil {
		return err
	}
	if policy.EnvironmentID == nil {
		return errors.New("policy environment_id is required")
	}
	if err := u.authorize(ctx, productID, *policy.EnvironmentID, "UPDATE"); err != nil {
		return err
	}
	if isActive {
		next, err := retentionschedule.Next(policy.CronSchedule, policy.ScheduleTimezone, time.Now().UTC())
		if err != nil {
			return err
		}
		policy.NextPurgeAt = &next
	}
	policy.IsActive = isActive
	return u.repo.Update(ctx, policy)
}

// TriggerNow hands a real retention run to the Retention Worker by moving the
// durable schedule deadline to now. The API never executes archive work itself.
func (u *usecase) TriggerNow(ctx context.Context, policyID int, productID int) (*dto.RetentionPolicyResponse, error) {
	policy, err := u.repo.GetByID(ctx, policyID, productID)
	if err != nil {
		return nil, err
	}
	if policy.EnvironmentID == nil || *policy.EnvironmentID <= 0 {
		return nil, errors.New("policy environment_id is required")
	}
	if err := u.authorize(ctx, productID, *policy.EnvironmentID, "UPDATE"); err != nil {
		return nil, err
	}
	if !policy.IsActive || !policy.ArchiveEnabled {
		return nil, errors.New("policy must be active with archive enabled to run now")
	}
	now := time.Now().UTC()
	policy.NextPurgeAt = &now
	if err := u.repo.Update(ctx, policy); err != nil {
		return nil, err
	}
	return toResponse(policy), nil
}

func (u *usecase) GetStats(ctx context.Context, productID int, environmentID *int) (*dto.PolicyStatsDTO, error) {
	if environmentID != nil {
		if *environmentID <= 0 {
			return nil, errors.New("environment_id must be greater than zero")
		}
		if err := u.authorize(ctx, productID, *environmentID, "READ"); err != nil {
			return nil, err
		}
		if err := u.validateEnvironment(ctx, productID, *environmentID); err != nil {
			return nil, err
		}
	}
	stats, err := u.repo.Stats(ctx, productID, environmentID)
	if err != nil {
		return nil, err
	}

	return &dto.PolicyStatsDTO{
		TotalPolicies:           stats.TotalPolicies,
		ActivePolicies:          stats.ActivePolicies,
		TotalStorageBytes:       stats.TotalStorageBytes,
		ActiveStorageBytes:      stats.ActiveStorageBytes,
		ArchiveGZIPStorageBytes: stats.ArchiveGZIPStorageBytes,
		ReadyZIPStorageBytes:    stats.ReadyZIPStorageBytes,
		TotalStorageGB:          float64(stats.TotalStorageBytes) / (1024 * 1024 * 1024),
		ScheduledPurges24h:      stats.ScheduledPurges24h,
		TotalFoldersCount:       stats.TotalFoldersCount,
	}, nil
}

func (u *usecase) Historical(ctx context.Context, req dto.RetentionHistoricalRequest) (*dto.RetentionHistoricalResponse, error) {
	if err := u.authorize(ctx, req.ProductID, req.EnvironmentID, "READ"); err != nil {
		return nil, err
	}
	if err := u.validateEnvironment(ctx, req.ProductID, req.EnvironmentID); err != nil {
		return nil, err
	}
	return u.repo.Historical(ctx, req)
}

func (u *usecase) Preview(ctx context.Context, req dto.RetentionPreviewRequest) (*dto.RetentionPreviewResponse, error) {
	if err := u.authorize(ctx, req.ProductID, req.EnvironmentID, "READ"); err != nil {
		return nil, err
	}
	if err := u.validateEnvironment(ctx, req.ProductID, req.EnvironmentID); err != nil {
		return nil, err
	}
	historical, err := u.repo.Historical(ctx, req.RetentionHistoricalRequest)
	if err != nil {
		return nil, err
	}
	result := &dto.RetentionPreviewResponse{Historical: *historical, DeleteRequiresVerified: true}
	result.ArchiveCandidateCount = historical.DocumentCount
	if req.DateFrom != nil && req.DateTo != nil {
		result.DeleteCandidateCount, err = u.repo.CountCandidates(ctx, req.RetentionHistoricalRequest)
		if err != nil {
			return nil, err
		}
	}
	if req.PolicyID != nil {
		policy, policyErr := u.repo.GetByID(ctx, *req.PolicyID, req.ProductID)
		if policyErr != nil {
			return nil, policyErr
		}
		now := time.Now().UTC()
		if policy.ActiveRetentionValue > 0 {
			cutoff, addErr := helper.Subtract(now, policy.ActiveRetentionValue, policy.ActiveRetentionUnit)
			if addErr != nil {
				return nil, addErr
			}
			result.ActiveCutoff = &cutoff
		}
		if policy.ArchiveEnabled && policy.ArchiveAfterValue != nil {
			cutoff, addErr := helper.Subtract(now, *policy.ArchiveAfterValue, policy.ArchiveAfterUnit)
			if addErr != nil {
				return nil, addErr
			}
			result.ArchiveCutoff = &cutoff
		}
	}
	return result, nil
}

func (u *usecase) Simulate(ctx context.Context, req dto.SimulateRetentionRequest) (*dto.SimulationResultDTO, error) {
	mode := req.RetentionMode
	if mode == "" {
		mode = "WEEKLY"
	}
	unit := req.RetentionUnit
	if unit == "" {
		unit = "WEEKS"
	}
	val := req.RetentionValue
	if val <= 0 {
		val = 1
	}
	structure := req.FolderStructure
	if structure == "" {
		structure = defaultFolderStructure(mode)
	}

	now := time.Now().UTC()
	effectiveUnit := unit
	if strings.TrimSpace(effectiveUnit) == "" {
		switch mode {
		case "DAILY":
			effectiveUnit = helper.UnitDay
		case "WEEKLY":
			effectiveUnit = helper.UnitWeek
		case "MONTHLY":
			effectiveUnit = helper.UnitMonth
		default:
			effectiveUnit = helper.UnitDay
		}
	}
	cutoff, cutoffErr := helper.Subtract(now, val, effectiveUnit)
	if cutoffErr != nil {
		return nil, cutoffErr
	}
	totalDays := int(now.Sub(cutoff).Hours() / 24)
	cutoffStr := cutoff.Format("2006-01-02")

	rootNode := dto.FolderNodeDTO{
		Key:         "root",
		Title:       "/var/log/omnilogs-archive",
		Path:        "/omnilogs-archive",
		Type:        "root",
		FolderCount: 0,
		FileCount:   0,
		EstimatedMB: 0,
		Status:      "ACTIVE",
	}

	var children []dto.FolderNodeDTO
	totalFolders := 0
	activeFolders := 0
	purgeFolders := 0
	estimatedSavingsMB := 0.0

	if mode == "WEEKLY" || structure == "7_DAILY_FOLDERS" {
		// 1 Week contains 7 daily folders (e.g. day-01 to day-07)
		weekNode := dto.FolderNodeDTO{
			Key:         "week-current",
			Title:       fmt.Sprintf("Week 1 (Retained %d Days / 7 Daily Folders)", totalDays),
			Path:        "/omnilogs-archive/week-01",
			Type:        "week",
			Status:      "ACTIVE",
			FolderCount: 7,
		}
		var dayNodes []dto.FolderNodeDTO
		for day := 1; day <= 7; day++ {
			dayDate := now.AddDate(0, 0, -(day - 1))
			isPurged := dayDate.Before(cutoff)
			status := "ACTIVE"
			if isPurged {
				status = "TO_BE_PURGED"
				purgeFolders++
				estimatedSavingsMB += 250.5
			} else {
				activeFolders++
			}
			totalFolders++

			dayNodes = append(dayNodes, dto.FolderNodeDTO{
				Key:         fmt.Sprintf("day-%d", day),
				Title:       fmt.Sprintf("Day %02d (%s)", day, dayDate.Format("Mon 02 Jan")),
				Path:        fmt.Sprintf("/omnilogs-archive/week-01/day-%02d", day),
				Type:        "day",
				FolderCount: 0,
				FileCount:   24,
				EstimatedMB: 250.5,
				Status:      status,
				ExpiryDate:  cutoffStr,
			})
		}
		weekNode.Children = dayNodes
		children = append(children, weekNode)

	} else if mode == "MONTHLY" || structure == "4_WEEKLY_SUBFOLDERS" {
		// 1 Month contains 4 weekly subfolders, each containing 7 daily folders (28 day folders total)
		monthPath := now.Format("2006-01")
		monthNode := dto.FolderNodeDTO{
			Key:         "month-current",
			Title:       fmt.Sprintf("Month 1 (Retained %d Days / 4 Weekly Subfolders)", totalDays),
			Path:        "/omnilogs-archive/" + monthPath,
			Type:        "month",
			Status:      "ACTIVE",
			FolderCount: 4,
		}

		var weekNodes []dto.FolderNodeDTO
		dayCounter := 1

		for week := 1; week <= 4; week++ {
			wNode := dto.FolderNodeDTO{
				Key:         fmt.Sprintf("week-%d", week),
				Title:       fmt.Sprintf("Week %d (7 Daily Subfolders)", week),
				Path:        fmt.Sprintf("/omnilogs-archive/%s/week-%02d", monthPath, week),
				Type:        "week",
				Status:      "ACTIVE",
				FolderCount: 7,
			}
			var subDays []dto.FolderNodeDTO
			for d := 1; d <= 7; d++ {
				dayDate := now.AddDate(0, 0, -(dayCounter - 1))
				isPurged := dayDate.Before(cutoff)
				status := "ACTIVE"
				if isPurged {
					status = "TO_BE_PURGED"
					purgeFolders++
					estimatedSavingsMB += 210.0
				} else {
					activeFolders++
				}
				totalFolders++

				subDays = append(subDays, dto.FolderNodeDTO{
					Key:         fmt.Sprintf("w%d-d%d", week, d),
					Title:       fmt.Sprintf("Day %02d (%s)", dayCounter, dayDate.Format("02 Jan")),
					Path:        fmt.Sprintf("/omnilogs-archive/%s/week-%02d/day-%02d", monthPath, week, d),
					Type:        "day",
					FolderCount: 0,
					FileCount:   24,
					EstimatedMB: 210.0,
					Status:      status,
					ExpiryDate:  cutoffStr,
				})
				dayCounter++
			}
			wNode.Children = subDays
			weekNodes = append(weekNodes, wNode)
		}
		monthNode.Children = weekNodes
		children = append(children, monthNode)

	} else {
		// Daily / Custom mode
		var dayNodes []dto.FolderNodeDTO
		for day := 1; day <= totalDays; day++ {
			dayDate := now.AddDate(0, 0, -(day - 1))
			isPurged := dayDate.Before(cutoff)
			status := "ACTIVE"
			if isPurged {
				status = "TO_BE_PURGED"
				purgeFolders++
				estimatedSavingsMB += 180.0
			} else {
				activeFolders++
			}
			totalFolders++

			dayNodes = append(dayNodes, dto.FolderNodeDTO{
				Key:         fmt.Sprintf("custom-d%d", day),
				Title:       fmt.Sprintf("Folder %s", dayDate.Format("YYYY-MM-DD")),
				Path:        fmt.Sprintf("/omnilogs-archive/%s", dayDate.Format("2006-01-02")),
				Type:        "day",
				FolderCount: 0,
				FileCount:   24,
				EstimatedMB: 180.0,
				Status:      status,
				ExpiryDate:  cutoffStr,
			})
		}
		children = dayNodes
	}

	rootNode.Children = children
	rootNode.FolderCount = totalFolders

	return &dto.SimulationResultDTO{
		RetentionMode:        mode,
		TotalRetentionDays:   totalDays,
		CalculatedCutoffDate: cutoffStr,
		TotalFoldersCount:    totalFolders,
		ActiveFoldersCount:   activeFolders,
		PurgeFoldersCount:    purgeFolders,
		EstimatedSavingsMB:   estimatedSavingsMB,
		TreeData:             rootNode,
	}, nil
}

// Helpers
func calculateTotalDays(mode, unit string, val int) int {
	if val <= 0 {
		val = 1
	}
	switch mode {
	case "DAILY":
		return val
	case "WEEKLY":
		return val * 7
	case "MONTHLY":
		return calendarDays(val, helper.UnitMonth)
	case "CUSTOM":
		switch unit {
		case "DAYS":
			return val
		case "WEEKS":
			return val * 7
		case "MONTHS", "MONTH":
			return calendarDays(val, helper.UnitMonth)
		case "YEARS", "YEAR":
			return calendarDays(val, helper.UnitYear)
		default:
			return val
		}
	default:
		return val * 7
	}
}

func calendarDays(value int, unit string) int {
	if value <= 0 {
		return 0
	}
	now := time.Now().UTC()
	cutoff, err := helper.Subtract(now, value, unit)
	if err != nil {
		return 0
	}
	return int(now.Sub(cutoff).Hours() / 24)
}

func normalizeActiveDuration(value int, unit string, legacyValue int, legacyUnit string) (int, string, error) {
	if value <= 0 {
		value = legacyValue
	}
	if value <= 0 {
		value = 1
	}
	if strings.TrimSpace(unit) == "" {
		unit = legacyUnit
	}
	if strings.TrimSpace(unit) == "" {
		unit = helper.UnitDay
	}
	normalized, err := helper.NormalizeUnit(unit)
	if err != nil {
		return 0, "", err
	}
	return value, normalized, nil
}

func normalizeOptionalUnit(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	normalized, err := helper.NormalizeUnit(value)
	if err != nil {
		return strings.ToUpper(strings.TrimSpace(value))
	}
	return normalized
}

func validateArchiveFields(enabled bool, afterValue *int, afterUnit string, archiveValue *int, archiveUnit string, neverDelete bool) error {
	if !enabled {
		return nil
	}
	if afterValue == nil || *afterValue <= 0 {
		return errors.New("archive_after_value is required when archive is enabled")
	}
	if _, err := helper.NormalizeUnit(afterUnit); err != nil {
		return errors.New("archive_after_unit is required when archive is enabled")
	}
	if neverDelete {
		return nil
	}
	if archiveValue == nil || *archiveValue <= 0 {
		return errors.New("archive retention is required unless archive_never_delete is true")
	}
	if _, err := helper.NormalizeUnit(archiveUnit); err != nil {
		return errors.New("archive_retention_unit is required unless archive_never_delete is true")
	}
	return nil
}

func intPointer(value int) *int { return &value }

func validateCalendarPolicy(policy *models.LogRetentionPolicy) error {
	if policy.EnvironmentID == nil || *policy.EnvironmentID <= 0 {
		return errors.New("environment_id is required")
	}
	if err := helper.Validate(policy.ActiveRetentionValue, policy.ActiveRetentionUnit, false); err != nil {
		return err
	}
	return validateArchiveFields(policy.ArchiveEnabled, policy.ArchiveAfterValue, policy.ArchiveAfterUnit, policy.ArchiveRetentionValue, policy.ArchiveRetentionUnit, policy.ArchiveNeverDelete)
}

func (u *usecase) validateEnvironment(ctx context.Context, productID, environmentID int) error {
	if productID <= 0 || environmentID <= 0 {
		return errors.New("product_id and environment_id are required")
	}
	ok, err := u.repo.EnvironmentBelongsToProduct(ctx, productID, environmentID)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("environment does not belong to product")
	}
	return nil
}

func (u *usecase) authorize(ctx context.Context, productID, environmentID int, action string) error {
	actor, ok := actorFromContext(ctx)
	if !ok {
		return responses.ErrForbidden
	}
	if u.authz == nil {
		return responses.ErrForbidden
	}
	product := productID
	environment := environmentID
	value, err := u.authz.CheckPermission(globalauth.Actor{UserID: actor.UserID, PlatformAdmin: actor.PlatformAdmin}, dto.PermissionCheckRequest{ProductID: &product, EnvironmentID: &environment, ResourceType: "RETENTION", Action: strings.ToUpper(action)})
	if err != nil {
		return err
	}
	if !value.Allowed {
		return responses.ErrForbidden
	}
	return nil
}

func defaultFolderStructure(mode string) string {
	switch mode {
	case "WEEKLY":
		return "7_DAILY_FOLDERS"
	case "MONTHLY":
		return "4_WEEKLY_SUBFOLDERS"
	case "DAILY":
		return "DAILY_FOLDERS"
	default:
		return "CUSTOM"
	}
}

func toResponse(p *models.LogRetentionPolicy) *dto.RetentionPolicyResponse {
	return &dto.RetentionPolicyResponse{
		PolicyID:                 p.PolicyID,
		ProductID:                p.ProductID,
		EnvironmentID:            p.EnvironmentID,
		ProjectID:                p.ProjectID,
		Name:                     p.Name,
		Description:              p.Description,
		RetentionMode:            p.RetentionMode,
		RetentionUnit:            p.RetentionUnit,
		RetentionValue:           p.RetentionValue,
		TotalDays:                p.TotalDays,
		FolderStructure:          p.FolderStructure,
		AutoPurgeAction:          p.AutoPurgeAction,
		StorageProvider:          p.StorageProvider,
		BucketName:               p.BucketName,
		CronSchedule:             p.CronSchedule,
		ScheduleTimezone:         p.ScheduleTimezone,
		IsActive:                 p.IsActive,
		LastPurgeAt:              p.LastPurgeAt,
		NextPurgeAt:              p.NextPurgeAt,
		CreatedAt:                p.CreatedAt,
		UpdatedAt:                p.UpdatedAt,
		ActiveRetentionValue:     p.ActiveRetentionValue,
		ActiveRetentionUnit:      p.ActiveRetentionUnit,
		ArchiveEnabled:           p.ArchiveEnabled,
		ArchiveAfterValue:        p.ArchiveAfterValue,
		ArchiveAfterUnit:         p.ArchiveAfterUnit,
		DeleteActiveAfterArchive: p.DeleteActiveAfterArchive,
		ArchiveRetentionValue:    p.ArchiveRetentionValue,
		ArchiveRetentionUnit:     p.ArchiveRetentionUnit,
		ArchiveNeverDelete:       p.ArchiveNeverDelete,
		ApplyToExistingLogs:      p.ApplyToExistingLogs,
		EffectiveFrom:            p.EffectiveFrom,
	}
}
