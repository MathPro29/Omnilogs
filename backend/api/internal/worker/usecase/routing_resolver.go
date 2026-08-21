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

type routingCondition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
	Min      any    `json:"min,omitempty"`
	Max      any    `json:"max,omitempty"`
}

type hierarchyCandidate struct {
	values     map[string]any
	allowCamel bool
}

func (u *usecase) resolveConfiguredRouting(ctx context.Context, envelope dto.LogMessage, payload map[string]any) error {
	if envelope.ProductID == nil || envelope.EnvironmentID == nil {
		return nil
	}
	productID, environmentID := *envelope.ProductID, *envelope.EnvironmentID
	explicitHierarchy := normalizeExplicitHierarchy(payload)
	// A feature/sub-feature path without an OmniLogs project is incomplete: it
	// still needs source-project routing to determine the project it belongs to.
	// Do not return early here, otherwise Assetwise logs bypass a configured
	// source_project_id mapping and fall through to the single-project default.
	if intPtrFromAny(payload["source_project_id"]) != nil {
		matched, err := u.resolveSourceProjectRouting(ctx, productID, environmentID, envelope.SourceID, payload)
		if err != nil {
			return err
		}
		if matched {
			return nil
		}
	}
	if explicitHierarchy && hasProjectDiscriminator(payload) {
		localProject, err := u.hasLocalProjectDiscriminator(ctx, productID, payload)
		if err != nil {
			return err
		}
		if localProject {
			// The authenticated API key fixes the product scope. resolveLogHierarchy
			// subsequently validates the project/feature against that product.
			applyRoutingMetadata(payload, "CLASSIFIED", "CLIENT_EXPLICIT")
			payload["routing_reason"] = "Used explicit client hierarchy within the API key product"
			return nil
		}
		// External project IDs must not bypass configured routing rules.
		if intPtrFromAny(payload["source_project_id"]) != nil {
			delete(payload, "project_id")
			delete(payload, "project_code")
			explicitHierarchy = false
		}
	}
	routeKey := firstString(payload, "route_key")

	if routeKey != "" {
		var route models.LogRoute
		query := u.repo.DB().WithContext(ctx).
			Where("product_id = ? AND environment_id = ? AND is_active = TRUE AND LOWER(BTRIM(route_key)) = LOWER(BTRIM(?))", productID, environmentID, routeKey).
			Where("source_id IS NULL")
		if envelope.SourceID != nil {
			query = u.repo.DB().WithContext(ctx).
				Where("product_id = ? AND environment_id = ? AND is_active = TRUE AND LOWER(BTRIM(route_key)) = LOWER(BTRIM(?))", productID, environmentID, routeKey).
				Where("source_id IS NULL OR source_id = ?", *envelope.SourceID).
				Order("CASE WHEN source_id IS NULL THEN 1 ELSE 0 END")
		}
		if err := query.Order("priority DESC, route_id").First(&route).Error; err == nil {
			applyRouteTarget(payload, route.ProjectID, route.CategoryID, "ROUTE_KEY")
			payload["matched_route_key"] = route.RouteKey
			return nil
		}
		var disabledRoute models.LogRoute
		if err := u.repo.DB().WithContext(ctx).
			Where("product_id = ? AND environment_id = ? AND is_active = FALSE AND LOWER(BTRIM(route_key)) = LOWER(BTRIM(?))", productID, environmentID, routeKey).
			First(&disabledRoute).Error; err == nil {
			applyDeletedMappingFallback(payload, "Deleted route-key mapping blocks automatic classification")
			return nil
		}
	}

	var rules []models.LogRoutingRule
	ruleQuery := u.repo.DB().WithContext(ctx).
		Where("product_id = ? AND is_active = TRUE AND (environment_id IS NULL OR environment_id = ?)", productID, environmentID).
		Where("source_id IS NULL")
	if envelope.SourceID != nil {
		ruleQuery = ruleQuery.Where("source_id IS NULL OR source_id = ?", *envelope.SourceID)
	}
	if err := ruleQuery.Order("priority DESC, rule_id").Find(&rules).Error; err != nil && !isMissingRoutingTableError(err) {
		return fmt.Errorf("load routing rules: %w", err)
	}
	matchedRules := make([]models.LogRoutingRule, 0, 2)
	matchedPriority := 0
	for i := range rules {
		matched, err := routingRuleMatches(rules[i].Conditions, payload)
		if err != nil {
			return fmt.Errorf("invalid routing rule %d: %w", rules[i].RuleID, err)
		}
		if matched {
			if len(matchedRules) == 0 {
				matchedPriority = rules[i].Priority
			}
			if rules[i].Priority != matchedPriority {
				break
			}
			matchedRules = append(matchedRules, rules[i])
		}
	}
	if len(matchedRules) > 0 {
		winner := matchedRules[0]
		conflictingIDs := make([]int, 0, len(matchedRules))
		conflict := false
		for _, rule := range matchedRules {
			conflictingIDs = append(conflictingIDs, rule.RuleID)
			if rule.TargetProjectID != winner.TargetProjectID || !sameOptionalInt(rule.TargetCategoryID, winner.TargetCategoryID) {
				conflict = true
			}
		}
		if !conflict {
			applyRouteTarget(payload, winner.TargetProjectID, winner.TargetCategoryID, "ROUTING_RULE")
			payload["matched_rule_id"] = winner.RuleID
			return nil
		}
		payload["routing_conflict_rule_ids"] = conflictingIDs
		payload["routing_reason"] = "Multiple highest-priority rules matched different hierarchy targets; used default hierarchy"
	}
	if routingMatchesDeletedRule(ctx, u, productID, environmentID, envelope.SourceID, payload) {
		applyDeletedMappingFallback(payload, "Deleted HTTP routing rule blocks automatic classification")
		return nil
	}

	if service := firstString(payload, "service", "source"); service != "" {
		var source models.LogSource
		if err := u.repo.DB().WithContext(ctx).
			Where("product_id = ? AND is_active = TRUE AND (environment_id IS NULL OR environment_id = ?) AND (LOWER(BTRIM(source_code)) = LOWER(BTRIM(?)) OR LOWER(BTRIM(source_name)) = LOWER(BTRIM(?)))", productID, environmentID, service, service).
			First(&source).Error; err == nil {
			payload["resolved_source_id"] = source.SourceID
			if source.DefaultCategoryID != nil && source.DefaultProjectID != nil {
				applyRouteTarget(payload, *source.DefaultProjectID, source.DefaultCategoryID, "SERVICE_MAPPING")
				return nil
			}
			if source.DefaultProjectID != nil {
				applyRouteTarget(payload, *source.DefaultProjectID, nil, "SERVICE_MAPPING")
				return nil
			}
		}
	}
	if envelope.SourceID != nil {
		var source models.LogSource
		if err := u.repo.DB().WithContext(ctx).
			Where("source_id = ? AND product_id = ? AND is_active = TRUE", *envelope.SourceID, productID).
			First(&source).Error; err == nil {
			if source.DefaultCategoryID != nil && source.DefaultProjectID != nil {
				applyRouteTarget(payload, *source.DefaultProjectID, source.DefaultCategoryID, "SERVICE_MAPPING")
				return nil
			}
			if source.DefaultProjectID != nil {
				applyRouteTarget(payload, *source.DefaultProjectID, nil, "SERVICE_MAPPING")
				return nil
			}
		}
	}

	// A default is safe only when the authenticated product has exactly one
	// active project. API-key metadata never selects a project.
	var projects []models.Project
	if err := u.repo.DB().WithContext(ctx).
		Where("product_id = ? AND is_active = TRUE", productID).
		Order("project_id").Limit(2).Find(&projects).Error; err != nil {
		return fmt.Errorf("load product projects: %w", err)
	}
	if project := singleProjectDefault(projects); project != nil {
		applyRouteTarget(payload, project.ProjectID, nil, "PRODUCT_SINGLE_PROJECT")
		payload["routing_reason"] = "Used the product's only active project"
		return nil
	}

	applyRoutingMetadata(payload, "UNCLASSIFIED", "PRODUCT_UNCLASSIFIED")
	if _, conflict := payload["routing_conflict_rule_ids"]; conflict {
		applyRoutingMetadata(payload, "UNCLASSIFIED", "ROUTING_CONFLICT")
	} else if routeKey != "" {
		payload["routing_reason"] = "No route mapping matched route_key: " + routeKey
	} else {
		payload["routing_reason"] = "project discriminator missing"
	}
	return nil
}

