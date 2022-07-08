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

type Version struct {
	ID                    uuid.UUID `gorm:"column:id;type:uuid;default:uuid_generate_v4();primary_key;unique"`
	AppID                 uuid.UUID `gorm:"column:id;type:uuid""`
	Name                  string    `gorm:"column:name;type:varchar"`
	Components            db.JSONB  `gorm:"column:components;type:jsonb"`
	DependenciesState     db.JSONB  `gorm:"column:dependencies_state;type:jsonb"`
	ExecutionState        db.JSONB  `gorm:"column:execution_state;type:jsonb"`
	DragShadowState       db.JSONB  `gorm:"column:drag_shadow_state;type:jsonb"`
	DottedLineSquareState db.JSONB  `gorm:"column:dotted_line_square_state;type:jsonb"`
	DisplaynameState      db.JSONB  `gorm:"column:displayname_state;type:jsonb"`
	CreatedBy             uuid.UUID `gorm:"column:created_by;type:uuid"`
	CreatedAt             time.Time `gorm:"column:created_at;type:timestamp"`
	UpdatedBy             uuid.UUID `gorm:"column:updated_by;type:uuid"`
	UpdatedAt             time.Time `gorm:"column:updated_at;type:timestamp"`
}

type AppVersionRepository interface {
	Create(appVersion *Version) error
	Delete(appVersionId uuid.UUID) error
	Update(appVersion *Version) error
	FetchVersionById(appVersionId uuid.UUID) (*Version, error)
}

type AppVersionRepositoryImpl struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewAppVersionRepositoryImpl(logger *zap.SugaredLogger, db *gorm.DB) *AppVersionRepositoryImpl {
	return &AppVersionRepositoryImpl{
		logger: logger,
		db:     db,
	}
}

func (impl *AppVersionRepositoryImpl) Create(appVersion *Version) error {
	if err := impl.db.Create(appVersion).Error; err != nil {
		return err
	}
	return nil
}

func (impl *AppVersionRepositoryImpl) Delete(appVersionId uuid.UUID) error {
	if err := impl.db.Delete(&App{}, appVersionId).Error; err != nil {
		return err
	}
	return nil
}

func (impl *AppVersionRepositoryImpl) Update(appVersion *Version) error {
	if err := impl.db.Model(appVersion).Updates(Version{
		Name:                  appVersion.Name,
		Components:            appVersion.Components,
		DependenciesState:     appVersion.DependenciesState,
		ExecutionState:        appVersion.ExecutionState,
		DragShadowState:       appVersion.DragShadowState,
		DottedLineSquareState: appVersion.DottedLineSquareState,
		DisplaynameState:      appVersion.DisplaynameState,
		UpdatedBy:             appVersion.UpdatedBy,
		UpdatedAt:             appVersion.UpdatedAt,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *AppVersionRepositoryImpl) FetchVersionById(appVersionId uuid.UUID) (*Version, error) {
	version := &Version{}
	if err := impl.db.First(version, appVersionId).Error; err != nil {
		return &Version{}, err
	}
	return version, nil
}
