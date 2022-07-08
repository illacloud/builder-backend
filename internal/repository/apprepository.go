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
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type App struct {
	ID             uuid.UUID `gorm:"column:id;type:uuid;default:uuid_generate_v4();primary_key;unique"`
	Name           string    `gorm:"column:name;type:varchar"`
	CurrentVersion uuid.UUID `gorm:"column:current_version;type:uuid"`
	CreatedBy      uuid.UUID `gorm:"column:created_by;type:uuid"`
	CreatedAt      time.Time `gorm:"column:created_at;type:timestamp"`
	UpdatedBy      uuid.UUID `gorm:"column:updated_by;type:uuid"`
	UpdatedAt      time.Time `gorm:"column:updated_at;type:timestamp"`
}

type AppRepository interface {
	Create(app *App) error
	Delete(appId uuid.UUID) error
	Update(app *App) error
	RetrieveAll() ([]*App, error)
}

type AppRepositoryImpl struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewAppRepositoryImpl(logger *zap.SugaredLogger, db *gorm.DB) *AppRepositoryImpl {
	return &AppRepositoryImpl{
		logger: logger,
		db:     db,
	}
}

func (impl *AppRepositoryImpl) Create(app *App) error {
	if err := impl.db.Create(app).Error; err != nil {
		return err
	}
	return nil
}

func (impl *AppRepositoryImpl) Delete(appId uuid.UUID) error {
	if err := impl.db.Delete(&App{}, appId).Error; err != nil {
		return err
	}
	return nil
}

func (impl *AppRepositoryImpl) Update(app *App) error {
	if err := impl.db.Model(app).Updates(App{
		Name:           app.Name,
		CurrentVersion: app.CurrentVersion,
		UpdatedBy:      app.UpdatedBy,
		UpdatedAt:      app.UpdatedAt,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *AppRepositoryImpl) RetrieveAll() ([]*App, error) {
	var apps []*App
	if err := impl.db.Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}