func (u *usecase) hasLocalProjectDiscriminator(ctx context.Context, productID int, payload map[string]any) (bool, error) {
	// Header hierarchy is an explicit OmniLogs contract. Keep validation strict.
	if intPtrFromAny(payload["_header_project_id"]) != nil || firstString(payload, "_header_project_code") != "" {
		return true, nil
	}
	customFields := customFieldsMap(payload)
	metadata, _ := payload["metadata"].(map[string]any)
	actor, _ := payload["actor"].(map[string]any)
	metadataActor, _ := metadata["actor"].(map[string]any)

	projectID := intPtrFromAny(payload["project_id"])
	if projectID == nil {
		projectID = intPtrFromAny(customFields["project_id"])
	}
	if projectID == nil {
		projectID = intPtrFromAny(metadata["project_id"])
	}
	if projectID == nil {
		projectID = intPtrFromAny(actor["project_id"])
	}
	if projectID == nil {
		projectID = intPtrFromAny(metadataActor["project_id"])
	}

	projectCode := firstString(payload, "project_code")
	if projectCode == "" {
		projectCode = firstString(customFields, "project_code")
	}
	if projectCode == "" {
		projectCode = firstString(metadata, "project_code")
	}
	if projectCode == "" {
		projectCode = firstString(actor, "project_code")
	}
	if projectCode == "" {
		projectCode = firstString(metadataActor, "project_code")
	}

	query := u.repo.DB().WithContext(ctx).Model(&models.Project{}).
		Where("product_id = ? AND is_active = TRUE", productID)
	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	} else if projectCode != "" {
		query = query.Where("LOWER(BTRIM(project_code)) = LOWER(BTRIM(?))", projectCode)
	} else {
		return false, nil
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("check explicit project hierarchy: %w", err)
	}
	return count > 0, nil
}
func (u *usecase) resolveSourceProjectRouting(ctx context.Context, productID, environmentID int, sourceID *int, payload map[string]any) (bool, error) {
	query := u.repo.DB().WithContext(ctx).
		Where("product_id = ? AND is_active = TRUE AND (environment_id IS NULL OR environment_id = ?)", productID, environmentID)
	if sourceID != nil {
		query = query.Where("source_id IS NULL OR source_id = ?", *sourceID)
	} else {
		query = query.Where("source_id IS NULL")
	}
	var rules []models.LogRoutingRule
	if err := query.Order("priority DESC, rule_id").Find(&rules).Error; err != nil && !isMissingRoutingTableError(err) {
		return false, fmt.Errorf("load source project routing rules: %w", err)
	}
	for _, rule := range rules {
		if !routingRuleHasField(rule.Conditions, "source_project_id") {
			continue
		}
		matched, err := routingRuleMatches(rule.Conditions, payload)
		if err != nil {
			return false, fmt.Errorf("invalid routing rule %d: %w", rule.RuleID, err)
		}
		if matched {
			applyRouteTarget(payload, rule.TargetProjectID, rule.TargetCategoryID, "SOURCE_PROJECT_MAPPING")
			payload["matched_rule_id"] = rule.RuleID
			payload["routing_reason"] = "Mapped source_project_id to the configured OmniLogs hierarchy"
			return true, nil
		}
	}
	return false, nil
}

