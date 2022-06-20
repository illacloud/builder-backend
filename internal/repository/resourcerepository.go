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
	"github.com/jackc/pgtype"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Resource struct {
	ID      uuid.UUID    `gorm:"column:id;type:uuid;default:uuid_generate_v4();primary_key;unique"`
	Name    string       `gorm:"column:name;type:varchar"`
	Kind    string       `gorm:"column:kind;type:varchar"`
	Options pgtype.JSONB `gorm:"column:options;type:jsonb"`
	db.AuditLog
}

type ResourceRepository interface {
	Save(resource *Resource) error
	Delete(resourceId string) error
	Update(resource *Resource) (*Resource, error)
	RetrieveById(resourceId string) (*Resource, error)
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

func (impl *ResourceRepositoryImpl) Save(resource *Resource) error {
	if err := impl.db.Create(resource).Error; err != nil {
		return err
	}
	return nil
}

func (impl *ResourceRepositoryImpl) Delete(resourceId string) error {
	return nil
}

func (impl *ResourceRepositoryImpl) Update(resource *Resource) (*Resource, error) {
	return &Resource{}, nil
}

func (impl *ResourceRepositoryImpl) RetrieveById(resourceId string) (*Resource, error) {
	return &Resource{}, nil
}

func (impl *ResourceRepositoryImpl) RetrieveAll() ([]*Resource, error) {
	return nil, nil
}
