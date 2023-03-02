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

package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const TREE_STATE_SUMMIT_ID = 0        // the id placeholder for tree state summit node.
const TREE_STATE_SUMMIT_NAME = "root" // the tree state summit node name.

type TreeState struct {
	ID                 int       `json:"id" 							 gorm:"column:id;type:bigserial"`
	UID                uuid.UUID `json:"uid" 							 gorm:"column:uid;type:uuid;not null"`
	TeamID             int       `json:"teamID" 						 gorm:"column:team_id;type:bigserial"`
	StateType          int       `json:"state_type" 					 gorm:"column:state_type;type:bigint"`
	ParentNode         string    `json:"-" 								 gorm:"-" sql:"-"`
	ParentNodeRefID    int       `json:"parent_node_ref_id" 			 gorm:"column:parent_node_ref_id;type:bigint"`
	ChildrenNodeRefIDs string    `json:"children_node_ref_ids" 		     gorm:"column:children_node_ref_ids;type:jsonb"`
	AppRefID           int       `json:"app_ref_id" 					 gorm:"column:app_ref_id;type:bigint"`
	Version            int       `json:"version" 					     gorm:"column:version;type:bigint"`
	Name               string    `json:"name" 						     gorm:"column:name;type:text"`
	Content            string    `json:"content"    					 gorm:"column:content;type:jsonb"`
	CreatedAt          time.Time `json:"created_at" 					 gorm:"column:created_at;type:timestamp"`
	CreatedBy          int       `json:"created_by" 					 gorm:"column:created_by;type:bigint"`
	UpdatedAt          time.Time `json:"updated_at" 					 gorm:"column:updated_at;type:timestamp"`
	UpdatedBy          int       `json:"updated_by" 					 gorm:"column:updated_by;type:bigint"`
}

func NewTreeState() *TreeState {
	return &TreeState{}
}

func NewTeeeStateByAppAndComponentNode(app *App, cnode *ComponentNode) *TreeState {
	treeState := &TreeState{
		TeamID:    app.ExportTeamID(),
		StateType: TREE_STATE_TYPE_COMPONENTS,
		AppRefID:  app.ExportID(),
		Version:   APP_EDIT_VERSION,
		Name:      cnode.ExportName(),
		Content:   cnode.SerializationForDatabase(),
	}
	treeState.InitUID()
	return treeState
}

func (ts *TreeState) InitUID() {
	ts.UID = uuid.New()
}

func (ts *TreeState) SetTeamID(teamID int) {
	ts.TeamID = teamID
}

func (ts *TreeState) SetStateType(stateType int) {
	ts.StateType = stateType
}

func (treeState *TreeState) AppendChildrenNodeRefIDs(id int) error {
	var ids []int
	json.Unmarshal([]byte(treeState.ChildrenNodeRefIDs), &ids)
	ids = append(ids, id)
	idsjsonb, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	treeState.ChildrenNodeRefIDs = string(idsjsonb)
	return nil
}

// build the component tree and insert each node into tree_state storage.
func CreateComponentTree(app *App, parentNodeID int, componentNodeTree *ComponentNode) error {
	// summit node
	if parentNodeID == 0 {
		parentNodeID = TREE_STATE_SUMMIT_ID
	}

	// convert ComponentNode to TreeState
	currentNode := NewTeeeStateByAppAndComponentNode(app, componentNodeTree)

	// get parentNode
	parentTreeState := NewTreeState()
	isSummitNode := true
	if parentNodeID != 0 || currentNode.ParentNode == TREE_STATE_SUMMIT_NAME { // parentNode is in database
		isSummitNode = false
		var errInRetrieveParentTree error
		parentTreeState, errInRetrieveParentTree = impl.Storage.TreeStateStorage.RetrieveByID(app.ExportTeamID(), parentNodeID)
		if errInRetrieveParentTree != nil {
			return errInRetrieveParentTree
		}
	} else if componentNodeTree.ParentNode != "" && componentNodeTree.ParentNode != TREE_STATE_SUMMIT_NAME { // or parentNode is exist in context
		isSummitNode = false
		var errInRetrieveEditVersionParentTree error
		parentTreeState, errInRetrieveEditVersionParentTree = impl.Storage.TreeStateStorage.RetrieveEditVersionByAppAndName(app.ExportTeamID(), currentNode.AppRefID, currentNode.StateType, componentNodeTree.ParentNode)
		if errInRetrieveEditVersionParentTree != nil {
			return errInRetrieveEditVersionParentTree
		}
	}

	// no parentNode, currentNode is tree summit
	if isSummitNode && currentNode.Name != TREE_STATE_SUMMIT_NAME {
		// get root node
		var errInGetRootNode error
		parentTreeState, errInGetRootNode = impl.Storage.TreeStateStorage.RetrieveEditVersionByAppAndName(app.ExportTeamID(), currentNode.AppRefID, currentNode.StateType, TREE_STATE_SUMMIT_NAME)
		if errInGetRootNode != nil {
			return errInGetRootNode
		}
	}
	currentNode.ParentNodeRefID = parentTreeState.ID

	// insert currentNode and get id
	if errInCreateTreeState := impl.Storage.TreeStateStorage.CreateTreeState(currentNode); errInCreateTreeState != nil {
		return errInCreateTreeState
	}

	// insert currentNode id into parentNode.ChildrenNodeRefIDs
	if currentNode.Name != TREE_STATE_SUMMIT_NAME {
		// insert
		parentTreeState.AppendChildrenNodeRefIDs(currentNode.ID)
		// save parentNode
		errInUpdateTreeState := impl.Storage.TreeStateStorage.Update(parentTreeState)
		if errInUpdateTreeState != nil {
			return errInUpdateTreeState
		}
	}

	// create currentNode.ChildrenNode
	for _, childrenComponentNode := range componentNodeTree.ChildrenNode {
		if err := impl.CreateComponentTree(app, currentNode.ID, childrenComponentNode); err != nil {
			return err
		}
	}
	return nil
}

func NewTreeStateByComponentState(appDto *app.AppDto, cnode *repository.ComponentNode) *TreeStateDto {
	var cnodeserilized []byte
	var err error
	if cnodeserilized, err = cnode.SerializationForDatabase(); err != nil {
		return nil, err
	}

	treestatedto := &TreeStateDto{
		UID:       uuid.New(),
		TeamID:    appDto.TeamID,
		StateType: repository.TREE_STATE_TYPE_COMPONENTS,
		AppRefID:  appDto.ID,
		Version:   repository.APP_EDIT_VERSION,
		Name:      cnode.DisplayName,
		Content:   string(cnodeserilized),
	}
	return treestatedto, nil
}
