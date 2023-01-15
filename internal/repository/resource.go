// Copyright 2022 The ILLA Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/illacloud/builder-backend/pkg/db"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Resource struct {
	ID        int       `gorm:"column:id;type:bigserial;primary_key"`
	UID       uuid.UUID `gorm:"column:uid;type:uuid;not null"`
	TeamID    int       `gorm:"column:team_id;type:bigserial"`
	Name      string    `gorm:"column:name;type:varchar;size:200;not null"`
	Type      int       `gorm:"column:type;type:smallint;not null"`
	Options   db.JSONB  `gorm:"column:options;type:jsonb"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;not null"`
	CreatedBy int       `gorm:"column:created_by;type:bigint;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;not null"`
	UpdatedBy int       `gorm:"column:updated_by;type:bigint;not null"`
}

func (resource *Resource) ExportUpdatedAt() time.Time {
	return resource.UpdatedAt
}

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
	if err := impl.db.Delete(&Resource{}).Where("id = ? AND team_id = ?", resourceID, teamID).Error; err != nil {
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
