package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"omnilogs-api/internal/main_logs/usecase"
	"omnilogs-api/middleware"
	"omnilogs-api/models"
	"omnilogs-api/responses"

	"github.com/gin-gonic/gin"
)

type usecaseStub struct {
	authorize func(context.Context, uint, bool, int64) error
}

func (stub usecaseStub) Search(context.Context, usecase.SearchInput, *string, *string, *string, *string) (*usecase.SearchResult, error) {
	return nil, errors.New("not implemented")
}
func (stub usecaseStub) FindByID(context.Context, uint, bool, int64, string, *string, *string, *string, *string) (*usecase.MainLogDocument, error) {
	return nil, errors.New("not implemented")
}
func (stub usecaseStub) FindByAudit(context.Context, uint, bool, string, *string, *string, *string, *string) (*models.SystemAuditLog, *usecase.MainLogDocument, string, error) {
	return nil, nil, "", errors.New("not implemented")
}
func (stub usecaseStub) AuthorizeProductAccess(ctx context.Context, userID uint, admin bool, productID int64) error {
	return stub.authorize(ctx, userID, admin, productID)
}

func TestLiveTailRejectsUserWithoutProductAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(usecaseStub{authorize: func(context.Context, uint, bool, int64) error {
		return responses.ErrForbidden
	}}, nil)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/logs/live?product_id=99", nil)
	ctx.Set(middleware.ContextUserID, uint(5))

	handler.LiveTail(ctx)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestMapLiveMessageFiltersMissingOrDifferentEnvironment(t *testing.T) {
	withoutEnvironment := []byte(`{"log_id":"a","product_id":1,"payload":{}}`)
	if _, ok := mapLiveMessage(withoutEnvironment, "2"); ok {
		t.Fatal("message without environment must not bypass an environment filter")
	}

	matching := []byte(`{"log_id":"a","product_id":1,"environment_id":2,"payload":{"duration_ms":12}}`)
	mapped, ok := mapLiveMessage(matching, "2")
	if !ok || mapped["latency_ms"] != int64(12) {
		t.Fatalf("expected matching message, got %#v, %v", mapped, ok)
	}
}
