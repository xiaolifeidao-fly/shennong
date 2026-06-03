package repository

import (
	"common/middleware/db"
	"fmt"
	userDTO "service/manager_user/dto"
	"strings"

	"gorm.io/gorm"
)

type UserRepository struct {
	db.Repository[*User]
}

func (r *UserRepository) EnsureTable() error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return r.Db.AutoMigrate(&User{})
}

func (r *UserRepository) FindByUsername(username string) (*User, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	var entity User
	if err := r.QueryOneBySQL(&entity, "SELECT * FROM user WHERE username = ? AND active = 1 ORDER BY id ASC LIMIT 1", strings.TrimSpace(username)); err != nil {
		return nil, err
	}
	if entity.Id == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &entity, nil
}

func (r *UserRepository) FindByUsernameOrPhone(account string) (*User, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	var entity User
	account = strings.TrimSpace(account)
	if err := r.QueryOneBySQL(&entity, "SELECT * FROM user WHERE (username = ? OR phone = ?) AND active = 1 ORDER BY id ASC LIMIT 1", account, account); err != nil {
		return nil, err
	}
	if entity.Id == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &entity, nil
}

func (r *UserRepository) CountActiveByRoles(roles []string) (int64, error) {
	if r.Db == nil {
		return 0, fmt.Errorf("database is not initialized")
	}
	if len(roles) == 0 {
		return 0, nil
	}
	return r.CountBySQL("SELECT id FROM user WHERE active = 1 AND role IN ?", roles)
}

func (r *UserRepository) CountRecentLoginUsers() (int64, error) {
	if r.Db == nil {
		return 0, fmt.Errorf("database is not initialized")
	}
	return r.CountBySQL(`SELECT id
	FROM user
	WHERE active = 1
		AND last_login_time IS NOT NULL
		AND last_login_time > ?`, "1970-01-02 00:00:00")
}

func (r *UserRepository) CountActiveAccounts() (int64, error) {
	if r.Db == nil {
		return 0, fmt.Errorf("database is not initialized")
	}
	return r.CountBySQL("SELECT id FROM account WHERE active = 1")
}

func (r *UserRepository) CountUsersByQuery(query userDTO.UserQueryDTO) (int64, error) {
	if r.Db == nil {
		return 0, fmt.Errorf("database is not initialized")
	}
	whereSQL, values := buildUserListWhere(query)
	sql := "SELECT u.id FROM user u " + whereSQL
	return r.CountBySQL(sql, values...)
}

func (r *UserRepository) ListUsersByQuery(query userDTO.UserQueryDTO, pageIndex, pageSize int) ([]UserListRow, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	whereSQL, values := buildUserListWhere(query)
	sql := `SELECT
		u.id, u.active, u.created_time, u.updated_time, u.created_by, u.updated_by,
		u.name, u.username, u.email, u.phone, u.department, u.tenant_id, u.role, u.password,
		u.origin_password, u.status, u.last_login_time, u.secret_key, u.remark,
		u.pub_token, u.ban_count
	FROM user u ` + whereSQL + ` ORDER BY u.id DESC LIMIT ? OFFSET ?`
	values = append(values, pageSize, (pageIndex-1)*pageSize)
	var rows []UserListRow
	if err := r.QueryBySQL(&rows, sql, values...); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *UserRepository) ListUserAccounts(userIDs []int) ([]UserAccountRow, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	if len(userIDs) == 0 {
		return []UserAccountRow{}, nil
	}
	sql := `SELECT id, user_id, account_status, balance_amount
	FROM account
	WHERE active = 1 AND user_id IN ?
	ORDER BY id DESC`
	var rows []UserAccountRow
	if err := r.QueryBySQL(&rows, sql, userIDs); err != nil {
		return nil, err
	}
	return rows, nil
}

