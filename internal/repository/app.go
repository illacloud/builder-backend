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

	"go.uber.org/zap"
	"gorm.io/gorm"
)

const APP_EDIT_VERSION = 0 // the editable version app ID always be 0

type App struct {
	ID              int       `json:"id" 				gorm:"column:id;type:uuid;default:uuid_generate_v4();primary_key;unique"`
	Name            string    `json:"name" 				gorm:"column:name;type:varchar"`
	ReleaseVersion  int       `json:"release_version" 	gorm:"column:release_version;type:uuid"`
	MainlineVersion int       `json:"mainline_version" 	gorm:"column:mainline_version;type:uuid"`
	CreatedAt       time.Time `json:"created_at" 		gorm:"column:created_at;type:timestamp"`
	CreatedBy       int       `json:"created_by" 		gorm:"column:created_by;type:uuid"`
	UpdatedAt       time.Time `json:"updated_at" 		gorm:"column:updated_at;type:timestamp"`
	UpdatedBy       int       `json:"updated_by" 		gorm:"column:updated_by;type:uuid"`
}

type AppRepository interface {
	Create(app *App) (int, error)
	Delete(appID int) error
	Update(app *App) error
	RetrieveAll() ([]*App, error)
	RetrieveAppByID(appID int) (*App, error)
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

func (impl *AppRepositoryImpl) Create(app *App) (int, error) {
	if err := impl.db.Create(app).Error; err != nil {
		return 0, err
	}
	return app.ID, nil
}

func (impl *AppRepositoryImpl) Delete(appID int) error {
	if err := impl.db.Delete(&App{}, appID).Error; err != nil {
		return err
	}
	return nil
}

func (impl *AppRepositoryImpl) Update(app *App) error {
	if err := impl.db.Model(app).Updates(App{
		Name:            app.Name,
		ReleaseVersion:  app.ReleaseVersion,
		MainlineVersion: app.MainlineVersion,
		UpdatedBy:       app.UpdatedBy,
		UpdatedAt:       app.UpdatedAt,
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

func (impl *AppRepositoryImpl) RetrieveAppByID(id int) (*App, error) {
	var app *App
	if err := impl.db.Where("id = ?", id).Find(&app).Error; err != nil {
		return nil, err
	}
	return app, nil
}
