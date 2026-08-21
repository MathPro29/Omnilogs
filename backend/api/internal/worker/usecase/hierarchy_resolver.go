package usecase

import (
	"context"
	"fmt"
	"strings"

	"omnilogs-api/models"
)

type resolvedLogHierarchy struct {
	projectID int
	featureID *int
	pathIDs   *string
	fullPath  *string
}

// resolveLogHierarchy normalizes identifiers emitted by SDKs and the setup wizard.
// Clients may send stable codes instead of database IDs; indexed documents always
// receive canonical IDs and the complete feature path used by scoped searches.
func (u *usecase) resolveLogHierarchy(ctx context.Context, productID, environmentID int, payload map[string]any, cache *batchCache) error {
	if skip, _ := payload["skip_automatic_hierarchy"].(bool); skip {
		fallbackToProductEnvironment(payload, "A deleted route mapping explicitly keeps this log at Product + Environment")
		applyRoutingMetadata(payload, "UNCLASSIFIED", "MAPPING_DELETED")
		return nil
	}
	customFields := customFieldsMap(payload)
	metadata, _ := payload["metadata"].(map[string]any)
	actor, _ := payload["actor"].(map[string]any)
	metadataActor, _ := metadata["actor"].(map[string]any)

	// Discriminator precedence: 1. Header (_header_project_id) 2. Payload project_id 3. Payload project_code
	headerProjectID := intPtrFromAny(payload["_header_project_id"])
	headerProjectCode := firstString(payload, "_header_project_code")

	payloadProjectID := intPtrFromAny(payload["project_id"])
	if payloadProjectID == nil {
		payloadProjectID = intPtrFromAny(customFields["project_id"])
	}
	if payloadProjectID == nil {
		payloadProjectID = intPtrFromAny(metadata["project_id"])
	}
	if payloadProjectID == nil {
		payloadProjectID = intPtrFromAny(actor["project_id"])
	}
	if payloadProjectID == nil {
		payloadProjectID = intPtrFromAny(metadataActor["project_id"])
	}

	projectID := headerProjectID
	if projectID != nil && payloadProjectID != nil && *projectID != *payloadProjectID {
		payload["routing_conflict"] = true
		payload["header_project_id"] = fmt.Sprintf("%d", *projectID)
		payload["payload_project_id"] = fmt.Sprintf("%d", *payloadProjectID)
		payload["routing_reason"] = "Header X-Project-ID takes precedence over payload project_id"
	}
	if projectID == nil {
		projectID = payloadProjectID
	}

	categoryID := intPtrFromAny(payload["category_id"])
	if categoryID == nil {
		categoryID = intPtrFromAny(customFields["category_id"])
	}
	if categoryID == nil {
		categoryID = intPtrFromAny(metadata["category_id"])
	}
	featureID := intPtrFromAny(payload["feature_id"])
	subFeatureID := intPtrFromAny(payload["sub_feature_id"])

	projectCode := headerProjectCode
	if projectCode == "" {
		projectCode = firstString(payload, "project_code")
	}
	if projectCode == "" {
		projectCode = firstString(customFields, "project_code", "projectCode", "project")
	}
	if projectCode == "" {
		projectCode = firstString(metadata, "project_code")
	}
	if projectCode == "" {
		projectCode = firstString(actor, "project_code", "projectCode", "project")
	}
	if projectCode == "" {
		projectCode = firstString(metadataActor, "project_code", "projectCode", "project")
	}
	featureCode := firstString(payload, "feature_code")
	subFeatureCode := firstString(payload, "sub_feature_code")
	legacyCategoryCode := firstString(payload, "category_code")
	if featureCode == "" {
		featureCode = firstString(customFields, "feature_code", "featureCode")
	}
	if subFeatureCode == "" {
		subFeatureCode = firstString(customFields, "sub_feature_code", "subFeatureCode")
	}
	if legacyCategoryCode == "" {
		legacyCategoryCode = firstString(customFields, "category_code", "categoryCode")
	}
	if featureCode == "" {
		featureCode = firstString(metadata, "feature_code")
	}
	if subFeatureCode == "" {
		subFeatureCode = firstString(metadata, "sub_feature_code")
	}
	if legacyCategoryCode == "" {
		legacyCategoryCode = firstString(metadata, "category_code")
	}
	if projectID == nil && projectCode == "" && categoryID == nil && featureID == nil && subFeatureID == nil && featureCode == "" && subFeatureCode == "" && legacyCategoryCode == "" {
		return nil
	}

	targetID := categoryID
	if featureID != nil {
		targetID = featureID
	}
	if subFeatureID != nil {
		targetID = subFeatureID
	}
	hierarchyCodes := strings.Join([]string{featureCode, subFeatureCode, legacyCategoryCode}, ">")
	key := hierarchyCacheKey(productID, environmentID, projectID, projectCode, targetID, hierarchyCodes)
	if cached, ok := cache.resolvedHierarchy.Load(key); ok {
		resolved := cached.(resolvedLogHierarchy)
		if err := applyResolvedHierarchy(payload, resolved); err != nil {
			return err
		}
		if resolved.featureID != nil && firstString(payload, "routing_status") != "CLASSIFIED" {
			applyRoutingMetadata(payload, "CLASSIFIED", "AUTO_HIERARCHY")
		}
		return nil
	}

	db := u.repo.DB().WithContext(ctx)
	if projectID != nil {
		var project models.Project
		if err := db.Where("project_id = ? AND product_id = ? AND is_active = TRUE", *projectID, productID).First(&project).Error; err != nil {
			payload["source_project_id"] = *projectID
			delete(payload, "project_id")
			delete(payload, "category_id")
			payload["classification_status"] = "UNCLASSIFIED"
			if _, exists := payload["routing_reason"]; !exists {
				payload["routing_reason"] = fmt.Sprintf("Project ID %d is unclassified (not registered in database for product %d)", *projectID, productID)
			}
			applyRoutingMetadata(payload, "UNCLASSIFIED", "UNREGISTERED_PROJECT")
			return nil
		}
	}
	if projectID == nil && projectCode != "" {
		var projects []models.Project
		if err := db.Where("LOWER(BTRIM(project_code)) = LOWER(BTRIM(?)) AND product_id = ? AND is_active = TRUE", projectCode, productID).
			Limit(2).Find(&projects).Error; err != nil {
			return fmt.Errorf("resolve project_code: %w", err)
		}
		if len(projects) == 0 {
			payload["source_project_code"] = projectCode
			delete(payload, "project_code")
			delete(payload, "project_id")
			delete(payload, "category_id")
			payload["classification_status"] = "UNCLASSIFIED"
			if _, exists := payload["routing_reason"]; !exists {
				payload["routing_reason"] = fmt.Sprintf("Project Code %q is unclassified (not registered in database for product %d)", projectCode, productID)
			}
			applyRoutingMetadata(payload, "UNCLASSIFIED", "UNREGISTERED_PROJECT")
			return nil
		}
		if len(projects) > 1 {
			fallbackToProductEnvironment(payload, "AMBIGUOUS_PROJECT: project_code matched multiple projects")
			applyRoutingMetadata(payload, "UNCLASSIFIED", "AMBIGUOUS_HIERARCHY")
			return nil
		}
		projectID = &projects[0].ProjectID
	}

	var feature *models.ProjectFeature
	if targetID != nil {
		if projectID == nil {
			fallbackToProductEnvironment(payload, "CATEGORY_NOT_IN_PROJECT: hierarchy ID requires an explicit or mapped project")
			applyRoutingMetadata(payload, "UNCLASSIFIED", "INCOMPLETE_EXPLICIT_HIERARCHY")
			return nil
		}
		features, err := u.loadActiveProjectFeatures(ctx, productID, *projectID, cache)
		if err != nil {
			return err
		}
		value := findFeatureByID(features, *targetID)
		if value == nil {
			return fmt.Errorf("CATEGORY_NOT_IN_PROJECT: category_id %d does not belong to product %d project %d", *targetID, productID, *projectID)
		}
		feature = value
		categoryID = &value.CategoryID
		if featureID != nil && subFeatureID != nil && (value.ParentID == nil || *value.ParentID != *featureID) {
			return fmt.Errorf("SUB_FEATURE_NOT_IN_FEATURE: sub_feature_id %d is not a child of feature_id %d", *subFeatureID, *featureID)
		}
		if featureID != nil && subFeatureID == nil && value.ParentID != nil {
			return fmt.Errorf("FEATURE_NOT_AT_PROJECT_ROOT: feature_id %d has a parent", *featureID)
		}
	}

	// Query หา sub_feature_code ใน DB
	if feature == nil && (featureCode != "" || subFeatureCode != "" || legacyCategoryCode != "") {
		if projectID == nil {
			fallbackToProductEnvironment(payload, "CATEGORY_NOT_IN_PROJECT: hierarchy codes require an explicit or mapped project")
			applyRoutingMetadata(payload, "UNCLASSIFIED", "INCOMPLETE_EXPLICIT_HIERARCHY")
			return nil
		}
		features, err := u.loadActiveProjectFeatures(ctx, productID, *projectID, cache)
		if err != nil {
			return err
		}
		if featureCode != "" {
			featureSegments := splitHierarchyPath(featureCode)
			var curr *models.ProjectFeature
			for _, seg := range featureSegments {
				match := findDescendantFeature(features, curr, seg)
				if match != nil {
					curr = match
				}
			}
			if curr == nil {
				curr = bestAutomaticFeatureMatch(features, []string{featureCode})
			}
			if curr != nil {
				feature = curr
				categoryID = &curr.CategoryID
			}
			if subFeatureCode != "" {
				subSegments := splitHierarchyPath(subFeatureCode)
				for _, seg := range subSegments {
					child := findDescendantFeature(features, feature, seg)
					if child != nil {
						feature = child
						categoryID = &child.CategoryID
					}
				}
			}
		} else if subFeatureCode != "" {
			subSegments := splitHierarchyPath(subFeatureCode)
			var curr *models.ProjectFeature
			for _, seg := range subSegments {
				match := findDescendantFeature(features, curr, seg)
				if match != nil {
					curr = match
				}
			}
			if curr == nil {
				curr = bestAutomaticFeatureMatch(features, []string{subFeatureCode})
			}
			if curr != nil {
				feature = curr
				categoryID = &curr.CategoryID
			}
		} else if legacyCategoryCode != "" {
			matches := findFeaturesByCode(features, legacyCategoryCode, 2)
			if len(matches) == 1 {
				feature = &matches[0]
				categoryID = &matches[0].CategoryID
			}
		}
	}

	resolved := resolvedLogHierarchy{featureID: categoryID}
	if projectID != nil {
		resolved.projectID = *projectID
	}
	if feature != nil {
		resolved.pathIDs = feature.PathIDs
		resolved.fullPath = feature.FullPath
	} else if featureCode != "" || subFeatureCode != "" {
		parts := make([]string, 0)
		if featureCode != "" {
			for _, p := range splitHierarchyPath(featureCode) {
				parts = append(parts, formatTitleCase(p))
			}
		}
		if subFeatureCode != "" {
			for _, p := range splitHierarchyPath(subFeatureCode) {
				parts = append(parts, formatTitleCase(p))
			}
		}
		if len(parts) > 0 {
			path := strings.Join(parts, " / ")
			resolved.fullPath = &path
		}
	}

	if feature == nil && categoryID == nil && resolved.fullPath == nil {
		if projectID != nil {
			fallbackToProject(payload, *projectID, "No feature in the scoped project matched route or hierarchy hints")
		} else {
			fallbackToProductEnvironment(payload, "No existing hierarchy matched route or hierarchy hints")
		}
		return nil
	}

	applyRoutingMetadata(payload, "CLASSIFIED", "AUTO_HIERARCHY")
	cache.resolvedHierarchy.Store(key, resolved)
	return applyResolvedHierarchy(payload, resolved)
}

