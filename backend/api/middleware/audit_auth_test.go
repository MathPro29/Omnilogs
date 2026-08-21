package middleware

import (
	"testing"

	"omnilogs-api/models"

	"github.com/gin-gonic/gin"
)

func TestAPIKeyAllowsPermission(t *testing.T) {
	if !apiKeyAllowsPermission([]byte(`["log_ingest_create"]`), "LOG_INGEST_CREATE") {
		t.Fatal("ingest permission should be case insensitive")
	}
	if apiKeyAllowsPermission([]byte(`["LOG_READ"]`), "LOG_INGEST_CREATE") {
		t.Fatal("read-only API key must not ingest logs")
	}
	if apiKeyAllowsPermission([]byte(`{"all":true}`), "LOG_INGEST_CREATE") {
		t.Fatal("unexpected permission shapes must fail closed")
	}
}

func TestAPIKeyIdentityStopsAtProduct(t *testing.T) {
	projectID, categoryID := 10, 20
	ctx, _ := gin.CreateTestContext(nil)
	applyAPIKeyIdentity(ctx, models.ProductAPIKey{
		KeyID: 1, ProductID: 2, DefaultProjectID: &projectID, DefaultCategoryID: &categoryID,
	})
	if value, _ := ctx.Get("service_product_id"); value != 2 {
		t.Fatalf("product identity missing: %#v", value)
	}
	if _, exists := ctx.Get("service_project_id"); exists {
		t.Fatal("API key default_project_id must not become request project identity")
	}
	if _, exists := ctx.Get("service_category_id"); exists {
		t.Fatal("API key default_category_id must not become request hierarchy identity")
	}
}
