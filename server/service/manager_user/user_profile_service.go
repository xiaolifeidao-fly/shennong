package user

import (
	baseDTO "common/base/dto"
	"common/middleware/db"
	"fmt"
	userDTO "service/manager_user/dto"
	userPassword "service/manager_user/password"
	userRepository "service/manager_user/repository"
	"strings"
	"time"

	"gorm.io/gorm"
)

func (s *UserService) GetCurrentUserProfile(id uint) (*userDTO.CurrentUserProfileDTO, error) {
	entity, err := s.userRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	if entity.Active == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return toCurrentUserProfileDTO(entity), nil
}

func (s *UserService) UpdateCurrentUserProfile(id uint, req *userDTO.UpdateCurrentUserProfileDTO) (*userDTO.CurrentUserProfileDTO, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	entity, err := s.userRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	if entity.Active == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	name := strings.TrimSpace(req.Name)
	username := strings.TrimSpace(req.Username)
	email := strings.TrimSpace(req.Email)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if err := validateEmail(email); err != nil {
		return nil, err
	}

	if username != entity.Username {
		existing, err := s.userRepository.FindByUsername(username)
		if err == nil && existing != nil && existing.Active == 1 && existing.Id != entity.Id {
			return nil, fmt.Errorf("username already exists")
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}
		entity.Username = username
		if strings.TrimSpace(entity.OriginPassword) != "" {
			entity.Password = userPassword.Encrypt(entity.Username, entity.OriginPassword)
		}
	}

	entity.Name = name
	entity.Email = email
	entity.Phone = strings.TrimSpace(req.Phone)
	entity.Department = strings.TrimSpace(req.Department)
	entity.Remark = strings.TrimSpace(req.Remark)

	saved, err := s.userRepository.SaveOrUpdate(entity)
	if err != nil {
		return nil, err
	}
	return toCurrentUserProfileDTO(saved), nil
}

func (s *UserService) ChangeCurrentUserPassword(id uint, req *userDTO.ChangeCurrentUserPasswordDTO) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}
	entity, err := s.userRepository.FindById(id)
	if err != nil {
		return err
	}
	if entity.Active == 0 {
		return gorm.ErrRecordNotFound
	}

	oldPassword := strings.TrimSpace(req.OldPassword)
	newPassword := strings.TrimSpace(req.NewPassword)
	if oldPassword == "" {
		return fmt.Errorf("old password is required")
	}
	if err := validateUserPassword(newPassword); err != nil {
		return err
	}
	if strings.EqualFold(oldPassword, newPassword) {
		return fmt.Errorf("new password must be different from old password")
	}

	expectedPassword := userPassword.Encrypt(entity.Username, oldPassword)
	if !strings.EqualFold(expectedPassword, strings.TrimSpace(entity.Password)) {
		return fmt.Errorf("old password is incorrect")
	}

	return s.saveCurrentUserPassword(entity, newPassword)
}

func (s *UserService) saveCurrentUserPassword(entity *userRepository.User, rawPassword string) error {
	if s.userRepository.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	if entity == nil {
		return fmt.Errorf("user is nil")
	}
	rawPassword = strings.TrimSpace(rawPassword)
	if err := validateUserPassword(rawPassword); err != nil {
		return err
	}

	entity.OriginPassword = rawPassword
	entity.Password = userPassword.Encrypt(entity.Username, rawPassword)
	_, err := s.userRepository.SaveOrUpdate(entity)
	return err
}

