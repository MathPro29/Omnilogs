package usecase

import (
	"fmt"
	"strings"
	"time"

	"omnilogs-api/models"
	"omnilogs-api/responses"
)

func (u *usecase) authorize(actor Actor, target AccessTarget, resource, action string) error {
	var isGlobalAdmin bool
	var pm models.PlatformMembership
	if err := u.repository.DB().Where("user_id = ? AND is_active = TRUE AND platform_role_id IN (1, 2, 3, 5)", actor.UserID).Limit(1).Find(&pm).Error; err == nil && pm.PlatformMembershipID != 0 {
		isGlobalAdmin = true
	}
	if isGlobalAdmin {
		return nil
	}
	if target.ProductID <= 0 || !u.exists(&models.Product{}, "product_id = ?", target.ProductID) {
		return responses.ErrNotFound
	}
	var memberships []models.ProductMembership
	if err := u.repository.DB().Where("user_id = ? AND product_id = ? AND is_active = TRUE AND (expires_at IS NULL OR expires_at > ?)", actor.UserID, target.ProductID, time.Now()).Find(&memberships).Error; err != nil {
		return err
	}
	if len(memberships) == 0 {
		return responses.ErrForbidden
	}
	var rules []models.UserRolePermissionRule
	if err := u.repository.DB().Where("user_id = ? AND resource_type = ? AND action = ? AND is_active = TRUE AND (expires_at IS NULL OR expires_at > ?) AND (product_id IS NULL OR product_id = ?)", actor.UserID, strings.ToUpper(resource), strings.ToUpper(action), time.Now(), target.ProductID).Find(&rules).Error; err != nil {
		return err
	}
	allowedByRule := false
	for _, membership := range memberships {
		if !u.scopeAllows(membership.MembershipID, target, strings.EqualFold(action, "READ")) {
			continue
		}
		if target.EnvironmentID != nil {
			allowed, scoped := u.environmentPermissionAllows(actor.UserID, target.ProductID, *target.EnvironmentID, resource, action)
			if scoped && !allowed {
				continue
			}
		}
		for _, rule := range rules {
			if !ruleMatches(rule, membership.RoleID, target) {
				continue
			}
			if rule.Effect == "DENY" {
				return responses.ErrForbidden
			}
			if rule.Effect == "ALLOW" {
				allowedByRule = true
			}
		}
		if allowedByRule || (strings.EqualFold(action, "READ") && membershipReadAllowed(resource)) {
			return nil
		}
		var role models.ProductRole
		if err := u.repository.DB().Preload("Permissions").Where("role_id = ? AND product_id = ?", membership.RoleID, target.ProductID).First(&role).Error; err == nil {
			if strings.ToUpper(resource) == "FEATURE" || strings.ToUpper(resource) == "CATEGORY" {
				if role.RoleCode != "owner" {
					continue
				}
			}
			if permissionListAllows(role.Permissions, resource, action) {
				return nil
			}
		}
	}
	return responses.ErrForbidden
}

// environmentPermissionAllows bridges the older environment-specific
// permission table with the product-role authorization model. If an explicit
// row exists for the requested environment it becomes an additional gate;
// when no row exists, the product-role decision remains backward compatible.
func (u *usecase) environmentPermissionAllows(userID, productID, environmentID int, resource, action string) (allowed, scoped bool) {
	if !u.repository.DB().Migrator().HasTable(&models.UserPermission{}) {
		return false, false
	}
	var value models.UserPermission
	err := u.repository.DB().Where("user_id = ? AND product_id = ? AND environment_id = ?", userID, productID, environmentID).First(&value).Error
	if err != nil {
		return false, false
	}
	return permissionMapAllows(value.Permissions, resource, action), true
}

func permissionMapAllows(values map[string]interface{}, resource, action string) bool {
	resource = strings.ToUpper(strings.TrimSpace(resource))
	action = strings.ToUpper(strings.TrimSpace(action))
	for key, raw := range values {
		upperKey := strings.ToUpper(strings.TrimSpace(key))
		if upperKey == "ALL" {
			if allowed, ok := raw.(bool); ok && allowed {
				return true
			}
		}
		if upperKey == resource+"."+action {
			if allowed, ok := raw.(bool); ok && allowed {
				return true
			}
		}
		if upperKey != resource {
			continue
		}
		switch typed := raw.(type) {
		case map[string]interface{}:
			if allowed, ok := typed[action].(bool); ok && allowed {
				return true
			}
		case []interface{}:
			for _, item := range typed {
				if strings.EqualFold(fmt.Sprint(item), action) {
					return true
				}
			}
		}
	}
	return false
}

func (u *usecase) scopeAllows(membershipID int, target AccessTarget, allowParentRead bool) bool {
	var scopes []models.ProductMembershipScope
	if err := u.repository.DB().Where("membership_id = ? AND product_id = ? AND is_active = TRUE", membershipID, target.ProductID).Find(&scopes).Error; err != nil {
		return false
	}
	for _, scope := range scopes {
		if allowParentRead && target.ProjectID == nil && target.CategoryID == nil {
			return true
		}
		switch scope.ScopeLevel {
		case "PRODUCT":
			return true
		case "PROJECT":
			if target.ProjectID != nil && scope.ProjectID != nil && *target.ProjectID == *scope.ProjectID {
				return true
			}
		case "CATEGORY":
			if target.CategoryID != nil && scope.CategoryID != nil && *target.CategoryID == *scope.CategoryID {
				return true
			}
			if target.CategoryID != nil && scope.CategoryID != nil {
				var feature models.ProjectFeature
				if err := u.repository.DB().Select("path_ids").First(&feature, *target.CategoryID).Error; err == nil && pathContains(feature.PathIDs, *scope.CategoryID) {
					return true
				}
			}
			if allowParentRead && target.ProjectID != nil && scope.ProjectID != nil && *target.ProjectID == *scope.ProjectID {
				return true
			}
		}
	}
	return false
}

func ruleMatches(rule models.UserRolePermissionRule, roleID int, target AccessTarget) bool {
	if rule.RoleID != nil && *rule.RoleID != roleID {
		return false
	}
	if rule.ProjectID != nil && (target.ProjectID == nil || *rule.ProjectID != *target.ProjectID) {
		return false
	}
	if rule.CategoryID != nil && (target.CategoryID == nil || *rule.CategoryID != *target.CategoryID) {
		return false
	}
	return true
}
