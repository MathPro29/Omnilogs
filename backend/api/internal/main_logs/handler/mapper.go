package handler

func mapESDocToMainLogDocument(source map[string]any) map[string]any {
	result := map[string]any{
		"log_id": source["log_id"],
		"raw":    source,
	}

	result["product_id"] = getInt64(source["product_id"])
	if env, ok := source["environment_id"]; ok {
		result["environment_id"] = getInt64(env)
	}
	if src, ok := source["source_id"]; ok {
		result["source_id"] = getInt64(src)
	}
	if ts, ok := source["@timestamp"]; ok {
		result["timestamp"] = ts
	}

	payload, _ := source["payload"].(map[string]any)
	if payload != nil {
		if lvl := payload["log_level"]; lvl != nil && lvl != "" {
			result["level"] = lvl
		} else if lvl := payload["level"]; lvl != nil && lvl != "" {
			result["level"] = lvl
		} else if lvl := payload["severity"]; lvl != nil && lvl != "" {
			result["level"] = lvl
		}

		result["log_type"] = payload["event_type"]
		result["message"] = payload["message"]
		result["request_id"] = payload["source_request_id"]
		result["trace_id"] = payload["trace_id"]
		result["method"] = payload["request_method"]

		if p := payload["request_path"]; p != nil && p != "" {
			result["path"] = p
		} else if p := payload["path"]; p != nil && p != "" {
			result["path"] = p
		} else if p := payload["url"]; p != nil && p != "" {
			result["path"] = p
		} else if p := payload["endpoint"]; p != nil && p != "" {
			result["path"] = p
		}

		result["url"] = payload["url"]
		if sc, ok := payload["status_code"]; ok {
			result["status_code"] = getInt64(sc)
		}

		if lat, ok := payload["duration_ms"]; ok {
			result["latency_ms"] = getInt64(lat)
		} else if lat, ok := payload["latency_ms"]; ok {
			result["latency_ms"] = getInt64(lat)
		} else if lat, ok := payload["duration"]; ok {
			result["latency_ms"] = getInt64(lat)
		}

		result["error_code"] = payload["error_code"]
		result["error_message"] = payload["error_message"]
		result["stack_trace"] = payload["stack_trace"]
		result["request_headers"] = payload["request_headers"]
		result["response_headers"] = payload["response_headers"]
		result["request_payload"] = payload["request_payload"]
		result["response_payload"] = payload["response_payload"]
		result["custom_fields"] = payload["custom_fields"]
		result["actor"] = payload["actor"]
	}

	return result
}

func getInt64(val any) int64 {
	switch v := val.(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	default:
		return 0
	}
}
