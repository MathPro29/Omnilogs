package usecase

import (
	"strings"
	"time"

	"omnilogs-api/dto"
	"omnilogs-api/models"
	"omnilogs-api/responses"
	"omnilogs-api/utils"
)

func (u *usecase) ListProducts(actor Actor) ([]models.Product, error) {
	var products []models.Product
	q := u.repository.DB().Model(&models.Product{}).Preload("ProductEnvironments").Order("product_id")

	isGlobalAdmin := true

	if !isGlobalAdmin {
		q = q.Where(`EXISTS (SELECT 1 FROM product_memberships pm JOIN product_membership_scopes pms ON pms.membership_id = pm.membership_id WHERE pm.product_id = products.product_id AND pm.user_id = ? AND pm.is_active = TRUE AND pms.is_active = TRUE AND (pm.expires_at IS NULL OR pm.expires_at > ?))`, actor.UserID, time.Now())
	}
	if err := q.Find(&products).Error; err != nil {
		return nil, err
	}
	if isGlobalAdmin {
		for i := range products {
			products[i].MembershipRole = "platform_admin"
			products[i].IsOwner = true
		}
		return products, nil
	}
	var memberships []struct {
		ProductID    int
		MembershipID int
		RoleCode     string
	}
	if err := u.repository.DB().Table("product_memberships AS pm").
		Select("pm.product_id, pm.membership_id, pr.role_code").
		Joins("JOIN product_roles AS pr ON pr.role_id = pm.role_id").
		Joins("JOIN product_membership_scopes AS pms ON pms.membership_id = pm.membership_id AND pms.is_active = TRUE").
		Where("pm.user_id = ? AND pm.is_active = TRUE AND (pm.expires_at IS NULL OR pm.expires_at > ?)", actor.UserID, time.Now()).
		Order("pm.updated_at DESC, pm.membership_id DESC").Find(&memberships).Error; err != nil {
		return nil, err
	}
	accessByProduct := map[int]struct {
		membershipID int
		roleCode     string
	}{}
	for _, membership := range memberships {
		current, exists := accessByProduct[membership.ProductID]
		if !exists || (strings.EqualFold(membership.RoleCode, "owner") && !strings.EqualFold(current.roleCode, "owner")) {
			accessByProduct[membership.ProductID] = struct {
				membershipID int
				roleCode     string
			}{membership.MembershipID, membership.RoleCode}
		}
	}
	for i := range products {
		if access, ok := accessByProduct[products[i].ProductID]; ok {
			products[i].MembershipID = access.membershipID
			products[i].MembershipRole = access.roleCode
			products[i].IsOwner = strings.EqualFold(access.roleCode, "owner")
		}
	}
	return products, nil
}

func (u *usecase) GetProduct(actor Actor, productID int) (*models.Product, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "PRODUCT", "READ"); err != nil {
		return nil, err
	}
	var value models.Product
	if err := u.repository.DB().Preload("ProductEnvironments").First(&value, productID).Error; err != nil {
		return nil, classifyDBError(err)
	}
	return &value, nil
}

func (u *usecase) UpdateProduct(actor Actor, productID int, req dto.UpdateProductRequest) (*models.Product, error) {
	if err := u.requireProductOwner(actor, productID); err != nil {
		return nil, err
	}
	updates := map[string]any{}
	if req.ProductName != nil {
		name := strings.TrimSpace(*req.ProductName)
		if name == "" {
			return nil, responses.ErrInvalid
		}
		updates["product_name"] = name
	}
	if req.Description != nil {
		updates["description"] = strings.TrimSpace(*req.Description)
	}
	if req.SetupStatus != nil {
		updates["setup_status"] = *req.SetupStatus
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if len(updates) > 0 {
		if err := u.repository.DB().Model(&models.Product{}).Where("product_id = ?", productID).Updates(updates).Error; err != nil {
			return nil, classifyDBError(err)
		}
	}
	return u.GetProduct(actor, productID)
}

func (u *usecase) DeleteProduct(actor Actor, productID int) error {
	if err := u.requireProductOwner(actor, productID); err != nil {
		return err
	}
	if err := u.repository.DB().Delete(&models.Product{}, productID).Error; err != nil {
		return classifyDBError(err)
	}
	return nil
}

func (u *usecase) RestoreProduct(actor Actor, productID int) (*models.Product, error) {
	if err := u.requireProductOwner(actor, productID); err != nil {
		return nil, err
	}
	var deleted models.Product
	if err := u.repository.DB().Unscoped().Where("product_id = ? AND deleted_at IS NOT NULL", productID).First(&deleted).Error; err != nil {
		return nil, classifyDBError(err)
	}
	var conflict models.Product
	query := u.repository.DB().Where("product_id <> ? AND (LOWER(BTRIM(product_name)) = ? OR LOWER(BTRIM(product_code)) = ?)", productID, utils.NormalizeName(deleted.ProductName), utils.NormalizeCode(deleted.ProductCode))
	if err := query.First(&conflict).Error; err == nil {
		return nil, &responses.ScopedConflictError{Field: "Product name or code", Value: deleted.ProductName, Scope: "active products", Message: "Product cannot be restored because its name or code is already in use", Cause: responses.ErrConflict}
	}
	if err := u.repository.DB().Unscoped().Model(&models.Product{}).Where("product_id = ? AND deleted_at IS NOT NULL", productID).Update("deleted_at", nil).Error; err != nil {
		return nil, classifyDBError(err)
	}
	return u.GetProduct(actor, productID)
}
func (u *usecase) BulkDeleteProducts(actor Actor, productID []int) error {
	for _, id := range productID {
		if err := u.requireProductOwner(actor, id); err != nil {
			return err
		}
	}
	if err := u.repository.DB().Delete(&models.Product{}, "product_id IN ?", productID).Error; err != nil {
		return classifyDBError(err)
	}
	return nil
}
