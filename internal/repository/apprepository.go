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
	"database/sql"
	"go.uber.org/zap"
)

type App struct {
	ID int
}

type AppRepository interface {
	Save(app *App) error
	Delete(appId string) error
	Update(app *App) error
	RetrieveAll() ([]*App, error)
}

type AppRepositoryImpl struct {
	logger *zap.SugaredLogger
	db     *sql.DB
}

func NewAppRepositoryImpl(logger *zap.SugaredLogger, db *sql.DB) *AppRepositoryImpl {
	return &AppRepositoryImpl{
		logger: logger,
		db:     db,
	}
}

func (impl *AppRepositoryImpl) Save(app *App) error {

	return nil
}

func (impl *AppRepositoryImpl) Delete(appId string) error {
	return nil
}

func (impl *AppRepositoryImpl) Update(app *App) error {
	return nil
}

func (impl *AppRepositoryImpl) RetrieveAll() ([]*App, error) {
	return nil, nil
}
