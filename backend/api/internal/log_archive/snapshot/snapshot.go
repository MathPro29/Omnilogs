package snapshot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

// Config controls the Elasticsearch repository used as the primary archive.
// The repository itself must already be registered in Elasticsearch. This
// service deliberately does not create a repository from application input,
// because repository credentials and storage settings belong to Elasticsearch.
type Config struct {
	Enabled    bool
	Repository string
	Timeout    time.Duration
}

func ConfigFromEnv() Config {
	enabled, _ := strconv.ParseBool(strings.TrimSpace(os.Getenv("ELASTICSEARCH_SNAPSHOT_ENABLED")))
	timeout := 30 * time.Minute
	if raw := strings.TrimSpace(os.Getenv("ELASTICSEARCH_SNAPSHOT_TIMEOUT_SECONDS")); raw != "" {
		if seconds, err := strconv.Atoi(raw); err == nil && seconds > 0 {
			timeout = time.Duration(seconds) * time.Second
		}
	}
	return Config{
		Enabled:    enabled,
		Repository: strings.TrimSpace(os.Getenv("ELASTICSEARCH_SNAPSHOT_REPOSITORY")),
		Timeout:    timeout,
	}
}

type Result struct {
	Name    string
	UUID    string
	State   string
	Indices []string
}

type Client struct {
	es     *elasticsearch.Client
	config Config
}

func New(es *elasticsearch.Client, config Config) *Client {
	return &Client{es: es, config: config}
}

func (c *Client) Enabled() bool {
	return c != nil && c.config.Enabled
}

func (c *Client) Repository() string {
	if c == nil {
		return ""
	}
	return c.config.Repository
}

func (c *Client) Create(ctx context.Context, name string, indices []string) (*Result, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("snapshot name is required")
	}
	if len(indices) == 0 {
		return nil, errors.New("at least one Elasticsearch index is required for snapshot")
	}

	body, err := json.Marshal(map[string]any{
		"indices":              strings.Join(indices, ","),
		"include_global_state": false,
	})
	if err != nil {
		return nil, err
	}
	requestCtx, cancel := context.WithTimeout(ctx, c.timeout())
	defer cancel()
	// Retention must not mark an archive ready for verification/deletion while
	// the snapshot is still running. Keep this synchronous for the lifecycle.
	wait := true
	response, err := c.es.Snapshot.Create(
		c.config.Repository,
		name,
		c.es.Snapshot.Create.WithContext(requestCtx),
		c.es.Snapshot.Create.WithBody(bytes.NewReader(body)),
		c.es.Snapshot.Create.WithWaitForCompletion(wait),
	)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	responseBody, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return nil, readErr
	}
	if response.IsError() {
		return nil, elasticError("create snapshot", response.StatusCode, responseBody)
	}

	var payload struct {
		Snapshot struct {
			Name    string   `json:"snapshot"`
			UUID    string   `json:"uuid"`
			State   string   `json:"state"`
			Indices []string `json:"indices"`
		} `json:"snapshot"`
	}
	if err := json.Unmarshal(responseBody, &payload); err != nil {
		return nil, fmt.Errorf("decode create snapshot response: %w", err)
	}
	result := &Result{
		Name:    firstNonEmpty(payload.Snapshot.Name, name),
		UUID:    payload.Snapshot.UUID,
		State:   strings.ToUpper(payload.Snapshot.State),
		Indices: payload.Snapshot.Indices,
	}
	if !wait {
		return result, nil
	}
	if result.State != "SUCCESS" {
		return nil, fmt.Errorf("snapshot %s completed with state %s", result.Name, result.State)
	}
	return result, nil
}

