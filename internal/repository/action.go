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

	"github.com/illacloud/builder-backend/pkg/db"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Action struct {
	ID          int       `gorm:"column:id;type:bigserial;primary_key"`
	App         int       `gorm:"column:app_ref_id;type:bigint;not null"`
	Version     int       `gorm:"column:version;type:bigint;not null"`
	Resource    int       `gorm:"column:resource_ref_id;type:bigint;not null"`
	Name        string    `gorm:"column:name;type:varchar;size:255;not null"`
	Type        int       `gorm:"column:type;type:smallint;not null"`
	TriggerMode string    `gorm:"column:trigger_mode;type:varchar;size:16;not null"`
	Transformer db.JSONB  `gorm:"column:transformer;type:jsonb"`
	Template    db.JSONB  `gorm:"column:template;type:jsonb"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp;not null"`
	CreatedBy   int       `gorm:"column:created_by;type:bigint;not null"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamp;not null"`
	UpdatedBy   int       `gorm:"column:updated_by;type:bigint;not null"`
}

type ActionRepository interface {
	Create(action *Action) (int, error)
	Delete(id int) error
	Update(action *Action) error
	RetrieveByID(id int) (*Action, error)
	RetrieveActionsByAppVersion(app, version int) ([]*Action, error)
	DeleteActionsByApp(appID int) error
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

func (impl *ActionRepositoryImpl) Create(action *Action) (int, error) {
	if err := impl.db.Create(action).Error; err != nil {
		return 0, err
	}
	return action.ID, nil
}

func (impl *ActionRepositoryImpl) Delete(id int) error {
	if err := impl.db.Delete(&Action{}, id).Error; err != nil {
		return err
	}
	return nil
}

func (impl *ActionRepositoryImpl) Update(action *Action) error {
	if err := impl.db.Model(action).UpdateColumns(Action{
		Resource:    action.Resource,
		Type:        action.Type,
		Name:        action.Name,
		TriggerMode: action.TriggerMode,
		Transformer: action.Transformer,
		Template:    action.Template,
		UpdatedBy:   action.UpdatedBy,
		UpdatedAt:   action.UpdatedAt,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *ActionRepositoryImpl) RetrieveByID(id int) (*Action, error) {
	action := &Action{}
	if err := impl.db.First(action, id).Error; err != nil {
		return &Action{}, err
	}
	return action, nil
}

func (impl *ActionRepositoryImpl) RetrieveActionsByAppVersion(app, version int) ([]*Action, error) {
	var actions []*Action
	if err := impl.db.Where("app_ref_id = ? AND version = ?", app, version).Find(&actions).Error; err != nil {
		return nil, err
	}
	return actions, nil
}

func (impl *ActionRepositoryImpl) DeleteActionsByApp(appID int) error {
	if err := impl.db.Where("app_ref_id = ?", appID).Delete(&Action{}).Error; err != nil {
		return err
	}
	return nil
}
