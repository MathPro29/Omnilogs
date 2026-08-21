package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"omnilogs-api/dto"
	"omnilogs-api/models"
)

type environmentHint struct {
	ID   *int
	Code string
}

type commonRoutingCondition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
	Min      any    `json:"min,omitempty"`
	Max      any    `json:"max,omitempty"`
}

func (u *usecase) resolveEnvironmentScope(ctx context.Context, req *dto.IngestLogBatchRequest) error {
	productID, err := u.resolveProductID(ctx, req)
	if err != nil {
		return err
	}
	candidates := map[int]string{}
	add := func(id int, source string) {
		if id > 0 {
			candidates[id] = source
		}
	}
	if req.EnvironmentID != nil {
		add(*req.EnvironmentID, "request.environment_id")
	}
	if strings.TrimSpace(req.EnvironmentCode) != "" {
		id, err := u.environmentIDByCode(ctx, productID, req.EnvironmentCode)
		if err != nil {
			return err
		}
		add(id, "request.environment_code")
	}

	for i := range req.Logs {
		var payload map[string]any
		if err := json.Unmarshal(req.Logs[i].InputPayload, &payload); err != nil {
			return fmt.Errorf("%w: logs[%d] payload is invalid", ErrInvalidScope, i)
		}
		for _, hint := range payloadEnvironmentHints(payload) {
			if hint.ID != nil {
				add(*hint.ID, fmt.Sprintf("logs[%d].environment_id", i))
			}
			if hint.Code != "" {
				id, err := u.environmentIDByCode(ctx, productID, hint.Code)
				if err != nil {
					return fmt.Errorf("logs[%d]: %w", i, err)
				}
				add(id, fmt.Sprintf("logs[%d].environment_code", i))
			}
		}
	}

	if req.APIKeyID != nil {
		var key models.ProductAPIKey
		if err := u.repo.DB().WithContext(ctx).
			Where("key_id = ? AND product_id = ? AND is_active = TRUE", *req.APIKeyID, productID).
			First(&key).Error; err != nil {
			return fmt.Errorf("%w: API key does not belong to product", ErrInvalidScope)
		}
		if key.EnvironmentID != nil {
			add(*key.EnvironmentID, "api_key.environment_id")
		}
	}
	if req.SourceID != nil {
		var source models.LogSource
		if err := u.repo.DB().WithContext(ctx).
			Where("source_id = ? AND product_id = ? AND is_active = TRUE", *req.SourceID, productID).
			First(&source).Error; err != nil {
			return fmt.Errorf("%w: source does not belong to product", ErrInvalidScope)
		}
		if source.EnvironmentID != nil {
			add(*source.EnvironmentID, "source.environment_id")
		}
	}

	if len(candidates) > 1 {
		// Environment is optional for ingestion. An unscoped API key/source and a
		// payload without an environment hint must still be accepted; the batch is
		// kept at product scope and can be routed later by the worker. Requiring a
		// unique environment here made otherwise valid legacy log requests fail
		// before they reached the queue.
		if len(candidates) == 0 {
			return nil
		}

		return fmt.Errorf("%w: ENVIRONMENT_MISMATCH: explicit/API-key/source scopes resolved to different environments", ErrInvalidScope)
	}

	environmentID, err := uniqueEnvironmentCandidate(candidates)
	if err != nil {
		return err
	}
	var count int64
	if err := u.repo.DB().WithContext(ctx).Model(&models.ProductEnvironment{}).
		Where("environment_id = ? AND product_id = ? AND is_active = TRUE AND deleted_at IS NULL", environmentID, productID).
		Count(&count).Error; err != nil {
		return fmt.Errorf("validate environment: %w", err)
	}
	if count != 1 {
		return fmt.Errorf("%w: ENVIRONMENT_NOT_ALLOWED: environment %d does not belong to active product %d", ErrInvalidScope, environmentID, productID)
	}
	req.EnvironmentID = &environmentID
	return nil
}

