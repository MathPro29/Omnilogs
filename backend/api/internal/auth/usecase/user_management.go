package usecase

import (
	"errors"
	"strings"

	"omnilogs-api/dto"
	"omnilogs-api/models"
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

func (u *usecase) EditUserRole(userID uint, req dto.EditUserRoleRequest) error {
	// 1. Check if user exists
	_, err := u.repo.FindByID(req.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return responses.ErrorUserCode["USER_NOT_FOUND"]
	}
	if err != nil {
		return err
	}

	// 2. Platform role check
	if req.PlatformRoleID != 0 {
		if req.PlatformRoleID < 1 || req.PlatformRoleID > 4 {
			return ErrInvalidRoleAssignment
		}

		actor, err := u.repo.FindByID(userID)
		if err != nil {
			return err
		}
		actorRole := getUserRoleName(actor)

		switch actorRole {
		case "god":
			// GOD may assign any platform role.
		case "owner":
			if req.PlatformRoleID == 1 { // Cannot assign GOD role.
				return ErrForbiddenRoleAssignment
			}
			allowed, err := u.repo.CanOwnerManageUser(userID, req.ID)
			if err != nil {
				return err
			}
			if !allowed {
				return ErrForbiddenRoleAssignment
			}
		default:
			return ErrForbiddenRoleAssignment
		}

		if err := u.repo.UpdatePlatformRole(req.ID, req.PlatformRoleID); err != nil {
			return err
		}
	}

	// 3. Update user fields if provided
	if req.FirstName != "" {
		if err := u.repo.UpdateUserField(req.ID, "first_name", req.FirstName); err != nil {
			return err
		}
	}
	if req.LastName != "" {
		if err := u.repo.UpdateUserField(req.ID, "last_name", req.LastName); err != nil {
			return err
		}
	}
	if req.Email != "" {
		if err := u.repo.UpdateUserField(req.ID, "email", req.Email); err != nil {
			return err
		}
	}
	if req.IsActive != nil {
		if err := u.repo.UpdateUserField(req.ID, "is_active", *req.IsActive); err != nil {
			return err
		}
	}

	// 4. Permission level update
	if req.PermissionLevel > 0 {
		if req.ProductRoleID > 0 || req.EnvironmentID > 0 {
			if req.ProductRoleID == 0 || req.EnvironmentID == 0 {
				return ErrInvalidRoleAssignment
			}

			productRole, err := u.repo.FindProductRoleByID(req.ProductRoleID)
			if err != nil {
				return ErrInvalidRoleAssignment
			}

			productID := uint(productRole.ProductID)

			permMap := make(map[string]any)
			for _, p := range productRole.Permissions {
				permMap[p.ResourceType+"."+p.Action] = true
			}
			permMap["level"] = req.PermissionLevel

			userPerm, err := u.repo.GetUserPermissionByID(req.ID, productID, req.EnvironmentID)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := u.repo.CreateUserPermission(req.ID, productID, req.EnvironmentID, permMap); err != nil {
					return err
				}
				return nil
			}
			if err != nil {
				return err
			}

			return u.repo.UpdateUserPermission(uint(userPerm.UserID), uint(userPerm.ProductID), uint(userPerm.EnvironmentID), permMap)
		}
	}
	return nil
}

func (u *usecase) CreateUser(req dto.CreateUserRequest) (*dto.UserResponse, error) {
	if req.Username != nil && *req.Username != "" {
		if err := u.CheckUserExistsByUsername(*req.Username); err != nil {
			return nil, err
		}
	}

	if _, err := u.repo.FindByEmail(req.Email); err == nil {
		return nil, responses.ErrorUserCode["EMAIL_ALREADY_EXISTS"]
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	passwordHash, err := u.passwordService.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	parts := strings.SplitN(req.FullName, " ", 2)
	firstName := parts[0]
	var lastName *string
	if len(parts) > 1 {
		lastName = &parts[1]
	}

	user := &models.User{
		Username:     req.Username,
		FirstName:    &firstName,
		LastName:     lastName,
		Email:        req.Email,
		PhoneNumber:  req.PhoneNumber,
		PasswordHash: passwordHash,
		IsActive:     true,
	}

	if err := u.repo.Create(user); err != nil {
		return nil, err
	}

	// Update platform role if provided and not default
	if req.Role != "" {
		pRole, err := u.repo.FindPlatformRoleByCode(req.Role)
		if err == nil {
			if err := u.repo.UpdatePlatformRole(uint(user.UserID), uint(pRole.PlatformRoleID)); err != nil {
				return nil, err
			}
		}
	}

	// Reload user to populate role information
	reloadedUser, err := u.repo.FindByID(uint(user.UserID))
	if err != nil {
		return nil, err
	}

	return toUserResponse(reloadedUser), nil
}

func (u *usecase) UpdateUser(userID uint, req dto.UpdateUserAdminRequest) (*dto.UserResponse, error) {
	user, err := u.repo.FindByID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, responses.ErrorUserCode["USER_NOT_FOUND"]
	}
	if err != nil {
		return nil, err
	}

	// 1. Update FullName -> Split to FirstName & LastName
	if req.FullName != "" {
		parts := strings.SplitN(req.FullName, " ", 2)
		firstName := parts[0]
		if err := u.repo.UpdateUserField(userID, "first_name", firstName); err != nil {
			return nil, err
		}
		if len(parts) > 1 {
			lastName := parts[1]
			if err := u.repo.UpdateUserField(userID, "last_name", lastName); err != nil {
				return nil, err
			}
		} else {
			if err := u.repo.UpdateUserField(userID, "last_name", nil); err != nil {
				return nil, err
			}
		}
	}

	// 2. Update Email
	if req.Email != "" && req.Email != user.Email {
		// Check if email already exists
		if _, err := u.repo.FindByEmail(req.Email); err == nil {
			return nil, responses.ErrorUserCode["EMAIL_ALREADY_EXISTS"]
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if err := u.repo.UpdateUserField(userID, "email", req.Email); err != nil {
			return nil, err
		}
	}

	// 3. Update PhoneNumber
	if req.PhoneNumber != nil {
		if err := u.repo.UpdateUserField(userID, "phone_number", *req.PhoneNumber); err != nil {
			return nil, err
		}
	}

	// 4. Update Status (IsActive)
	if req.Status != "" {
		isActive := req.Status == "active"
		if err := u.repo.UpdateUserField(userID, "is_active", isActive); err != nil {
			return nil, err
		}
	}

	// 5. Update Platform Role
	if req.Role != "" {
		pRole, err := u.repo.FindPlatformRoleByCode(req.Role)
		if err == nil {
			if err := u.repo.UpdatePlatformRole(userID, uint(pRole.PlatformRoleID)); err != nil {
				return nil, err
			}
		}
	}

	// Reload user to populate role information
	reloadedUser, err := u.repo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	return toUserResponse(reloadedUser), nil
}

func (u *usecase) GetUserByID(userID uint) (*dto.UserResponse, error) {
	user, err := u.repo.FindByID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, responses.ErrorUserCode["USER_NOT_FOUND"]
	}
	if err != nil {
		return nil, err
	}
	return toUserResponse(user), nil
}

func (u *usecase) DeleteUserByID(userID uint) error {
	// Check if user exists
	_, err := u.repo.FindByID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return responses.ErrorUserCode["USER_NOT_FOUND"]
	}
	if err != nil {
		return err
	}

	// Delete user
	if err := u.repo.DeleteUserByID(userID); err != nil {
		return err
	}
	return nil
}
