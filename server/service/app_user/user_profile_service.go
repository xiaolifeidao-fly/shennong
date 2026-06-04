package app_user

import (
	baseDTO "common/base/dto"
	"common/middleware/db"
	"fmt"
	appUserDTO "service/app_user/dto"
	appUserPassword "service/app_user/password"
	appUserRepository "service/app_user/repository"
	grainConfigRepository "service/grain_config/repository"
	"strings"

	"gorm.io/gorm"
)

func (s *AppUserService) GetUserStats() (*appUserDTO.AppUserStatsDTO, error) {
	visibleUsers, err := s.appUserRepository.CountVisibleUsers()
	if err != nil {
		return nil, err
	}
	activeUsers, err := s.appUserRepository.CountActiveUsers()
	if err != nil {
		return nil, err
	}
	recentLoginUsers, err := s.appUserRepository.CountRecentLoginUsers()
	if err != nil {
		return nil, err
	}
	return &appUserDTO.AppUserStatsDTO{
		VisibleUsers:     int(visibleUsers),
		RecentLoginUsers: int(recentLoginUsers),
		ActiveUsers:      int(activeUsers),
	}, nil
}

func (s *AppUserService) GetCurrentUserProfile(id uint) (*appUserDTO.CurrentAppUserProfileDTO, error) {
	entity, err := s.appUserRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	if entity.Active == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return toCurrentAppUserProfileDTO(entity), nil
}

func (s *AppUserService) UpdateCurrentUserProfile(id uint, req *appUserDTO.UpdateCurrentAppUserProfileDTO) (*appUserDTO.CurrentAppUserProfileDTO, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	entity, err := s.appUserRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	if entity.Active == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	name := strings.TrimSpace(req.Name)
	wxNickname := strings.TrimSpace(req.WxNickname)
	if name == "" {
		name = wxNickname
	}
	if name == "" {
		name = entity.Name
	}
	email := strings.TrimSpace(req.Email)
	if err := validateAppUserEmail(email); err != nil {
		return nil, err
	}

	entity.Name = name
	entity.Email = email
	entity.Phone = strings.TrimSpace(req.Phone)
	entity.Department = strings.TrimSpace(req.Department)
	entity.Remark = strings.TrimSpace(req.Remark)
	if wxNickname != "" {
		entity.WxNickname = wxNickname
	}
	if wxAvatar := strings.TrimSpace(req.WxAvatar); wxAvatar != "" {
		entity.WxAvatar = wxAvatar
	}

	saved, err := s.appUserRepository.SaveOrUpdate(entity)
	if err != nil {
		return nil, err
	}
	return toCurrentAppUserProfileDTO(saved), nil
}

func (s *AppUserService) UpdateCurrentUserPhone(id uint, phone string) (*appUserDTO.CurrentAppUserProfileDTO, error) {
	entity, err := s.appUserRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	if entity.Active == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	entity.Phone = strings.TrimSpace(phone)
	saved, err := s.appUserRepository.SaveOrUpdate(entity)
	if err != nil {
		return nil, err
	}
	return toCurrentAppUserProfileDTO(saved), nil
}

func (s *AppUserService) ChangeCurrentUserPassword(id uint, req *appUserDTO.ChangeCurrentAppUserPasswordDTO) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}
	entity, err := s.appUserRepository.FindById(id)
	if err != nil {
		return err
	}
	if entity.Active == 0 {
		return gorm.ErrRecordNotFound
	}

	oldPassword := strings.TrimSpace(req.OldPassword)
	newPassword := strings.TrimSpace(req.NewPassword)
	hasOriginPassword := strings.TrimSpace(entity.OriginPassword) != ""
	if hasOriginPassword && oldPassword == "" {
		return fmt.Errorf("old password is required")
	}
	if err := validateAppUserPassword(newPassword); err != nil {
		return err
	}
	if oldPassword != "" && strings.EqualFold(oldPassword, newPassword) {
		return fmt.Errorf("new password must be different from old password")
	}

	if hasOriginPassword {
		expectedPassword := appUserPassword.Encrypt(entity.Username, oldPassword)
		if !strings.EqualFold(expectedPassword, strings.TrimSpace(entity.Password)) {
			return fmt.Errorf("old password is incorrect")
		}
	}

	entity.OriginPassword = newPassword
	entity.Password = appUserPassword.Encrypt(entity.Username, newPassword)
	_, err = s.appUserRepository.SaveOrUpdate(entity)
	return err
}

