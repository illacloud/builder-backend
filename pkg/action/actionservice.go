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
	"time"

	"github.com/illa-family/builder-backend/internal/repository"
	"github.com/illa-family/builder-backend/pkg/connector"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type ActionService interface {
	CreateAction(versionId uuid.UUID, action ActionDto) (ActionDto, error)
	DeleteAction(actionId uuid.UUID) error
	UpdateAction(versionId uuid.UUID, action ActionDto) (ActionDto, error)
	GetAction(actionID uuid.UUID) (ActionDto, error)
	FindActionsByVersion(versionId uuid.UUID) ([]ActionDto, error)
	RunAction(action ActionDto) (interface{}, error)
}

type ActionDto struct {
	ActionId       uuid.UUID              `json:"actionId"`
	ResourceId     uuid.UUID              `json:"resourceId,omitempty"`
	DisplayName    string                 `json:"displayName,omitempty" validate:"required"`
	ActionType     string                 `json:"actionType,omitempty" validate:"required"`
	ActionTemplate map[string]interface{} `json:"actionTemplate,omitempty" validate:"required"`
	CreatedBy      uuid.UUID              `json:"createdBy,omitempty"`
	CreatedAt      time.Time              `json:"createdAt,omitempty"`
	UpdatedBy      uuid.UUID              `json:"updatedBy,omitempty"`
	UpdatedAt      time.Time              `json:"updatedAt,omitempty"`
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

func (impl *ActionServiceImpl) CreateAction(versionId uuid.UUID, action ActionDto) (ActionDto, error) {
	// TODO: validate the versionId
	validate := validator.New()
	if err := validate.Struct(action); err != nil {
		return ActionDto{}, err
	}
	action.CreatedAt = time.Now().UTC()
	action.UpdatedAt = time.Now().UTC()
	if err := impl.actionRepository.Create(&repository.Action{
		ID:             action.ActionId,
		VersionID:      versionId,
		ResourceID:     action.ResourceId,
		Name:           action.DisplayName,
		Type:           action.ActionType,
		ActionTemplate: action.ActionTemplate,
		CreatedBy:      action.CreatedBy,
		CreatedAt:      action.CreatedAt,
		UpdatedBy:      action.UpdatedBy,
		UpdatedAt:      action.UpdatedAt,
	}); err != nil {
		return ActionDto{}, err
	}
	return action, nil
}

func (impl *ActionServiceImpl) DeleteAction(actionId uuid.UUID) error {
	if err := impl.actionRepository.Delete(actionId); err != nil {
		return err
	}
	return nil
}

func (impl *ActionServiceImpl) UpdateAction(versionId uuid.UUID, action ActionDto) (ActionDto, error) {
	validate := validator.New()
	if err := validate.Struct(action); err != nil {
		return ActionDto{}, err
	}
	action.UpdatedAt = time.Now().UTC()
	if err := impl.actionRepository.Update(&repository.Action{
		ID:             action.ActionId,
		VersionID:      versionId,
		ResourceID:     action.ResourceId,
		Name:           action.DisplayName,
		Type:           action.ActionType,
		ActionTemplate: action.ActionTemplate,
		CreatedBy:      action.CreatedBy,
		CreatedAt:      action.CreatedAt,
		UpdatedBy:      action.UpdatedBy,
		UpdatedAt:      action.UpdatedAt,
	}); err != nil {
		return ActionDto{}, err
	}
	return action, nil
}

func (impl *ActionServiceImpl) GetAction(actionId uuid.UUID) (ActionDto, error) {
	res, err := impl.actionRepository.RetrieveById(actionId)
	if err != nil {
		return ActionDto{}, err
	}
	resDto := ActionDto{
		ActionId:       res.ID,
		ResourceId:     res.ResourceID,
		DisplayName:    res.Name,
		ActionType:     res.Type,
		ActionTemplate: res.ActionTemplate,
		CreatedBy:      res.CreatedBy,
		CreatedAt:      res.CreatedAt,
		UpdatedBy:      res.UpdatedBy,
		UpdatedAt:      res.UpdatedAt,
	}
	return resDto, nil
}

func (impl *ActionServiceImpl) FindActionsByVersion(versionId uuid.UUID) ([]ActionDto, error) {
	res, err := impl.actionRepository.RetrieveActionsByVersion(versionId)
	if err != nil {
		return nil, err
	}
	resDtoSlice := make([]ActionDto, 0, len(res))
	for _, value := range res {
		resDtoSlice = append(resDtoSlice, ActionDto{
			ActionId:       value.ID,
			ResourceId:     value.ResourceID,
			DisplayName:    value.Name,
			ActionType:     value.Type,
			ActionTemplate: value.ActionTemplate,
			CreatedBy:      value.CreatedBy,
			CreatedAt:      value.CreatedAt,
			UpdatedBy:      value.UpdatedBy,
			UpdatedAt:      value.UpdatedAt,
		})
	}
	return resDtoSlice, nil
}

func (impl *ActionServiceImpl) RunAction(action ActionDto) (interface{}, error) {
	var actionFactory *Factory
	if action.ResourceId != uuid.Nil {
		rsc, err := impl.resourceRepository.RetrieveById(action.ResourceId)
		if err != nil {
			return nil, err
		}
		resourceConn := &connector.Connector{
			Type:    rsc.Kind,
			Options: rsc.Options,
		}
		actionFactory = &Factory{
			Type:     action.ActionType,
			Template: action.ActionTemplate,
			Resource: resourceConn,
		}
	} else {
		actionFactory = &Factory{
			Type:     action.ActionType,
			Template: action.ActionTemplate,
			Resource: nil,
		}
	}
	actionAssemblyline := actionFactory.Build()
	if actionAssemblyline == nil {
		return nil, errors.New("invalid ActionType:: unsupported type")
	}
	res, err := actionAssemblyline.Run()
	if err != nil {
		return nil, err
	}
	return res, nil
}
