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
	"github.com/google/uuid"
	"github.com/illa-family/builder-backend/pkg/db"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

type Resource struct {
	ID        uuid.UUID `gorm:"column:id;type:uuid;default:uuid_generate_v4();primary_key;unique"`
	Name      string    `gorm:"column:name;type:varchar"`
	Kind      string    `gorm:"column:kind;type:varchar"`
	Options   db.JSONB  `gorm:"column:options;type:jsonb"`
	CreatedBy uuid.UUID `gorm:"column:created_by;type:uuid"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp"`
	UpdatedBy uuid.UUID `gorm:"column:updated_by;type:uuid"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp"`
}

type ResourceRepository interface {
	Create(resource *Resource) error
	Delete(resourceId uuid.UUID) error
	Update(resource *Resource) error
	RetrieveById(resourceId uuid.UUID) (*Resource, error)
	RetrieveAll() ([]*Resource, error)
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

func (impl *ResourceRepositoryImpl) Create(resource *Resource) error {
	if err := impl.db.Create(resource).Error; err != nil {
		return err
	}
	return nil
}

func (impl *ResourceRepositoryImpl) Delete(resourceId uuid.UUID) error {
	if err := impl.db.Delete(&Resource{}, resourceId).Error; err != nil {
		return err
	}
	return nil
}

func (impl *ResourceRepositoryImpl) Update(resource *Resource) error {
	if err := impl.db.Model(resource).Updates(Resource{
		Name:      resource.Name,
		Options:   resource.Options,
		UpdatedBy: resource.UpdatedBy,
		UpdatedAt: resource.UpdatedAt,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *ResourceRepositoryImpl) RetrieveById(resourceId uuid.UUID) (*Resource, error) {
	resource := &Resource{}
	if err := impl.db.First(resource, resourceId).Error; err != nil {
		return &Resource{}, err
	}
	return resource, nil
}

func (impl *ResourceRepositoryImpl) RetrieveAll() ([]*Resource, error) {
	var resources []*Resource
	if err := impl.db.Find(&resources).Error; err != nil {
		return nil, err
	}
	return resources, nil
}
