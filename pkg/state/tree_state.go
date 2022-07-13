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
	"github.com/illa-family/builder-backend/internal/repository"
	"github.com/illa-family/builder-backend/pkg/connector"
	"go.uber.org/zap"
)

type TreeStateService interface {
	CreateTreeState(versionId uuid.UUID, treestate TreeStateDto) (TreeStateDto, error)
	DeleteTreeState(treestateId uuid.UUID) error
	UpdateTreeState(versionId uuid.UUID, treestate TreeStateDto) (TreeStateDto, error)
	GetTreeState(treestateID uuid.UUID) (TreeStateDto, error)
	FindTreeStatesByVersion(versionId uuid.UUID) ([]TreeStateDto, error)
	RunTreeState(treestate TreeStateDto) (interface{}, error)
}

type TreeStateDto struct {
	ID                 int       `json:"id"`
	StateType          int       `json:"state_type"`
	ParentNodeRefID    int       `json:"parent_node_ref_id"`
	ChildrenNodeRefIDs int       `json:"children_node_ref_ids"`
	AppRefID           int       `json:"app_ref_id"`
	Version            int       `json:"version"`
	Name               string    `json:"name"`
	Content            string    `json:"content"`
	CreatedAt          time.Time `json:"created_at"`
	CreatedBy          int       `json:"created_by"`
	UpdatedAt          time.Time `json:"updated_at"`
	UpdatedBy          int       `json:"updated_by"`
}

type TreeStateServiceImpl struct {
	logger              *zap.SugaredLogger
	treestateRepository repository.TreeStateRepository
	resourceRepository  repository.ResourceRepository
}

func NewTreeStateServiceImpl(logger *zap.SugaredLogger, treestateRepository repository.TreeStateRepository,
	resourceRepository repository.ResourceRepository) *TreeStateServiceImpl {
	return &TreeStateServiceImpl{
		logger:              logger,
		treestateRepository: treestateRepository,
		resourceRepository:  resourceRepository,
	}
}

func (impl *TreeStateServiceImpl) CreateTreeState(treestate TreeStateDto) (TreeStateDto, error) {
	// TODO: validate the versionId
	validate := validator.New()
	if err := validate.Struct(treestate); err != nil {
		return TreeStateDto{}, err
	}
	treestate.CreatedAt = time.Now().UTC()
	treestate.UpdatedAt = time.Now().UTC()
	if err := impl.treestateRepository.Create(&repository.TreeState{
		StateType:          treestate.StateType,
		ParentNodeRefID:    treestate.ParentNodeRefID,
		ChildrenNodeRefIDs: treestate.ChildrenNodeRefIDs,
		AppRefID:           treestate.AppRefID,
		Version:            treestate.Version,
		Name:               treestate.Name,
		Content:            treestate.Content,
		CreatedAt:          treestate.CreatedAt,
		CreatedBy:          treestate.CreatedBy,
		UpdatedAt:          treestate.UpdatedAt,
		UpdatedBy:          treestate.UpdatedBy,
	}); err != nil {
		return TreeStateDto{}, err
	}
	return treestate, nil
}

func (impl *TreeStateServiceImpl) DeleteTreeState(treestateID int) error {
	if err := impl.treestateRepository.Delete(treestateID); err != nil {
		return err
	}
	return nil
}

func (impl *TreeStateServiceImpl) UpdateTreeState(treestate TreeStateDto) (TreeStateDto, error) {
	validate := validator.New()
	if err := validate.Struct(treestate); err != nil {
		return TreeStateDto{}, err
	}
	treestate.UpdatedAt = time.Now().UTC()
	if err := impl.treestateRepository.Update(&repository.TreeState{
		ID:                 treestate.ID,
		StateType:          treestate.StateType,
		ParentNodeRefID:    treestate.ParentNodeRefID,
		ChildrenNodeRefIDs: treestate.ChildrenNodeRefIDs,
		AppRefID:           treestate.AppRefID,
		Version:            treestate.Version,
		Name:               treestate.Name,
		Content:            treestate.Content,
		UpdatedAt:          treestate.UpdatedAt,
		UpdatedBy:          treestate.UpdatedBy,
	}); err != nil {
		return TreeStateDto{}, err
	}
	return treestate, nil
}

func (impl *TreeStateServiceImpl) GetTreeState(treestateID int) (TreeStateDto, error) {
	res, err := impl.treestateRepository.RetrieveById(treestateID)
	if err != nil {
		return TreeStateDto{}, err
	}
	resDto := TreeStateDto{
		ID:                 res.ID,
		StateType:          res.StateType,
		ParentNodeRefID:    res.ParentNodeRefID,
		ChildrenNodeRefIDs: res.ChildrenNodeRefIDs,
		AppRefID:           res.AppRefID,
		Version:            res.Version,
		Name:               res.Name,
		Content:            res.Content,
		CreatedAt:          res.CreatedAt,
		CreatedBy:          res.CreatedBy,
		UpdatedAt:          res.UpdatedAt,
		UpdatedBy:          res.UpdatedBy,
	}
	return resDto, nil
}

func (impl *TreeStateServiceImpl) FindTreeStatesByVersion(versionId uuid.UUID) ([]TreeStateDto, error) {
	res, err := impl.treestateRepository.RetrieveTreeStatesByVersion(versionId)
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

func (impl *TreeStateServiceImpl) RunTreeState(treestate TreeStateDto) (interface{}, error) {
	var treestateFactory *Factory
	if treestate.ResourceId != uuid.Nil {
		rsc, err := impl.resourceRepository.RetrieveById(treestate.ResourceId)
		if err != nil {
			return nil, err
		}
		resourceConn := &connector.Connector{
			Type:    rsc.Kind,
			Options: rsc.Options,
		}
		treestateFactory = &Factory{
			Type:     treestate.TreeStateType,
			Template: treestate.TreeStateTemplate,
			Resource: resourceConn,
		}
	} else {
		treestateFactory = &Factory{
			Type:     treestate.TreeStateType,
			Template: treestate.TreeStateTemplate,
			Resource: nil,
		}
	}
	treestateAssemblyline := treestateFactory.Build()
	if treestateAssemblyline == nil {
		return nil, errors.New("invalid TreeStateType:: unsupported type")
	}
	res, err := treestateAssemblyline.Run()
	if err != nil {
		return nil, err
	}
	return res, nil
}
