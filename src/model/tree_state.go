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
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	util "github.com/illacloud/builder-backend/src/utils/extendslice"
)

const TREE_STATE_SUMMIT_ID = 0
const TREE_STATE_SUMMIT_NAME = "root"

const (
	TREE_STATE_FIELD_DISPLAY_NAME = "displayName"
	TREE_STATE_FIELD_PARENT_NODE  = "parentNode"
)

type TreeState struct {
	ID                 int       `json:"id" 							 gorm:"column:id;type:bigserial"`
	UID                uuid.UUID `json:"uid" 							 gorm:"column:uid;type:uuid;not null"`
	TeamID             int       `json:"teamID" 						 gorm:"column:team_id;type:bigserial"`
	StateType          int       `json:"state_type" 					 gorm:"column:state_type;type:bigint"`
	ParentNode         string    `json:"parentNode" sql:"-" gorm:"-"`
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

func NewTreeStateByApp(app *App, stateType int) *TreeState {
	treeState := &TreeState{
		StateType:          stateType,
		ParentNodeRefID:    0,
		ChildrenNodeRefIDs: "[]",
		TeamID:             app.ExportTeamID(),
		AppRefID:           app.ExportID(),
		Version:            APP_EDIT_VERSION,
		Name:               TREE_STATE_SUMMIT_NAME,
		Content:            "",
		CreatedBy:          app.ExportUpdatedBy(),
		UpdatedBy:          app.ExportUpdatedBy(),
	}
	treeState.InitUID()
	treeState.InitCreatedAt()
	treeState.InitUpdatedAt()
	return treeState
}

// data sample: ------------------------
//
//	                                   |
//		{                              |
//		    "signal": 5,               |
//		    "target": 1,               |
//		    "option": 0,               |
//		    "broadcast": null,         |
//		    "payload": [               |
//		        {                      ↓
//		            "before": { <- data
//		                "displayName": "input1"
//		            },                 ↓
//		            "after": {  <- data
//		                "w": 6,
//		                "h": 5,
//		                "minW": 1,
//		                "minH": 3,
//		                "x": 19,
//		                "y": 22,
//		                "z": 0,
//		                "showName": "input",
//		                "type": "INPUT_WIDGET",
//		                "displayName": "input1",
//		                "containerType": "EDITOR_SCALE_SQUARE",
//		                "parentNode": "bodySection1-bodySectionContainer1",
//		                "childrenNode": [],
//		                "props": {
//		                    "value": "",
//		                    "label": "Label",
//		                    "labelAlign": "left",
//		                    "labelPosition": "left",
//		                    "labelWidth": "{{33}}",
//		                    "colorScheme": "blue",
//		                    "hidden": false,
//		                    "formDataKey": "{{input1.displayName}}",
//		                    "placeholder": "input sth",
//		                    "$dynamicAttrPaths": [
//		                        "labelWidth",
//		                        "formDataKey",
//		                        "showVisibleButton"
//		                    ],
//		                    "type": "input",
//		                    "showVisibleButton": "{{true}}"
//		                },
//		                "version": 0
//		            }
//		        }
//		    ],
//		    "teamID": "ILAfx4p1C7bN",
//		    "uid": "ILAfx4p1C7bN"
//		}
func NewTreeStateByWebsocketMessage(app *App, stateType int, data interface{}) (*TreeState, error) {
	treeState := NewTreeStateByApp(app, stateType)
	udata, ok := data.(map[string]interface{})
	if !ok {
		return nil, errors.New("can not init tree state")
	}
	for k, v := range udata {
		switch k {
		case TREE_STATE_FIELD_DISPLAY_NAME:
			treeState.Name, _ = v.(string)
		case TREE_STATE_FIELD_PARENT_NODE:
			treeState.ParentNode, _ = v.(string)
		}
	}
	marshaledData, marshalError := json.Marshal(data)
	if marshalError != nil {
		treeState.Content = string(marshaledData)
	}
	return treeState, nil
}

// data sample:
//
//	{
//	    "signal": 6,
//	    "target": 1,
//	    "option": 1,
//	    "broadcast": {
//	        "type": "components/updateComponentContainerReducer",
//	        "payload": {
//	            "oldParentNodeDisplayName": "bodySection1-bodySectionContainer1",
//	            "newParentNodeDisplayName": "canvas1",
//	            "updateSlices": [
//	                {
//	                    "displayName": "text1",
//	                    "x": 8,
//	                    "y": 4,
//	                    "w": 20,
//	                    "h": 5
//	                }
//	            ]
//	        }
//	    },
//	    "payload": [
//	        {
//	            "displayName": "text1",
//	            "parentNode": "canvas1",
//	            "childrenNode": null
//	        }
//	    ],
//	    "teamID": "ILAfx4p1C7dL",
//	    "uid": "ILAfx4p1C7dL"
//	}
func NewTreeStateByMoveStateWebsocketMessage(app *App, stateType int, data interface{}) (*TreeState, error) {
	treeState := NewTreeStateByApp(app, stateType)
	udata, ok := data.(map[string]interface{})
	if !ok {
		return nil, errors.New("can not init tree state")
	}
	for k, v := range udata {
		switch k {
		case TREE_STATE_FIELD_DISPLAY_NAME:
			treeState.Name, _ = v.(string)
		case TREE_STATE_FIELD_PARENT_NODE:
			treeState.ParentNode, _ = v.(string)
		}
	}
	marshaledData, marshalError := json.Marshal(data)
	if marshalError != nil {
		treeState.Content = string(marshaledData)
	}
	return treeState, nil
}

// delete component tree, message like:
//
//	{
//	    "signal": 4,
//	    "target": 1,
//	    "option": 1,
//	    "broadcast": {
//	        "type": "components/deleteComponentNodeReducer",
//	        "payload": {
//	            "displayNames": [
//	                "image1"
//	            ],
//	            "source": "manage_delete"
//	        }
//	    },
//	    "payload": [
//	        "image1" <- data
//	    ],
//	    "teamID": "ILAfx4p1C7bN",
//	    "uid": "ILAfx4p1C7bN"
//	}
func NewTreeStateByDeleteComponentsWebsocketMessage(app *App, stateType int, data interface{}) (*TreeState, error) {
	treeState := NewTreeStateByApp(app, stateType)
	udata, ok := data.(string)
	if !ok {
		return nil, errors.New("can not init tree state")
	}
	treeState.Name = udata
	return treeState, nil
}

func NewTreeStateByAppAndComponentState(app *App, stateType int, componentNode *ComponentNode) (*TreeState, error) {
	componentNodeSerilized, errInSerialization := componentNode.SerializationForDatabase()
	if errInSerialization != nil {
		return nil, errInSerialization
	}
	treeState := NewTreeStateByApp(app, stateType)
	treeState.Name = componentNode.DisplayName
	treeState.ParentNode = componentNode.ParentNode
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

func (treeState *TreeState) UpdateByNewTreeState(newTreeState *TreeState) {
	treeState.ParentNode = newTreeState.ParentNode
	treeState.ParentNodeRefID = newTreeState.ParentNodeRefID
	treeState.ChildrenNodeRefIDs = newTreeState.ChildrenNodeRefIDs
	treeState.Name = newTreeState.Name
	treeState.Content = newTreeState.Content
	treeState.UpdatedBy = newTreeState.UpdatedBy
	treeState.InitUpdatedAt()
}

func (treeState *TreeState) ExportChildrenNodeRefIDs() ([]int, error) {
	var ids []int
	if err := json.Unmarshal([]byte(treeState.ChildrenNodeRefIDs), &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

func (treeState *TreeState) ExportName() string {
	return treeState.Name
}

func (treeState *TreeState) ExportStateType() int {
	return treeState.StateType
}

func (treeState *TreeState) SetParentNodeRefID(id int) {
	treeState.ParentNodeRefID = id
}

func (treeState *TreeState) AppendChildrenNodeRefIDs(targetIDs []int) error {
	var ids []int
	json.Unmarshal([]byte(treeState.ChildrenNodeRefIDs), &ids)
	for _, targetID := range targetIDs {
		ids = append(ids, targetID)
	}
	idsjsonb, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	treeState.ChildrenNodeRefIDs = string(idsjsonb)
	return nil
}

func (treeState *TreeState) RemoveChildrenNodeRefIDs(targetIDs []int) error {
	var ids []int
	if err := json.Unmarshal([]byte(treeState.ChildrenNodeRefIDs), &ids); err != nil {
		return err
	}
	for _, targetID := range targetIDs {
		ids = util.DeleteElement(ids, targetID)
	}
	idsjsonb, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	treeState.ChildrenNodeRefIDs = string(idsjsonb)
	return nil
}

func (treeState *TreeState) RemapChildrenNodeRefIDs(idMap map[int]int) {
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

func (treeState *TreeState) UpdateNameAndContent(newTreeState *TreeState) {
	treeState.Name = newTreeState.Name
	treeState.Content = newTreeState.Content
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

// the tree_state content field like:
//
//	{
//	    "h": 110,
//	    "w": 64,
//	    "x": 0,
//	    "y": 11,
//	    "z": 0,
//	    "minH": 3,
//	    "minW": 2,
//	    "type": "CONTAINER_WIDGET",
//	    "error": false,
//	    "props": {
//	        "radius": "4px",
//	        "shadow": "small",
//	        "viewList": [
//	            {
//	                "id": "b58e92d5-22c3-4ffd-a4bc-aab23675387f",
//	                "key": "View 1",
//	                "label": "View 1"
//	            }
//	        ],
//	        "currentKey": "View 1",
//	        "borderColor": "#ffffffff",
//	        "borderWidth": "0px",
//	        "currentIndex": 0,
//	        "dynamicHeight": "fixed",
//	        "backgroundColor": "#f7f7f7ff",
//	        "resizeDirection": "ALL",
//	        "$dynamicAttrPaths": []
//	    },
//	    "unitH": 8,
//	    "unitW": 18.109375,
//	    "showName": "container",
//	    "isDragging": false,
//	    "parentNode": "",
//	    "displayName": "container9",
//	    "panelConfig": null,
//	    "childrenNode": null,
//	    "containerType": "EDITOR_SCALE_SQUARE",
//	    "verticalResize": false
//	}
//
// we need extract type field which suffixed with "_WIDGET" and put them in to slice and return.
func ExtractComponentsNameList(treeStates []*TreeState) []string {
	widgetLT := make(map[string]bool)
	ret := make([]string, 0)
	for _, treeState := range treeStates {
		var content map[string]interface{}
		unmarshalError := json.Unmarshal([]byte(treeState.Content), &content)
		if unmarshalError != nil {
			continue
		}
		typeField, hit := content["type"]
		if !hit {
			continue
		}
		typeFieldAsserted, assertPass := typeField.(string)
		if !assertPass {
			continue
		}
		widgetSuffixPos := strings.Index(typeFieldAsserted, "_WIDGET")
		if widgetSuffixPos < 0 {
			continue
		}
		widgetLT[typeFieldAsserted] = true
	}
	for key, _ := range widgetLT {
		ret = append(ret, key)
	}
	return ret
}
