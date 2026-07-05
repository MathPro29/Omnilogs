package usecase

import (
	"errors"
	"omnilogs-api/dto"
	"omnilogs-api/internal/elastic_index_policy/repository"
	"omnilogs-api/models"
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
		schemaVer = *req.SchemaVersion
	}
	retentionEnabled := req.RetentionDays != nil

	policy := &models.ElasticIndexPolicy{
		ProductID:        req.ProductID,
		EnvironmentID:    req.EnvironmentID,
		ProjectID:        req.ProjectID,
		CategoryID:       req.CategoryID,
		IndexPrefix:      req.IndexPrefix,
		IndexPattern:     req.IndexPattern,
		WriteAlias:       req.WriteAlias,
		RolloverType:     req.RolloverType,
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
		policy.RolloverType = *req.RolloverType
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
		policy.SchemaVersion = *req.SchemaVersion
	}
	if req.IsActive != nil {
		policy.IsActive = *req.IsActive
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
