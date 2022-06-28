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

	"github.com/illa-family/builder-backend/pkg/db"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Action struct {
	ID             uuid.UUID `gorm:"column:id;type:uuid;default:uuid_generate_v4();primary_key;unique"`
	VersionID      uuid.UUID `gorm:"column:version_id;type:uuid"`
	ResourceID     uuid.UUID `gorm:"column:resource_id;type:uuid"`
	Name           string    `gorm:"column:name;type:varchar"`
	Type           string    `gorm:"column:type;type:varchar"`
	ActionTemplate db.JSONB  `gorm:"column:action_template;type:jsonb"`
	CreatedBy      uuid.UUID `gorm:"column:created_by;type:uuid"`
	CreatedAt      time.Time `gorm:"column:created_at;type:timestamp"`
	UpdatedBy      uuid.UUID `gorm:"column:updated_by;type:uuid"`
	UpdatedAt      time.Time `gorm:"column:updated_at;type:timestamp"`
}

type ActionRepository interface {
	Create(action *Action) error
	Delete(actionId uuid.UUID) error
	Update(action *Action) error
	RetrieveById(actionId uuid.UUID) (*Action, error)
	RetrieveActionsByVersion(versionId uuid.UUID) ([]*Action, error)
}

type ActionRepositoryImpl struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewActionRepositoryImpl(logger *zap.SugaredLogger, db *gorm.DB) *ActionRepositoryImpl {
	return &ActionRepositoryImpl{
		logger: logger,
		db:     db,
	}
}

func (impl *ActionRepositoryImpl) Create(action *Action) error {
	if err := impl.db.Create(action).Error; err != nil {
		return err
	}
	return nil
}

func (impl *ActionRepositoryImpl) Delete(actionId uuid.UUID) error {
	if err := impl.db.Delete(&Action{}, actionId).Error; err != nil {
		return err
	}
	return nil
}

func (impl *ActionRepositoryImpl) Update(action *Action) error {
	if err := impl.db.Model(action).Updates(Action{
		Name:           action.Name,
		ActionTemplate: action.ActionTemplate,
		UpdatedBy:      action.UpdatedBy,
		UpdatedAt:      action.UpdatedAt,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *ActionRepositoryImpl) RetrieveById(actionId uuid.UUID) (*Action, error) {
	action := &Action{}
	if err := impl.db.First(action, actionId).Error; err != nil {
		return &Action{}, err
	}
	return action, nil
}

func (impl *ActionRepositoryImpl) RetrieveActionsByVersion(versionId uuid.UUID) ([]*Action, error) {
	var actions []*Action
	if err := impl.db.Where("version_id = ?", versionId).Find(&actions).Error; err != nil {
		return nil, err
	}
	return actions, nil
}
