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

type Resource struct {
	Id int
}

type ResourceRepository interface {
	Save(action *Resource) (*Resource, error)
	Delete(actionId string) error
	Update(action *Resource) (*Resource, error)
}

type ResourceRepositoryImpl struct {
	logger *zap.SugaredLogger
	db     *sql.DB
}

func NewResourceRepositoryImpl(logger *zap.SugaredLogger, db *sql.DB) *ResourceRepositoryImpl {
	return &ResourceRepositoryImpl{
		logger: logger,
		db:     db,
	}
}

func (impl *ResourceRepositoryImpl) Save(resource *Resource) (*Resource, error) {
	return &Resource{}, nil
}

func (impl *ResourceRepositoryImpl) Delete(resourceId string) error {
	return nil
}

func (impl *ResourceRepositoryImpl) Update(resource *Resource) (*Resource, error) {
	return &Resource{}, nil
}
