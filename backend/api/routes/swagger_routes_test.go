package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"omnilogs-api/configs"

	"github.com/gin-gonic/gin"
)

func TestSwaggerRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("swagger disabled returns not found", func(t *testing.T) {
		router := gin.New()
		SwaggerRoutes(router, &configs.Env{SwaggerEnabled: false})

		request := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", recorder.Code)
		}
	})

	t.Run("swagger enabled serves spec", func(t *testing.T) {
		router := gin.New()
		SwaggerRoutes(router, &configs.Env{SwaggerEnabled: true})

		request := httptest.NewRequest(http.MethodGet, "/swagger/openapi.yaml", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
	})
}
