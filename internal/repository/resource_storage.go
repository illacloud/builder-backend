package repository

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ResourceRepository interface {
	Create(resource *Resource) (int, error)
	Delete(teamID int, resourceID int) error
	Update(resource *Resource) error
	RetrieveByID(teamID int, resourceID int) (*Resource, error)
	RetrieveAll(teamID int) ([]*Resource, error)
	RetrieveAllByUpdatedTime(teamID int) ([]*Resource, error)
	CountResourceByTeamID(teamID int) (int, error)
	RetrieveResourceLastModifiedTime(teamID int) (time.Time, error)
}

type ResourceRepositoryImpl struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewResourceRepositoryImpl(logger *zap.SugaredLogger, db *gorm.DB) *ResourceRepositoryImpl {
	return &ResourceRepositoryImpl{
		logger: logger,
		db:     db,
	}
}

func (impl *ResourceRepositoryImpl) Create(resource *Resource) (int, error) {
	if err := impl.db.Create(resource).Error; err != nil {
		return 0, err
	}
	return resource.ID, nil
}

func (impl *ResourceRepositoryImpl) Delete(teamID int, resourceID int) error {
	if err := impl.db.Where("id = ? AND team_id = ?", resourceID, teamID).Delete(&Resource{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *ResourceRepositoryImpl) Update(resource *Resource) error {
	if err := impl.db.Model(resource).UpdateColumns(Resource{
		Name:      resource.Name,
		Options:   resource.Options,
		UpdatedBy: resource.UpdatedBy,
		UpdatedAt: resource.UpdatedAt,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *ResourceRepositoryImpl) RetrieveByID(teamID int, resourceID int) (*Resource, error) {
	var resource *Resource
	if err := impl.db.Where("id = ? AND team_id = ?", resourceID, teamID).First(&resource).Error; err != nil {
		return &Resource{}, err
	}
	return resource, nil
}

func (impl *ResourceRepositoryImpl) RetrieveAll(teamID int) ([]*Resource, error) {
	var resources []*Resource
	if err := impl.db.Where("team_id = ?", teamID).Find(&resources).Error; err != nil {
		return nil, err
	}
	return resources, nil
}

func (impl *ResourceRepositoryImpl) RetrieveAllByUpdatedTime(teamID int) ([]*Resource, error) {
	var resources []*Resource
	if err := impl.db.Where("team_id = ?", teamID).Order("updated_at desc").Find(&resources).Error; err != nil {
		return nil, err
	}
	return resources, nil
}

func (impl *ResourceRepositoryImpl) CountResourceByTeamID(teamID int) (int, error) {
	var count int64
	if err := impl.db.Model(&Resource{}).Where("team_id = ?", teamID).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func (impl *ResourceRepositoryImpl) RetrieveResourceLastModifiedTime(teamID int) (time.Time, error) {
	var resource *Resource
	if err := impl.db.Where("team_id = ?", teamID).Order("updated_at desc").First(&resource).Error; err != nil {
		return time.Time{}, err
	}
	return resource.ExportUpdatedAt(), nil
}
