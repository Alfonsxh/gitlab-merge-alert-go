package services

import (
	"errors"
	"fmt"

	"gitlab-merge-alert-go/internal/models"
	"gorm.io/gorm"
)

type ResourceManagerService struct {
	db *gorm.DB
}

func NewResourceManagerService(db *gorm.DB) *ResourceManagerService {
	return &ResourceManagerService{
		db: db,
	}
}

func (s *ResourceManagerService) AssignManager(accountID uint, req *models.AssignManagerRequest) error {
	var manager models.Account
	if err := s.db.First(&manager, req.ManagerID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("管理员账户不存在")
		}
		return err
	}

	if !s.checkResourceExists(req.ResourceID, req.ResourceType) {
		return errors.New("资源不存在")
	}

	var existing models.ResourceManager
	err := s.db.Where("resource_id = ? AND resource_type = ? AND manager_id = ?",
		req.ResourceID, req.ResourceType, req.ManagerID).First(&existing).Error
	
	if err == nil {
		return errors.New("该管理员已经被分配到此资源")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	resourceManager := &models.ResourceManager{
		ResourceID:   req.ResourceID,
		ResourceType: req.ResourceType,
		ManagerID:    req.ManagerID,
		CreatedBy:    accountID,
	}

	return s.db.Create(resourceManager).Error
}

func (s *ResourceManagerService) RemoveManager(accountID uint, req *models.RemoveManagerRequest) error {
	result := s.db.Where("resource_id = ? AND resource_type = ? AND manager_id = ?",
		req.ResourceID, req.ResourceType, req.ManagerID).Delete(&models.ResourceManager{})
	
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("未找到相应的管理员分配记录")
	}

	return nil
}

func (s *ResourceManagerService) GetResourceManagers(resourceID uint, resourceType models.ResourceType) ([]*models.ResourceManagerResponse, error) {
	var managers []models.ResourceManager
	
	err := s.db.Preload("Manager").
		Where("resource_id = ? AND resource_type = ?", resourceID, resourceType).
		Find(&managers).Error
	
	if err != nil {
		return nil, err
	}

	responses := make([]*models.ResourceManagerResponse, len(managers))
	for i, m := range managers {
		responses[i] = &models.ResourceManagerResponse{
			ID:           m.ID,
			ResourceID:   m.ResourceID,
			ResourceType: m.ResourceType,
			ManagerID:    m.ManagerID,
			Manager:      m.Manager.ToResponse(),
			CreatedAt:    m.CreatedAt,
		}
	}

	return responses, nil
}

func (s *ResourceManagerService) GetManagedResources(accountID uint, resourceType models.ResourceType) ([]uint, error) {
	var resourceIDs []uint
	
	err := s.db.Model(&models.ResourceManager{}).
		Where("manager_id = ? AND resource_type = ?", accountID, resourceType).
		Pluck("resource_id", &resourceIDs).Error
	
	return resourceIDs, err
}

func (s *ResourceManagerService) IsManager(accountID uint, resourceID uint, resourceType models.ResourceType) bool {
	var count int64
	s.db.Model(&models.ResourceManager{}).
		Where("manager_id = ? AND resource_id = ? AND resource_type = ?",
			accountID, resourceID, resourceType).
		Count(&count)
	
	return count > 0
}

func (s *ResourceManagerService) checkResourceExists(resourceID uint, resourceType models.ResourceType) bool {
	var count int64
	
	switch resourceType {
	case models.ResourceTypeProject:
		s.db.Model(&models.Project{}).Where("id = ?", resourceID).Count(&count)
	case models.ResourceTypeWebhook:
		s.db.Model(&models.Webhook{}).Where("id = ?", resourceID).Count(&count)
	case models.ResourceTypeUser:
		s.db.Model(&models.User{}).Where("id = ?", resourceID).Count(&count)
	default:
		return false
	}
	
	return count > 0
}

func (s *ResourceManagerService) GetBatchResourceManagers(resourceIDs []uint, resourceType models.ResourceType) (map[uint][]*models.AccountResponse, error) {
	var managers []models.ResourceManager
	
	err := s.db.Preload("Manager").
		Where("resource_id IN ? AND resource_type = ?", resourceIDs, resourceType).
		Find(&managers).Error
	
	if err != nil {
		return nil, err
	}

	result := make(map[uint][]*models.AccountResponse)
	for _, m := range managers {
		if result[m.ResourceID] == nil {
			result[m.ResourceID] = make([]*models.AccountResponse, 0)
		}
		result[m.ResourceID] = append(result[m.ResourceID], m.Manager.ToResponse())
	}

	return result, nil
}

func (s *ResourceManagerService) GetResourceList(accountID uint, role string, resourceType models.ResourceType, db *gorm.DB) *gorm.DB {
	if role == models.RoleAdmin {
		return db
	}

	var managedResourceIDs []uint
	s.db.Model(&models.ResourceManager{}).
		Where("manager_id = ? AND resource_type = ?", accountID, resourceType).
		Pluck("resource_id", &managedResourceIDs)

	if len(managedResourceIDs) > 0 {
		switch resourceType {
		case models.ResourceTypeProject:
			return db.Where("created_by = ? OR id IN ?", accountID, managedResourceIDs)
		case models.ResourceTypeWebhook:
			return db.Where("created_by = ? OR id IN ?", accountID, managedResourceIDs)
		case models.ResourceTypeUser:
			return db.Where("created_by = ? OR id IN ?", accountID, managedResourceIDs)
		}
	}

	return db.Where("created_by = ?", accountID)
}

func (s *ResourceManagerService) HasPermission(accountID uint, role string, resourceID uint, resourceType models.ResourceType) bool {
	if role == models.RoleAdmin {
		return true
	}

	var createdBy *uint

	switch resourceType {
	case models.ResourceTypeProject:
		var project models.Project
		if err := s.db.First(&project, resourceID).Error; err != nil {
			return false
		}
		createdBy = project.CreatedBy
	case models.ResourceTypeWebhook:
		var webhook models.Webhook
		if err := s.db.First(&webhook, resourceID).Error; err != nil {
			return false
		}
		createdBy = webhook.CreatedBy
	case models.ResourceTypeUser:
		var user models.User
		if err := s.db.First(&user, resourceID).Error; err != nil {
			return false
		}
		createdBy = user.CreatedBy
	default:
		return false
	}

	if createdBy != nil && *createdBy == accountID {
		return true
	}

	return s.IsManager(accountID, resourceID, resourceType)
}

func (s *ResourceManagerService) TransferOwnership(resourceID uint, resourceType models.ResourceType, newOwnerID uint) error {
	var newOwner models.Account
	if err := s.db.First(&newOwner, newOwnerID).Error; err != nil {
		return fmt.Errorf("新所有者不存在: %v", err)
	}

	switch resourceType {
	case models.ResourceTypeProject:
		return s.db.Model(&models.Project{}).Where("id = ?", resourceID).Update("created_by", newOwnerID).Error
	case models.ResourceTypeWebhook:
		return s.db.Model(&models.Webhook{}).Where("id = ?", resourceID).Update("created_by", newOwnerID).Error
	case models.ResourceTypeUser:
		return s.db.Model(&models.User{}).Where("id = ?", resourceID).Update("created_by", newOwnerID).Error
	default:
		return errors.New("不支持的资源类型")
	}
}