package usecase

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"omnilogs-api/models"
)

func (u *usecase) authorize(actor Actor, productID int, action string) error {
	if actor.PlatformAdmin {
		return nil
	}
	if productID <= 0 || !existsDB(u.repository.DB(), &models.Product{}, "product_id = ?", productID) {
		return ErrNotFound
	}

	var memberships []models.ProductMembership
	if err := u.repository.DB().
		Where("user_id = ? AND product_id = ? AND is_active = TRUE AND (expires_at IS NULL OR expires_at > ?)", actor.UserID, productID, time.Now()).
		Find(&memberships).Error; err != nil {
		return err
	}
	if len(memberships) == 0 {
		return ErrForbidden
	}

	var rules []models.UserRolePermissionRule
	if err := u.repository.DB().
		Where("user_id = ? AND resource_type = ? AND action = ? AND is_active = TRUE AND (expires_at IS NULL OR expires_at > ?) AND (product_id IS NULL OR product_id = ?)", actor.UserID, "API_KEY", strings.ToUpper(action), time.Now(), productID).
		Find(&rules).Error; err != nil {
		return err
	}

	allowedByRule := false
	for _, membership := range memberships {
		if !scopeAllows(u, membership.MembershipID, productID) {
			continue
		}
		for _, rule := range rules {
			if rule.RoleID != nil && *rule.RoleID != membership.RoleID {
				continue
			}
			if rule.Effect == "DENY" {
				return ErrForbidden
			}
			if rule.Effect == "ALLOW" {
				allowedByRule = true
			}
		}
		if allowedByRule {
			return nil
		}

		var role models.ProductRole
		if err := u.repository.DB().Where("role_id = ? AND product_id = ?", membership.RoleID, productID).First(&role).Error; err == nil && permissionJSONAllows(role.Permissions, "API_KEY", action) {
			return nil
		}
	}

	return ErrForbidden
}

func scopeAllows(u *usecase, membershipID, productID int) bool {
	var scopes []models.ProductMembershipScope
	if err := u.repository.DB().Where("membership_id = ? AND product_id = ? AND is_active = TRUE", membershipID, productID).Find(&scopes).Error; err != nil {
		return false
	}
	return len(scopes) > 0
}

func permissionJSONAllows(raw json.RawMessage, resource, action string) bool {
	var value map[string]any
	if json.Unmarshal(raw, &value) != nil {
		return false
	}
	if all, ok := value["all"].(bool); ok && all {
		return true
	}

	resource, action = strings.ToUpper(resource), strings.ToUpper(action)
	for _, key := range []string{resource, strings.ToLower(resource)} {
		entry, ok := value[key]
		if !ok {
			continue
		}
		switch typed := entry.(type) {
		case bool:
			if typed {
				return true
			}
		case []any:
			for _, item := range typed {
				if strings.EqualFold(fmt.Sprint(item), action) {
					return true
				}
			}
		case map[string]any:
			for k, item := range typed {
				if strings.EqualFold(k, action) {
					allowed, _ := item.(bool)
					return allowed
				}
			}
		}
	}
	for key, item := range value {
		if strings.EqualFold(key, resource+"."+action) {
			allowed, _ := item.(bool)
			return allowed
		}
	}
	return false
}
