package usecase

import (
	"context"
	"testing"

	"omnilogs-api/internal/main_logs/repository"
	"omnilogs-api/models"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type accessRepositoryStub struct {
	hasAccess bool
}

var _ repository.Repository = (*accessRepositoryStub)(nil)

func (s *accessRepositoryStub) SearchLogs(context.Context, []string, []byte) (*esapi.Response, error) {
	return nil, nil
}

func (s *accessRepositoryStub) GetLogIndexRef(context.Context, int64, string) (*models.LogIndexRef, error) {
	return nil, nil
}

func (s *accessRepositoryStub) GetMainLogFromES(context.Context, string, string) (*esapi.Response, error) {
	return nil, nil
}

func (s *accessRepositoryStub) GetPostgresPayload(context.Context, string) (*models.LogObjectStorageRef, error) {
	return nil, nil
}

func (s *accessRepositoryStub) HasMainLogAccess(context.Context, uint, int64) (bool, error) {
	return s.hasAccess, nil
}

func (s *accessRepositoryStub) GetActiveIndexPolicies(context.Context, int64) ([]models.ElasticIndexPolicy, error) {
	return nil, nil
}

func (s *accessRepositoryStub) GetRestoredArchiveIndexPattern(context.Context, int64, *int64, string) (string, error) {
	return "", nil
}

func TestEnsureProductAccessRequiresProductAccessForPlatformAdmin(t *testing.T) {
	u := &usecase{repo: &accessRepositoryStub{hasAccess: false}}

	if err := u.ensureProductAccess(context.Background(), 1, true, 99); err == nil {
		t.Fatal("platform admin without product access must be denied")
	}
}

func TestEnsureProductAccessAllowsAssignedProduct(t *testing.T) {
	u := &usecase{repo: &accessRepositoryStub{hasAccess: true}}

	if err := u.ensureProductAccess(context.Background(), 1, true, 99); err != nil {
		t.Fatalf("platform admin with product access must be allowed: %v", err)
	}
}

func TestResolveSearchIndicesSupportsAllEnvironments(t *testing.T) {
	u := &usecase{repo: &accessRepositoryStub{}}

	all := u.resolveSearchIndices(context.Background(), 99, nil)
	if len(all) != 1 || all[0] != "omnilogs-product-99-env-*-search-v3-*" {
		t.Fatalf("unexpected all-environment indices: %#v", all)
	}

	environmentID := int64(7)
	scoped := u.resolveSearchIndices(context.Background(), 99, &environmentID)
	if len(scoped) != 1 || scoped[0] != "omnilogs-product-99-env-7-search-v3-*" {
		t.Fatalf("unexpected scoped indices: %#v", scoped)
	}
}
