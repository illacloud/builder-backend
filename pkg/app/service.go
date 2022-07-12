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

package app

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/illa-family/builder-backend/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AppService interface {
	CreateApp(app AppDto) (AppDto, error)
	DeleteApp(appId uuid.UUID) error
	UpdateApp(app AppDto) (AppDto, error)
	GetAllApp() ([]AppDto, error)
	GetAppEditingVersion(appId uuid.UUID) (AppVersionDto, error)
}

type AppServiceImpl struct {
	logger        *zap.SugaredLogger
	appRepository repository.AppRepository
}

type AppDto struct {
	AppId            uuid.UUID `json:"appId"`
	AppName          string    `json:"appName" validate:"required"`
	CurrentVersionID uuid.UUID `json:"currentVersionId"`
	CreatedBy        uuid.UUID `json:"createdBy" `
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedBy        uuid.UUID `json:"updatedBy"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type AppVersionDto struct {
}

func NewAppServiceImpl(logger *zap.SugaredLogger, appRepository repository.AppRepository) *AppServiceImpl {
	return &AppServiceImpl{
		logger:        logger,
		appRepository: appRepository,
	}
}

func (impl *AppServiceImpl) CreateApp(app AppDto) (AppDto, error) {
	validate := validator.New()
	if err := validate.Struct(app); err != nil {
		return AppDto{}, err
	}
	app.CreatedAt = time.Now().UTC()
	app.UpdatedAt = time.Now().UTC()
	app.AppId = uuid.New()
	if err := impl.appRepository.Create(&repository.App{
		ID:             app.AppId,
		Name:           app.AppName,
		CurrentVersion: uuid.Nil,
		CreatedBy:      app.CreatedBy,
		CreatedAt:      app.CreatedAt,
		UpdatedBy:      app.UpdatedBy,
		UpdatedAt:      app.UpdatedAt,
	}); err != nil {
		return AppDto{}, err
	}
	versionId, err := impl.CreateAppVersion(app.CreatedBy, app.AppId)
	if err != nil {
		return app, err
	}
	app.CurrentVersionID = versionId
	app.UpdatedAt = time.Now().UTC()
	if err := impl.appRepository.Update(&repository.App{
		ID:             app.AppId,
		Name:           app.AppName,
		CurrentVersion: app.CurrentVersionID,
		CreatedBy:      app.CreatedBy,
		CreatedAt:      app.CreatedAt,
		UpdatedBy:      app.UpdatedBy,
		UpdatedAt:      app.UpdatedAt,
	}); err != nil {
		return app, err
	}
	return app, nil
}

func (impl *AppServiceImpl) UpdateApp(app AppDto) (AppDto, error) {
	app.UpdatedAt = time.Now().UTC()
	if err := impl.appRepository.Update(&repository.App{
		ID:             app.AppId,
		Name:           app.AppName,
		CurrentVersion: app.CurrentVersionID,
		CreatedBy:      app.CreatedBy,
		CreatedAt:      app.CreatedAt,
		UpdatedBy:      app.UpdatedBy,
		UpdatedAt:      app.UpdatedAt,
	}); err != nil {
		return app, err
	}
	return AppDto{}, nil
}

func (impl *AppServiceImpl) DeleteApp(appId uuid.UUID) error {
	return nil
}

func (impl *AppServiceImpl) GetAllApp() ([]AppDto, error) {
	return nil, nil
}

func (impl *AppServiceImpl) GetAppEditingVersion(appId uuid.UUID) (AppVersionDto, error) {
	return AppVersionDto{}, nil
}

func (impl *AppServiceImpl) CreateAppVersion(userId, appId uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, nil
}
