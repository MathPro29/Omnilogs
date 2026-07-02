package usecase

import (
	"errors"

	"omnilogs-api/dto"
	"omnilogs-api/responses"

	"gorm.io/gorm"
)

func (u *usecase) GetMe(userID uint) (*dto.UserResponse, error) {
	user, err := u.repo.FindByID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, responses.ErrorUserCode["USER_NOT_FOUND"]
	}
	if err != nil {
		return nil, err
	}
	return toUserResponse(user), nil
}

func (u *usecase) ListAllUsers() ([]dto.UserResponse, error) {
	users, err := u.repo.ListAllUsers()
	if err != nil {
		return nil, err
	}
	var userResponses []dto.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, *toUserResponse(&user))
	}
	return userResponses, nil
}

func (u *usecase) GiveAdminAccess(userID uint, adminID uint, roleID uint) (*dto.GiveAdminAccessResponse, error) {
	actor, err := u.repo.FindByID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, responses.ErrorUserCode["USER_NOT_FOUND"]
	}
	if err != nil {
		return nil, err
	}

	target, err := u.repo.FindByID(adminID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, responses.ErrorUserCode["USER_NOT_FOUND"]
	}
	if err != nil {
		return nil, err
	}

	actorRole := getUserRoleName(actor)
	switch actorRole {
	case "god":
		// GOD may assign any platform role.
	case "owner":
		if roleID == 1 {
			return nil, ErrForbiddenRoleAssignment
		}
		allowed, err := u.repo.CanOwnerManageUser(userID, adminID)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, ErrForbiddenRoleAssignment
		}
	case "superadmin":
		return nil, ErrForbiddenRoleAssignment
	default:
		return nil, ErrForbiddenRoleAssignment
	}

	previousRoleID := target.RoleID
	previousRoleCode := getUserRoleName(target)

	newRole, err := u.repo.FindPlatformRoleByID(roleID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidRoleAssignment
	}
	if err != nil {
		return nil, err
	}

	target.RoleID = roleID
	if err := u.repo.UpdateRoleID(uint(target.UserID), roleID); err != nil {
		return nil, err
	}

	return &dto.GiveAdminAccessResponse{
		UserID:           uint(target.UserID),
		StorageTable:     "platform_memberships",
		PreviousRoleID:   previousRoleID,
		PreviousRoleCode: previousRoleCode,
		NewRoleID:        uint(newRole.PlatformRoleID),
		NewRoleCode:      newRole.RoleCode,
		Message:          "platform role updated successfully",
	}, nil
}

func (u *usecase) CheckUserExistsByUsername(username string) error {
	if _, err := u.repo.FindByUsername(username); err == nil {
		return responses.ErrorUserCode["USERNAME_ALREADY_EXISTS"]
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return nil
}
