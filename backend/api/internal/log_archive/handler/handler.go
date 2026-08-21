package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"omnilogs-api/dto"
	"omnilogs-api/internal/log_archive/service"
	"omnilogs-api/middleware"
	"omnilogs-api/models"
	"omnilogs-api/responses"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct{ service service.Service }

// NewHandler keeps the existing route wiring compatible. The optional
// authorizer is the established product access usecase.
func NewHandler(db *gorm.DB, esClient *elasticsearch.Client, authorizers ...interface {
	CheckPermission(any, dto.PermissionCheckRequest) (*dto.PermissionCheckResponse, error)
}) *Handler {
	// The concrete global-auth interface has a typed Actor, so route wiring uses
	// NewHandlerWithService below. This constructor remains for old callers.
	return &Handler{service: service.New(db, esClient, nil)}
}

func NewHandlerWithService(value service.Service) *Handler { return &Handler{service: value} }

func (h *Handler) List(c *gin.Context) {
	productID, environmentID, ok := ids(c)
	if !ok {
		return
	}
	categoryID, err := optionalPositiveInt(c.Query("category_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
		return
	}
	featureID, err := optionalPositiveInt(c.Query("feature_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid feature_id"})
		return
	}
	subFeatureID, err := optionalPositiveInt(c.Query("sub_feature_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sub_feature_id"})
		return
	}
	from, err := optionalTime(c.Query("date_from"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date_from"})
		return
	}
	to, err := optionalTime(c.Query("date_to"))
	if err != nil || (from != nil && to != nil && !to.After(*from)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date_to"})
		return
	}
	archives, err := h.service.List(withActor(c), dto.ListArchiveRequest{
		ProductID: productID, EnvironmentID: environmentID, DateFrom: from, DateTo: to,
		CategoryID: categoryID, FeatureID: featureID, SubFeatureID: subFeatureID,
		BackupType: c.Query("backup_type"), BackupTag: c.Query("backup_tag"),
	})
	if err != nil {
		writeError(c, err)
		return
	}
	if group := strings.ToLower(strings.TrimSpace(c.Query("group_by"))); group != "" && group != "day" {
		result, groupErr := groupArchives(archives, group)
		if groupErr != nil {
			writeError(c, groupErr)
			return
		}
		c.JSON(http.StatusOK, result)
		return
	}
	c.JSON(http.StatusOK, archives)
}

func (h *Handler) Create(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("productId"))
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	var req dto.CreateArchiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ProductID = productID
	c.Set("audit_action", "ARCHIVE_STARTED")
	job, archives, err := h.service.Create(withActor(c), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Set("audit_action", "ARCHIVE_COMPLETED")
	c.JSON(http.StatusAccepted, gin.H{"job": job, "archives": archives})
}

