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

package action

import (
	"errors"
	"time"

	"github.com/illa-family/builder-backend/internal/repository"
	"github.com/illa-family/builder-backend/pkg/connector"

	"go.uber.org/zap"
)

var type_array = [8]string{"transformer", "restapi", "graphql", "redis", "mysql", "mariadb", "postgresql", "mongodb"}
var type_map = map[string]int{
	"transformer": 0,
	"restapi":     1,
	"graphql":     2,
	"redis":       3,
	"mysql":       4,
	"mariadb":     5,
	"postgresql":  6,
	"mongodb":     7,
}

type ActionService interface {
	CreateAction(action ActionDto) (ActionDto, error)
	DeleteAction(id int) error
	UpdateAction(action ActionDto) (ActionDto, error)
	GetAction(id int) (ActionDto, error)
	FindActionsByAppVersion(app, version int) ([]ActionDto, error)
	RunAction(action ActionDto) (interface{}, error)
}

type ActionDto struct {
	ID          int                    `json:"actionId"`
	App         int                    `json:"-"`
	Version     int                    `json:"-"`
	Resource    int                    `json:"resourceId,omitempty"`
	DisplayName string                 `json:"displayName,omitempty" validate:"required"`
	Type        string                 `json:"actionType,omitempty" validate:"oneof=transformer restapi graphql redis mysql mariadb postgresql mongodb"`
	Template    map[string]interface{} `json:"actionTemplate,omitempty" validate:"required"`
	CreatedAt   time.Time              `json:"createdAt,omitempty"`
	CreatedBy   int                    `json:"createdBy,omitempty"`
	UpdatedAt   time.Time              `json:"updatedAt,omitempty"`
	UpdatedBy   int                    `json:"updatedBy,omitempty"`
}

type ActionServiceImpl struct {
	logger             *zap.SugaredLogger
	actionRepository   repository.ActionRepository
	resourceRepository repository.ResourceRepository
}

func NewActionServiceImpl(logger *zap.SugaredLogger, actionRepository repository.ActionRepository,
	resourceRepository repository.ResourceRepository) *ActionServiceImpl {
	return &ActionServiceImpl{
		logger:             logger,
		actionRepository:   actionRepository,
		resourceRepository: resourceRepository,
	}
}

func (impl *ActionServiceImpl) CreateAction(action ActionDto) (ActionDto, error) {
	// TODO: guarantee `action` DisplayName unique
	id, err := impl.actionRepository.Create(&repository.Action{
		ID:        action.ID,
		App:       action.App,
		Version:   action.Version,
		Resource:  action.Resource,
		Name:      action.DisplayName,
		Type:      type_map[action.Type],
		Template:  action.Template,
		CreatedAt: action.CreatedAt,
		CreatedBy: action.CreatedBy,
		UpdatedAt: action.UpdatedAt,
		UpdatedBy: action.UpdatedBy,
	})
	if err != nil {
		return ActionDto{}, err
	}
	action.ID = id

	return action, nil
}

func (impl *ActionServiceImpl) DeleteAction(id int) error {
	if err := impl.actionRepository.Delete(id); err != nil {
		return err
	}
	return nil
}

func (impl *ActionServiceImpl) UpdateAction(action ActionDto) (ActionDto, error) {
	// TODO: guarantee `action` DisplayName unique
	if err := impl.actionRepository.Update(&repository.Action{
		ID:        action.ID,
		Resource:  action.Resource,
		Name:      action.DisplayName,
		Type:      type_map[action.Type],
		Template:  action.Template,
		UpdatedAt: action.UpdatedAt,
		UpdatedBy: action.UpdatedBy,
	}); err != nil {
		return ActionDto{}, err
	}
	return action, nil
}

func (impl *ActionServiceImpl) GetAction(id int) (ActionDto, error) {
	res, err := impl.actionRepository.RetrieveByID(id)
	if err != nil {
		return ActionDto{}, err
	}
	resDto := ActionDto{
		ID:          res.ID,
		Resource:    res.Resource,
		DisplayName: res.Name,
		Type:        type_array[res.Type],
		Template:    res.Template,
		CreatedBy:   res.CreatedBy,
		CreatedAt:   res.CreatedAt,
		UpdatedBy:   res.UpdatedBy,
		UpdatedAt:   res.UpdatedAt,
	}
	return resDto, nil
}

func (impl *ActionServiceImpl) FindActionsByAppVersion(app, version int) ([]ActionDto, error) {
	res, err := impl.actionRepository.RetrieveActionsByAppVersion(app, version)
	if err != nil {
		return nil, err
	}

	resDtoSlice := make([]ActionDto, 0, len(res))
	for _, value := range res {
		resDtoSlice = append(resDtoSlice, ActionDto{
			ID:          value.ID,
			Resource:    value.Resource,
			DisplayName: value.Name,
			Type:        type_array[value.Type],
			Template:    value.Template,
			CreatedBy:   value.CreatedBy,
			CreatedAt:   value.CreatedAt,
			UpdatedBy:   value.UpdatedBy,
			UpdatedAt:   value.UpdatedAt,
		})
	}
	return resDtoSlice, nil
}

func (impl *ActionServiceImpl) RunAction(action ActionDto) (interface{}, error) {
	var actionFactory *Factory
	if action.ID != 0 {
		rsc, err := impl.resourceRepository.RetrieveByID(action.ID)
		if err != nil {
			return nil, err
		}
		resourceConn := &connector.Connector{
			Type:    type_array[rsc.Type],
			Options: rsc.Options,
		}
		actionFactory = &Factory{
			Type:     action.Type,
			Template: action.Template,
			Resource: resourceConn,
		}
	} else {
		actionFactory = &Factory{
			Type:     action.Type,
			Template: action.Template,
			Resource: nil,
		}
	}
	actionAssemblyLine := actionFactory.Build()
	if actionAssemblyLine == nil {
		return nil, errors.New("invalid ActionType:: unsupported type")
	}
	res, err := actionAssemblyLine.Run()
	if err != nil {
		return nil, err
	}
	return res, nil
}
