package user

import (
	"fmt"
	appUserPassword "service/app_user/password"
	appUserRepository "service/app_user/repository"
	authService "service/manager_auth"
	permission "service/manager_permission"
	permissionRepository "service/manager_permission/repository"
	userPassword "service/manager_user/password"
	userRepository "service/manager_user/repository"
	"strings"
	"time"

	"gorm.io/gorm"
)

func (s *UserService) EnsureSalesmanUser(username, name, phone, rawPassword string) error {
	if s.userRepository.Db == nil || s.userRoleRepository.Db == nil || s.roleRepository.Db == nil {
		return fmt.Errorf("database is not initialized")
	}

	username = strings.TrimSpace(username)
	name = strings.TrimSpace(name)
	phone = strings.TrimSpace(phone)
	rawPassword = strings.TrimSpace(rawPassword)
	if username == "" {
		return fmt.Errorf("username is required")
	}
	if name == "" {
		name = username
	}
	if rawPassword != "" {
		if err := validateUserPassword(rawPassword); err != nil {
			return err
		}
	}

	entity, err := s.userRepository.FindByUsername(username)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
		entity = &userRepository.User{
			Name:          name,
			Username:      username,
			Phone:         phone,
			Status:        "active",
			LastLoginTime: func() *time.Time { t := time.Now(); return &t }(),
		}
		applySyncedUserPassword(entity, rawPassword)
		created, err := s.userRepository.Create(entity)
		if err != nil {
			return err
		}
		entity = created
	} else {
		entity.Name = name
		entity.Phone = phone
		if strings.TrimSpace(entity.Status) == "" {
			entity.Status = "active"
		}
		if entity.LastLoginTime == nil {
			t := time.Now()
			entity.LastLoginTime = &t
		}
		if rawPassword != "" {
			applySyncedUserPassword(entity, rawPassword)
		}
		saved, err := s.userRepository.SaveOrUpdate(entity)
		if err != nil {
			return err
		}
		entity = saved
	}

	return s.ensureSalesmanRole(uint64(entity.Id))
}

func (s *UserService) SyncSalesmanUserPassword(username, rawPassword string) error {
	if s.userRepository.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	username = strings.TrimSpace(username)
	rawPassword = strings.TrimSpace(rawPassword)
	if username == "" {
		return fmt.Errorf("username is required")
	}
	if err := validateUserPassword(rawPassword); err != nil {
		return err
	}

	entity, err := s.userRepository.FindByUsername(username)
	if err != nil {
		return err
	}
	if entity.Active == 0 {
		return gorm.ErrRecordNotFound
	}
	applySyncedUserPassword(entity, rawPassword)
	_, err = s.userRepository.SaveOrUpdate(entity)
	authService.ClearUserRoleCache(uint64(entity.Id))
	return err
}

func (s *UserService) saveUserPasswordAndSyncSalesmanAppUser(entity *userRepository.User, rawPassword string) error {
	if s.userRepository.Db == nil || s.userRoleRepository.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	if entity == nil {
		return fmt.Errorf("user is nil")
	}
	rawPassword = strings.TrimSpace(rawPassword)
	if err := validateUserPassword(rawPassword); err != nil {
		return err
	}

	return s.userRepository.Db.Transaction(func(tx *gorm.DB) error {
		hasSalesmanRole, err := userHasSalesmanRole(tx, entity)
		if err != nil {
			return err
		}

		applySyncedUserPassword(entity, rawPassword)
		entity.Init()
		if err := tx.Save(entity).Error; err != nil {
			return err
		}
		if !hasSalesmanRole {
			return nil
		}
		return syncAppUserPasswordByUsername(tx, entity.Username, rawPassword)
	})
}

func userHasSalesmanRole(tx *gorm.DB, entity *userRepository.User) (bool, error) {
	var count int64
	err := tx.Table("user_role").
		Joins("INNER JOIN `role` r ON r.id = user_role.role_id").
		Where("user_role.user_id = ? AND user_role.active = ? AND r.active = ? AND r.code = ?",
			uint64(entity.Id), 1, 1, string(permission.RoleCodeSalesman)).
		Count(&count).Error
	return count > 0, err
}

func syncAppUserPasswordByUsername(tx *gorm.DB, username, rawPassword string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("username is required")
	}

	var entity appUserRepository.AppUser
	err := tx.Where("username = ? AND active = ?", username, 1).First(&entity).Error
	if err == gorm.ErrRecordNotFound {
		return nil
	}
	if err != nil {
		return err
	}

	rawPassword = strings.TrimSpace(rawPassword)
	entity.OriginPassword = rawPassword
	entity.Password = appUserPassword.Encrypt(entity.Username, rawPassword)
	entity.Init()
	return tx.Save(&entity).Error
}

func (s *UserService) ensureSalesmanRole(userID uint64) error {
	if userID == 0 {
		return fmt.Errorf("userId must be positive")
	}
	role, err := s.findOrCreateSalesmanRole()
	if err != nil {
		return err
	}

	var existing userRepository.UserRole
	err = s.userRoleRepository.Db.
		Where("user_id = ? AND role_id = ? AND active = ?", userID, uint64(role.Id), 1).
		Order("id ASC").
		First(&existing).Error
	if err == nil && existing.Id > 0 {
		authService.ClearUserRoleCache(userID)
		return nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if _, err := s.userRoleRepository.Create(&userRepository.UserRole{
		UserID: userID,
		RoleID: uint64(role.Id),
	}); err != nil {
		return err
	}
	authService.ClearUserRoleCache(userID)
	return nil
}

func (s *UserService) findOrCreateSalesmanRole() (*permissionRepository.Role, error) {
	var role permissionRepository.Role
	err := s.roleRepository.Db.
		Where("code = ? AND active = ?", string(permission.RoleCodeSalesman), 1).
		Order("id ASC").
		First(&role).Error
	if err == nil && role.Id > 0 {
		return &role, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return s.roleRepository.Create(&permissionRepository.Role{
		Name: permission.RoleNameSalesman,
		Code: string(permission.RoleCodeSalesman),
	})
}

func applySyncedUserPassword(entity *userRepository.User, rawPassword string) {
	rawPassword = strings.TrimSpace(rawPassword)
	if rawPassword == "" {
		return
	}
	entity.OriginPassword = rawPassword
	entity.Password = userPassword.Encrypt(entity.Username, rawPassword)
}
