package user

import (
	baseDTO "common/base/dto"
	"common/middleware/db"
	"common/utils"
	"fmt"
	userDTO "service/manager_user/dto"
	userPassword "service/manager_user/password"
	userRepository "service/manager_user/repository"
	"strings"
	"time"

	"gorm.io/gorm"
)

// decryptPassword 尝试解密前端混淆的密码，解密失败则返回原值
func decryptPassword(encrypted string) string {
	if plain, err := utils.DecryptContent(encrypted); err == nil {
		return plain
	}
	return encrypted
}

// fillUserTenants attaches TenantIDs and TenantNames to each UserDTO.
func (s *UserService) fillUserTenants(items []*userDTO.UserDTO) error {
	if len(items) == 0 || s.userTenantRepository.Db == nil {
		return nil
	}
	userIDs := make([]int, 0, len(items))
	itemByID := make(map[int]*userDTO.UserDTO, len(items))
	for _, item := range items {
		userIDs = append(userIDs, item.Id)
		itemByID[item.Id] = item
	}
	var rows []struct {
		UserID     int
		TenantID   uint64
		TenantName string
	}
	err := s.userTenantRepository.Db.Table("user_tenant AS ut").
		Select("ut.user_id AS user_id, ut.tenant_id AS tenant_id, t.tenant_name AS tenant_name").
		Joins("LEFT JOIN tenant AS t ON t.id = ut.tenant_id AND t.active = ?", 1).
		Where("ut.active = ? AND ut.user_id IN ?", 1, userIDs).
		Order("ut.id ASC").
		Scan(&rows).Error
	if err != nil {
		return err
	}
	for _, row := range rows {
		if item, ok := itemByID[row.UserID]; ok {
			item.TenantIDs = append(item.TenantIDs, row.TenantID)
			if strings.TrimSpace(row.TenantName) != "" {
				item.TenantNames = append(item.TenantNames, row.TenantName)
			}
		}
	}
	return nil
}

func (s *UserService) GetCurrentUserProfile(id uint) (*userDTO.CurrentUserProfileDTO, error) {
	entity, err := s.userRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	if entity.Active == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return s.toCurrentUserProfileDTO(entity), nil
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
	return s.toCurrentUserProfileDTO(saved), nil
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

	oldPassword := decryptPassword(strings.TrimSpace(req.OldPassword))
	newPassword := decryptPassword(strings.TrimSpace(req.NewPassword))
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
			Role:           row.RoleCode,
			Password:       row.Password,
			OriginPassword: row.OriginPassword,
			Status:         row.Status,
			LastLoginTime:  row.LastLoginTime,
			SecretKey:      row.SecretKey,
			Remark:         row.Remark,
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
	if err := s.fillUserTenants(items); err != nil {
		return nil, err
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
	result := db.ToDTO[userDTO.UserDTO](entity)
	result.Role = s.userRepository.FindFirstRoleCodeByUserID(entity.Id)
	if err := s.fillUserTenants([]*userDTO.UserDTO{result}); err != nil {
		return nil, err
	}
	return result, nil
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
	status := normalizeUserStatus(req.Status)
	originPassword := decryptPassword(strings.TrimSpace(req.OriginPassword))
	password := userPassword.Encrypt(strings.TrimSpace(req.Username), originPassword)
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
	var lastLoginTime *time.Time
	if !req.LastLoginTime.IsZero() {
		t := req.LastLoginTime
		lastLoginTime = &t
	}
	roleCode := normalizeUserRole(req.Role)
	created, err := s.userRepository.Create(&userRepository.User{
		Name:           name,
		Username:       username,
		Email:          email,
		Phone:          phone,
		Department:     department,
		TenantID:       tenantID,
		Password:       password,
		OriginPassword: originPassword,
		Status:         status,
		LastLoginTime:  lastLoginTime,
		SecretKey:      secretKey,
		Remark:         remark,
		BanCount:       req.BanCount,
	})
	if err != nil {
		return nil, err
	}
	if roleCode != "" {
		if err := s.replaceUserRole(uint64(created.Id), roleCode); err != nil {
			return nil, err
		}
	}
	if len(req.TenantIDs) > 0 {
		if err := s.userTenantRepository.ReplaceActiveTenants(uint64(created.Id), req.TenantIDs); err != nil {
			return nil, err
		}
	}
	return s.GetUserByID(uint(created.Id))
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
		if err := s.replaceUserRole(uint64(entity.Id), *req.Role); err != nil {
			return nil, err
		}
	}
	if req.OriginPassword != nil {
		raw := decryptPassword(strings.TrimSpace(*req.OriginPassword))
		entity.OriginPassword = raw
		entity.Password = userPassword.Encrypt(entity.Username, raw)
	} else if req.Password != nil {
		entity.Password = strings.TrimSpace(*req.Password)
	}
	if req.Status != nil {
		status := normalizeUserStatus(*req.Status)
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
	saved, err := s.userRepository.SaveOrUpdate(entity)
	if err != nil {
		return nil, err
	}
	if req.TenantIDs != nil {
		if err := s.userTenantRepository.ReplaceActiveTenants(uint64(saved.Id), *req.TenantIDs); err != nil {
			return nil, err
		}
	}
	return s.GetUserByID(uint(saved.Id))
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

func (s *UserService) toCurrentUserProfileDTO(entity *userRepository.User) *userDTO.CurrentUserProfileDTO {
	return &userDTO.CurrentUserProfileDTO{
		Id:         entity.Id,
		Name:       entity.Name,
		Username:   entity.Username,
		Email:      entity.Email,
		Phone:      entity.Phone,
		Department: entity.Department,
		TenantID:   entity.TenantID,
		Status:     entity.Status,
		Remark:     entity.Remark,
	}
}
