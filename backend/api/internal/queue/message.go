package queue

import (
	"encoding/json"
	"hash/fnv"
	"strconv"
	"time"
)

type LogMessage struct {
	LogID               string          `json:"log_id"`
	QueueItemID         int64           `json:"queue_item_id"`
	BatchID             string          `json:"batch_id"`
	ProductID           *int            `json:"product_id,omitempty"`
	SourceID            *int            `json:"source_id,omitempty"`
	EnvironmentID       *int            `json:"environment_id,omitempty"`
	QueueKey            string          `json:"queue_key"`
	SourceType          string          `json:"source_type"`
	SourcePlatform      string          `json:"source_platform"`
	IdempotencyKey      *string         `json:"idempotency_key,omitempty"`
	DetectedProductCode *string         `json:"detected_product_code,omitempty"`
	Priority            int             `json:"priority"`
	SequenceNo          int             `json:"sequence_no"`
	InputPayload        json.RawMessage `json:"input_payload"`
	RetentionUntil      *time.Time      `json:"retention_until,omitempty"`
	RetryCount          int             `json:"retry_count"`
	MaxRetryCount       int             `json:"max_retry_count"`
	PublishedAt         time.Time       `json:"published_at"`
}

func SyntheticQueueItemID(batchID string, sequenceNo int) int64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(batchID))
	_, _ = hasher.Write([]byte{':'})
	_, _ = hasher.Write([]byte(strconv.Itoa(sequenceNo)))
	return int64(hasher.Sum64() & 0x7fffffffffffffff)
}

func MarshalMessage(message LogMessage) ([]byte, error) {
	return json.Marshal(message)
}