func (s *AppUserService) ListUsers(query appUserDTO.AppUserQueryDTO) (*baseDTO.PageDTO[appUserDTO.AppUserDTO], error) {
	pageIndex, pageSize := normalizeAppUserPage(query.Page, query.PageIndex, query.PageSize)
	total, err := s.appUserRepository.CountUsersByQuery(query)
	if err != nil {
		return nil, err
	}
	rows, err := s.appUserRepository.ListUsersByQuery(query, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]*appUserDTO.AppUserDTO, 0, len(rows))
	for i := range rows {
		if err := decryptAppUserListRowIDNumber(&rows[i]); err != nil {
			return nil, err
		}
		row := rows[i]
		items = append(items, &appUserDTO.AppUserDTO{
			BaseDTO: baseDTO.BaseDTO{
				Id:          row.Id,
				Active:      row.Active,
				CreatedTime: row.CreatedTime,
				CreatedBy:   row.CreatedBy,
				UpdatedTime: row.UpdatedTime,
				UpdatedBy:   row.UpdatedBy,
			},
			Name:            row.Name,
			Username:        row.Username,
			Email:           row.Email,
			Phone:           row.Phone,
			Department:      row.Department,
			Password:        row.Password,
			OriginPassword:  row.OriginPassword,
			Status:          row.Status,
			LastLoginTime:   row.LastLoginTime,
			SecretKey:       row.SecretKey,
			Remark:          row.Remark,
			BanCount:        row.BanCount,
			OpenUID:         row.OpenUID,
			UnionID:         row.UnionID,
			WxSessionKey:    row.WxSessionKey,
			WxNickname:      row.WxNickname,
			WxAvatar:        row.WxAvatar,
			WxGender:        row.WxGender,
			WxCountry:       row.WxCountry,
			WxProvince:      row.WxProvince,
			WxCity:          row.WxCity,
			WxLanguage:      row.WxLanguage,
			WxLastLoginTime: row.WxLastLoginTime,
			StationID:       row.StationID,
			StationName:     row.StationName,
			IDNumber:        row.IDNumber,
			IDCardFrontURL:  row.IDCardFrontURL,
			IDCardFrontKey:  row.IDCardFrontKey,
			IDCardBackURL:   row.IDCardBackURL,
			IDCardBackKey:   row.IDCardBackKey,
		})
	}
	return baseDTO.BuildPage(int(total), items), nil
}

func (s *AppUserService) GetUserByID(id uint) (*appUserDTO.AppUserDTO, error) {
	entity, err := s.appUserRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	if entity.Active == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return s.toAppUserDTO(entity)
}

