package routes

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("live returns ok", func(t *testing.T) {
		router := gin.New()
		HealthRoutes(router, func(context.Context) error { return nil })

		request := httptest.NewRequest(http.MethodGet, "/health/live", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
	})

	t.Run("ready returns unavailable when dependency fails", func(t *testing.T) {
		router := gin.New()
		HealthRoutes(router, func(context.Context) error { return errors.New("db down") })

		request := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected 503, got %d", recorder.Code)
		}
	})
}
