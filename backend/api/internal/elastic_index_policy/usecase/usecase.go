package usecase

import (
	"errors"
	"fmt"
	"omnilogs-api/dto"
	"omnilogs-api/internal/elastic_index_policy/repository"
	"omnilogs-api/models"
	"strings"
)

type Usecase interface {
	Create(req dto.CreateElasticIndexPolicyRequest) (*dto.ElasticIndexPolicyResponse, error)
	Update(req dto.UpdateElasticIndexPolicyRequest) (*dto.ElasticIndexPolicyResponse, error)
	Delete(req dto.DeleteElasticIndexPolicyRequest) (*dto.ElasticIndexPolicyResponse, error)
	List(req dto.ListElasticIndexPolicyRequest) ([]dto.ElasticIndexPolicyResponse, error)
	GetByID(req dto.GetElasticIndexPolicyRequest) (*dto.ElasticIndexPolicyResponse, error)
	PushToArchives(req dto.PushToArchivesRequest) (*dto.PushToArchivesResponse, error)
	ClearAllLogs(productID int) error
}

type usecase struct {
	repo repository.Repository
}

func NewUsecase(repo repository.Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) Create(req dto.CreateElasticIndexPolicyRequest) (*dto.ElasticIndexPolicyResponse, error) {
	// ตรวจรูปแบบก่อนบันทึก เพื่อให้ชื่อ index และ policy ถูกใช้ต่อได้จริงใน worker
	normalizedPrefix, err := normalizeElasticIndexPrefix(req.IndexPrefix)
	if err != nil {
		return nil, err
	}
	rolloverType, err := normalizeRolloverType(req.RolloverType)
	if err != nil {
		return nil, err
	}

	shards := 1
	if req.NumberOfShards != nil {
		shards = *req.NumberOfShards
	}
	replicas := 1
	if req.NumberOfReplicas != nil {
		replicas = *req.NumberOfReplicas
	}
	schemaVer := "v1"
	if req.SchemaVersion != nil {
		schemaVer = strings.TrimSpace(*req.SchemaVersion)
	}
	if err := validateElasticIndexNumbers(shards, replicas); err != nil {
		return nil, err
	}
	if schemaVer == "" {
		return nil, errors.New("schema version is required")
	}
	retentionEnabled := req.RetentionDays != nil

	policy := &models.ElasticIndexPolicy{
		ProductID:        req.ProductID,
		EnvironmentID:    req.EnvironmentID,
		ProjectID:        req.ProjectID,
		CategoryID:       req.CategoryID,
		IndexPrefix:      normalizedPrefix,
		IndexPattern:     req.IndexPattern,
		WriteAlias:       req.WriteAlias,
		RolloverType:     rolloverType,
		NumberOfShards:   shards,
		NumberOfReplicas: replicas,
		RetentionEnabled: retentionEnabled,
		RetentionDays:    req.RetentionDays,
		SchemaVersion:    schemaVer,
		IsActive:         true,
	}

	if err := u.repo.Create(policy); err != nil {
		return nil, err
	}

	return mapToResponse(policy), nil
}

func (u *usecase) Update(req dto.UpdateElasticIndexPolicyRequest) (*dto.ElasticIndexPolicyResponse, error) {
	policy, err := u.repo.GetByID(req.ElasticPolicyID)
	if err != nil {
		return nil, err
	}

	if req.ProjectID != nil {
		policy.ProjectID = req.ProjectID
	}
	if req.CategoryID != nil {
		policy.CategoryID = req.CategoryID
	}
	if req.IndexPattern != nil {
		policy.IndexPattern = req.IndexPattern
	}
	if req.WriteAlias != nil {
		policy.WriteAlias = req.WriteAlias
	}
	if req.RolloverType != nil {
		rolloverType, normalizeErr := normalizeRolloverType(*req.RolloverType)
		if normalizeErr != nil {
			return nil, normalizeErr
		}
		policy.RolloverType = rolloverType
	}
	if req.NumberOfShards != nil {
		policy.NumberOfShards = *req.NumberOfShards
	}
	if req.NumberOfReplicas != nil {
		policy.NumberOfReplicas = *req.NumberOfReplicas
	}
	if req.RetentionDays != nil {
		policy.RetentionDays = req.RetentionDays
		policy.RetentionEnabled = true
	}
	if req.SchemaVersion != nil {
		policy.SchemaVersion = strings.TrimSpace(*req.SchemaVersion)
	}
	if req.IsActive != nil {
		policy.IsActive = *req.IsActive
	}
	if err := validateElasticIndexNumbers(policy.NumberOfShards, policy.NumberOfReplicas); err != nil {
		return nil, err
	}
	if strings.TrimSpace(policy.SchemaVersion) == "" {
		return nil, errors.New("schema version is required")
	}

	if err := u.repo.Update(policy); err != nil {
		return nil, err
	}

	return mapToResponse(policy), nil
}

