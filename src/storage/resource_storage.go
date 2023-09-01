package storage

import (
	"time"

	"github.com/illacloud/builder-backend/src/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ResourceStorage struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewResourceStorage(logger *zap.SugaredLogger, db *gorm.DB) *ResourceStorage {
	return &ResourceStorage{
		logger: logger,
		db:     db,
	}
}

func (impl *ResourceStorage) Create(resource *model.Resource) (int, error) {
	if err := impl.db.Create(resource).Error; err != nil {
		return 0, err
	}
	return resource.ID, nil
}

func (impl *ResourceStorage) UpdateWholeResource(resource *model.Resource) error {
	if err := impl.db.Model(resource).Where("id = ?", resource.ID).UpdateColumns(resource).Error; err != nil {
		return err
	}
	return nil
}

func (impl *ResourceStorage) RetrieveByTeamIDAndResourceID(teamID int, resourceID int) (*model.Resource, error) {
	var resource *model.Resource
	if err := impl.db.Where("id = ? AND team_id = ?", resourceID, teamID).First(&resource).Error; err != nil {
		return &model.Resource{}, err
	}
	return resource, nil
}

func (impl *ResourceStorage) RetrieveByTeamID(teamID int) ([]*model.Resource, error) {
	var resources []*model.Resource
	if err := impl.db.Where("team_id = ?", teamID).Find(&resources).Error; err != nil {
		return nil, err
	}
	return resources, nil
}

func (impl *ResourceStorage) RetrieveAllByUpdatedTime(teamID int) ([]*model.Resource, error) {
	var resources []*model.Resource
	if err := impl.db.Where("team_id = ?", teamID).Order("updated_at desc").Find(&resources).Error; err != nil {
		return nil, err
	}
	return resources, nil
}

func (impl *ResourceStorage) CountResourceByTeamID(teamID int) (int, error) {
	var count int64
	if err := impl.db.Model(&model.Resource{}).Where("team_id = ?", teamID).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func (impl *ResourceStorage) RetrieveResourceLastModifiedTime(teamID int) (time.Time, error) {
	var resource *model.Resource
	if err := impl.db.Where("team_id = ?", teamID).Order("updated_at desc").First(&resource).Error; err != nil {
		return time.Time{}, err
	}
	return resource.ExportUpdatedAt(), nil
}

func (impl *ResourceStorage) Delete(teamID int, resourceID int) error {
	if err := impl.db.Where("id = ? AND team_id = ?", resourceID, teamID).Delete(&model.Resource{}).Error; err != nil {
		return err
	}
	return nil
}
