package usecase

import (
	"strconv"
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
	product := &models.Product{
		ProductName: strings.TrimSpace(req.ProductName),
		ProductCode: "AUTO_" + strconv.FormatInt(time.Now().UnixNano(), 10),
		Description: strings.TrimSpace(req.Description),
		SetupStatus: "DRAFT",
		IsActive:    true,
	}
	if product.ProductName == "" {
		return nil, responses.ErrInvalid
	}

	// Environments are optional.  A product starts without environments unless
	// the caller explicitly supplies them.
	envsToCreate := req.Environments

	err := u.repository.Transaction(func(tx *gorm.DB) error {
		var duplicate models.Product
		if err := tx.Where("LOWER(BTRIM(product_name)) = ?", utils.NormalizeName(product.ProductName)).First(&duplicate).Error; err == nil {
			return &responses.ScopedConflictError{Field: "Product name", Value: product.ProductName, Scope: "active products", Cause: responses.ErrConflict}
		}
		if err := tx.Create(product).Error; err != nil {
			return classifyDBError(err)
		}
		product.ProductCode = strconv.Itoa(product.ProductID)
		if err := tx.Model(product).Update("product_code", product.ProductCode).Error; err != nil {
			return classifyDBError(err)
		}
		for _, environmentRequest := range envsToCreate {
			code := normalizeCode(environmentRequest.EnvironmentCode)
			name := strings.TrimSpace(environmentRequest.EnvironmentName)
			if code == "" || name == "" {
				return responses.ErrInvalid
			}
			envType := "DEVELOPMENT"
			if strings.TrimSpace(environmentRequest.EnvironmentType) != "" {
				envType = strings.ToUpper(strings.TrimSpace(environmentRequest.EnvironmentType))
			}

			environment := models.ProductEnvironment{
				ProductID:       product.ProductID,
				EnvironmentCode: code,
				EnvironmentName: name,
				EnvironmentType: envType,
				IsDefault:       false,
				IsActive:        true,
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
		resources := []string{"PRODUCT", "PROJECT", "ENVIRONMENT", "FEATURE", "CATEGORY", "LOG", "AUDIT", "API_KEY", "ROLE", "ACCESS"}
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

		// Assign the selected owner (or creator) to the Owner role of this product
		targetOwnerID := actor.UserID
		if req.OwnerID != nil && *req.OwnerID > 0 {
			targetOwnerID = *req.OwnerID
		}

		membership := models.ProductMembership{
			UserID:    targetOwnerID,
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

		// If creator is different from target owner, also grant creator access as owner
		if actor.UserID != targetOwnerID {
			creatorMembership := models.ProductMembership{
				UserID:    actor.UserID,
				ProductID: product.ProductID,
				RoleID:    ownerRole.RoleID,
				IsActive:  true,
			}
			if err := tx.Create(&creatorMembership).Error; err == nil {
				creatorScope := models.ProductMembershipScope{
					MembershipID: creatorMembership.MembershipID,
					ProductID:    product.ProductID,
					ScopeLevel:   "PRODUCT",
					IsActive:     true,
				}
				_ = tx.Create(&creatorScope).Error
			}
		}

		return nil
	})
	return product, err
}