func routingRuleHasField(raw json.RawMessage, field string) bool {
	var conditions []routingCondition
	if err := json.Unmarshal(raw, &conditions); err != nil {
		return false
	}
	for _, condition := range conditions {
		if strings.EqualFold(strings.TrimSpace(condition.Field), field) {
			return true
		}
	}
	return false
}
func hasProjectDiscriminator(payload map[string]any) bool {
	customFields := customFieldsMap(payload)
	metadata, _ := payload["metadata"].(map[string]any)
	actor, _ := payload["actor"].(map[string]any)
	metadataActor, _ := metadata["actor"].(map[string]any)

	return intPtrFromAny(payload["project_id"]) != nil ||
		intPtrFromAny(customFields["project_id"]) != nil ||
		intPtrFromAny(metadata["project_id"]) != nil ||
		intPtrFromAny(actor["project_id"]) != nil ||
		intPtrFromAny(metadataActor["project_id"]) != nil ||
		firstString(payload, "project_code") != "" ||
		firstString(customFields, "project_code") != "" ||
		firstString(metadata, "project_code") != "" ||
		firstString(actor, "project_code") != "" ||
		intPtrFromAny(payload["_header_project_id"]) != nil ||
		firstString(payload, "_header_project_code") != ""
}

func singleProjectDefault(projects []models.Project) *models.Project {
	if len(projects) != 1 {
		return nil
	}
	return &projects[0]
}
func routingMatchesDeletedRule(ctx context.Context, u *usecase, productID, environmentID int, sourceID *int, payload map[string]any) bool {
	query := u.repo.DB().WithContext(ctx).
		Where("product_id = ? AND is_active = FALSE AND (environment_id IS NULL OR environment_id = ?)", productID, environmentID)
	if sourceID != nil {
		query = query.Where("source_id IS NULL OR source_id = ?", *sourceID)
	} else {
		query = query.Where("source_id IS NULL")
	}
	var rules []models.LogRoutingRule
	if err := query.Find(&rules).Error; err != nil {
		return false
	}
	for _, rule := range rules {
		matched, err := routingRuleMatches(rule.Conditions, payload)
		if err == nil && matched {
			return true
		}
	}
	return false
}

