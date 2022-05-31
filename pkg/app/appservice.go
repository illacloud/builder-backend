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
	"github.com/illa-family/builder-backend/internal/repository"
	"go.uber.org/zap"
)

type AppService interface {
	CreateApp(app AppDto) (AppDto, error)
	DeleteApp(appId string) error
	UpdateApp(app AppDto) (AppDto, error)
	GetAllApp() ([]AppDto, error)
}

type AppServiceImpl struct {
	logger        *zap.SugaredLogger
	appRepository repository.AppRepository
}

type AppDto struct {
	AppId       int    `json:"appId,omitempty"`
	AppName     string `json:"appName"`
	AppActivity string `json:"appActivity"`
}

func NewAppServiceImpl(logger *zap.SugaredLogger, appRepository repository.AppRepository) *AppServiceImpl {
	return &AppServiceImpl{
		logger:        logger,
		appRepository: appRepository,
	}
}

func (appServiceImpl *AppServiceImpl) CreateApp(app AppDto) (AppDto, error) {
	return AppDto{}, nil
}

func (appServiceImpl *AppServiceImpl) UpdateApp(app AppDto) (AppDto, error) {
	return AppDto{}, nil
}

func (appServiceImpl *AppServiceImpl) DeleteApp(appId string) error {
	return nil
}

func (appServiceImpl *AppServiceImpl) GetAllApp() ([]AppDto, error) {
	return nil, nil
}
