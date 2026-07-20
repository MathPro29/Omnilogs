package repository

import (
	"log/slog"
	"time"
)

func (r *repository) logSlowElasticsearch(operation string, startedAt time.Time, indices []string, limit int) {
	duration := time.Since(startedAt)
	if duration < r.slowQueryThreshold {
		return
	}
	slog.Warn("slow elasticsearch query",
		"operation", operation,
		"duration_ms", duration.Milliseconds(),
		"slow_threshold_ms", r.slowQueryThreshold.Milliseconds(),
		"index_count", len(indices),
		"limit", limit,
	)
}