func (s *UserService) GetUserStats() (*userDTO.UserStatsDTO, error) {
	if s.userRepository.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	var visibleUsers int64
	var activeUsers int64
	if err := s.userRepository.Db.Model(&userRepository.User{}).Where("active = ?", 1).Count(&visibleUsers).Error; err != nil {
		return nil, err
	}
	if err := s.userRepository.Db.Model(&userRepository.User{}).Where("active = ?", 1).Where("status = ?", "active").Count(&activeUsers).Error; err != nil {
		return nil, err
	}
	privilegedUsers, err := s.userRepository.CountActiveByRoles([]string{"admin", "manager"})
	if err != nil {
		return nil, err
	}
	recentLoginUsers, err := s.userRepository.CountRecentLoginUsers()
	if err != nil {
		return nil, err
	}
	accountCount, err := s.userRepository.CountActiveAccounts()
	if err != nil {
		return nil, err
	}
	return &userDTO.UserStatsDTO{
		VisibleUsers:     int(visibleUsers),
		AccountCount:     int(accountCount),
		PrivilegedUsers:  int(privilegedUsers),
		RecentLoginUsers: int(recentLoginUsers),
		ActiveUsers:      int(activeUsers),
	}, nil
}

func (s *UserService) ListUsers(query userDTO.UserQueryDTO) (*baseDTO.PageDTO[userDTO.UserDTO], error) {
	if s.userRepository.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	pageIndex, pageSize := normalizeUserPage(query.Page, query.PageIndex, query.PageSize)
	total, err := s.userRepository.CountUsersByQuery(query)
	if err != nil {
		return nil, err
	}
	rows, err := s.userRepository.ListUsersByQuery(query, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]*userDTO.UserDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, &userDTO.UserDTO{
			BaseDTO: baseDTO.BaseDTO{
				Id:          row.Id,
				Active:      row.Active,
				CreatedTime: row.CreatedTime,
				CreatedBy:   row.CreatedBy,
				UpdatedTime: row.UpdatedTime,
				UpdatedBy:   row.UpdatedBy,
			},
			Name:           row.Name,
			Username:       row.Username,
			Email:          row.Email,
			Phone:          row.Phone,
			Department:     row.Department,
			TenantID:       row.TenantID,
			Role:           row.Role,
			Password:       row.Password,
			OriginPassword: row.OriginPassword,
			Status:         row.Status,
			LastLoginTime:  row.LastLoginTime,
			SecretKey:      row.SecretKey,
			Remark:         row.Remark,
			PubToken:       row.PubToken,
			BanCount:       row.BanCount,
		})
	}
	if len(items) == 0 {
		return baseDTO.BuildPage(int(total), items), nil
	}

	userIDs := make([]int, 0, len(rows))
	for _, row := range rows {
		userIDs = append(userIDs, row.Id)
	}
	accountRows, err := s.userRepository.ListUserAccounts(userIDs)
	if err != nil {
		return nil, err
	}
	accountByUserID := make(map[int]userRepository.UserAccountRow, len(accountRows))
	for _, row := range accountRows {
		if _, exists := accountByUserID[row.UserID]; !exists {
			accountByUserID[row.UserID] = row
		}
	}
	for _, item := range items {
		if account, ok := accountByUserID[item.Id]; ok {
			item.AccountID = account.ID
			item.AccountStatus = account.AccountStatus
			item.BalanceAmount = account.BalanceAmount
		}
	}
	return baseDTO.BuildPage(int(total), items), nil
}

func (s *UserService) GetUserByID(id uint) (*userDTO.UserDTO, error) {
	if s.userRepository.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	entity, err := s.userRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	if entity.Active == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return db.ToDTO[userDTO.UserDTO](entity), nil
}