func (u *usecase) Delete(req dto.DeleteElasticIndexPolicyRequest) (*dto.ElasticIndexPolicyResponse, error) {
	policy, err := u.repo.GetByID(req.ElasticPolicyID)
	if err != nil {
		return nil, err
	}

	if err := u.repo.Delete(req.ElasticPolicyID); err != nil {
		return nil, err
	}

	return mapToResponse(policy), nil
}

func (u *usecase) List(req dto.ListElasticIndexPolicyRequest) ([]dto.ElasticIndexPolicyResponse, error) {
	policies, err := u.repo.List(req.ProductID, req.EnvironmentID)
	if err != nil {
		return nil, err
	}

	res := make([]dto.ElasticIndexPolicyResponse, len(policies))
	for i := range policies {
		res[i] = *mapToResponse(&policies[i])
	}
	return res, nil
}

func (u *usecase) GetByID(req dto.GetElasticIndexPolicyRequest) (*dto.ElasticIndexPolicyResponse, error) {
	policy, err := u.repo.GetByID(req.ElasticPolicyID)
	if err != nil {
		return nil, err
	}
	if policy.ProductID != req.ProductID {
		return nil, errors.New("policy does not belong to product")
	}
	if req.EnvironmentID != nil {
		if policy.EnvironmentID == nil || *policy.EnvironmentID != *req.EnvironmentID {
			return nil, errors.New("policy does not belong to environment")
		}
	}

	return mapToResponse(policy), nil
}

func (u *usecase) PushToArchives(req dto.PushToArchivesRequest) (*dto.PushToArchivesResponse, error) {
	policy, err := u.repo.GetByID(req.ElasticPolicyID)
	if err != nil {
		return nil, err
	}

	if policy.ProductID != req.ProductID {
		return nil, errors.New("policy does not belong to product")
	}
	return u.repo.PushToArchives(req)
}

func (u *usecase) ClearAllLogs(productID int) error {
	return u.repo.ClearAllLogs(productID)
}

func mapToResponse(p *models.ElasticIndexPolicy) *dto.ElasticIndexPolicyResponse {
	return &dto.ElasticIndexPolicyResponse{
		ElasticPolicyID:  p.ElasticPolicyID,
		ProductID:        p.ProductID,
		EnvironmentID:    p.EnvironmentID,
		ProjectID:        p.ProjectID,
		CategoryID:       p.CategoryID,
		IndexPrefix:      p.IndexPrefix,
		IndexPattern:     p.IndexPattern,
		WriteAlias:       p.WriteAlias,
		RolloverType:     p.RolloverType,
		NumberOfShards:   p.NumberOfShards,
		NumberOfReplicas: p.NumberOfReplicas,
		RetentionDays:    p.RetentionDays,
		SchemaVersion:    p.SchemaVersion,
		IsActive:         p.IsActive,
		TimestampResponse: dto.TimestampResponse{
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		},
	}
}

func normalizeElasticIndexPrefix(value string) (string, error) {
	prefix := strings.ToLower(strings.TrimSpace(value))
	if prefix == "" {
		return "", errors.New("index prefix is required")
	}

	// จำกัดรูปแบบให้ปลอดภัยสำหรับชื่อ index ของ Elasticsearch
	for _, r := range prefix {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			continue
		}
		return "", fmt.Errorf("index prefix contains invalid character: %q", r)
	}

	return prefix, nil
}

func normalizeRolloverType(value string) (string, error) {
	rolloverType := strings.ToLower(strings.TrimSpace(value))
	switch rolloverType {
	case "size", "age":
		return rolloverType, nil
	default:
		return "", errors.New("invalid rollover type: allowed values are size or age")
	}
}

func validateElasticIndexNumbers(shards int, replicas int) error {
	if shards <= 0 {
		return errors.New("number of shards must be greater than 0")
	}
	if replicas < 0 {
		return errors.New("number of replicas must be greater than or equal to 0")
	}
	return nil
}
