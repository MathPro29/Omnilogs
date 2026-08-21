package archive_utils

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

const (
	ArchiveFormatNDJSON    = "NDJSON"
	ArchiveCompressionGZIP = "GZIP"
	archivePageSize        = 500
)

type ArchiveManifest struct {
	Type          string    `json:"_type"`
	Version       int       `json:"version"`
	ProductID     int       `json:"product_id"`
	EnvironmentID int       `json:"environment_id"`
	CategoryID    *int      `json:"category_id,omitempty"`
	FeatureID     *int      `json:"feature_id,omitempty"`
	SubFeatureID  *int      `json:"sub_feature_id,omitempty"`
	DateFrom      time.Time `json:"date_from"`
	DateTo        time.Time `json:"date_to"`
	DocumentCount int64     `json:"document_count"`
	OriginalBytes int64     `json:"original_size_bytes"`
	CreatedAt     time.Time `json:"created_at"`
}

type StreamArchiveRequest struct {
	Client        *elasticsearch.Client
	Context       context.Context
	Indices       []string
	Query         map[string]any
	ProductID     int
	EnvironmentID int
	CategoryID    *int
	FeatureID     *int
	SubFeatureID  *int
	DateFrom      time.Time
	DateTo        time.Time
	FilePath      string
}

type StreamArchiveResult struct {
	DocumentCount   int64
	OriginalBytes   int64
	CompressedBytes int64
	Checksum        string
	FilePath        string
	Manifest        ArchiveManifest
	Empty           bool
}

type streamHit struct {
	ID     string         `json:"_id"`
	Index  string         `json:"_index"`
	Source map[string]any `json:"_source"`
}

type streamPage struct {
	ScrollID string `json:"_scroll_id"`
	Hits     struct {
		Hits []streamHit `json:"hits"`
	} `json:"hits"`
}

// BuildDailyArchivePath is the only physical layout used by the retention
// backend. Higher-level views group these daily records from PostgreSQL.
func BuildDailyArchivePath(root string, productID, environmentID int, day time.Time) string {
	return BuildScopedDailyArchivePath(root, productID, environmentID, nil, nil, nil, day)
}

// BuildScopedDailyArchivePath keeps each feature scope in a separate
// directory so a portable archive can be copied or restored independently.
func BuildScopedDailyArchivePath(root string, productID, environmentID int, categoryID, featureID, subFeatureID *int, day time.Time) string {
	return buildScopedDailyArchivePath(root, productID, environmentID, categoryID, featureID, subFeatureID, day, "")
}

// BuildScopedDailyArchiveVersionPath gives every archive generation its own
// immutable file. This is required when logs are restored and archived again:
// the previous verified GZIP must remain untouched until the replacement has
// been written and verified successfully.
func BuildScopedDailyArchiveVersionPath(root string, productID, environmentID int, categoryID, featureID, subFeatureID *int, day time.Time, archiveID string) string {
	return buildScopedDailyArchivePath(root, productID, environmentID, categoryID, featureID, subFeatureID, day, strings.TrimSpace(archiveID))
}

func buildScopedDailyArchivePath(root string, productID, environmentID int, categoryID, featureID, subFeatureID *int, day time.Time, archiveID string) string {
	root = filepath.Clean(root)
	category := "all"
	if categoryID != nil {
		category = strconv.Itoa(*categoryID)
	}
	feature := "all"
	if featureID != nil {
		feature = strconv.Itoa(*featureID)
	}
	subFeature := "all"
	if subFeatureID != nil {
		subFeature = strconv.Itoa(*subFeatureID)
	}
	filename := day.UTC().Format("02.ndjson.gz")
	if archiveID != "" {
		filename = fmt.Sprintf("%s-%s.ndjson.gz", day.UTC().Format("02"), archiveID)
	}
	return filepath.Join(root, "archives", fmt.Sprintf("product-%d", productID), fmt.Sprintf("environment-%d", environmentID), "category-"+category, "feature-"+feature, "sub-feature-"+subFeature, day.UTC().Format("2006"), day.UTC().Format("01"), filename)
}