func buildUserListWhere(query userDTO.UserQueryDTO) (string, []interface{}) {
	clauses := []string{"WHERE u.active = 1"}
	values := make([]interface{}, 0, 16)

	if value := strings.TrimSpace(query.Search); value != "" {
		likeValue := "%" + value + "%"
		clauses = append(clauses, `(u.name LIKE ? OR u.username LIKE ? OR u.email LIKE ? OR u.phone LIKE ? OR u.department LIKE ? OR u.remark LIKE ?)`)
		values = append(values, likeValue, likeValue, likeValue, likeValue, likeValue, likeValue)
	}
	if value := strings.TrimSpace(query.Name); value != "" {
		clauses = append(clauses, "u.name LIKE ?")
		values = append(values, "%"+value+"%")
	}
	if value := strings.TrimSpace(query.Username); value != "" {
		clauses = append(clauses, "u.username LIKE ?")
		values = append(values, "%"+value+"%")
	}
	if value := strings.TrimSpace(query.Email); value != "" {
		clauses = append(clauses, "u.email LIKE ?")
		values = append(values, "%"+value+"%")
	}
	if value := strings.TrimSpace(query.Phone); value != "" {
		clauses = append(clauses, "u.phone LIKE ?")
		values = append(values, "%"+value+"%")
	}
	if value := strings.TrimSpace(query.Department); value != "" {
		clauses = append(clauses, "u.department LIKE ?")
		values = append(values, "%"+value+"%")
	}
	if query.TenantID > 0 {
		clauses = append(clauses, "u.tenant_id = ?")
		values = append(values, query.TenantID)
	}
	if value := strings.TrimSpace(query.Role); value != "" {
		clauses = append(clauses, "u.role = ?")
		values = append(values, value)
	}
	if value := strings.TrimSpace(query.Status); value != "" {
		clauses = append(clauses, `(u.status = ? OR EXISTS (
			SELECT 1 FROM account a
			WHERE a.user_id = u.id AND a.active = 1 AND a.account_status = ?
		))`)
		values = append(values, value, value)
	}
	if value := strings.TrimSpace(query.SecretKey); value != "" {
		clauses = append(clauses, "u.secret_key = ?")
		values = append(values, value)
	}
	if value := strings.TrimSpace(query.PubToken); value != "" {
		clauses = append(clauses, "u.pub_token = ?")
		values = append(values, value)
	}

	return strings.Join(clauses, " AND "), values
}

type UserLoginRecordRepository struct {
	db.Repository[*UserLoginRecord]
}

func (r *UserLoginRecordRepository) EnsureTable() error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return r.Db.AutoMigrate(&UserLoginRecord{})
}

type UserTenantRepository struct {
	db.Repository[*UserTenant]
}

func (r *UserTenantRepository) EnsureTable() error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return r.Db.AutoMigrate(&UserTenant{})
}

func (r *UserTenantRepository) ReplaceActiveTenants(userID uint64, tenantIDs []uint64) error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	if userID == 0 {
		return fmt.Errorf("userId must be positive")
	}
	if err := r.Db.Model(&UserTenant{}).
		Where("active = ? AND user_id = ?", 1, userID).
		Updates(map[string]interface{}{"active": 0}).Error; err != nil {
		return err
	}
	for _, tenantID := range uniquePositiveUint64(tenantIDs) {
		if _, err := r.Create(&UserTenant{UserID: userID, TenantID: tenantID}); err != nil {
			return err
		}
	}
	return nil
}

func (r *UserTenantRepository) ListActiveTenantIDs(userID uint64) ([]uint64, error) {
	if r.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	var rows []struct {
		TenantID uint64
	}
	err := r.Db.Raw(
		`SELECT tenant_id FROM user_tenant WHERE active = ? AND user_id = ? ORDER BY id ASC`,
		1, userID,
	).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]uint64, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.TenantID)
	}
	return result, nil
}

func uniquePositiveUint64(values []uint64) []uint64 {
	seen := make(map[uint64]struct{}, len(values))
	result := make([]uint64, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

type UserRoleRepository struct {
	db.Repository[*UserRole]
}

func (r *UserRoleRepository) EnsureTable() error {
	if r.Db == nil {
		return fmt.Errorf("database is not initialized")
	}
	return r.Db.AutoMigrate(&UserRole{})
}
