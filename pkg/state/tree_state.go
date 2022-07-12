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

package state

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid"
	"github.com/illa-family/builder-backend/pkg/connector"
	"go.uber.org/zap"
)

type TreeStateService interface {
	CreateTreeState(versionId uuid.UUID, streestate TreeStateDto) (TreeStateDto, error)
	DeleteTreeState(streestateId uuid.UUID) error
	UpdateTreeState(versionId uuid.UUID, streestate TreeStateDto) (TreeStateDto, error)
	GetTreeState(streestateID uuid.UUID) (TreeStateDto, error)
	FindTreeStatesByVersion(versionId uuid.UUID) ([]TreeStateDto, error)
	RunTreeState(streestate TreeStateDto) (interface{}, error)
}

type TreeStateDto struct {
	ID                 int       `json:"id"`
	StateType          int       `json:"state_type"`
	ParentNodeRefID    int       `json:"parent_node_ref_id"`
	ChildrenNodeRefIDs int       `json:"children_node_ref_ids"`
	AppRefID           int       `json:"app_ref_id"`
	Version            int       `json:"version"`
	Content            string    `json:"content"`
	CreatedAt          time.Time `json:"created_at"`
	CreatedBy          int       `json:"created_by"`
	UpdatedAt          time.Time `json:"updated_at"`
	UpdatedBy          int       `json:"updated_by"`
}

type TreeStateServiceImpl struct {
	logger               *zap.SugaredLogger
	streestateRepository repository.TreeStateRepository
	resourceRepository   repository.ResourceRepository
}

func NewTreeStateServiceImpl(logger *zap.SugaredLogger, streestateRepository repository.TreeStateRepository,
	resourceRepository repository.ResourceRepository) *TreeStateServiceImpl {
	return &TreeStateServiceImpl{
		logger:               logger,
		streestateRepository: streestateRepository,
		resourceRepository:   resourceRepository,
	}
}

func (impl *TreeStateServiceImpl) CreateTreeState(versionId uuid.UUID, streestate TreeStateDto) (TreeStateDto, error) {
	// TODO: validate the versionId
	validate := validator.New()
	if err := validate.Struct(streestate); err != nil {
		return TreeStateDto{}, err
	}
	streestate.CreatedAt = time.Now().UTC()
	streestate.UpdatedAt = time.Now().UTC()
	if err := impl.streestateRepository.Create(&repository.TreeState{
		ID:                streestate.TreeStateId,
		VersionID:         versionId,
		ResourceID:        streestate.ResourceId,
		Name:              streestate.DisplayName,
		Type:              streestate.TreeStateType,
		TreeStateTemplate: streestate.TreeStateTemplate,
		CreatedBy:         streestate.CreatedBy,
		CreatedAt:         streestate.CreatedAt,
		UpdatedBy:         streestate.UpdatedBy,
		UpdatedAt:         streestate.UpdatedAt,
	}); err != nil {
		return TreeStateDto{}, err
	}
	return streestate, nil
}

func (impl *TreeStateServiceImpl) DeleteTreeState(streestateId uuid.UUID) error {
	if err := impl.streestateRepository.Delete(streestateId); err != nil {
		return err
	}
	return nil
}

func (impl *TreeStateServiceImpl) UpdateTreeState(versionId uuid.UUID, streestate TreeStateDto) (TreeStateDto, error) {
	validate := validator.New()
	if err := validate.Struct(streestate); err != nil {
		return TreeStateDto{}, err
	}
	streestate.UpdatedAt = time.Now().UTC()
	if err := impl.streestateRepository.Update(&repository.TreeState{
		ID:                streestate.TreeStateId,
		VersionID:         versionId,
		ResourceID:        streestate.ResourceId,
		Name:              streestate.DisplayName,
		Type:              streestate.TreeStateType,
		TreeStateTemplate: streestate.TreeStateTemplate,
		CreatedBy:         streestate.CreatedBy,
		CreatedAt:         streestate.CreatedAt,
		UpdatedBy:         streestate.UpdatedBy,
		UpdatedAt:         streestate.UpdatedAt,
	}); err != nil {
		return TreeStateDto{}, err
	}
	return streestate, nil
}

func (impl *TreeStateServiceImpl) GetTreeState(streestateId uuid.UUID) (TreeStateDto, error) {
	res, err := impl.streestateRepository.RetrieveById(streestateId)
	if err != nil {
		return TreeStateDto{}, err
	}
	resDto := TreeStateDto{
		TreeStateId:       res.ID,
		ResourceId:        res.ResourceID,
		DisplayName:       res.Name,
		TreeStateType:     res.Type,
		TreeStateTemplate: res.TreeStateTemplate,
		CreatedBy:         res.CreatedBy,
		CreatedAt:         res.CreatedAt,
		UpdatedBy:         res.UpdatedBy,
		UpdatedAt:         res.UpdatedAt,
	}
	return resDto, nil
}

func (impl *TreeStateServiceImpl) FindTreeStatesByVersion(versionId uuid.UUID) ([]TreeStateDto, error) {
	res, err := impl.streestateRepository.RetrieveTreeStatesByVersion(versionId)
	if err != nil {
		return nil, err
	}
	resDtoSlice := make([]TreeStateDto, 0, len(res))
	for _, value := range res {
		resDtoSlice = append(resDtoSlice, TreeStateDto{
			TreeStateId:       value.ID,
			ResourceId:        value.ResourceID,
			DisplayName:       value.Name,
			TreeStateType:     value.Type,
			TreeStateTemplate: value.TreeStateTemplate,
			CreatedBy:         value.CreatedBy,
			CreatedAt:         value.CreatedAt,
			UpdatedBy:         value.UpdatedBy,
			UpdatedAt:         value.UpdatedAt,
		})
	}
	return resDtoSlice, nil
}

func (impl *TreeStateServiceImpl) RunTreeState(streestate TreeStateDto) (interface{}, error) {
	var streestateFactory *Factory
	if streestate.ResourceId != uuid.Nil {
		rsc, err := impl.resourceRepository.RetrieveById(streestate.ResourceId)
		if err != nil {
			return nil, err
		}
		resourceConn := &connector.Connector{
			Type:    rsc.Kind,
			Options: rsc.Options,
		}
		streestateFactory = &Factory{
			Type:     streestate.TreeStateType,
			Template: streestate.TreeStateTemplate,
			Resource: resourceConn,
		}
	} else {
		streestateFactory = &Factory{
			Type:     streestate.TreeStateType,
			Template: streestate.TreeStateTemplate,
			Resource: nil,
		}
	}
	streestateAssemblyline := streestateFactory.Build()
	if streestateAssemblyline == nil {
		return nil, errors.New("invalid TreeStateType:: unsupported type")
	}
	res, err := streestateAssemblyline.Run()
	if err != nil {
		return nil, err
	}
	return res, nil
}