func StreamArchive(req StreamArchiveRequest) (StreamArchiveResult, error) {
	if req.Client == nil || req.Context == nil {
		return StreamArchiveResult{}, fmt.Errorf("archive client and context are required")
	}
	if req.ProductID <= 0 || req.EnvironmentID <= 0 {
		return StreamArchiveResult{}, fmt.Errorf("product_id and environment_id are required")
	}
	if len(req.Indices) == 0 {
		return StreamArchiveResult{}, fmt.Errorf("no Elasticsearch indices matched archive range")
	}
	if req.FilePath == "" {
		return StreamArchiveResult{}, fmt.Errorf("archive file path is required")
	}
	if !hasArchiveScopeGuard(req.Query, req.ProductID, req.EnvironmentID) {
		return StreamArchiveResult{}, fmt.Errorf("archive query must constrain product_id and environment_id")
	}

	if err := os.MkdirAll(filepath.Dir(req.FilePath), 0755); err != nil {
		return StreamArchiveResult{}, err
	}
	file, err := os.Create(req.FilePath)
	if err != nil {
		return StreamArchiveResult{}, err
	}
	removeOnError := true
	defer func() {
		_ = file.Close()
		if removeOnError {
			_ = os.Remove(req.FilePath)
		}
	}()
	gzipWriter := gzip.NewWriter(file)
	manifest := ArchiveManifest{
		Type: "manifest", Version: 1, ProductID: req.ProductID,
		EnvironmentID: req.EnvironmentID, DateFrom: req.DateFrom.UTC(),
		DateTo: req.DateTo.UTC(), CategoryID: req.CategoryID, FeatureID: req.FeatureID,
		SubFeatureID: req.SubFeatureID, CreatedAt: time.Now().UTC(),
	}
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		_ = gzipWriter.Close()
		return StreamArchiveResult{}, err
	}
	if _, err := gzipWriter.Write(append(manifestBytes, '\n')); err != nil {
		_ = gzipWriter.Close()
		return StreamArchiveResult{}, err
	}

	var scrollID string
	defer func() {
		if scrollID == "" {
			return
		}
		clearCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		res, clearErr := req.Client.ClearScroll(req.Client.ClearScroll.WithContext(clearCtx), req.Client.ClearScroll.WithScrollID(scrollID))
		if clearErr == nil {
			_ = res.Body.Close()
		}
	}()

	query := make(map[string]any, len(req.Query)+2)
	for key, value := range req.Query {
		query[key] = value
	}
	query["size"] = archivePageSize
	// _doc is supported by Elasticsearch/OpenSearch scroll contexts across
	// the versions used by Omnilogs. _shard_doc is intended primarily for
	// search_after/PIT and can produce a 400 when used with scroll.
	query["sort"] = []string{"_doc"}
	queryBytes, err := json.Marshal(query)
	if err != nil {
		return StreamArchiveResult{}, err
	}
	response, err := req.Client.Search(
		req.Client.Search.WithContext(req.Context),
		req.Client.Search.WithIndex(req.Indices...),
		req.Client.Search.WithBody(bytes.NewReader(queryBytes)),
		req.Client.Search.WithScroll(archiveScrollTTL),
	)
	if err != nil {
		return StreamArchiveResult{}, err
	}
	if response.StatusCode == http.StatusNotFound {
		_ = response.Body.Close()
	} else if response.IsError() {
		status := response.Status()
		body, _ := io.ReadAll(response.Body)
		_ = response.Body.Close()
		return StreamArchiveResult{}, fmt.Errorf("elasticsearch archive search returned status %s: %s", status, elasticErrorReason(body))
	} else {
		for {
			var page streamPage
			decodeErr := json.NewDecoder(response.Body).Decode(&page)
			_ = response.Body.Close()
			if decodeErr != nil {
				return StreamArchiveResult{}, decodeErr
			}
			scrollID = page.ScrollID
			for _, hit := range page.Hits.Hits {
				line := map[string]any{"_type": "log", "_id": hit.ID, "_index": hit.Index, "_source": hit.Source}
				lineBytes, marshalErr := json.Marshal(line)
				if marshalErr != nil {
					return StreamArchiveResult{}, marshalErr
				}
				lineBytes = append(lineBytes, '\n')
				if _, writeErr := gzipWriter.Write(lineBytes); writeErr != nil {
					return StreamArchiveResult{}, writeErr
				}
				manifest.DocumentCount++
				manifest.OriginalBytes += int64(len(lineBytes))
			}
			if len(page.Hits.Hits) == 0 {
				break
			}
			response, err = req.Client.Scroll(req.Client.Scroll.WithContext(req.Context), req.Client.Scroll.WithScrollID(scrollID), req.Client.Scroll.WithScroll(archiveScrollTTL))
			if err != nil {
				return StreamArchiveResult{}, err
			}
			if response.IsError() {
				status := response.Status()
				_ = response.Body.Close()
				return StreamArchiveResult{}, fmt.Errorf("elasticsearch archive scroll returned status %s", status)
			}
		}
	}

	// Never leave a manifest-only export behind. This check lives in the
	// writer, so every caller (including legacy archive paths) gets the same
	// protection even if it did not perform a preflight count.
	if manifest.DocumentCount == 0 {
		if err := gzipWriter.Close(); err != nil {
			return StreamArchiveResult{}, err
		}
		if err := file.Close(); err != nil {
			return StreamArchiveResult{}, err
		}
		if err := os.Remove(req.FilePath); err != nil && !os.IsNotExist(err) {
			return StreamArchiveResult{}, err
		}
		removeOnError = false
		return StreamArchiveResult{Manifest: manifest, Empty: true}, nil
	}

	// Append a final manifest trailer with the exact count. The first manifest
	// lets readers validate the scope before reading logs; the trailer avoids
	// buffering or rewriting a potentially very large archive.
	trailer, err := json.Marshal(manifest)
	if err != nil {
		return StreamArchiveResult{}, err
	}
	if _, err := gzipWriter.Write(append(trailer, '\n')); err != nil {
		return StreamArchiveResult{}, err
	}
	if err := gzipWriter.Close(); err != nil {
		return StreamArchiveResult{}, err
	}
	if err := file.Close(); err != nil {
		return StreamArchiveResult{}, err
	}
	checksum, err := ChecksumFile(req.FilePath)
	if err != nil {
		return StreamArchiveResult{}, err
	}
	info, err := os.Stat(req.FilePath)
	if err != nil {
		return StreamArchiveResult{}, err
	}
	removeOnError = false
	return StreamArchiveResult{
		DocumentCount: manifest.DocumentCount, OriginalBytes: manifest.OriginalBytes,
		CompressedBytes: info.Size(), Checksum: checksum, FilePath: req.FilePath,
		Manifest: manifest,
	}, nil
}

