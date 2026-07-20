package usecase

import "omnilogs-api/dto"

const maxDashboardPageSize = 100

func NormalizeLogQuery(query dto.LogQuery) dto.LogQuery {
	if query.Limit <= 0 {
		query.Limit = 50
	} else if query.Limit > maxDashboardPageSize {
		query.Limit = maxDashboardPageSize
	}
	if query.Offset < 0 {
		query.Offset = 0
	}
	return query
}

func NormalizeAuditLogQuery(query dto.AuditLogQuery) dto.AuditLogQuery {
	if query.Limit <= 0 {
		query.Limit = 50
	} else if query.Limit > maxDashboardPageSize {
		query.Limit = maxDashboardPageSize
	}
	if query.Offset < 0 {
		query.Offset = 0
	}
	return query
}
