package permission

import (
	"common/middleware/db"
	"fmt"
	permissionDTO "service/manager_permission/dto"
	permissionRepository "service/manager_permission/repository"
	userRepository "service/manager_user/repository"
)

func (s *PermissionService) ListCurrentUserMenuResources(userID uint64) ([]*permissionDTO.ResourceDTO, error) {
	if userID == 0 {
		return []*permissionDTO.ResourceDTO{}, nil
	}
	if s.resourceRepository.Db == nil || s.roleRepository.Db == nil ||
		s.roleResourceRepository.Db == nil || s.userRoleRepository.Db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}

	roleIDs, err := s.findActiveRoleIDsByUserID(userID)
	if err != nil {
		return nil, err
	}
	if len(roleIDs) == 0 {
		return []*permissionDTO.ResourceDTO{}, nil
	}

	var allMenuResources []*permissionRepository.Resource
	if err := s.resourceRepository.Db.
		Model(&permissionRepository.Resource{}).
		Where("active = ? AND resource_type IN ?", 1, []string{"menu", "page"}).
		Order("sort_id ASC, id ASC").
		Find(&allMenuResources).Error; err != nil {
		return nil, err
	}

	if s.hasSuperAdminRole(roleIDs) {
		return db.ToDTOs[permissionDTO.ResourceDTO](allMenuResources), nil
	}

	var authorizedResourceIDs []uint64
	if err := s.roleResourceRepository.Db.
		Model(&permissionRepository.RoleResource{}).
		Where("active = ? AND role_id IN ?", 1, roleIDs).
		Pluck("resource_id", &authorizedResourceIDs).Error; err != nil {
		return nil, err
	}

	return db.ToDTOs[permissionDTO.ResourceDTO](
		filterMenuResourcesWithAncestors(allMenuResources, authorizedResourceIDs),
	), nil
}

func (s *PermissionService) findActiveRoleIDsByUserID(userID uint64) ([]uint64, error) {
	var entities []*userRepository.UserRole
	if err := s.userRoleRepository.Db.
		Where("user_id = ? AND active = ?", userID, 1).
		Order("id DESC").
		Find(&entities).Error; err != nil {
		return nil, err
	}

	roleIDs := make([]uint64, 0, len(entities))
	seen := make(map[uint64]struct{}, len(entities))
	for _, entity := range entities {
		if entity.RoleID == 0 {
			continue
		}
		if _, exists := seen[entity.RoleID]; exists {
			continue
		}
		seen[entity.RoleID] = struct{}{}
		roleIDs = append(roleIDs, entity.RoleID)
	}
	return roleIDs, nil
}

func (s *PermissionService) hasSuperAdminRole(roleIDs []uint64) bool {
	if len(roleIDs) == 0 {
		return false
	}

	var count int64
	if err := s.roleRepository.Db.
		Model(&permissionRepository.Role{}).
		Where("active = ? AND id IN ? AND code IN ?", 1, roleIDs, []string{
			string(RoleCodeSuperAdmin),
			"admin",
		}).
		Count(&count).Error; err != nil {
		return false
	}
	return count > 0
}

func filterMenuResourcesWithAncestors(
	resources []*permissionRepository.Resource,
	authorizedResourceIDs []uint64,
) []*permissionRepository.Resource {
	authorized := make(map[uint64]struct{}, len(authorizedResourceIDs))
	included := make(map[uint64]struct{}, len(authorizedResourceIDs))
	resourceByID := make(map[uint64]*permissionRepository.Resource, len(resources))

	for _, resource := range resources {
		if resource == nil || resource.Id == 0 {
			continue
		}
		resourceByID[uint64(resource.Id)] = resource
	}
	for _, resourceID := range authorizedResourceIDs {
		if resourceID > 0 {
			authorized[resourceID] = struct{}{}
		}
	}
	for _, resource := range resources {
		if resource == nil {
			continue
		}
		if _, ok := authorized[uint64(resource.Id)]; !ok {
			continue
		}
		included[uint64(resource.Id)] = struct{}{}
		parentID := resource.ParentID
		for parentID > 0 {
			parent, ok := resourceByID[parentID]
			if !ok || parent == nil {
				break
			}
			if parent.ResourceType == "menu" {
				included[uint64(parent.Id)] = struct{}{}
			}
			parentID = parent.ParentID
		}
	}

	result := make([]*permissionRepository.Resource, 0, len(included))
	for _, resource := range resources {
		if resource == nil {
			continue
		}
		if _, ok := included[uint64(resource.Id)]; ok {
			result = append(result, resource)
		}
	}
	return result
}