func (c *Client) Verify(ctx context.Context, name string) (*Result, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("snapshot name is required")
	}
	requestCtx, cancel := context.WithTimeout(ctx, c.timeout())
	defer cancel()
	verifyResponse, err := c.es.Snapshot.VerifyRepository(
		c.config.Repository,
		c.es.Snapshot.VerifyRepository.WithContext(requestCtx),
	)
	if err != nil {
		return nil, err
	}
	defer verifyResponse.Body.Close()
	verifyBody, readErr := io.ReadAll(verifyResponse.Body)
	if readErr != nil {
		return nil, readErr
	}
	if verifyResponse.IsError() {
		return nil, elasticError("verify snapshot repository", verifyResponse.StatusCode, verifyBody)
	}

	snapshotResponse, err := c.es.Snapshot.Get(
		c.config.Repository,
		[]string{name},
		c.es.Snapshot.Get.WithContext(requestCtx),
	)
	if err != nil {
		return nil, err
	}
	defer snapshotResponse.Body.Close()
	snapshotBody, readErr := io.ReadAll(snapshotResponse.Body)
	if readErr != nil {
		return nil, readErr
	}
	if snapshotResponse.IsError() {
		return nil, elasticError("get snapshot", snapshotResponse.StatusCode, snapshotBody)
	}
	var payload struct {
		Snapshots []struct {
			Name    string   `json:"snapshot"`
			UUID    string   `json:"uuid"`
			State   string   `json:"state"`
			Indices []string `json:"indices"`
		} `json:"snapshots"`
	}
	if err := json.Unmarshal(snapshotBody, &payload); err != nil {
		return nil, fmt.Errorf("decode get snapshot response: %w", err)
	}
	if len(payload.Snapshots) == 0 {
		return nil, fmt.Errorf("snapshot %s was not found", name)
	}
	value := payload.Snapshots[0]
	result := &Result{Name: value.Name, UUID: value.UUID, State: strings.ToUpper(value.State), Indices: value.Indices}
	if result.State != "SUCCESS" {
		return nil, fmt.Errorf("snapshot %s is not restorable: state %s", result.Name, result.State)
	}
	return result, nil
}

func (c *Client) Restore(ctx context.Context, name string, indices []string, prefix string) (string, error) {
	if err := c.validate(); err != nil {
		return "", err
	}
	if strings.TrimSpace(name) == "" {
		return "", errors.New("snapshot name is required")
	}
	if len(indices) == 0 {
		return "", errors.New("snapshot has no indices to restore")
	}
	prefix = strings.Trim(strings.TrimSpace(prefix), "-")
	if prefix == "" {
		return "", errors.New("restore index prefix is required")
	}
	replacement := prefix + "-$1"
	body, err := json.Marshal(map[string]any{
		"indices":              strings.Join(indices, ","),
		"ignore_unavailable":   false,
		"include_global_state": false,
		"include_aliases":      false,
		"rename_pattern":       "(.+)",
		"rename_replacement":   replacement,
		"index_settings": map[string]any{
			"index.blocks.write": true,
		},
	})
	if err != nil {
		return "", err
	}
	requestCtx, cancel := context.WithTimeout(ctx, c.timeout())
	defer cancel()
	wait := true
	response, err := c.es.Snapshot.Restore(
		c.config.Repository,
		name,
		c.es.Snapshot.Restore.WithContext(requestCtx),
		c.es.Snapshot.Restore.WithBody(bytes.NewReader(body)),
		c.es.Snapshot.Restore.WithWaitForCompletion(wait),
	)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	responseBody, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return "", readErr
	}
	if response.IsError() {
		return "", elasticError("restore snapshot", response.StatusCode, responseBody)
	}
	return prefix + "-*", nil
}

func (c *Client) validate() error {
	if c == nil || !c.config.Enabled {
		return errors.New("Elasticsearch snapshot archive is disabled")
	}
	if c.es == nil {
		return errors.New("Elasticsearch snapshot client is not configured")
	}
	if strings.TrimSpace(c.config.Repository) == "" {
		return errors.New("ELASTICSEARCH_SNAPSHOT_REPOSITORY is required when snapshot archive is enabled")
	}
	return nil
}

func (c *Client) timeout() time.Duration {
	if c.config.Timeout > 0 {
		return c.config.Timeout
	}
	return 30 * time.Minute
}

func elasticError(operation string, status int, body []byte) error {
	var payload struct {
		Error any `json:"error"`
	}
	if err := json.Unmarshal(body, &payload); err == nil && payload.Error != nil {
		encoded, _ := json.Marshal(payload.Error)
		return fmt.Errorf("%s returned status %d: %s", operation, status, string(encoded))
	}
	return fmt.Errorf("%s returned status %d: %s", operation, status, strings.TrimSpace(string(body)))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
