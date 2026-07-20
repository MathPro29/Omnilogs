package handler

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"omnilogs-api/internal/archive_utils"
	"omnilogs-api/models"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db       *gorm.DB
	esClient *elasticsearch.Client
}

func NewHandler(db *gorm.DB, esClient *elasticsearch.Client) *Handler {
	return &Handler{
		db:       db,
		esClient: esClient,
	}
}

// List exposes archive results for the retention test screen.
func (h *Handler) List(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("productId"))
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	query := h.db.Where("product_id = ?", productID).Order("created_at DESC")
	if value := c.Query("environment_id"); value != "" {
		environmentID, err := strconv.Atoi(value)
		if err != nil || environmentID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid environment id"})
			return
		}
		query = query.Where("environment_id = ?", environmentID)
	}

	var archives []models.LogArchive
	if err := query.Find(&archives).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, archives)
}

// Restore decompresses the archived GZIP file, bulk-indexes logs back to ES, and restores PostgreSQL audits.
func (h *Handler) Restore(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("productId"))
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	archiveID := c.Param("archiveId")
	if archiveID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid archive id"})
		return
	}

	var archive models.LogArchive
	if err := h.db.Where("archive_id = ? AND product_id = ?", archiveID, productID).First(&archive).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "archive not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if archive.FilePath == nil || *archive.FilePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "archive file path is empty"})
		return
	}

	// Resolve the original target index name from the archive file name.
	baseName := filepath.Base(*archive.FilePath)
	targetIndexName := strings.TrimSuffix(strings.TrimSuffix(baseName, ".json.gz"), ".csv.gz")
	if archive.EnvironmentID != nil {
		targetIndexName = strings.TrimSuffix(targetIndexName, fmt.Sprintf("-archive-env-%d", *archive.EnvironmentID))
	}

	restoredCount, err := archive_utils.RestoreIndex(c.Request.Context(), h.db, h.esClient, *archive.FilePath, targetIndexName)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":       "archive file not found",
				"file_path":   *archive.FilePath,
				"archive_id":  archive.ArchiveID,
				"next_action": "recreate the archive after mounting the shared archive storage",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to restore archive: %v", err)})
		return
	}

	now := time.Now().UTC()
	restoredStatus := "RESTORED"
	archive.Status = &restoredStatus
	archive.RestoredAt = &now

	if err := h.db.Save(&archive).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to update archive status: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "successfully restored logs and audit logs",
		"restored_count": restoredCount,
		"index_name":     targetIndexName,
	})
}