func elasticErrorReason(body []byte) string {
	var payload struct {
		Error struct {
			Reason string `json:"reason"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &payload); err == nil && strings.TrimSpace(payload.Error.Reason) != "" {
		return strings.TrimSpace(payload.Error.Reason)
	}
	return strings.TrimSpace(string(body))
}

func hasArchiveScopeGuard(query map[string]any, productID, environmentID int) bool {
	encoded, err := json.Marshal(query)
	if err != nil {
		return false
	}
	text := string(encoded)
	return bytes.Contains([]byte(text), []byte(`"product_id"`)) && bytes.Contains([]byte(text), []byte(`"environment_id"`)) && productID > 0 && environmentID > 0
}

func ChecksumFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func VerifyArchive(path string, expectedProductID, expectedEnvironmentID int) (ArchiveManifest, int64, error) {
	return VerifyArchiveScope(path, expectedProductID, expectedEnvironmentID, nil, nil, nil)
}

func VerifyArchiveScope(path string, expectedProductID, expectedEnvironmentID int, expectedCategoryID, expectedFeatureID, expectedSubFeatureID *int) (ArchiveManifest, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return ArchiveManifest{}, 0, err
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		return ArchiveManifest{}, 0, err
	}
	defer gz.Close()
	decoder := json.NewDecoder(gz)
	var manifest ArchiveManifest
	if err := decoder.Decode(&manifest); err != nil {
		return ArchiveManifest{}, 0, fmt.Errorf("invalid archive manifest: %w", err)
	}
	if manifest.Type != "manifest" || manifest.Version != 1 || manifest.ProductID != expectedProductID || manifest.EnvironmentID != expectedEnvironmentID ||
		!optionalIntMatches(manifest.CategoryID, expectedCategoryID) || !optionalIntMatches(manifest.FeatureID, expectedFeatureID) || !optionalIntMatches(manifest.SubFeatureID, expectedSubFeatureID) {
		return ArchiveManifest{}, 0, fmt.Errorf("archive manifest scope is invalid")
	}
	var count int64
	finalManifest := manifest
	for {
		var line struct {
			Type          string         `json:"_type"`
			ID            string         `json:"_id"`
			Source        map[string]any `json:"_source"`
			DocumentCount int64          `json:"document_count"`
			OriginalBytes int64          `json:"original_size_bytes"`
		}
		err := decoder.Decode(&line)
		if err == io.EOF {
			break
		}
		if err != nil {
			return ArchiveManifest{}, 0, fmt.Errorf("invalid archive log line: %w", err)
		}
		if line.Type == "manifest" {
			finalManifest.DocumentCount = line.DocumentCount
			finalManifest.OriginalBytes = line.OriginalBytes
			continue
		}
		if line.Type != "log" || line.ID == "" || line.Source == nil {
			return ArchiveManifest{}, 0, fmt.Errorf("invalid archive log record")
		}
		count++
	}
	if count != finalManifest.DocumentCount {
		return ArchiveManifest{}, 0, fmt.Errorf("archive manifest count mismatch")
	}
	return finalManifest, count, nil
}

func optionalIntMatches(actual, expected *int) bool {
	if expected == nil {
		return true
	}
	return actual != nil && *actual == *expected
}