func uniqueEnvironmentCandidate(candidates map[int]string) (int, error) {
	if len(candidates) == 0 {
		return 0, fmt.Errorf("%w: provide environment_id/environment_code or configure one unique common mapping", ErrUnmappedScope)
	}
	if len(candidates) > 1 {
		ids := make([]string, 0, len(candidates))
		for id, source := range candidates {
			ids = append(ids, fmt.Sprintf("%d (%s)", id, source))
		}
		return 0, fmt.Errorf("%w: matched %s", ErrAmbiguousScope, strings.Join(ids, ", "))
	}
	for id := range candidates {
		return id, nil
	}
	return 0, ErrUnmappedScope
}

func (u *usecase) environmentIDByCode(ctx context.Context, productID int, code string) (int, error) {
	var values []models.ProductEnvironment
	if err := u.repo.DB().WithContext(ctx).
		Where("product_id = ? AND is_active = TRUE AND deleted_at IS NULL AND LOWER(BTRIM(environment_code)) = LOWER(BTRIM(?))", productID, code).
		Limit(2).Find(&values).Error; err != nil {
		return 0, fmt.Errorf("resolve environment code: %w", err)
	}
	if len(values) == 0 {
		return 0, fmt.Errorf("%w: ENVIRONMENT_NOT_FOUND: environment_code %q does not belong to product", ErrInvalidScope, code)
	}
	if len(values) > 1 {
		return 0, fmt.Errorf("%w: environment_code %q matched multiple environments", ErrAmbiguousScope, code)
	}
	return values[0].EnvironmentID, nil
}

func payloadEnvironmentHints(payload map[string]any) []environmentHint {
	values := []map[string]any{payload}
	if customFields, ok := payload["custom_fields"].(map[string]any); ok {
		values = append(values, customFields)
	}
	if customFields, ok := payload["customFields"].(map[string]any); ok {
		values = append(values, customFields)
	}
	if metadata, ok := payload["metadata"].(map[string]any); ok {
		values = append(values, metadata)
	}
	if routing, ok := payload["routing"].(map[string]any); ok {
		for _, key := range []string{"explicit_hierarchy", "explicitHierarchy"} {
			if explicit, ok := routing[key].(map[string]any); ok {
				values = append(values, explicit)
			}
		}
	}
	hints := make([]environmentHint, 0, len(values))
	for _, value := range values {
		hint := environmentHint{}
		for _, key := range []string{"environment_id", "environmentId"} {
			if id := intFromScopeValue(value[key]); id != nil {
				hint.ID = id
				break
			}
		}
		for _, key := range []string{"environment_code", "environmentCode", "environment"} {
			if code, ok := value[key].(string); ok && strings.TrimSpace(code) != "" {
				hint.Code = strings.TrimSpace(code)
				break
			}
		}
		if hint.ID != nil || hint.Code != "" {
			hints = append(hints, hint)
		}
	}
	return hints
}

func intFromScopeValue(value any) *int {
	var result int
	switch typed := value.(type) {
	case float64:
		result = int(typed)
	case int:
		result = typed
	case json.Number:
		parsed, err := typed.Int64()
		if err == nil {
			result = int(parsed)
		}
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		if err == nil {
			result = parsed
		}
	}
	if result <= 0 {
		return nil
	}
	return &result
}