func (u *usecase) loadActiveProjectFeatures(ctx context.Context, productID, projectID int, cache *batchCache) ([]models.ProjectFeature, error) {
	key := fmt.Sprintf("%d|%d", productID, projectID)
	if cached, ok := cache.projectFeatures.Load(key); ok {
		return cached.([]models.ProjectFeature), nil
	}
	var features []models.ProjectFeature
	if err := u.repo.DB().WithContext(ctx).
		Where("product_id = ? AND project_id = ? AND is_active = TRUE", productID, projectID).
		Order("category_id").Find(&features).Error; err != nil {
		return nil, fmt.Errorf("load project hierarchy: %w", err)
	}
	cache.projectFeatures.Store(key, features)
	return features, nil
}

func hierarchyCacheKey(productID, environmentID int, projectID *int, projectCode string, categoryID *int, featureCode string) string {
	return fmt.Sprintf("%d|%d|%s|%s|%s|%s",
		productID,
		environmentID,
		pointerIntKey(projectID),
		strings.ToLower(strings.TrimSpace(projectCode)),
		pointerIntKey(categoryID),
		strings.ToLower(strings.TrimSpace(featureCode)),
	)
}

func fallbackToProject(payload map[string]any, projectID int, reason string) {
	for _, key := range []string{"category_id", "feature_id", "sub_feature_id", "feature_path_ids", "feature_full_path"} {
		delete(payload, key)
	}
	payload["project_id"] = projectID
	applyRoutingMetadata(payload, "CLASSIFIED", "PROJECT_SCOPE_DEFAULT")
	payload["routing_reason"] = reason
}

func fallbackToProductEnvironment(payload map[string]any, reason string) {
	for _, key := range []string{"project_id", "category_id", "feature_id", "sub_feature_id", "feature_path_ids", "feature_full_path"} {
		delete(payload, key)
	}
	applyRoutingMetadata(payload, "UNCLASSIFIED", "PRODUCT_ENVIRONMENT_DEFAULT")
	payload["routing_reason"] = reason
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func applyResolvedHierarchy(payload map[string]any, value resolvedLogHierarchy) error {
	if value.projectID > 0 {
		payload["project_id"] = value.projectID
	}
	if value.featureID != nil {
		payload["category_id"] = *value.featureID
	}
	if value.pathIDs != nil {
		payload["feature_path_ids"] = *value.pathIDs
	}
	if value.fullPath != nil {
		payload["feature_full_path"] = *value.fullPath
	}
	return nil
}

func firstString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := payload[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func pointerIntKey(value *int) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(*value)
}

func applyRoutingMetadata(payload map[string]any, status, method string) {
	payload["routing_status"] = status
	payload["routing_method"] = method
}
