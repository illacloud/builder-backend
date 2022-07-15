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
	"time"

	"github.com/illa-family/builder-backend/internal/repository"
	"github.com/illa-family/builder-backend/pkg/app"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type TreeStateService interface {
	CreateTreeState(treestate TreeStateDto) (TreeStateDto, error)
	DeleteTreeState(treestateId int) error
	UpdateTreeState(treestate TreeStateDto) (TreeStateDto, error)
	GetTreeStateByID(treestateID int) (TreeStateDto, error)
	GetAllTypeTreeStateByApp(app *app.AppDto, version int) ([]*TreeStateDto, error)
	GetTreeStateByApp(app *app.AppDto, statetype int, version int) ([]*TreeStateDto, error)
	ReleaseTreeStateByApp(app *app.AppDto) error
}

type TreeStateDto struct {
	ID                 int       `json:"id"`
	StateType          int       `json:"state_type"`
	ParentNodeRefID    int       `json:"parent_node_ref_id"`
	ChildrenNodeRefIDs []int     `json:"children_node_ref_ids"`
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
}

func NewTreeStateServiceImpl(logger *zap.SugaredLogger, treestateRepository repository.TreeStateRepository) *TreeStateServiceImpl {
	return &TreeStateServiceImpl{
		logger:              logger,
		treestateRepository: treestateRepository,
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

func (impl *TreeStateServiceImpl) GetTreeStateByID(treestateID int) (TreeStateDto, error) {
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

func (impl *TreeStateServiceImpl) GetAllTypeTreeStateByApp(app *app.AppDto, version int) ([]*TreeStateDto, error) {
	treestates, err := impl.treestateRepository.RetrieveAllTypeTreeStatesByApp(app.ID, version)
	if err != nil {
		return nil, err
	}
	treestatesdto := make([]*TreeStateDto, len(treestates))
	for _, treestate := range treestates {
		treestatesdto = append(treestatesdto, &TreeStateDto{
			ID:                 treestate.ID,
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
		})
	}
	return treestatesdto, nil
}

func (impl *TreeStateServiceImpl) GetTreeStateByApp(app *app.AppDto, statetype int, version int) ([]*TreeStateDto, error) {
	treestates, err := impl.treestateRepository.RetrieveTreeStatesByApp(app.ID, statetype, version)
	if err != nil {
		return nil, err
	}
	treestatesdto := make([]*TreeStateDto, len(treestates))
	for _, treestate := range treestates {
		treestatesdto = append(treestatesdto, &TreeStateDto{
			ID:                 treestate.ID,
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
		})
	}
	return treestatesdto, nil
}

// @todo: should this method be in a transaction?
func (impl *TreeStateServiceImpl) ReleaseTreeStateByApp(app *app.AppDto) error {
	// get edit version K-V state from database
	treestates, err := impl.treestateRepository.RetrieveAllTypeTreeStatesByApp(app.ID, repository.APP_EDIT_VERSION)
	if err != nil {
		return err
	}
	// set version as mainline version
	for serial, _ := range treestates {
		treestates[serial].Version = app.MainlineVersion
	}
	// and put them to the database as duplicate
	for _, treestate := range treestates {
		if err := impl.treestateRepository.Create(treestate); err != nil {
			return err
		}
	}
	return nil
}
