package user

import (
	baseDTO "common/base/dto"
	"common/middleware/db"
	"fmt"
	authService "service/manager_auth"
	permissionRepository "service/manager_permission/repository"
	userDTO "service/manager_user/dto"
	userRepository "service/manager_user/repository"
	"strings"

	"gorm.io/gorm"
)

func (s *UserService) ListUserRoles(query userDTO.UserRoleQueryDTO) (*baseDTO.PageDTO[userDTO.UserRoleDTO], error) {
	if s.userRoleRepository.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	pageIndex, pageSize := normalizeUserPage(query.Page, query.PageIndex, query.PageSize)
	dbQuery := s.userRoleRepository.Db.Model(&userRepository.UserRole{}).Where("active = ?", 1)
	if query.UserID > 0 {
		dbQuery = dbQuery.Where("user_id = ?", query.UserID)
	}
	if query.RoleID > 0 {
		dbQuery = dbQuery.Where("role_id = ?", query.RoleID)
	}
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, err
	}
	var entities []*userRepository.UserRole
	if err := dbQuery.Order("id DESC").Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&entities).Error; err != nil {
		return nil, err
	}
	return baseDTO.BuildPage(int(total), db.ToDTOs[userDTO.UserRoleDTO](entities)), nil
}

func (s *UserService) GetUserRoleByID(id uint) (*userDTO.UserRoleDTO, error) {
	if s.userRoleRepository.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	entity, err := s.userRoleRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	if entity.Active == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return db.ToDTO[userDTO.UserRoleDTO](entity), nil
}

func (s *UserService) CreateUserRole(req *userDTO.CreateUserRoleDTO) (*userDTO.UserRoleDTO, error) {
	if s.userRoleRepository.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	if err := ensureUserExists(s.userRepository, req.UserID); err != nil {
		return nil, err
	}
	if err := s.userRoleRepository.Db.
		Where("user_id = ?", req.UserID).
		Delete(&userRepository.UserRole{}).Error; err != nil {
		return nil, err
	}
	newEntity := &userRepository.UserRole{
		UserID: req.UserID,
		RoleID: req.RoleID,
	}
	newEntity.Active = 1
	created, err := s.userRoleRepository.Create(newEntity)
	if err != nil {
		return nil, err
	}
	authService.ClearUserRoleCache(req.UserID)
	return db.ToDTO[userDTO.UserRoleDTO](created), nil
}

func (s *UserService) UpdateUserRole(id uint, req *userDTO.UpdateUserRoleDTO) (*userDTO.UserRoleDTO, error) {
	if s.userRoleRepository.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	entity, err := s.userRoleRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	if entity.Active == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	if req.UserID != nil {
		if err := ensureUserExists(s.userRepository, *req.UserID); err != nil {
			return nil, err
		}
		authService.ClearUserRoleCache(entity.UserID)
		entity.UserID = *req.UserID
	}
	if req.RoleID != nil {
		entity.RoleID = *req.RoleID
	}
	saved, err := s.userRoleRepository.SaveOrUpdate(entity)
	if err != nil {
		return nil, err
	}
	authService.ClearUserRoleCache(entity.UserID)
	return db.ToDTO[userDTO.UserRoleDTO](saved), nil
}

func (s *UserService) SetUserRoles(req *userDTO.SetUserRolesDTO) error {
	if s.userRoleRepository.Db == nil || s.roleRepository.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	if req == nil {
		return fmt.Errorf("request is nil")
	}
	if err := ensureUserExists(s.userRepository, req.UserID); err != nil {
		return err
	}
	if err := s.userRoleRepository.Db.
		Where("user_id = ?", req.UserID).
		Delete(&userRepository.UserRole{}).Error; err != nil {
		return err
	}
	for _, code := range req.RoleCodes {
		code = strings.ToLower(strings.TrimSpace(code))
		if code == "" {
			continue
		}
		var role permissionRepository.Role
		if err := s.roleRepository.Db.Where("code = ? AND active = ?", code, 1).First(&role).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("role '%s' not found", code)
			}
			return err
		}
		newEntity := &userRepository.UserRole{
			UserID: req.UserID,
			RoleID: uint64(role.Id),
		}
		newEntity.Active = 1
		if _, err := s.userRoleRepository.Create(newEntity); err != nil {
			return err
		}
	}
	authService.ClearUserRoleCache(req.UserID)
	return nil
}

func (s *UserService) DeleteUserRole(id uint) error {
	if s.userRoleRepository.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	entity, err := s.userRoleRepository.FindById(id)
	if err != nil {
		return err
	}
	if entity.Active == 0 {
		return gorm.ErrRecordNotFound
	}
	entity.Active = 0
	_, err = s.userRoleRepository.SaveOrUpdate(entity)
	authService.ClearUserRoleCache(entity.UserID)
	return err
}
