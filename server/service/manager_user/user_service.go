package user

import (
	"common/middleware/db"
	"fmt"
	"net/mail"
	permission "service/manager_permission"
	permissionRepository "service/manager_permission/repository"
	userRepository "service/manager_user/repository"
	"strings"

	"gorm.io/gorm"
)

type UserService struct {
	userRepository            *userRepository.UserRepository
	userLoginRecordRepository *userRepository.UserLoginRecordRepository
	userRoleRepository        *userRepository.UserRoleRepository
	userTenantRepository      *userRepository.UserTenantRepository
	roleRepository            *permissionRepository.RoleRepository
}

func NewUserService() *UserService {
	return &UserService{
		userRepository:            db.GetRepository[userRepository.UserRepository](),
		userLoginRecordRepository: db.GetRepository[userRepository.UserLoginRecordRepository](),
		userRoleRepository:        db.GetRepository[userRepository.UserRoleRepository](),
		userTenantRepository:      db.GetRepository[userRepository.UserTenantRepository](),
		roleRepository:            db.GetRepository[permissionRepository.RoleRepository](),
	}
}

func (s *UserService) EnsureTable() error {
	if err := s.userRepository.EnsureTable(); err != nil {
		return err
	}
	if err := s.userLoginRecordRepository.EnsureTable(); err != nil {
		return err
	}
	if err := s.userRoleRepository.EnsureTable(); err != nil {
		return err
	}
	if err := s.userTenantRepository.EnsureTable(); err != nil {
		return err
	}
	if err := s.roleRepository.EnsureTable(); err != nil {
		return err
	}
	return nil
}

func normalizeUserPage(page, pageIndex, pageSize int) (int, int) {
	if pageIndex <= 0 {
		pageIndex = page
	}
	if pageIndex <= 0 {
		pageIndex = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return pageIndex, pageSize
}

func normalizeUserStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "", "active":
		return "active"
	case "inactive":
		return "inactive"
	case "locked":
		return "locked"
	default:
		return ""
	}
}

func normalizeUserRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "", "member":
		return "member"
	case "admin":
		return "admin"
	case "manager":
		return "manager"
	case "auditor":
		return "auditor"
	case string(permission.RoleCodeSalesman):
		return string(permission.RoleCodeSalesman)
	default:
		return ""
	}
}

func validateEmail(email string) error {
	if email == "" {
		return nil
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("email format is invalid")
	}
	return nil
}

func validateUserPassword(password string) error {
	password = strings.TrimSpace(password)
	if len(password) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}
	if len(password) > 50 {
		return fmt.Errorf("password must be at most 50 characters")
	}
	return nil
}

func ensureUserExists(repo *userRepository.UserRepository, userID uint64) error {
	if userID == 0 {
		return fmt.Errorf("userId must be positive")
	}
	entity, err := repo.FindById(uint(userID))
	if err != nil {
		return err
	}
	if entity.Active == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
