package usecase

import (
	"testing"

	"omnilogs-api/dto"
)

func TestNormalizeLogQueryBoundsPagination(t *testing.T) {
	query := NormalizeLogQuery(dto.LogQuery{Limit: 1000, Offset: -20})
	if query.Limit != 100 || query.Offset != 0 {
		t.Fatalf("unexpected pagination: limit=%d offset=%d", query.Limit, query.Offset)
	}

	defaults := NormalizeLogQuery(dto.LogQuery{})
	if defaults.Limit != 50 || defaults.Offset != 0 {
		t.Fatalf("unexpected defaults: limit=%d offset=%d", defaults.Limit, defaults.Offset)
	}
}

func TestNormalizeAuditLogQueryBoundsPagination(t *testing.T) {
	query := NormalizeAuditLogQuery(dto.AuditLogQuery{Limit: 101, Offset: -1})
	if query.Limit != 100 || query.Offset != 0 {
		t.Fatalf("unexpected pagination: limit=%d offset=%d", query.Limit, query.Offset)
	}
}
