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
	ID              int       `json:"id" 				gorm:"column:id;type:bigserial;primary_key;unique"`
	UID             uuid.UUID `json:"uid"   		    gorm:"column:uid;type:uuid;not null"`
	TeamID          int       `json:"team_id" 		    gorm:"column:team_id;type:bigserial"`
	Name            string    `json:"name" 				gorm:"column:name;type:varchar"`
	ReleaseVersion  int       `json:"release_version" 	gorm:"column:release_version;type:bigserial"`
	MainlineVersion int       `json:"mainline_version" 	gorm:"column:mainline_version;type:bigserial"`
	CreatedAt       time.Time `json:"created_at" 		gorm:"column:created_at;type:timestamp"`
	CreatedBy       int       `json:"created_by" 		gorm:"column:created_by;type:bigserial"`
	UpdatedAt       time.Time `json:"updated_at" 		gorm:"column:updated_at;type:timestamp"`
	UpdatedBy       int       `json:"updated_by" 		gorm:"column:updated_by;type:bigserial"`
}

func (app *App) ExportUpdatedAt() time.Time {
	return app.UpdatedAt
}

type AppRepository interface {
	Create(app *App) (int, error)
	Delete(teamID int, appID int) error
	Update(app *App) error
	UpdateUpdatedAt(app *App) error
	RetrieveAll(teamID int) ([]*App, error)
	RetrieveAppByID(teamID int, appID int) (*App, error)
	RetrieveAllByUpdatedTime(teamID int) ([]*App, error)
	CountAPPByTeamID(teamID int) (int, error)
	RetrieveAppLastModifiedTime(teamID int) (time.Time, error)
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

func (impl *AppRepositoryImpl) Delete(teamID int, appID int) error {
	if err := impl.db.Delete(&App{}).Where("id = ? AND team_id = ?", app, teamID).Error; err != nil {
		return err
	}
	return nil
}

func (impl *AppRepositoryImpl) Update(app *App) error {
	if err := impl.db.Model(app).UpdateColumns(App{
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

func (impl *AppRepositoryImpl) RetrieveAll(teamID int) ([]*App, error) {
	var apps []*App
	if err := impl.db.Where("team_id = ?", teamID).Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (impl *AppRepositoryImpl) RetrieveAppByID(teamID int,appID int) (*App, error) {
	var app *App
	if err := impl.db.Where("id = ? AND team_id = ?", appID, teamID).Find(&app).Error; err != nil {
		return nil, err
	}
	return app, nil
}

func (impl *AppRepositoryImpl) RetrieveAllByUpdatedTime(teamID int) ([]*App, error) {
	var apps []*App
	if err := impl.db.Where("team_id = ?", teamID).Order("updated_at desc").Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (impl *AppRepositoryImpl) UpdateUpdatedAt(app *App) error {
	if err := impl.db.Model(app).UpdateColumns(App{
		UpdatedBy: app.UpdatedBy,
		UpdatedAt: app.UpdatedAt,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *AppRepositoryImpl) CountAPPByTeamID(teamID int) (int, error) {
	var count int64
	if err := impl.db.Where("team_id = ?", teamID).Count(&count).Error; err !=nil {
		return 0, err
	}
	return int(count), nil
}

func (impl *AppRepositoryImpl) RetrieveAppLastModifiedTime(teamID int) (time.Time, error) {
	var app *App
	if err := impl.db.Where("team_id = ?", teamID).Order("updated_at desc").First(&app).Error; err != nil {
		return time.Time{}, err
	}
	return app.ExportUpdatedAt(), nil
}
