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
	util "github.com/illacloud/builder-backend/src/utils/extendslice"
)

const TREE_STATE_SUMMIT_ID = 0
const TREE_STATE_SUMMIT_NAME = "root"

type TreeState struct {
	ID                 int       `json:"id" 							 gorm:"column:id;type:bigserial"`
	UID                uuid.UUID `json:"uid" 							 gorm:"column:uid;type:uuid;not null"`
	TeamID             int       `json:"teamID" 						 gorm:"column:team_id;type:bigserial"`
	StateType          int       `json:"state_type" 					 gorm:"column:state_type;type:bigint"`
	ParentNode         string    `json:"parentNode" 					 gorm"-"`
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

func NewTreeStateByTeamIDAndStateType(teamID int, stateType int) *TreeState {
	treeState := &TreeState{
		TeamID:    teamID,
		StateType: stateType,
	}
	treeState.InitUID()
	treeState.InitCreatedAt()
	treeState.InitUpdatedAt()
	return treeState
}

func NewTreeStateByApp(app *App) *TreeState {
	treeState := &TreeState{
		StateType:          TREE_STATE_TYPE_COMPONENTS,
		ParentNodeRefID:    0,
		ChildrenNodeRefIDs: "[]",
		TeamID:             app.ExportTeamID(),
		AppRefID:           app.ExportID(),
		Version:            APP_EDIT_VERSION,
		Name:               TREE_STATE_SUMMIT_NAME,
		Content:            "",
		CreatedBy:          app.ExportCreatedBy(),
		UpdatedBy:          app.ExportUpdatedBy(),
	}
	treeState.InitUID()
	treeState.InitCreatedAt()
	treeState.InitUpdatedAt()
	return treeState
}

func NewTreeStateByAppAndComponentState(app *App, componentNode *ComponentNode) (*TreeState, error) {
	componentNodeSerilized, errInSerialization := componentNode.SerializationForDatabase()
	if errInSerialization != nil {
		return nil, errInSerialization
	}
	treeState := NewTreeStateByApp(app)
	treeState.Name = componentNode.DisplayName
	treeState.Content = componentNodeSerilized
	return treeState, nil
}

func (treeState *TreeState) CleanID() {
	treeState.ID = 0
}

func (treeState *TreeState) InitForFork(teamID int, appID int, version int, userID int) {
	treeState.TeamID = teamID
	treeState.AppRefID = appID
	treeState.Version = version
	treeState.CreatedBy = userID
	treeState.UpdatedBy = userID
	treeState.CleanID()
	treeState.InitUID()
	treeState.InitCreatedAt()
	treeState.InitUpdatedAt()
}

func (treeState *TreeState) InitUID() {
	treeState.UID = uuid.New()
}

func (treeState *TreeState) InitCreatedAt() {
	treeState.CreatedAt = time.Now().UTC()
}

func (treeState *TreeState) InitUpdatedAt() {
	treeState.UpdatedAt = time.Now().UTC()
}

func (treeState *TreeState) SetTeamID(teamID int) {
	treeState.TeamID = teamID
}

func (treeState *TreeState) AppendNewVersion(newVersion int) {
	treeState.CleanID()
	treeState.InitUID()
	treeState.Version = newVersion
}

func (treeState *TreeState) ExportID() int {
	return treeState.ID
}

func (treeState *TreeState) ExportContentAsComponentState() (*ComponentNode, error) {
	cnode, err := NewComponentNodeFromJSON([]byte(treeState.Content))
	if err != nil {
		return nil, err
	}
	return cnode, nil
}

func (treeState *TreeState) ExportChildrenNodeRefIDs() ([]int, error) {
	var ids []int
	if err := json.Unmarshal([]byte(treeState.ChildrenNodeRefIDs), &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

func (treeState *TreeState) SetParentNodeRefID(id int) {
	treeState.ParentNodeRefID = id
}

func (treeState *TreeState) AppendChildrenNodeRefIDs(id int) error {
	var ids []int
	if err := json.Unmarshal([]byte(treeState.ChildrenNodeRefIDs), &ids); err != nil {
		return err
	}
	ids = append(ids, id)
	idsjsonb, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	treeState.ChildrenNodeRefIDs = string(idsjsonb)
	return nil
}

func (treeState *TreeState) RemoveChildrenNodeRefIDs(id int) error {
	var ids []int
	if err := json.Unmarshal([]byte(treeState.ChildrenNodeRefIDs), &ids); err != nil {
		return err
	}
	ids = util.DeleteElement(ids, id)
	idsjsonb, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	treeState.ChildrenNodeRefIDs = string(idsjsonb)
	return nil
}

// reset parent id by idmap[oldID]newID
func (treeState *TreeState) ResetParentNodeRefIDByMap(idMap map[int]int) {
	treeState.ParentNodeRefID = idMap[treeState.ParentNodeRefID]
}

// reset children ids by idmap[oldID]newID
func (treeState *TreeState) ResetChildrenNodeRefIDsByMap(idMap map[int]int) {
	// convert string to []int
	var oldIDs []int
	if err := json.Unmarshal([]byte(treeState.ChildrenNodeRefIDs), &oldIDs); err != nil {
		return
	}

	// map old id to new id
	newIDs := make([]int, 0, len(oldIDs))
	for _, oldID := range oldIDs {
		newIDs = append(newIDs, idMap[oldID])
	}

	// convert []int to string
	idsjsonb, err := json.Marshal(newIDs)
	if err != nil {
		return
	}

	// set new
	treeState.ChildrenNodeRefIDs = string(idsjsonb)
}

func BuildTreeStateLookupTable(treeStates []*TreeState) map[int]*TreeState {
	tempMap := make(map[int]*TreeState, len(treeStates))
	for _, component := range treeStates {
		tempMap[component.ID] = component
	}
	return tempMap
}

func PickUpTreeStatesRootNode(treeStates []*TreeState) *TreeState {
	root := &TreeState{}
	for _, component := range treeStates {
		if component.Name == TREE_STATE_SUMMIT_NAME {
			root = component
			break
		}
	}
	return root
}