func (u *usecase) commonEnvironmentCandidates(ctx context.Context, productID int, sourceID *int, logs []dto.IngestLogItemRequest) (map[int]struct{}, error) {
	result := map[int]struct{}{}
	for i := range logs {
		var payload map[string]any
		if err := json.Unmarshal(logs[i].InputPayload, &payload); err != nil {
			return nil, fmt.Errorf("%w: logs[%d] payload is invalid", ErrInvalidScope, i)
		}
		normalizeCommonHTTPFields(payload)
		if routeKey := commonString(payload, "route_key"); routeKey != "" {
			var routes []models.LogRoute
			query := u.repo.DB().WithContext(ctx).
				Where("product_id = ? AND is_active = TRUE AND LOWER(BTRIM(route_key)) = LOWER(BTRIM(?))", productID, routeKey)
			if sourceID != nil {
				query = query.Where("source_id IS NULL OR source_id = ?", *sourceID)
			} else {
				query = query.Where("source_id IS NULL")
			}
			if err := query.Find(&routes).Error; err != nil {
				return nil, fmt.Errorf("load common route mappings: %w", err)
			}
			for _, route := range routes {
				result[route.EnvironmentID] = struct{}{}
			}
		}

		var rules []models.LogRoutingRule
		query := u.repo.DB().WithContext(ctx).
			Where("product_id = ? AND environment_id IS NOT NULL AND is_active = TRUE", productID)
		if sourceID != nil {
			query = query.Where("source_id IS NULL OR source_id = ?", *sourceID)
		} else {
			query = query.Where("source_id IS NULL")
		}
		if err := query.Find(&rules).Error; err != nil {
			return nil, fmt.Errorf("load common HTTP mappings: %w", err)
		}
		for _, rule := range rules {
			matched, err := commonRuleMatches(rule.Conditions, payload)
			if err != nil {
				return nil, fmt.Errorf("invalid common mapping rule %d: %w", rule.RuleID, err)
			}
			if matched && rule.EnvironmentID != nil {
				result[*rule.EnvironmentID] = struct{}{}
			}
		}

		if service := commonString(payload, "service", "source"); service != "" {
			var sources []models.LogSource
			if err := u.repo.DB().WithContext(ctx).
				Where("product_id = ? AND environment_id IS NOT NULL AND is_active = TRUE AND (LOWER(BTRIM(source_code)) = LOWER(BTRIM(?)) OR LOWER(BTRIM(source_name)) = LOWER(BTRIM(?)))", productID, service, service).
				Find(&sources).Error; err != nil {
				return nil, fmt.Errorf("load common service mappings: %w", err)
			}
			for _, source := range sources {
				if source.EnvironmentID != nil {
					result[*source.EnvironmentID] = struct{}{}
				}
			}
		}
	}
	return result, nil
}

func normalizeCommonHTTPFields(payload map[string]any) {
	if commonString(payload, "request_method") == "" {
		payload["request_method"] = firstNestedString(payload, []string{"method"}, []string{"request", "method"}, []string{"http", "method"})
	}
	if commonString(payload, "request_path") == "" {
		payload["request_path"] = firstNestedString(payload, []string{"path"}, []string{"endpoint"}, []string{"request", "path"}, []string{"http", "path"})
	}
	if commonString(payload, "service") == "" {
		payload["service"] = firstNestedString(payload, []string{"service_name"}, []string{"source"})
	}
}

