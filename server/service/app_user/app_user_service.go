package app_user

import (
	"common/middleware/db"
	"fmt"
	"net/mail"
	appUserRepository "service/app_user/repository"
	grainConfigRepository "service/grain_config/repository"
	managerUserService "service/manager_user"
	"strings"

	"gorm.io/gorm"
)

type AppUserService struct {
	appUserRepository            *appUserRepository.AppUserRepository
	appUserLoginRecordRepository *appUserRepository.AppUserLoginRecordRepository
	stationUserRepository        *grainConfigRepository.GrainStationUserRepository
	managerUserService           *managerUserService.UserService
}

func NewAppUserService() *AppUserService {
	return &AppUserService{
		appUserRepository:            db.GetRepository[appUserRepository.AppUserRepository](),
		appUserLoginRecordRepository: db.GetRepository[appUserRepository.AppUserLoginRecordRepository](),
		stationUserRepository:        db.GetRepository[grainConfigRepository.GrainStationUserRepository](),
		managerUserService:           managerUserService.NewUserService(),
	}
}

func (s *AppUserService) EnsureTable() error {
	if err := s.appUserRepository.EnsureTable(); err != nil {
		return err
	}
	if err := s.appUserLoginRecordRepository.EnsureTable(); err != nil {
		return err
	}
	if err := s.stationUserRepository.EnsureTable(); err != nil {
		return err
	}
	if err := s.managerUserService.EnsureTable(); err != nil {
		return err
	}
	return nil
}

func normalizeAppUserPage(page, pageIndex, pageSize int) (int, int) {
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

func normalizeAppUserStatus(status string) string {
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

func validateAppUserEmail(email string) error {
	if email == "" {
		return nil
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("email format is invalid")
	}
	return nil
}

func validateAppUserPassword(password string) error {
	if len(strings.TrimSpace(password)) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}
	return nil
}

func ensureAppUserExists(repo *appUserRepository.AppUserRepository, appUserID uint64) error {
	if appUserID == 0 {
		return fmt.Errorf("appUserId must be positive")
	}
	entity, err := repo.FindById(uint(appUserID))
	if err != nil {
		return err
	}
	if entity.Active == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *AppUserService) FindActiveUserByPhone(phone string) (*appUserRepository.AppUser, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return nil, fmt.Errorf("手机号不能为空")
	}
	user, err := s.appUserRepository.FindByPhone(phone)
	if err != nil {
		return nil, err
	}
	if user.Active == 0 || !strings.EqualFold(strings.TrimSpace(user.Status), "active") {
		return nil, gorm.ErrRecordNotFound
	}
	return user, nil
}

func (s *AppUserService) FindActiveWechatUser(openUID string) (*appUserRepository.AppUser, error) {
	openUID = strings.TrimSpace(openUID)
	if openUID == "" {
		return nil, fmt.Errorf("openUid is required")
	}
	user, err := s.appUserRepository.FindByOpenUID(openUID)
	if err != nil {
		return nil, err
	}
	if user.Active == 0 || !strings.EqualFold(strings.TrimSpace(user.Status), "active") {
		return nil, gorm.ErrRecordNotFound
	}
	return user, nil
}
