package responses

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestInternalErrorDoesNotExposeCause(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	Error(ctx, "INTERNAL_ERROR", "request failed", errors.New("password=do-not-expose"))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if strings.Contains(recorder.Body.String(), "do-not-expose") {
		t.Fatalf("internal cause leaked in response: %s", recorder.Body.String())
	}
}