func applyDeletedMappingFallback(payload map[string]any, reason string) {
	payload["skip_automatic_hierarchy"] = true
	applyRoutingMetadata(payload, "UNCLASSIFIED", "MAPPING_DELETED")
	payload["routing_reason"] = reason
}

// normalizeExplicitHierarchy accepts the new routing envelope and the legacy
// top-level/metadata hierarchy contract. Values are copied to canonical fields
// consumed by the existing hierarchy resolver.
func normalizeExplicitHierarchy(payload map[string]any) bool {
	candidates := make([]hierarchyCandidate, 0, 5)
	if headerProjID := intPtrFromAny(payload["_header_project_id"]); headerProjID != nil {
		payload["project_id"] = *headerProjID
	}
	if headerProjCode := firstString(payload, "_header_project_code"); headerProjCode != "" {
		payload["project_code"] = headerProjCode
	}
	if routing, ok := payload["routing"].(map[string]any); ok {
		candidates = append(candidates, hierarchyCandidate{values: routing, allowCamel: true})
		if explicit, ok := routing["explicit_hierarchy"].(map[string]any); ok {
			candidates = append(candidates, hierarchyCandidate{values: explicit, allowCamel: true})
		}
		if explicit, ok := routing["explicitHierarchy"].(map[string]any); ok {
			candidates = append(candidates, hierarchyCandidate{values: explicit, allowCamel: true})
		}
	}
	if metadata, ok := payload["metadata"].(map[string]any); ok {
		candidates = append(candidates, hierarchyCandidate{values: metadata})
	}
	candidates = append(candidates, hierarchyCandidate{values: payload, allowCamel: true})

	aliases := map[string][]string{
		"project_id":       {"project_id", "projectId"},
		"category_id":      {"category_id", "categoryId"},
		"feature_id":       {"feature_id", "featureId"},
		"sub_feature_id":   {"sub_feature_id", "subFeatureId"},
		"project_code":     {"project_code", "projectCode"},
		"feature_code":     {"feature_code", "featureCode"},
		"sub_feature_code": {"sub_feature_code", "subFeatureCode"},
	}
	found := false
	for canonical, keys := range aliases {
		for _, candidate := range candidates {
			for _, key := range keys {
				if !candidate.allowCamel && key != canonical {
					continue
				}
				value, ok := candidate.values[key]
				if !ok || value == nil || strings.TrimSpace(fmt.Sprint(value)) == "" {
					continue
				}
				if canonical == "project_id" && intPtrFromAny(value) == nil {
					if _, exists := payload["project_code"]; !exists {
						payload["project_code"] = value
					}
				} else {
					payload[canonical] = value
				}
				found = true
				break
			}
			if _, ok := payload[canonical]; ok {
				break
			}
		}
	}
	if path := explicitSubFeaturePath(candidates); len(path) > 0 {
		// The array is authoritative because it preserves each hierarchy level.
		payload["sub_feature_code"] = strings.Join(path, "/")
		found = true
	}
	return found
}

