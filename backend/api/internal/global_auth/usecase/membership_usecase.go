package usecase

import (
	"omnilogs-api/dto"
	"omnilogs-api/models"
	"omnilogs-api/responses"
	"strings"

	"gorm.io/gorm"
)

func (u *usecase) CreateMembership(actor Actor, productID int, req dto.CreateProductMembershipRequest) (*models.ProductMembership, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ACCESS", "GRANT"); err != nil {
		return nil, err
	}

	// Check if already a member
	var pm models.ProductMembership
	err := u.repository.DB().Where("user_id = ? AND product_id = ?", req.UserID, productID).Limit(1).Find(&pm).Error
	if err != nil {
		return nil, classifyDBError(err)
	}

	isMember := pm.MembershipID != 0
	if isMember {
		if req.RoleID > 0 {
			if !u.exists(&models.ProductRole{}, "role_id = ? AND product_id = ?", req.RoleID, productID) {
				return nil, responses.ErrInvalid
			}
			pm.RoleID = req.RoleID
		}
		if req.ExpiresAt != nil {
			pm.ExpiresAt = req.ExpiresAt
		}
		pm.IsActive = true
		if err := u.repository.DB().Save(&pm).Error; err != nil {
			return nil, classifyDBError(err)
		}
		return &pm, nil
	} else {
		// New member: MUST specify role_id and product_id
		if req.RoleID <= 0 || req.ProductID <= 0 || req.ProductID != productID {
			return nil, responses.ErrInvalid
		}
		if !u.exists(&models.User{}, "user_id = ? AND deleted_at IS NULL", req.UserID) {
			return nil, responses.ErrInvalid
		}
		if !u.exists(&models.ProductRole{}, "role_id = ? AND product_id = ?", req.RoleID, productID) {
			return nil, responses.ErrInvalid
		}

		value := &models.ProductMembership{
			UserID:    req.UserID,
			ProductID: productID,
			RoleID:    req.RoleID,
			ExpiresAt: req.ExpiresAt,
			IsActive:  true,
		}
		if err := u.repository.DB().Create(value).Error; err != nil {
			return nil, classifyDBError(err)
		}
		return value, nil
	}
}

func (u *usecase) CreateMemberships(actor Actor, productID int, req dto.CreateBulkProductMembershipRequest) ([]models.ProductMembership, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ACCESS", "GRANT"); err != nil {
		return nil, err
	}
	if len(req.Memberships) == 0 {
		return nil, responses.ErrInvalid
	}

	values := make([]models.ProductMembership, 0, len(req.Memberships))
	err := u.repository.Transaction(func(tx *gorm.DB) error {
		for _, item := range req.Memberships {
			var pm models.ProductMembership
			err := tx.Where("user_id = ? AND product_id = ?", item.UserID, productID).Limit(1).Find(&pm).Error
			if err != nil {
				return classifyDBError(err)
			}
			isMember := pm.MembershipID != 0

			if isMember {
				if item.RoleID > 0 {
					if !existsDB(tx, &models.ProductRole{}, "role_id = ? AND product_id = ?", item.RoleID, productID) {
						return responses.ErrInvalid
					}
					pm.RoleID = item.RoleID
				}
				if item.ExpiresAt != nil {
					pm.ExpiresAt = item.ExpiresAt
				}
				pm.IsActive = true
				if err := tx.Save(&pm).Error; err != nil {
					return classifyDBError(err)
				}

				// Ensure scope exists
				var scope models.ProductMembershipScope
				err = tx.Where("membership_id = ? AND product_id = ? AND scope_level = 'PRODUCT'", pm.MembershipID, productID).Limit(1).Find(&scope).Error
				if err != nil {
					return classifyDBError(err)
				}
				if scope.ScopeID == 0 {
					newScope := models.ProductMembershipScope{
						MembershipID: pm.MembershipID,
						ProductID:    productID,
						ScopeLevel:   "PRODUCT",
						IsActive:     true,
					}
					if err := tx.Create(&newScope).Error; err != nil {
						return classifyDBError(err)
					}
				}

				values = append(values, pm)
			} else {
				// New member: MUST specify role_id and product_id
				if item.RoleID <= 0 || item.ProductID <= 0 || item.ProductID != productID {
					return responses.ErrInvalid
				}
				if !existsDB(tx, &models.User{}, "user_id = ? AND deleted_at IS NULL", item.UserID) {
					return responses.ErrInvalid
				}
				if !existsDB(tx, &models.ProductRole{}, "role_id = ? AND product_id = ?", item.RoleID, productID) {
					return responses.ErrInvalid
				}

				value := models.ProductMembership{
					UserID:    item.UserID,
					ProductID: productID,
					RoleID:    item.RoleID,
					ExpiresAt: item.ExpiresAt,
					IsActive:  true,
				}
				if err := tx.Create(&value).Error; err != nil {
					return classifyDBError(err)
				}

				scope := models.ProductMembershipScope{
					MembershipID: value.MembershipID,
					ProductID:    productID,
					ScopeLevel:   "PRODUCT",
					IsActive:     true,
				}
				if err := tx.Create(&scope).Error; err != nil {
					return classifyDBError(err)
				}

				values = append(values, value)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return values, nil
}

func (u *usecase) ListMemberships(actor Actor, productID int) ([]models.ProductMembership, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ACCESS", "READ"); err != nil {
		return nil, err
	}
	var values []models.ProductMembership
	return values, u.repository.DB().Where("product_id = ?", productID).Order("membership_id").Find(&values).Error
}

func (u *usecase) UpdateMembership(actor Actor, productID, id int, req dto.UpdateProductMembershipRequest) (*models.ProductMembership, error) {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ACCESS", "GRANT"); err != nil {
		return nil, err
	}
	updates := map[string]any{}
	if req.RoleID != nil {
		if !u.exists(&models.ProductRole{}, "role_id = ? AND product_id = ?", *req.RoleID, productID) {
			return nil, responses.ErrInvalid
		}
		updates["role_id"] = *req.RoleID
	}
	if req.ExpiresAt != nil {
		updates["expires_at"] = req.ExpiresAt
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	res := u.repository.DB().Model(&models.ProductMembership{}).Where("membership_id = ? AND product_id = ?", id, productID).Updates(updates)
	if res.Error != nil {
		return nil, classifyDBError(res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, responses.ErrNotFound
	}
	var value models.ProductMembership
	if err := u.repository.DB().Where("membership_id = ? AND product_id = ?", id, productID).First(&value).Error; err != nil {
		return nil, classifyDBError(err)
	}
	return &value, nil
}

func (u *usecase) DeleteMembership(actor Actor, productID, id int) error {
	if err := u.authorize(actor, AccessTarget{ProductID: productID}, "ACCESS", "GRANT"); err != nil {
		return err
	}
	return u.repository.Transaction(func(tx *gorm.DB) error {
		var membership models.ProductMembership
		if err := tx.Where("membership_id = ? AND product_id = ?", id, productID).First(&membership).Error; err != nil {
			return classifyDBError(err)
		}
		if err := tx.Where("membership_id = ? AND product_id = ?", id, productID).Delete(&models.ProductMembershipScope{}).Error; err != nil {
			return classifyDBError(err)
		}
		if err := tx.Delete(&membership).Error; err != nil {
			return classifyDBError(err)
		}
		return nil
	})
}

func roleRequiresProductScope(roleCode string) bool {
	return strings.EqualFold(roleCode, "owner")
}