func firstNestedString(payload map[string]any, paths ...[]string) string {
	for _, path := range paths {
		current := any(payload)
		for _, part := range path {
			object, ok := current.(map[string]any)
			if !ok {
				current = nil
				break
			}
			current = object[part]
		}
		if value, ok := current.(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func commonString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := payload[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func commonRuleMatches(raw json.RawMessage, payload map[string]any) (bool, error) {
	var conditions []commonRoutingCondition
	if err := json.Unmarshal(raw, &conditions); err != nil {
		var one commonRoutingCondition
		if oneErr := json.Unmarshal(raw, &one); oneErr != nil {
			return false, err
		}
		conditions = []commonRoutingCondition{one}
	}
	if len(conditions) == 0 {
		return false, nil
	}
	for _, condition := range conditions {
		actual, ok := commonPathValue(payload, condition.Field)
		if !ok || !commonConditionMatches(actual, condition) {
			return false, nil
		}
	}
	return true, nil
}

func commonPathValue(payload map[string]any, path string) (any, bool) {
	current := any(payload)
	for _, part := range strings.Split(strings.TrimSpace(path), ".") {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = object[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func commonConditionMatches(actual any, condition commonRoutingCondition) bool {
	operator := strings.ToLower(strings.TrimSpace(condition.Operator))
	if operator == "" || operator == "equals" {
		return reflect.DeepEqual(commonComparable(actual), commonComparable(condition.Value))
	}
	actualText, expectedText := fmt.Sprint(actual), fmt.Sprint(condition.Value)
	switch operator {
	case "not_equals":
		return !reflect.DeepEqual(commonComparable(actual), commonComparable(condition.Value))
	case "starts_with":
		return strings.HasPrefix(actualText, expectedText)
	case "ends_with":
		return strings.HasSuffix(actualText, expectedText)
	case "contains":
		return strings.Contains(actualText, expectedText)
	case "in", "not_in":
		expected := reflect.ValueOf(condition.Value)
		matched := false
		if expected.IsValid() && (expected.Kind() == reflect.Slice || expected.Kind() == reflect.Array) {
			for i := 0; i < expected.Len(); i++ {
				if reflect.DeepEqual(commonComparable(actual), commonComparable(expected.Index(i).Interface())) {
					matched = true
					break
				}
			}
		}
		return matched != (operator == "not_in")
	case "regex":
		matched, err := regexp.MatchString(expectedText, actualText)
		return err == nil && matched
	case "range":
		number, numberErr := strconv.ParseFloat(actualText, 64)
		minimum, minErr := strconv.ParseFloat(fmt.Sprint(condition.Min), 64)
		maximum, maxErr := strconv.ParseFloat(fmt.Sprint(condition.Max), 64)
		return numberErr == nil && minErr == nil && maxErr == nil && number >= minimum && number <= maximum
	default:
		return false
	}
}

func commonComparable(value any) any {
	if text, ok := value.(string); ok {
		return strings.ToLower(strings.TrimSpace(text))
	}
	return value
}

func (u *usecase) resolveProductID(ctx context.Context, req *dto.IngestLogBatchRequest) (int, error) {
	if req.ProductID != nil && *req.ProductID > 0 {
		return *req.ProductID, nil
	}

	var candidateCode string
	if req.DetectedProductCode != nil && strings.TrimSpace(*req.DetectedProductCode) != "" {
		candidateCode = strings.TrimSpace(*req.DetectedProductCode)
	}

	if candidateCode == "" {
		for i := range req.Logs {
			var payload map[string]any
			if err := json.Unmarshal(req.Logs[i].InputPayload, &payload); err == nil {
				if customFields, ok := payload["custom_fields"].(map[string]any); ok {
					for _, key := range []string{"product_code", "productCode", "product"} {
						if val, ok := customFields[key].(string); ok && strings.TrimSpace(val) != "" {
							candidateCode = strings.TrimSpace(val)
							break
						}
					}
				}
				if candidateCode == "" {
					for _, key := range []string{"product_code", "productCode", "product"} {
						if val, ok := payload[key].(string); ok && strings.TrimSpace(val) != "" {
							candidateCode = strings.TrimSpace(val)
							break
						}
					}
				}
			}
			if candidateCode != "" {
				break
			}
		}
	}

	if candidateCode == "" {
		return 0, fmt.Errorf("%w: product_id is required", ErrInvalidScope)
	}

	var product models.Product
	if err := u.repo.DB().WithContext(ctx).
		Where("is_active = TRUE AND LOWER(BTRIM(product_code)) = LOWER(BTRIM(?))", candidateCode).
		First(&product).Error; err != nil {
		return 0, fmt.Errorf("%w: PRODUCT_NOT_FOUND: product_code %q does not exist", ErrInvalidScope, candidateCode)
	}

	req.ProductID = &product.ProductID
	return product.ProductID, nil
}