func (s *UserService) CreateUser(req *userDTO.CreateUserDTO) (*userDTO.UserDTO, error) {
	if s.userRepository.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	name := strings.TrimSpace(req.Name)
	username := strings.TrimSpace(req.Username)
	email := strings.TrimSpace(req.Email)
	phone := strings.TrimSpace(req.Phone)
	department := strings.TrimSpace(req.Department)
	tenantID := req.TenantID
	role := normalizeUserRole(req.Role)
	status := normalizeUserStatus(req.Status)
	password := strings.TrimSpace(req.Password)
	originPassword := strings.TrimSpace(req.OriginPassword)
	secretKey := strings.TrimSpace(req.SecretKey)
	remark := strings.TrimSpace(req.Remark)
	pubToken := strings.TrimSpace(req.PubToken)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if role == "" {
		return nil, fmt.Errorf("role is invalid")
	}
	if status == "" {
		return nil, fmt.Errorf("status is invalid")
	}
	if err := validateEmail(email); err != nil {
		return nil, err
	}
	existing, err := s.userRepository.FindByUsername(username)
	if err == nil && existing != nil && existing.Active == 1 {
		return nil, fmt.Errorf("username already exists")
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	lastLoginTime := req.LastLoginTime
	if lastLoginTime.IsZero() {
		lastLoginTime = time.Time{}
	}
	created, err := s.userRepository.Create(&userRepository.User{
		Name:           name,
		Username:       username,
		Email:          email,
		Phone:          phone,
		Department:     department,
		TenantID:       tenantID,
		Role:           role,
		Password:       password,
		OriginPassword: originPassword,
		Status:         status,
		LastLoginTime:  lastLoginTime,
		SecretKey:      secretKey,
		Remark:         remark,
		PubToken:       pubToken,
		BanCount:       req.BanCount,
	})
	if err != nil {
		return nil, err
	}
	return db.ToDTO[userDTO.UserDTO](created), nil
}

func (s *UserService) UpdateUser(id uint, req *userDTO.UpdateUserDTO) (*userDTO.UserDTO, error) {
	if s.userRepository.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	entity, err := s.userRepository.FindById(id)
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
		existing, err := s.userRepository.FindByUsername(value)
		if err == nil && existing != nil && existing.Active == 1 && existing.Id != entity.Id {
			return nil, fmt.Errorf("username already exists")
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}
		entity.Username = value
	}
	if req.Email != nil {
		value := strings.TrimSpace(*req.Email)
		if err := validateEmail(value); err != nil {
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
	if req.TenantID != nil {
		entity.TenantID = *req.TenantID
	}
	if req.Role != nil {
		role := normalizeUserRole(*req.Role)
		if role == "" {
			return nil, fmt.Errorf("role is invalid")
		}
		entity.Role = role
	}
	if req.Password != nil {
		entity.Password = strings.TrimSpace(*req.Password)
	}
	if req.OriginPassword != nil {
		entity.OriginPassword = strings.TrimSpace(*req.OriginPassword)
	}
	if req.Status != nil {
		status := normalizeUserStatus(*req.Status)
		if status == "" {
			return nil, fmt.Errorf("status is invalid")
		}
		entity.Status = status
	}
	if req.LastLoginTime != nil {
		entity.LastLoginTime = *req.LastLoginTime
	}
	if req.SecretKey != nil {
		entity.SecretKey = strings.TrimSpace(*req.SecretKey)
	}
	if req.Remark != nil {
		entity.Remark = strings.TrimSpace(*req.Remark)
	}
	if req.PubToken != nil {
		entity.PubToken = strings.TrimSpace(*req.PubToken)
	}
	if req.BanCount != nil {
		entity.BanCount = *req.BanCount
	}
	saved, err := s.userRepository.SaveOrUpdate(entity)
	if err != nil {
		return nil, err
	}
	return db.ToDTO[userDTO.UserDTO](saved), nil
}

func (s *UserService) DeleteUser(id uint) error {
	if s.userRepository.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	entity, err := s.userRepository.FindById(id)
	if err != nil {
		return err
	}
	if entity.Active == 0 {
		return gorm.ErrRecordNotFound
	}
	entity.Active = 0
	_, err = s.userRepository.SaveOrUpdate(entity)
	return err
}

func toCurrentUserProfileDTO(entity *userRepository.User) *userDTO.CurrentUserProfileDTO {
	return &userDTO.CurrentUserProfileDTO{
		Id:         entity.Id,
		Name:       entity.Name,
		Username:   entity.Username,
		Email:      entity.Email,
		Phone:      entity.Phone,
		Department: entity.Department,
		TenantID:   entity.TenantID,
		Role:       entity.Role,
		Status:     entity.Status,
		Remark:     entity.Remark,
	}
}
