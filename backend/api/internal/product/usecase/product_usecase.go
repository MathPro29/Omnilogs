package usecase

import (
	"encoding/json"
	"strings"
	"time"

	"omnilogs-api/dto"
	"omnilogs-api/models"

	"gorm.io/gorm"
)

func (u *usecase) CreateProduct(actor Actor, req dto.CreateProductRequest) (*models.Product, error) {
	if !actor.PlatformAdmin {
		return nil, ErrForbidden
	}
	product := &models.Product{ProductName: strings.TrimSpace(req.ProductName), ProductCode: normalizeCode(req.ProductCode), IsActive: true}
	if product.ProductName == "" || product.ProductCode == "" {
		return nil, ErrInvalid
	}
	err := u.repository.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(product).Error; err != nil {
			return classifyDBError(err)
		}
		for _, environmentRequest := range req.Environments {
			if environmentRequest.ProductID != nil && *environmentRequest.ProductID != product.ProductID {
				return ErrInvalid
			}
			environment := models.ProductEnvironment{
				ProductID:       product.ProductID,
				EnvironmentCode: normalizeCode(environmentRequest.EnvironmentCode),
				EnvironmentName: strings.TrimSpace(environmentRequest.EnvironmentName),
			}
			if environment.EnvironmentCode == "" || environment.EnvironmentName == "" {
				return ErrInvalid
			}
			if err := tx.Create(&environment).Error; err != nil {
				return classifyDBError(err)
			}
			product.ProductEnvironments = append(product.ProductEnvironments, environment)
		}
		role := models.ProductRole{ProductID: product.ProductID, RoleCode: "product_owner", RoleName: "Product Owner", Permissions: json.RawMessage(`{"all":true}`)}
		if err := tx.Create(&role).Error; err != nil {
			return classifyDBError(err)
		}
		membership := models.ProductMembership{UserID: actor.UserID, ProductID: product.ProductID, RoleID: role.RoleID, IsActive: true}
		if err := tx.Create(&membership).Error; err != nil {
			return classifyDBError(err)
		}
		scope := models.ProductMembershipScope{MembershipID: membership.MembershipID, ProductID: product.ProductID, ScopeLevel: "PRODUCT", IsActive: true}
		return tx.Create(&scope).Error
	})
	return product, err
}

func (u *usecase) ListProducts(actor Actor) ([]models.Product, error) {
	var products []models.Product
	q := u.repository.DB().Model(&models.Product{}).Preload("ProductEnvironments").Order("product_id")
	if !actor.PlatformAdmin {
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
			return nil, ErrInvalid
		}
		updates["product_name"] = name
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
