package usecase

import (
	"context"
	"fmt"
	"strings"

	"omnilogs-api/dto"
	"omnilogs-api/models"
)

func (u *usecase) enrichMappingContext(ctx context.Context, envelope dto.LogMessage, payload map[string]any) error {
	if envelope.ProductID == nil || envelope.EnvironmentID == nil {
		return fmt.Errorf("product_id and environment_id are required")
	}
	productID, environmentID := *envelope.ProductID, *envelope.EnvironmentID
	sourceHierarchy := hasCustomHierarchy(payload)
	customFields := customFieldsMap(payload)

	var product models.Product
	if err := u.repo.DB().WithContext(ctx).
		Select("product_id", "product_code").
		Where("product_id = ? AND is_active = TRUE", productID).
		First(&product).Error; err != nil {
		return fmt.Errorf("load product context: %w", err)
	}

	var environment models.ProductEnvironment
	if err := u.repo.DB().WithContext(ctx).
		Where("environment_id = ? AND product_id = ? AND is_active = TRUE AND deleted_at IS NULL", environmentID, productID).
		First(&environment).Error; err != nil {
		return fmt.Errorf("load environment context: %w", err)
	}

	payload["product_id"] = productID
	payload["product_code"] = product.ProductCode
	payload["environment_id"] = environmentID
	payload["environment_code"] = environment.EnvironmentCode
	setCustomHierarchyValue(payload, "product_id", productID)
	setCustomHierarchyValue(payload, "product_code", product.ProductCode)
	setCustomHierarchyValue(payload, "environment_id", environmentID)
	setCustomHierarchyValue(payload, "environment_code", environment.EnvironmentCode)

	projectID := intPtrFromAny(payload["project_id"])
	projectCode := firstString(customFields, "project_code", "projectCode", "project")
	if projectCode == "" {
		projectCode = firstString(payload, "project_code")
	}

	if projectID == nil && projectCode != "" {
		var project models.Project
		if err := u.repo.DB().WithContext(ctx).
			Select("project_id", "project_code").
			Where("product_id = ? AND is_active = TRUE AND LOWER(BTRIM(project_code)) = LOWER(BTRIM(?))", productID, projectCode).
			First(&project).Error; err == nil {
			projectID = &project.ProjectID
			payload["project_id"] = project.ProjectID
		}
	}

	if projectID != nil {
		var project models.Project
		if err := u.repo.DB().WithContext(ctx).
			Select("project_id", "project_code").
			Where("project_id = ? AND product_id = ? AND is_active = TRUE", *projectID, productID).
			First(&project).Error; err == nil {
			payload["project_code"] = project.ProjectCode
			setCustomHierarchyValue(payload, "project_id", *projectID)
			setCustomHierarchyValue(payload, "project_code", project.ProjectCode)
		}
	} else {
		delete(payload, "project_code")
	}
	/// ดึงค่า sub_feature_code or feature_code or category_code จาก payload
	categoryID := intPtrFromAny(payload["category_id"])
	/// ดึงค่า category_id จาก payload + เติม Feature ถ้าไม่มี
	if categoryID != nil {
		if projectID == nil {
			return fmt.Errorf("category context requires project_id")
		}
		var feature models.ProjectFeature
		if err := u.repo.DB().WithContext(ctx).
			Where("category_id = ? AND project_id = ? AND product_id = ? AND is_active = TRUE", *categoryID, *projectID, productID).
			First(&feature).Error; err == nil {
			payload["category_code"] = feature.CategoryCode
			setCustomHierarchyValue(payload, "category_id", *categoryID)
			setCustomHierarchyValue(payload, "category_code", feature.CategoryCode)
			applyFeatureCodes(ctx, u, payload, productID, *projectID, feature)
			setCustomHierarchyValue(payload, "feature_code", payload["feature_code"])
			setCustomHierarchyValue(payload, "sub_feature_code", payload["sub_feature_code"])
		}
	} else {
		for _, key := range []string{"category_code", "feature_code", "sub_feature_code"} {
			delete(payload, key)
		}
	}

	if sourceHierarchy {
		payload["mapping_source"] = "custom_fields"
		payload["mapping_status"] = "mapped"
		payload["routing_method"] = "CUSTOM_FIELDS"
		if categoryID == nil && projectID == nil {
			payload["mapping_status"] = "unmapped"
		}
	}
	return nil
}
func applyFeatureCodes(ctx context.Context, u *usecase, payload map[string]any, productID, projectID int, feature models.ProjectFeature) {
	if feature.Level <= 1 {
		payload["feature_id"] = feature.CategoryID
		payload["feature_code"] = feature.CategoryCode
		delete(payload, "sub_feature_id")
		delete(payload, "sub_feature_code")
		return
	}

	// Use the immediate parent as the feature and this node as its sub-feature.
	// This makes finance/invoices/clean_bills appear as invoices > clean_bills.
	payload["sub_feature_id"] = feature.CategoryID
	payload["sub_feature_code"] = feature.CategoryCode
	if feature.ParentID == nil || *feature.ParentID <= 0 {
		delete(payload, "feature_id")
		delete(payload, "feature_code")
		return
	}

	payload["feature_id"] = *feature.ParentID
	var parent models.ProjectFeature
	if err := u.repo.DB().WithContext(ctx).
		Select("category_code").
		Where("category_id = ? AND project_id = ? AND product_id = ? AND is_active = TRUE", *feature.ParentID, projectID, productID).
		First(&parent).Error; err == nil {
		payload["feature_code"] = parent.CategoryCode
	}
}

func mappingSource(method string) string {
	switch method {
	case "CLIENT_EXPLICIT":
		return "explicit"
	case "ROUTE_KEY", "ROUTING_RULE", "SERVICE_MAPPING", "AUTO_HIERARCHY", "PRODUCT_SINGLE_PROJECT":
		return "common_rule"
	default:
		return "unmapped"
	}
}

func mappingStatus(status, method string) string {
	if status == "CLASSIFIED" {
		return "mapped"
	}
	if strings.Contains(method, "CONFLICT") || strings.Contains(method, "AMBIGUOUS") {
		return "ambiguous"
	}
	return "unmapped"
}
