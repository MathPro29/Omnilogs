package audit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestShouldRecordAuditEventExcludesIngestionQueueEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, path := range []string{"/api/v1/queues", "/api/v1/queues/consume", "/api/v1/ingest/logs", "/api/v1/logs/ingest", "/api/v1/logs/search"} {
		t.Run(path, func(t *testing.T) {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = httptest.NewRequest(http.MethodPost, path, nil)
			if shouldRecordAuditEvent(ctx) {
				t.Fatalf("ingestion queue endpoint %q must not create a system audit log", path)
			}
		})
	}
}

func TestShouldRecordAuditEventStillRecordsUserMutation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/products/1/api-keys/2", nil)
	if !shouldRecordAuditEvent(ctx) {
		t.Fatal("user initiated API key revocation must remain audited")
	}
}