func (h *Handler) Get(c *gin.Context) {
	productID, environmentID, ok := ids(c)
	if !ok {
		return
	}
	archive, err := h.service.Get(withActor(c), productID, environmentID, c.Param("archiveId"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, archive)
}

func (h *Handler) Verify(c *gin.Context) {
	productID, environmentID, ok := ids(c)
	if !ok {
		return
	}
	var req dto.VerifyArchiveRequest
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	c.Set("audit_action", "ARCHIVE_VERIFIED")
	result, err := h.service.Verify(withActor(c), productID, environmentID, c.Param("archiveId"), req.PolicyID)
	if err != nil {
		c.Set("audit_action", "ARCHIVE_FAILED")
		writeError(c, err)
		return
	}
	if result.DeletedActive {
		c.Set("audit_action", "HISTORICAL_LOGS_DELETED")
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) Restore(c *gin.Context) {
	productID, environmentID, ok := ids(c)
	if !ok {
		return
	}
	var req dto.RestoreArchiveRequest
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	c.Set("audit_action", "RESTORE_STARTED")
	result, err := h.service.Restore(withActor(c), productID, environmentID, c.Param("archiveId"), req)
	if err != nil {
		c.Set("audit_action", "RESTORE_FAILED")
		writeError(c, err)
		return
	}
	c.Set("audit_action", "RESTORE_COMPLETED")
	c.JSON(http.StatusOK, result)
}

func (h *Handler) ListJobs(c *gin.Context) {
	productID, environmentID, ok := ids(c)
	if !ok {
		return
	}
	jobs, err := h.service.ListJobs(withActor(c), productID, environmentID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, jobs)
}

func (h *Handler) Download(c *gin.Context) {
	productID, environmentID, ok := ids(c)
	if !ok {
		return
	}
	var policyID *int
	if value := strings.TrimSpace(c.Query("policy_id")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy_id"})
			return
		}
		policyID = &parsed
	}
	categoryID, err := optionalPositiveInt(c.Query("category_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
		return
	}
	featureID, err := optionalPositiveInt(c.Query("feature_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid feature_id"})
		return
	}
	subFeatureID, err := optionalPositiveInt(c.Query("sub_feature_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sub_feature_id"})
		return
	}
	from, err := optionalTime(c.Query("date_from"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date_from"})
		return
	}
	to, err := optionalTime(c.Query("date_to"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date_to"})
		return
	}
	if from != nil && to != nil && !to.After(*from) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date_to must be after date_from"})
		return
	}
	download, err := h.service.Download(withActor(c), productID, environmentID, policyID, categoryID, featureID, subFeatureID, from, to)
	if err != nil {
		writeError(c, err)
		return
	}
	defer func() {
		_ = download.File.Close()
		_ = os.Remove(download.File.Name())
	}()
	c.Header("Content-Disposition", `attachment; filename="`+download.Filename+`"`)
	c.Header("Cache-Control", "no-store")
	c.DataFromReader(http.StatusOK, download.Size, "application/zip", download.File, nil)
}

func (h *Handler) ListReadyDownloads(c *gin.Context) {
	productID, environmentID, ok := ids(c)
	if !ok {
		return
	}
	downloads, err := h.service.ListReadyDownloads(withActor(c), productID, environmentID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, downloads)
}

func (h *Handler) DownloadReady(c *gin.Context) {
	productID, environmentID, ok := ids(c)
	if !ok {
		return
	}
	download, err := h.service.DownloadReady(withActor(c), productID, environmentID, c.Param("downloadId"))
	if err != nil {
		writeError(c, err)
		return
	}
	defer func() {
		_ = download.File.Close()
	}()
	c.Header("Content-Disposition", `attachment; filename="`+download.Filename+`"`)
	c.Header("Cache-Control", "no-store")
	c.DataFromReader(http.StatusOK, download.Size, "application/zip", download.File, nil)
}

func (h *Handler) Import(c *gin.Context) {
	productID, environmentID, ok := ids(c)
	if !ok {
		return
	}
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "multipart file field 'file' is required"})
		return
	}
	defer file.Close()
	temp, err := os.CreateTemp("", "omnilogs-upload-*.zip")
	if err != nil {
		writeError(c, err)
		return
	}
	tempPath := temp.Name()
	removeTemp := true
	defer func() {
		_ = temp.Close()
		if removeTemp {
			_ = os.Remove(tempPath)
		}
	}()
	if _, err := io.Copy(temp, file); err != nil {
		writeError(c, err)
		return
	}
	if err := temp.Close(); err != nil {
		writeError(c, err)
		return
	}
	categoryID, err := optionalPositiveInt(c.Query("category_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
		return
	}
	featureID, err := optionalPositiveInt(c.Query("feature_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid feature_id"})
		return
	}
	subFeatureID, err := optionalPositiveInt(c.Query("sub_feature_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sub_feature_id"})
		return
	}
	result, err := h.service.Import(withActor(c), productID, environmentID, tempPath, categoryID, featureID, subFeatureID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func optionalPositiveInt(value string) (*int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return nil, errors.New("invalid positive integer")
	}
	return &parsed, nil
}

func optionalTime(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err == nil {
		return &parsed, nil
	}
	parsed, err = time.Parse("2006-01-02", value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
func (h *Handler) GetJob(c *gin.Context) {
	productID, environmentID, ok := ids(c)
	if !ok {
		return
	}
	job, err := h.service.GetJob(withActor(c), productID, environmentID, c.Param("jobId"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, job)
}

func ids(c *gin.Context) (int, int, bool) {
	productID, err := strconv.Atoi(c.Param("productId"))
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return 0, 0, false
	}
	environmentID, err := strconv.Atoi(c.Query("environment_id"))
	if err != nil || environmentID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "environment_id query parameter is required"})
		return 0, 0, false
	}
	return productID, environmentID, true
}

