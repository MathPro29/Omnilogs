package searchindex

// ControlledLogMapping is shared by normal ingestion and archive rollback.
// Flexible payload fields stay in _source while flattened fields remain
// searchable without creating an unbounded Elasticsearch mapping.
const ControlledLogMapping = `{
  "dynamic": false,
  "properties": {
    "@timestamp": {"type": "date"},
    "received_at": {"type": "date"},
    "product_id": {"type": "integer"},
    "environment_id": {"type": "integer"},
    "product_code": {"type": "keyword"},
    "environment_code": {"type": "keyword"},
    "source_id": {"type": "integer"},
    "source_project_id": {"type": "integer"},
    "project_id": {"type": "integer"},
    "category_id": {"type": "integer"},
    "feature_id": {"type": "integer"},
    "sub_feature_id": {"type": "integer"},
    "project_code": {"type": "keyword"},
    "category_code": {"type": "keyword"},
    "feature_code": {"type": "keyword"},
    "sub_feature_code": {"type": "keyword"},
    "feature_path_ids": {"type": "keyword"},
    "feature_full_path": {"type": "keyword", "ignore_above": 1024},
    "level": {"type": "keyword"},
    "message": {"type": "text"},
    "service": {"type": "keyword"},
    "route_key": {"type": "keyword"},
    "event_name": {"type": "keyword"},
    "trace_id": {"type": "keyword"},
    "request_id": {"type": "keyword"},
    "method": {"type": "keyword"},
    "path": {"type": "keyword"},
    "status_code": {"type": "integer"},
    "duration_ms": {"type": "long"},
    "routing_status": {"type": "keyword"},
    "routing_method": {"type": "keyword"},
    "mapping_source": {"type": "keyword"},
    "mapping_status": {"type": "keyword"},
    "mapping_reason": {"type": "keyword", "ignore_above": 1024},
    "metadata": {"type": "flattened"},
    "custom_fields": {"type": "flattened"},
    "data": {"type": "flattened"},
    "payload": {"type": "object", "enabled": false},
    "search_text": {"type": "text"}
  }
}`
