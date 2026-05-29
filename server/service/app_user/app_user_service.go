package app_user

import (
	"common/middleware/db"
	"fmt"
	"net/mail"
	appUserDTO "service/app_user/dto"
	appUserRepository "service/app_user/repository"
	grainConfigRepository "service/grain_config/repository"
	"strings"
	"time"

	"gorm.io/gorm"
)

type AppUserService struct {
	appUserRepository            *appUserRepository.AppUserRepository
	appUserLoginRecordRepository *appUserRepository.AppUserLoginRecordRepository
	stationUserRepository        *grainConfigRepository.GrainStationUserRepository
}

func NewAppUserService() *AppUserService {
	return &AppUserService{
		appUserRepository:            db.GetRepository[appUserRepository.AppUserRepository](),
		appUserLoginRecordRepository: db.GetRepository[appUserRepository.AppUserLoginRecordRepository](),
		stationUserRepository:        db.GetRepository[grainConfigRepository.GrainStationUserRepository](),
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

func (s *AppUserService) UpsertWechatUser(openUID, unionID, sessionKey string, userInfo appUserDTO.WechatUserInfoDTO) (*appUserRepository.AppUser, error) {
	openUID = strings.TrimSpace(openUID)
	if openUID == "" {
		return nil, fmt.Errorf("openUid is required")
	}
	if s.appUserRepository.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}

	now := time.Now()
	entity, err := s.appUserRepository.FindByOpenUID(openUID)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		name := strings.TrimSpace(userInfo.NickName)
		if name == "" {
			name = "微信用户"
		}
		entity = &appUserRepository.AppUser{
			Name:            name,
			Username:        "wx_" + openUID,
			Status:          "active",
			OpenUID:         openUID,
			UnionID:         strings.TrimSpace(unionID),
			WxSessionKey:    strings.TrimSpace(sessionKey),
			WxLastLoginTime: &now,
		}
		applyWechatUserInfo(entity, userInfo)
		return s.appUserRepository.Create(entity)
	}

	if entity.Active == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	if strings.TrimSpace(entity.Status) == "" {
		entity.Status = "active"
	}
	if strings.TrimSpace(userInfo.NickName) != "" {
		entity.Name = strings.TrimSpace(userInfo.NickName)
	}
	entity.OpenUID = openUID
	entity.UnionID = strings.TrimSpace(unionID)
	entity.WxSessionKey = strings.TrimSpace(sessionKey)
	entity.WxLastLoginTime = &now
	applyWechatUserInfo(entity, userInfo)
	return s.appUserRepository.SaveOrUpdate(entity)
}

func applyWechatUserInfo(entity *appUserRepository.AppUser, userInfo appUserDTO.WechatUserInfoDTO) {
	if entity == nil {
		return
	}
	if value := strings.TrimSpace(userInfo.NickName); value != "" {
		entity.WxNickname = value
	}
	if value := strings.TrimSpace(userInfo.AvatarURL); value != "" {
		entity.WxAvatar = value
	}
	if userInfo.Gender > 0 {
		entity.WxGender = userInfo.Gender
	}
	entity.WxCountry = strings.TrimSpace(userInfo.Country)
	entity.WxProvince = strings.TrimSpace(userInfo.Province)
	entity.WxCity = strings.TrimSpace(userInfo.City)
	entity.WxLanguage = strings.TrimSpace(userInfo.Language)
}