func withActor(c *gin.Context) context.Context {
	id, _ := middleware.CurrentUserID(c)
	return service.WithActor(c.Request.Context(), service.Actor{UserID: int(id), PlatformAdmin: middleware.HasAdminPlatformRole(c)})
}

func writeError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	if errors.Is(err, responses.ErrForbidden) {
		status = http.StatusForbidden
	}
	if errors.Is(err, responses.ErrNotFound) {
		status = http.StatusNotFound
	}
	if strings.Contains(strings.ToLower(err.Error()), "no verified archives matched") {
		status = http.StatusNotFound
	}
	if strings.Contains(strings.ToLower(err.Error()), "restore incomplete") {
		status = http.StatusUnprocessableEntity
	}
	if strings.Contains(strings.ToLower(err.Error()), "required") || strings.Contains(strings.ToLower(err.Error()), "invalid") || strings.Contains(strings.ToLower(err.Error()), "must be") {
		status = http.StatusBadRequest
	}
	c.JSON(status, gin.H{"error": err.Error()})
}

func groupArchives(values []models.LogArchive, group string) ([]dto.ArchiveGroupResponse, error) {
	if group != "week" && group != "month" && group != "year" {
		return nil, errors.New("group_by must be day, week, month, or year")
	}
	type aggregate struct {
		dto.ArchiveGroupResponse
		key string
	}
	groups := map[string]*aggregate{}
	for _, archive := range values {
		if archive.DateFrom == nil || archive.DateTo == nil {
			continue
		}
		from := archive.DateFrom.UTC()
		keyFrom := from
		switch group {
		case "week":
			year, week := from.ISOWeek()
			keyFrom = time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, (week-1)*7)
		case "month":
			keyFrom = time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, time.UTC)
		case "year":
			keyFrom = time.Date(from.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		}
		key := keyFrom.Format(time.RFC3339) + ":" + archiveScopeKey(archive.CategoryID, archive.FeatureID, archive.SubFeatureID)
		value := groups[key]
		if value == nil {
			value = &aggregate{key: key, ArchiveGroupResponse: dto.ArchiveGroupResponse{Period: keyFrom.Format("2006-01-02"), DateFrom: keyFrom, DateTo: keyFrom.AddDate(0, 0, 1), CategoryID: archive.CategoryID, FeatureID: archive.FeatureID, SubFeatureID: archive.SubFeatureID}}
			switch group {
			case "week":
				value.DateTo = keyFrom.AddDate(0, 0, 7)
			case "month":
				value.DateTo = keyFrom.AddDate(0, 1, 0)
			case "year":
				value.DateTo = keyFrom.AddDate(1, 0, 0)
			}
			groups[key] = value
		}
		value.DocumentCount += archive.DocumentCount
		value.OriginalBytes += archive.OriginalSizeBytes
		if archive.CompressedSizeBytes != nil {
			value.CompressedBytes += *archive.CompressedSizeBytes
		}
		value.ArchiveCount++
		value.Status = mergeStatus(value.Status, archive.Status)
	}
	result := make([]dto.ArchiveGroupResponse, 0, len(groups))
	for _, value := range groups {
		result = append(result, value.ArchiveGroupResponse)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].DateFrom.After(result[j].DateFrom) })
	return result, nil
}

func archiveScopeKey(categoryID, featureID, subFeatureID *int) string {
	return nullableScopeValue(categoryID) + "/" + nullableScopeValue(featureID) + "/" + nullableScopeValue(subFeatureID)
}

func nullableScopeValue(value *int) string {
	if value == nil {
		return "all"
	}
	return strconv.Itoa(*value)
}

func mergeStatus(current string, next *string) string {
	if next == nil {
		return current
	}
	if current == "" {
		return *next
	}
	if current == *next {
		return current
	}
	return "MIXED"
}
