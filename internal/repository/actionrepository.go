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
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Action struct {
	ID int
}

type ActionRepository interface {
	Save(action *Action) (*Action, error)
	Delete(actionId string) error
	Update(action *Action) (*Action, error)
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

func (impl *ActionRepositoryImpl) Save(action *Action) (*Action, error) {
	return &Action{}, nil
}

func (impl *ActionRepositoryImpl) Delete(actionId string) error {
	return nil
}

func (impl *ActionRepositoryImpl) Update(action *Action) (*Action, error) {
	return &Action{}, nil
}