func (s *AppUserService) CreateUser(req *appUserDTO.CreateAppUserDTO) (*appUserDTO.AppUserDTO, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	name := strings.TrimSpace(req.Name)
	username := strings.TrimSpace(req.Username)
	email := strings.TrimSpace(req.Email)
	phone := strings.TrimSpace(req.Phone)
	department := strings.TrimSpace(req.Department)
	status := normalizeAppUserStatus(req.Status)
	password := strings.TrimSpace(req.Password)
	originPassword := strings.TrimSpace(req.OriginPassword)
	secretKey := strings.TrimSpace(req.SecretKey)
	remark := strings.TrimSpace(req.Remark)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if status == "" {
		return nil, fmt.Errorf("status is invalid")
	}
	if password == "" && originPassword != "" {
		password = appUserPassword.Encrypt(username, originPassword)
	}
	if originPassword == "" && password != "" {
		originPassword = password
	}
	if originPassword != "" {
		if err := validateAppUserPassword(originPassword); err != nil {
			return nil, err
		}
	}
	if err := validateAppUserEmail(email); err != nil {
		return nil, err
	}
	existing, err := s.appUserRepository.FindByUsername(username)
	if err == nil && existing != nil && existing.Active == 1 {
		return nil, fmt.Errorf("username already exists")
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	newUser := &appUserRepository.AppUser{
		Name:           name,
		Username:       username,
		Email:          email,
		Phone:          phone,
		Department:     department,
		Password:       password,
		OriginPassword: originPassword,
		Status:         status,
		SecretKey:      secretKey,
		Remark:         remark,
		BanCount:       req.BanCount,
		OpenUID:        strings.TrimSpace(req.OpenUID),
		UnionID:        strings.TrimSpace(req.UnionID),
		WxSessionKey:   strings.TrimSpace(req.WxSessionKey),
		WxNickname:     strings.TrimSpace(req.WxNickname),
		WxAvatar:       strings.TrimSpace(req.WxAvatar),
		WxGender:       req.WxGender,
		WxCountry:      strings.TrimSpace(req.WxCountry),
		WxProvince:     strings.TrimSpace(req.WxProvince),
		WxCity:         strings.TrimSpace(req.WxCity),
		WxLanguage:     strings.TrimSpace(req.WxLanguage),
		IDNumber:       strings.TrimSpace(req.IDNumber),
		IDCardFrontURL: strings.TrimSpace(req.IDCardFrontURL),
		IDCardFrontKey: strings.TrimSpace(req.IDCardFrontKey),
		IDCardBackURL:  strings.TrimSpace(req.IDCardBackURL),
		IDCardBackKey:  strings.TrimSpace(req.IDCardBackKey),
	}
	if err := prepareAppUserIDCardForSave(newUser); err != nil {
		return nil, err
	}
	if req.LastLoginTime != nil && !req.LastLoginTime.IsZero() {
		newUser.LastLoginTime = req.LastLoginTime
	}
	if req.WxLastLoginTime != nil && !req.WxLastLoginTime.IsZero() {
		newUser.WxLastLoginTime = req.WxLastLoginTime
	}
	created, err := s.appUserRepository.Create(newUser)
	if err != nil {
		return nil, err
	}
	if req.StationID > 0 {
		if _, err := s.stationUserRepository.SaveActiveStation(uint64(created.Id), req.StationID); err != nil {
			return nil, err
		}
	}
	return s.toAppUserDTO(created)
}

func (s *AppUserService) RegisterUser(req *appUserDTO.RegisterAppUserDTO) (*appUserDTO.AppUserDTO, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}

	name := strings.TrimSpace(req.Name)
	username := strings.TrimSpace(req.Username)
	rawPassword := strings.TrimSpace(req.Password)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if err := validateAppUserPassword(rawPassword); err != nil {
		return nil, err
	}

	existing, err := s.appUserRepository.FindByUsername(username)
	if err == nil && existing != nil && existing.Active == 1 {
		return nil, fmt.Errorf("username already exists")
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	created, err := s.appUserRepository.Create(&appUserRepository.AppUser{
		Name:           name,
		Username:       username,
		Password:       appUserPassword.Encrypt(username, rawPassword),
		OriginPassword: rawPassword,
		Status:         "active",
		LastLoginTime:  nil,
	})
	if err != nil {
		return nil, err
	}
	if err := s.managerUserService.EnsureSalesmanUser(created.Username, created.Name, created.Phone, rawPassword); err != nil {
		return nil, err
	}
	return s.toAppUserDTO(created)
}

func (s *AppUserService) UpdateUser(id uint, req *appUserDTO.UpdateAppUserDTO) (*appUserDTO.AppUserDTO, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	entity, err := s.appUserRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	if entity.Active == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	if req.Name != nil {
		value := strings.TrimSpace(*req.Name)
		if value == "" {
			return nil, fmt.Errorf("name is required")
		}
		entity.Name = value
	}
	if req.Username != nil {
		value := strings.TrimSpace(*req.Username)
		if value == "" {
			return nil, fmt.Errorf("username is required")
		}
		existing, err := s.appUserRepository.FindByUsername(value)
		if err == nil && existing != nil && existing.Active == 1 && existing.Id != entity.Id {
			return nil, fmt.Errorf("username already exists")
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}
		entity.Username = value
		if strings.TrimSpace(entity.OriginPassword) != "" {
			entity.Password = appUserPassword.Encrypt(entity.Username, entity.OriginPassword)
		}
	}
	if req.Email != nil {
		value := strings.TrimSpace(*req.Email)
		if err := validateAppUserEmail(value); err != nil {
			return nil, err
		}
		entity.Email = value
	}
	if req.Phone != nil {
		entity.Phone = strings.TrimSpace(*req.Phone)
	}
	if req.Department != nil {
		entity.Department = strings.TrimSpace(*req.Department)
	}
	if req.Password != nil {
		rawPassword := strings.TrimSpace(*req.Password)
		if rawPassword == "" {
			entity.Password = ""
		} else {
			if err := validateAppUserPassword(rawPassword); err != nil {
				return nil, err
			}
			entity.Password = appUserPassword.Encrypt(entity.Username, rawPassword)
		}
	}
	if req.OriginPassword != nil {
		rawPassword := strings.TrimSpace(*req.OriginPassword)
		if rawPassword != "" {
			if err := validateAppUserPassword(rawPassword); err != nil {
				return nil, err
			}
			entity.OriginPassword = rawPassword
			entity.Password = appUserPassword.Encrypt(entity.Username, rawPassword)
		} else {
			entity.OriginPassword = ""
			entity.Password = ""
		}
	}
	if req.Status != nil {
		status := normalizeAppUserStatus(*req.Status)
		if status == "" {
			return nil, fmt.Errorf("status is invalid")
		}
		entity.Status = status
	}
	if req.LastLoginTime != nil {
		entity.LastLoginTime = req.LastLoginTime
	}
	if req.SecretKey != nil {
		entity.SecretKey = strings.TrimSpace(*req.SecretKey)
	}
	if req.Remark != nil {
		entity.Remark = strings.TrimSpace(*req.Remark)
	}
	if req.BanCount != nil {
		entity.BanCount = *req.BanCount
	}
	if req.OpenUID != nil {
		entity.OpenUID = strings.TrimSpace(*req.OpenUID)
	}
	if req.UnionID != nil {
		entity.UnionID = strings.TrimSpace(*req.UnionID)
	}
	if req.WxSessionKey != nil {
		entity.WxSessionKey = strings.TrimSpace(*req.WxSessionKey)
	}
	if req.WxNickname != nil {
		entity.WxNickname = strings.TrimSpace(*req.WxNickname)
	}
	if req.WxAvatar != nil {
		entity.WxAvatar = strings.TrimSpace(*req.WxAvatar)
	}
	if req.WxGender != nil {
		entity.WxGender = *req.WxGender
	}
	if req.WxCountry != nil {
		entity.WxCountry = strings.TrimSpace(*req.WxCountry)
	}
	if req.WxProvince != nil {
		entity.WxProvince = strings.TrimSpace(*req.WxProvince)
	}
	if req.WxCity != nil {
		entity.WxCity = strings.TrimSpace(*req.WxCity)
	}
	if req.WxLanguage != nil {
		entity.WxLanguage = strings.TrimSpace(*req.WxLanguage)
	}
	if req.WxLastLoginTime != nil {
		entity.WxLastLoginTime = req.WxLastLoginTime
	}
	if req.IDNumber != nil {
		entity.IDNumber = strings.TrimSpace(*req.IDNumber)
	}
	if req.IDCardFrontURL != nil {
		entity.IDCardFrontURL = strings.TrimSpace(*req.IDCardFrontURL)
	}
	if req.IDCardFrontKey != nil {
		entity.IDCardFrontKey = strings.TrimSpace(*req.IDCardFrontKey)
	}
	if req.IDCardBackURL != nil {
		entity.IDCardBackURL = strings.TrimSpace(*req.IDCardBackURL)
	}
	if req.IDCardBackKey != nil {
		entity.IDCardBackKey = strings.TrimSpace(*req.IDCardBackKey)
	}
	if err := prepareAppUserIDCardForSave(entity); err != nil {
		return nil, err
	}
	saved, err := s.appUserRepository.SaveOrUpdate(entity)
	if err != nil {
		return nil, err
	}
	if req.StationID != nil && *req.StationID > 0 {
		if _, err := s.stationUserRepository.SaveActiveStation(uint64(saved.Id), *req.StationID); err != nil {
			return nil, err
		}
	}
	return s.toAppUserDTO(saved)
}

func (s *AppUserService) DeleteUser(id uint) error {
	entity, err := s.appUserRepository.FindById(id)
	if err != nil {
		return err
	}
	if entity.Active == 0 {
		return gorm.ErrRecordNotFound
	}
	entity.Active = 0
	_, err = s.appUserRepository.SaveOrUpdate(entity)
	return err
}

func toCurrentAppUserProfileDTO(entity *appUserRepository.AppUser) *appUserDTO.CurrentAppUserProfileDTO {
	if entity == nil {
		return nil
	}
	result := &appUserDTO.CurrentAppUserProfileDTO{
		Id:                entity.Id,
		Name:              entity.Name,
		Username:          entity.Username,
		Email:             entity.Email,
		Phone:             entity.Phone,
		Department:        entity.Department,
		Remark:            entity.Remark,
		LastLoginTime:     entity.LastLoginTime,
		OpenUID:           entity.OpenUID,
		UnionID:           entity.UnionID,
		WxNickname:        entity.WxNickname,
		WxAvatar:          entity.WxAvatar,
		WxLastLoginTime:   entity.WxLastLoginTime,
		HasOriginPassword: strings.TrimSpace(entity.OriginPassword) != "",
	}
	repo := db.GetRepository[grainConfigRepository.GrainStationUserRepository]()
	stationUser, err := repo.FindActiveByAppUserID(uint64(entity.Id))
	if err == nil && stationUser != nil {
		result.StationID = stationUser.StationID
		station, stationErr := db.GetRepository[grainConfigRepository.GrainStationRepository]().FindById(uint(stationUser.StationID))
		if stationErr == nil && station != nil && station.Active == 1 {
			result.StationName = station.StationName
		}
	}
	return result
}

func (s *AppUserService) toAppUserDTO(entity *appUserRepository.AppUser) (*appUserDTO.AppUserDTO, error) {
	if err := decryptAppUserIDNumber(entity); err != nil {
		return nil, err
	}
	result := db.ToDTO[appUserDTO.AppUserDTO](entity)
	if entity == nil || entity.Id == 0 {
		return result, nil
	}
	stationUser, err := s.stationUserRepository.FindActiveByAppUserID(uint64(entity.Id))
	if err == nil {
		result.StationID = stationUser.StationID
		return result, nil
	}
	if err == gorm.ErrRecordNotFound {
		return result, nil
	}
	return nil, err
}