func sameOptionalInt(left, right *int) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
func explicitSubFeaturePath(candidates []hierarchyCandidate) []string {
	for _, candidate := range candidates {
		for _, key := range []string{"sub_feature_path", "subFeaturePath"} {
			value, ok := candidate.values[key]
			if !ok || value == nil {
				continue
			}
			parts := make([]string, 0)
			switch typed := value.(type) {
			case []string:
				parts = append(parts, typed...)
			case []any:
				for _, segment := range typed {
					parts = append(parts, fmt.Sprint(segment))
				}
			case string:
				parts = append(parts, splitHierarchyPath(typed)...)
			}
			if normalized := splitHierarchyPath(strings.Join(parts, "/")); len(normalized) > 0 {
				return normalized
			}
		}
	}
	return nil
}

// Older installations may not have received the optional routing tables yet.
// Keep ingestion backward-compatible and fall through to default routing.
func isMissingRoutingTableError(err error) bool {
	if err == nil {
		return false
	}
	m := strings.ToLower(err.Error())
	return strings.Contains(m, "relation") && strings.Contains(m, "does not exist")
}
func applyRouteTarget(payload map[string]any, projectID int, categoryID *int, method string) {
	payload["project_id"] = projectID
	if categoryID != nil {
		payload["category_id"] = *categoryID
	} else {
		delete(payload, "category_id")
	}
	applyRoutingMetadata(payload, "CLASSIFIED", method)
}

func routingRuleMatches(raw json.RawMessage, payload map[string]any) (bool, error) {
	var conditions []routingCondition
	if err := json.Unmarshal(raw, &conditions); err != nil {
		var condition routingCondition
		if singleErr := json.Unmarshal(raw, &condition); singleErr != nil {
			return false, err
		}
		conditions = []routingCondition{condition}
	}
	if len(conditions) == 0 {
		return false, nil
	}
	for _, condition := range conditions {
		value, exists := routingValue(payload, condition.Field)
		if !exists || !conditionMatches(value, condition) {
			return false, nil
		}
	}
	return true, nil
}

func routingValue(payload map[string]any, path string) (any, bool) {
	current := any(payload)
	for _, segment := range strings.Split(strings.TrimSpace(path), ".") {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = object[segment]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func conditionMatches(actual any, condition routingCondition) bool {
	operator := strings.ToLower(strings.TrimSpace(condition.Operator))
	if operator == "" || operator == "equals" {
		return comparableEqual(actual, condition.Value)
	}
	actualText := fmt.Sprint(actual)
	expectedText := fmt.Sprint(condition.Value)
	switch operator {
	case "not_equals":
		return !comparableEqual(actual, condition.Value)
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
				if comparableEqual(actual, expected.Index(i).Interface()) {
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
		actualNumber, actualErr := strconv.ParseFloat(actualText, 64)
		minimum, minErr := strconv.ParseFloat(fmt.Sprint(condition.Min), 64)
		maximum, maxErr := strconv.ParseFloat(fmt.Sprint(condition.Max), 64)
		return actualErr == nil && minErr == nil && maxErr == nil && actualNumber >= minimum && actualNumber <= maximum
	default:
		return false
	}
}

func comparableEqual(actual, expected any) bool {
	actualNumber, actualErr := strconv.ParseFloat(strings.TrimSpace(fmt.Sprint(actual)), 64)
	expectedNumber, expectedErr := strconv.ParseFloat(strings.TrimSpace(fmt.Sprint(expected)), 64)
	if actualErr == nil && expectedErr == nil {
		return actualNumber == expectedNumber
	}
	return reflect.DeepEqual(normalizeComparable(actual), normalizeComparable(expected))
}
func normalizeComparable(value any) any {
	if text, ok := value.(string); ok {
		return strings.ToLower(strings.TrimSpace(text))
	}
	return value
}
