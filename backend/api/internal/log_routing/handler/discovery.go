package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"omnilogs-api/models"
)

// routingDiscoveryItem summarizes API paths observed from connector traffic.
// It lets an administrator create an explicit HTTP mapping from a real path,
// including paths that were already classified automatically.
type routingDiscoveryItem struct {
	EnvironmentID int       `json:"environment_id"`
	ServiceName   string    `json:"service_name"`
	RequestMethod string    `json:"request_method"`
	RequestPath   string    `json:"request_path"`
	RoutePattern  string    `json:"route_pattern,omitempty"`
	RoutingStatus string    `json:"routing_status"`
	RoutingMethod string    `json:"routing_method,omitempty"`
	TotalLogs     int64     `json:"total_logs"`
	LastSeenAt    time.Time `json:"last_seen_at"`
}

func (h *Handler) ListDiscovery(c *gin.Context) {
	productID := atoi(c.Param("productId"))
	if !h.allowed(c, productID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}

	var rows []routingDiscoveryItem
	err := h.db.Model(&models.LogIndexRef{}).
		Select(`environment_id, COALESCE(service_name, '') AS service_name, COALESCE(request_method, '') AS request_method, request_path, COALESCE(route_pattern, '') AS route_pattern, routing_status, COALESCE(routing_method, '') AS routing_method, COUNT(*) AS total_logs, MAX(timestamp) AS last_seen_at`).
		Where("product_id = ? AND request_path IS NOT NULL AND BTRIM(request_path) <> ''", productID).
		Group("environment_id, service_name, request_method, request_path, route_pattern, routing_status, routing_method").
		Order("total_logs DESC, last_seen_at DESC").
		Limit(100).
		Scan(&rows).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load unmapped log patterns"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rows})
}
