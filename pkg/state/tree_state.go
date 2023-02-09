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
	"encoding/json"
	"errors"
	"time"

	"github.com/illacloud/builder-backend/internal/repository"
	"github.com/illacloud/builder-backend/pkg/app"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type TreeStateService interface {
	CreateTreeState(treestate TreeStateDto) (TreeStateDto, error)
	DeleteTreeState(treestateId int) error
	UpdateTreeState(treestate *TreeStateDto) (*TreeStateDto, error)
	GetTreeStateByID(treestateID int) (TreeStateDto, error)
	GetAllTypeTreeStateByApp(app *app.AppDto, version int) ([]*TreeStateDto, error)
	GetTreeStateByApp(app *app.AppDto, statetype int, version int) ([]*TreeStateDto, error)
	ReleaseTreeStateByApp(app *app.AppDto) error
	CreateComponentTree(appDto *app.AppDto, parentNodeID int, componentNodeTree *repository.ComponentNode) error
}

type TreeStateDto struct {
	ID                 int       `json:"id"`
	StateType          int       `json:"state_type"`
	ParentNodeRefID    int       `json:"parent_node_ref_id"`
	ChildrenNodeRefIDs []int     `json:"children_node_ref_ids"`
	AppRefID           int       `json:"app_ref_id"`
	Version            int       `json:"version"`
	Name               string    `json:"name"`
	ParentNode         string    `json:"parentNode"`
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

func NewTreeStateDto() *TreeStateDto {
	return &TreeStateDto{}
}

func (tsd *TreeStateDto) ConstructByMap(data interface{}) {

	udata, ok := data.(map[string]interface{})
	if !ok {
		return
	}
	for k, v := range udata {
		switch k {
		case "displayName":
			tsd.Name, _ = v.(string)
		case "parentNode":
			tsd.ParentNode, _ = v.(string)
		}
	}
}

func (tsd *TreeStateDto) ConstructWithDisplayNameForDelete(displayNameInterface interface{}) error {
	dnis, ok := displayNameInterface.(string)
	if !ok {
		err := errors.New("ConstructWithDisplayNameForDelete() can not resolve displayName.")
		return err
	}
	tsd.Name = dnis
	return nil
}

func (tsd *TreeStateDto) ConstructByTreeState(treeState *repository.TreeState) error {
	cids, err := treeState.ExportChildrenNodeRefIDs()
	if err != nil {
		return err
	}
	tsd.ID = treeState.ID
	tsd.StateType = treeState.StateType
	tsd.ParentNodeRefID = treeState.ParentNodeRefID
	tsd.ChildrenNodeRefIDs = cids
	tsd.AppRefID = treeState.AppRefID
	tsd.Version = treeState.Version
	tsd.Name = treeState.Name
	tsd.ParentNode = ""
	tsd.Content = treeState.Content
	tsd.CreatedAt = treeState.CreatedAt
	tsd.CreatedBy = treeState.CreatedBy
	tsd.UpdatedBy = treeState.UpdatedBy
	tsd.UpdatedAt = treeState.UpdatedAt
	return nil
}

func (tsd *TreeStateDto) ConstructWithID(id int) {
	tsd.ID = id
}

func (tsd *TreeStateDto) ConstructWithType(stateType int) {
	tsd.StateType = stateType
}

func (tsd *TreeStateDto) ConstructByApp(app *app.AppDto) {
	tsd.AppRefID = app.ID
}

func (tsd *TreeStateDto) ConstructWithEditVersion() {
	tsd.Version = repository.APP_EDIT_VERSION
}

func (tsd *TreeStateDto) ConstructWithContent(content []byte) {
	tsd.Content = string(content)
}

func (tsd *TreeStateDto) ConstructWithNewStateContent(ntsd *TreeStateDto) {
	tsd.Content = ntsd.Content
	tsd.Name = ntsd.Name
}

func NewTreeStateServiceImpl(logger *zap.SugaredLogger, treestateRepository repository.TreeStateRepository) *TreeStateServiceImpl {
	return &TreeStateServiceImpl{
		logger:              logger,
		treestateRepository: treestateRepository,
	}
}

func (impl *TreeStateServiceImpl) NewTreeStateByComponentState(appDto *app.AppDto, cnode *repository.ComponentNode) (*TreeStateDto, error) {
	var cnodeserilized []byte
	var err error
	if cnodeserilized, err = cnode.SerializationForDatabase(); err != nil {
		return nil, err
	}

	treestatedto := &TreeStateDto{
		StateType: repository.TREE_STATE_TYPE_COMPONENTS,
		AppRefID:  appDto.ID,
		Version:   repository.APP_EDIT_VERSION,
		Name:      cnode.DisplayName,
		Content:   string(cnodeserilized),
	}
	return treestatedto, nil
}

func (impl *TreeStateServiceImpl) CreateTreeState(treestate TreeStateDto) (TreeStateDto, error) {
	// TODO: validate the versionId
	validate := validator.New()
	if err := validate.Struct(treestate); err != nil {
		return TreeStateDto{}, err
	}
	treestate.CreatedAt = time.Now().UTC()
	treestate.UpdatedAt = time.Now().UTC()
	treestateIDsJSON, err := json.Marshal(treestate.ChildrenNodeRefIDs)
	if err != nil {
		return TreeStateDto{}, err
	}
	treeStateForStorage := repository.TreeState{
		StateType:          treestate.StateType,
		ParentNodeRefID:    treestate.ParentNodeRefID,
		ChildrenNodeRefIDs: string(treestateIDsJSON),
		AppRefID:           treestate.AppRefID,
		Version:            treestate.Version,
		Name:               treestate.Name,
		Content:            treestate.Content,
		CreatedAt:          treestate.CreatedAt,
		CreatedBy:          treestate.CreatedBy,
		UpdatedAt:          treestate.UpdatedAt,
		UpdatedBy:          treestate.UpdatedBy,
	}
	if _, err := impl.treestateRepository.Create(&treeStateForStorage); err != nil {
		return TreeStateDto{}, err
	}
	// fill created id

	treestate.ID = treeStateForStorage.ID
	return treestate, nil
}

func (impl *TreeStateServiceImpl) DeleteTreeState(treestateID int) error {
	if err := impl.treestateRepository.Delete(treestateID); err != nil {
		return err
	}
	return nil
}

func (impl *TreeStateServiceImpl) UpdateTreeState(treestate *TreeStateDto) (*TreeStateDto, error) {
	validate := validator.New()
	if err := validate.Struct(treestate); err != nil {
		return nil, err
	}
	treestate.UpdatedAt = time.Now().UTC()
	treestateIDsJSON, err := json.Marshal(treestate.ChildrenNodeRefIDs)
	if err != nil {
		return nil, err
	}

	treeStateRepo := &repository.TreeState{
		ID:                 treestate.ID,
		StateType:          treestate.StateType,
		ParentNodeRefID:    treestate.ParentNodeRefID,
		ChildrenNodeRefIDs: string(treestateIDsJSON),
		AppRefID:           treestate.AppRefID,
		Version:            treestate.Version,
		Name:               treestate.Name,
		Content:            treestate.Content,
		UpdatedAt:          treestate.UpdatedAt,
		UpdatedBy:          treestate.UpdatedBy,
	}

	if err := impl.treestateRepository.Update(treeStateRepo); err != nil {
		return nil, err
	}
	return treestate, nil
}

func (impl *TreeStateServiceImpl) GetTreeStateByID(treestateID int) (TreeStateDto, error) {
	res, err := impl.treestateRepository.RetrieveByID(treestateID)
	if err != nil {
		return TreeStateDto{}, err
	}
	ids, err := res.ExportChildrenNodeRefIDs()
	if err != nil {
		return TreeStateDto{}, err
	}
	resDto := TreeStateDto{
		ID:                 res.ID,
		StateType:          res.StateType,
		ParentNodeRefID:    res.ParentNodeRefID,
		ChildrenNodeRefIDs: ids,
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
		ids, err := treestate.ExportChildrenNodeRefIDs()
		if err != nil {
			return nil, err
		}
		treestatesdto = append(treestatesdto, &TreeStateDto{
			ID:                 treestate.ID,
			StateType:          treestate.StateType,
			ParentNodeRefID:    treestate.ParentNodeRefID,
			ChildrenNodeRefIDs: ids,
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
		ids, err := treestate.ExportChildrenNodeRefIDs()
		if err != nil {
			return nil, err
		}
		treestatesdto = append(treestatesdto, &TreeStateDto{
			ID:                 treestate.ID,
			StateType:          treestate.StateType,
			ParentNodeRefID:    treestate.ParentNodeRefID,
			ChildrenNodeRefIDs: ids,
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
		if _, err := impl.treestateRepository.Create(treestate); err != nil {
			return err
		}
	}
	return nil
}

func (impl *TreeStateServiceImpl) GetTreeStateByName(currentNode *TreeStateDto) (*TreeStateDto, error) {
	// get id by displayName
	var err error
	var inDBTreeState *repository.TreeState
	if inDBTreeState, err = impl.treestateRepository.RetrieveEditVersionByAppAndName(currentNode.AppRefID, currentNode.StateType, currentNode.Name); err != nil {
		// not exists
		return nil, err
	}
	inDBTreeStateDto := NewTreeStateDto()
	inDBTreeStateDto.ConstructByTreeState(inDBTreeState)
	return inDBTreeStateDto, nil
}

// @todo: add tree ref circle checker.
func (impl *TreeStateServiceImpl) MoveTreeStateNode(currentNode *TreeStateDto) error {
	// prepare data
	oldParentTreeState := &repository.TreeState{}
	newParentTreeState := &repository.TreeState{}
	nowTreeState := &repository.TreeState{}
	var err error
	// get nowTreeState by name
	if nowTreeState, err = impl.treestateRepository.RetrieveEditVersionByAppAndName(currentNode.AppRefID, currentNode.StateType, currentNode.Name); err != nil {
		return err
	}

	// get oldParentTreeState by id
	if oldParentTreeState, err = impl.treestateRepository.RetrieveByID(nowTreeState.ParentNodeRefID); err != nil {
		return err
	}

	// get newParentTreeState by name
	switch currentNode.StateType {
	case repository.TREE_STATE_TYPE_COMPONENTS:
		if newParentTreeState, err = impl.treestateRepository.RetrieveEditVersionByAppAndName(currentNode.AppRefID, currentNode.StateType, currentNode.ParentNode); err != nil {
			return err
		}
	default:
		return nil
	}

	// fill into database
	// update nowTreeState
	nowTreeState.ParentNodeRefID = newParentTreeState.ID
	if err := impl.treestateRepository.Update(nowTreeState); err != nil {
		return err
	}

	// add now TreeState id into new parent TreeState.ChildrenNodeRefIDs
	newParentTreeState.AppendChildrenNodeRefIDs(nowTreeState.ID)

	// update newParentTreeState
	if err := impl.treestateRepository.Update(newParentTreeState); err != nil {
		return err
	}

	// remove now TreeState id from old parent TreeState.ChildrenNodeRefIDs
	oldParentTreeState.RemoveChildrenNodeRefIDs(nowTreeState.ID)

	// update oldParentTreeState
	if err := impl.treestateRepository.Update(oldParentTreeState); err != nil {
		return err
	}
	return nil
}

func (impl *TreeStateServiceImpl) DeleteTreeStateNodeRecursive(currentNode *TreeStateDto) error {
	// get nowTreeState by displayName from database
	var err error
	nowTreeState := &repository.TreeState{}
	if nowTreeState, err = impl.treestateRepository.RetrieveEditVersionByAppAndName(currentNode.AppRefID, currentNode.StateType, currentNode.Name); err != nil {
		return err
	}
	// unlink parentNode
	// get parentNode
	parentTreeState := &repository.TreeState{}
	if nowTreeState.ParentNodeRefID != 0 { // parentNode is in database
		if parentTreeState, err = impl.treestateRepository.RetrieveByID(nowTreeState.ParentNodeRefID); err != nil {
			return err
		}
		// update parentNode
		parentTreeState.RemoveChildrenNodeRefIDs(nowTreeState.ID)
		if err = impl.treestateRepository.Update(parentTreeState); err != nil {
			return err
		}
	}

	// get all sub nodes recursive
	targetNodes := []*repository.TreeState{}
	if err = impl.retrieveChildrenNodes(nowTreeState, &targetNodes); err != nil {
		return err
	}
	// do not forget delete now node
	targetNodes = append(targetNodes, nowTreeState)
	// delete them all
	// @todo: replace this Delete to a batch method

	for _, node := range targetNodes {

		if err = impl.treestateRepository.Delete(node.ID); err != nil {
			return err
		}
	}
	return nil
}

func (impl *TreeStateServiceImpl) retrieveChildrenNodes(treeState *repository.TreeState, childrenNodes *[]*repository.TreeState) error {
	// @todo: replace this RetrieveByID to a batch method
	ids, err := treeState.ExportChildrenNodeRefIDs()

	if err != nil {
		return err
	}
	for _, id := range ids {
		var node *repository.TreeState
		if node, err = impl.treestateRepository.RetrieveByID(id); err != nil {
			return err
		}
		if err := impl.retrieveChildrenNodes(node, childrenNodes); err != nil {
			return err
		}

		*childrenNodes = append(*childrenNodes, node)
	}
	return nil
}

func (impl *TreeStateServiceImpl) CreateComponentTree(appDto *app.AppDto, parentNodeID int, componentNodeTree *repository.ComponentNode) error {
	// summit node
	if parentNodeID == 0 {
		parentNodeID = repository.TREE_STATE_SUMMIT_ID
	}

	// convert ComponentNode to TreeState
	currentNode := NewTreeStateDto()
	currentNode.ConstructWithType(repository.TREE_STATE_TYPE_COMPONENTS)
	var err error
	if currentNode, err = impl.NewTreeStateByComponentState(appDto, componentNodeTree); err != nil {
		return err
	}

	// get parentNode
	parentTreeState := repository.NewTreeState()
	isSummitNode := true
	if parentNodeID != 0 || currentNode.ParentNode == repository.TREE_STATE_SUMMIT_NAME { // parentNode is in database
		isSummitNode = false
		if parentTreeState, err = impl.treestateRepository.RetrieveByID(parentNodeID); err != nil {
			return err
		}
	} else if componentNodeTree.ParentNode != "" && componentNodeTree.ParentNode != repository.TREE_STATE_SUMMIT_NAME { // or parentNode is exist
		isSummitNode = false
		if parentTreeState, err = impl.treestateRepository.RetrieveEditVersionByAppAndName(currentNode.AppRefID, currentNode.StateType, componentNodeTree.ParentNode); err != nil {
			return err
		}
	}

	// no parentNode, currentNode is tree summit
	if isSummitNode && currentNode.Name != repository.TREE_STATE_SUMMIT_NAME {

		// get root node
		if parentTreeState, err = impl.treestateRepository.RetrieveEditVersionByAppAndName(currentNode.AppRefID, currentNode.StateType, repository.TREE_STATE_SUMMIT_NAME); err != nil {
			return err
		}
	}
	currentNode.ParentNodeRefID = parentTreeState.ID

	// insert currentNode and get id
	var treeStateDtoInDB TreeStateDto
	if treeStateDtoInDB, err = impl.CreateTreeState(*currentNode); err != nil {
		return err
	}
	currentNode.ID = treeStateDtoInDB.ID

	// fill currentNode id into parentNode.ChildrenNodeRefIDs
	if currentNode.Name != repository.TREE_STATE_SUMMIT_NAME {

		parentTreeState.AppendChildrenNodeRefIDs(currentNode.ID)

		// save parentNode
		if err = impl.treestateRepository.Update(parentTreeState); err != nil {
			return err
		}
	}
	// create currentNode.ChildrenNode

	for _, childrenComponentNode := range componentNodeTree.ChildrenNode {
		if err := impl.CreateComponentTree(appDto, currentNode.ID, childrenComponentNode); err != nil {
			return err
		}
	}
	return nil
}
