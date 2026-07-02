package usecase

import (
	"strings"
	"time"

	"omnilogs-api/dto"
	"omnilogs-api/models"
	"omnilogs-api/responses"
	"omnilogs-api/utils"

	"gorm.io/gorm"
)

func (u *usecase) CreateProduct(actor Actor, req dto.CreateProductRequest) (*models.Product, error) {
	if !actor.PlatformAdmin {
		return nil, responses.ErrForbidden
	}
	productCode := normalizeCode(req.ProductCode)
	if productCode == "" {
		productCode = utils.GenerateCode(req.ProductName)
	}
	product := &models.Product{ProductName: strings.TrimSpace(req.ProductName), ProductCode: productCode, IsActive: true}
	if product.ProductName == "" || product.ProductCode == "" {
		return nil, responses.ErrInvalid
	}
	err := u.repository.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(product).Error; err != nil {
			return classifyDBError(err)
		}
		for _, environmentRequest := range req.Environments {
			if environmentRequest.ProductID != nil && *environmentRequest.ProductID != product.ProductID {
				return responses.ErrInvalid
			}
			environment := models.ProductEnvironment{
				ProductID:       product.ProductID,
				EnvironmentCode: normalizeCode(environmentRequest.EnvironmentCode),
				EnvironmentName: strings.TrimSpace(environmentRequest.EnvironmentName),
			}
			if environment.EnvironmentCode == "" || environment.EnvironmentName == "" {
				return responses.ErrInvalid
			}
			if err := tx.Create(&environment).Error; err != nil {
				return classifyDBError(err)
			}
			product.ProductEnvironments = append(product.ProductEnvironments, environment)
		}

		// Create default Owner role for this product
		ownerRole := models.ProductRole{
			ProductID: product.ProductID,
			RoleCode:  "owner",
			RoleName:  "Owner",
			IsDefault: true,
			IsActive:  true,
		}
		if err := tx.Create(&ownerRole).Error; err != nil {
			return classifyDBError(err)
		}

		// Create permissions for this role
		resources := []string{"PRODUCT", "PROJECT", "ENVIRONMENT", "FEATURE", "CATEGORY", "LOG", "API_KEY", "ROLE", "ACCESS"}
		actions := []string{"CREATE", "READ", "UPDATE", "DELETE", "GRANT", "REVOKE", "EXPORT", "VIEW_SENSITIVE"}
		var perms []models.ProductRolePermission
		for _, res := range resources {
			for _, act := range actions {
				perms = append(perms, models.ProductRolePermission{
					RoleID:       ownerRole.RoleID,
					ResourceType: res,
					Action:       act,
				})
			}
		}
		if err := tx.Create(&perms).Error; err != nil {
			return classifyDBError(err)
		}

		// Assign the creator (actor.UserID) to the Owner role of this product
		membership := models.ProductMembership{
			UserID:    actor.UserID,
			ProductID: product.ProductID,
			RoleID:    ownerRole.RoleID,
			IsActive:  true,
		}
		if err := tx.Create(&membership).Error; err != nil {
			return classifyDBError(err)
		}

		// Create a product scope for this membership
		scope := models.ProductMembershipScope{
			MembershipID: membership.MembershipID,
			ProductID:    product.ProductID,
			ScopeLevel:   "PRODUCT",
			IsActive:     true,
		}
		if err := tx.Create(&scope).Error; err != nil {
			return classifyDBError(err)
		}

		return nil
	})
	return product, err
}

func (u *usecase) ListProducts(actor Actor) ([]models.Product, error) {
	var products []models.Product
	q := u.repository.DB().Model(&models.Product{}).Preload("ProductEnvironments").Order("product_id")

	var isGlobalAdmin bool
	var pm models.PlatformMembership
	if err := u.repository.DB().Where("user_id = ? AND is_active = TRUE AND platform_role_id IN (1, 3)", actor.UserID).Limit(1).Find(&pm).Error; err == nil && pm.PlatformMembershipID != 0 {
		isGlobalAdmin = true
	}

	if !isGlobalAdmin {
		q = q.Where(`EXISTS (SELECT 1 FROM product_memberships pm JOIN product_membership_scopes pms ON pms.membership_id = pm.membership_id WHERE pm.product_id = products.product_id AND pm.user_id = ? AND pm.is_active = TRUE AND pms.is_active = TRUE AND (pm.expires_at IS NULL OR pm.expires_at > ?))`, actor.UserID, time.Now())
	}
	return products, q.Find(&products).Error
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
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "PRODUCT", "UPDATE"); err != nil {
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
	if req.ProductCode != nil {
		code := normalizeCode(*req.ProductCode)
		if code == "" {
			return nil, responses.ErrInvalid
		}
		updates["product_code"] = code
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
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "PRODUCT", "DELETE"); err != nil {
		return err
	}
	if err := u.repository.DB().Delete(&models.Product{}, productID).Error; err != nil {
		return classifyDBError(err)
	}
	return nil
}

// bulkdeleteproduct

func (u *usecase) BulkDeleteProducts(actor Actor, productID []int) error {
	for _, id := range productID {
		if err := u.authorize(actor, AccessTarget{ProductID: id}, "PRODUCT", "DELETE"); err != nil {
			return err
		}
	}
	if err := u.repository.DB().Delete(&models.Product{}, "product_id IN ?", productID).Error; err != nil {
		return classifyDBError(err)
	}
	return nil
}
