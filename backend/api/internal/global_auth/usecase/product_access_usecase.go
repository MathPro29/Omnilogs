package usecase

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"omnilogs-api/dto"
	"omnilogs-api/models"
	"omnilogs-api/responses"

	"gorm.io/gorm"
)

func roleAccessLevel(code string, permissions []models.ProductRolePermission) string {
	if strings.EqualFold(code, "owner") {
		return "FULL_ACCESS"
	}
	level := "READ_ONLY"
	for _, p := range permissions {
		switch strings.ToUpper(p.Action) {
		case "GRANT", "REVOKE", "DELETE":
			return "ADMIN"
		case "CREATE", "UPDATE":
			level = "EDITOR"
		}
	}
	return level
}

func (u *usecase) GetProductAccessOverview(actor Actor, productID int) (*dto.ProductAccessOverviewResponse, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ACCESS", "READ"); err != nil {
		return nil, err
	}
	return u.buildProductAccessOverview(productID)
}

func (u *usecase) buildProductAccessOverview(productID int) (*dto.ProductAccessOverviewResponse, error) {
	var roles []models.ProductRole
	if err := u.repository.DB().Preload("Permissions").Where("product_id = ?", productID).Order("role_name").Find(&roles).Error; err != nil {
		return nil, err
	}
	var memberships []models.ProductMembership
	if err := u.repository.DB().Where("product_id = ?", productID).Order("updated_at DESC, membership_id DESC").Find(&memberships).Error; err != nil {
		return nil, err
	}
	// Tolerate legacy duplicate rows until the migration has normalized them.
	seenUsers := map[int]bool{}
	unique := memberships[:0]
	for _, m := range memberships {
		if !seenUsers[m.UserID] {
			seenUsers[m.UserID] = true
			unique = append(unique, m)
		}
	}
	memberships = unique
	userIDs, membershipIDs := make([]int, 0, len(memberships)), make([]int, 0, len(memberships))
	for _, m := range memberships {
		userIDs = append(userIDs, m.UserID)
		membershipIDs = append(membershipIDs, m.MembershipID)
	}
	var users []models.User
	if len(userIDs) > 0 {
		if err := u.repository.DB().Where("user_id IN ?", userIDs).Find(&users).Error; err != nil {
			return nil, err
		}
	}
	var scopes []models.ProductMembershipScope
	if len(membershipIDs) > 0 {
		if err := u.repository.DB().Where("membership_id IN ? AND product_id = ? AND is_active = TRUE", membershipIDs, productID).Order("scope_level, project_id, category_id").Find(&scopes).Error; err != nil {
			return nil, err
		}
	}
	var rules []models.UserRolePermissionRule
	if len(userIDs) > 0 {
		if err := u.repository.DB().Where("product_id = ? AND user_id IN ? AND is_active = TRUE AND (expires_at IS NULL OR expires_at > ?)", productID, userIDs, time.Now()).Find(&rules).Error; err != nil {
			return nil, err
		}
	}
	roleByID := map[int]models.ProductRole{}
	userByID := map[int]models.User{}
	scopesByMembership := map[int][]models.ProductMembershipScope{}
	rulesByUser := map[int][]models.UserRolePermissionRule{}
	counts := map[int]int{}
	for _, r := range roles {
		roleByID[r.RoleID] = r
	}
	for _, v := range users {
		userByID[v.UserID] = v
	}
	for _, v := range scopes {
		scopesByMembership[v.MembershipID] = append(scopesByMembership[v.MembershipID], v)
	}
	for _, v := range rules {
		rulesByUser[v.UserID] = append(rulesByUser[v.UserID], v)
	}
	out := &dto.ProductAccessOverviewResponse{ProductID: productID, UpdatedAt: time.Now(), Roles: []dto.ProductAccessRoleResponse{}, Members: []dto.ProductAccessMemberResponse{}}
	for _, m := range memberships {
		r := roleByID[m.RoleID]
		usr := userByID[m.UserID]
		counts[m.RoleID]++
		perms := map[string]dto.EffectivePermissionResponse{}
		for _, p := range r.Permissions {
			key := strings.ToUpper(p.ResourceType) + ":" + strings.ToUpper(p.Action)
			perms[key] = dto.EffectivePermissionResponse{ResourceType: strings.ToUpper(p.ResourceType), Action: strings.ToUpper(p.Action), Source: "ROLE"}
		}
		for _, rule := range rulesByUser[m.UserID] {
			if rule.RoleID != nil && *rule.RoleID != m.RoleID {
				continue
			}
			key := strings.ToUpper(rule.ResourceType) + ":" + strings.ToUpper(rule.Action)
			if strings.EqualFold(rule.Effect, "DENY") {
				delete(perms, key)
			} else {
				perms[key] = dto.EffectivePermissionResponse{ResourceType: strings.ToUpper(rule.ResourceType), Action: strings.ToUpper(rule.Action), Source: "RULE_ALLOW"}
			}
		}
		effective := make([]dto.EffectivePermissionResponse, 0, len(perms))
		for _, p := range perms {
			effective = append(effective, p)
		}
		sort.Slice(effective, func(i, j int) bool {
			return effective[i].ResourceType+effective[i].Action < effective[j].ResourceType+effective[j].Action
		})
		fullName := strings.TrimSpace(fmt.Sprintf("%s %s", stringValue(usr.FirstName), stringValue(usr.LastName)))
		if fullName == "" {
			fullName = stringValue(usr.Username)
		}
		if fullName == "" {
			fullName = usr.Email
		}
		memberScopes := make([]dto.MembershipScopeResponse, 0, len(scopesByMembership[m.MembershipID]))
		for _, s := range scopesByMembership[m.MembershipID] {
			memberScopes = append(memberScopes, dto.MembershipScopeResponse{ScopeID: s.ScopeID, MembershipID: s.MembershipID, ProductID: s.ProductID, ProjectID: s.ProjectID, CategoryID: s.CategoryID, ScopeLevel: s.ScopeLevel, IsActive: s.IsActive})
		}
		out.Members = append(out.Members, dto.ProductAccessMemberResponse{MembershipID: m.MembershipID, UserID: m.UserID, Username: usr.Username, FullName: fullName, Email: usr.Email, RoleID: r.RoleID, RoleCode: r.RoleCode, RoleName: r.RoleName, AccessLevel: roleAccessLevel(r.RoleCode, r.Permissions), EffectivePermissions: effective, Scopes: memberScopes, ExpiresAt: m.ExpiresAt, IsActive: m.IsActive})
	}
	for _, r := range roles {
		ps := make([]dto.RolePermissionAssignment, 0, len(r.Permissions))
		for _, p := range r.Permissions {
			ps = append(ps, dto.RolePermissionAssignment{ResourceType: p.ResourceType, Action: p.Action})
		}
		out.Roles = append(out.Roles, dto.ProductAccessRoleResponse{RoleID: r.RoleID, RoleCode: r.RoleCode, RoleName: r.RoleName, Permissions: ps, AccessLevel: roleAccessLevel(r.RoleCode, r.Permissions), MemberCount: counts[r.RoleID], IsActive: r.IsActive})
	}
	return out, nil
}

func stringValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func (u *usecase) UpsertProductAccess(actor Actor, productID, userID int, req dto.UpsertProductAccessRequest) (*dto.ProductAccessMemberResponse, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ACCESS", "GRANT"); err != nil {
		return nil, err
	}
	if !u.exists(&models.User{}, "user_id = ? AND deleted_at IS NULL", userID) || !u.exists(&models.ProductRole{}, "role_id = ? AND product_id = ?", req.RoleID, productID) {
		return nil, responses.ErrInvalid
	}
	err := u.repository.Transaction(func(tx *gorm.DB) error {
		var m models.ProductMembership
		if err := tx.Where("user_id = ? AND product_id = ?", userID, productID).Order("updated_at DESC, membership_id DESC").Limit(1).Find(&m).Error; err != nil {
			return err
		}
		active := true
		if req.IsActive != nil {
			active = *req.IsActive
		}
		if m.MembershipID == 0 {
			m = models.ProductMembership{UserID: userID, ProductID: productID, RoleID: req.RoleID, ExpiresAt: req.ExpiresAt, IsActive: active}
			if err := tx.Create(&m).Error; err != nil {
				return classifyDBError(err)
			}
		} else {
			if err := tx.Model(&m).Updates(map[string]any{"role_id": req.RoleID, "expires_at": req.ExpiresAt, "is_active": active}).Error; err != nil {
				return classifyDBError(err)
			}
		}
		if err := tx.Where("membership_id = ? AND product_id = ?", m.MembershipID, productID).Delete(&models.ProductMembershipScope{}).Error; err != nil {
			return err
		}
		seen := map[string]bool{}
		for _, input := range req.Scopes {
			level := strings.ToUpper(input.ScopeLevel)
			key := fmt.Sprintf("%s:%v:%v", level, input.ProjectID, input.CategoryID)
			if seen[key] {
				continue
			}
			seen[key] = true
			if level == "PRODUCT" {
				input.ProjectID = nil
				input.CategoryID = nil
			} else if level == "PROJECT" {
				if input.ProjectID == nil || !existsDB(tx, &models.Project{}, "project_id = ? AND product_id = ?", *input.ProjectID, productID) {
					return responses.ErrInvalid
				}
				input.CategoryID = nil
			} else if level == "CATEGORY" {
				if input.ProjectID == nil || input.CategoryID == nil || !existsDB(tx, &models.Project{}, "project_id = ? AND product_id = ?", *input.ProjectID, productID) || !existsDB(tx, &models.ProjectFeature{}, "category_id = ? AND project_id = ?", *input.CategoryID, *input.ProjectID) {
					return responses.ErrInvalid
				}
			} else {
				return responses.ErrInvalid
			}
			s := models.ProductMembershipScope{MembershipID: m.MembershipID, ProductID: productID, ProjectID: input.ProjectID, CategoryID: input.CategoryID, ScopeLevel: level, IsActive: true}
			if err := tx.Create(&s).Error; err != nil {
				return classifyDBError(err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	overview, err := u.buildProductAccessOverview(productID)
	if err != nil {
		return nil, err
	}
	for i := range overview.Members {
		if overview.Members[i].UserID == userID {
			return &overview.Members[i], nil
		}
	}
	return nil, responses.ErrNotFound
}
