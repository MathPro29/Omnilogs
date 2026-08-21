package usecase

import (
	"time"

	"omnilogs-api/models"
	"omnilogs-api/responses"
)

func (u *usecase) requireProductOwner(actor Actor, productID int) error {
	if actor.PlatformAdmin {
		return nil
	}
	var membership models.ProductMembership
	err := u.repository.DB().Table("product_memberships AS pm").
		Joins("JOIN product_roles AS pr ON pr.role_id = pm.role_id").
		Joins("JOIN product_membership_scopes AS pms ON pms.membership_id = pm.membership_id AND pms.is_active = TRUE").
		Where("pm.user_id = ? AND pm.product_id = ? AND pm.is_active = TRUE AND LOWER(pr.role_code) = 'owner' AND (pm.expires_at IS NULL OR pm.expires_at > ?)", actor.UserID, productID, time.Now()).
		Limit(1).Find(&membership).Error
	if err != nil {
		return err
	}
	if membership.MembershipID == 0 {
		return responses.ErrForbidden
	}
	return nil
}
