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
	"encoding/json"
	"errors"
	"strconv"
)

const ROOT_COMPONET_NODE_IN_SERIAL = "{\"version\":0,\"displayName\":\"root\",\"parentNode\":\"\",\"showName\":\"root\",\"type\":\"DOT_PANEL\",\"childrenNode\":[],\"containerType\":\"EDITOR_DOT_PANEL\",\"h\":0,\"w\":0,\"minH\":0,\"minW\":0,\"x\":-1,\"y\":-1,\"z\":0,\"props\":{\"currentPageIndex\":0,\"pageSortedKey\":[\"page1\"]}}"

type ComponentNode struct {
	Version       float64                `json:"version"` // default is 0
	DisplayName   string                 `json:"displayName"`
	ParentNode    string                 `json:"parentNode"`
	ShowName      string                 `json:"showName"`
	ChildrenNode  []*ComponentNode       `json:"childrenNode"`
	Type          string                 `json:"type"`
	ContainerType string                 `json:"containerType"`
	H             float64                `json:"h"`
	W             float64                `json:"w"`
	MinH          float64                `json:"minH"`
	MinW          float64                `json:"minW"`
	X             float64                `json:"x"`
	Y             float64                `json:"y"`
	Z             float64                `json:"z"`
	Props         map[string]interface{} `json:"props"`
}

type ComponentStateForUpdate struct {
	Before interface{} `json:"before"`
	After  interface{} `json:"after"`
}

func GetInitialComponent() string {
	return ROOT_COMPONET_NODE_IN_SERIAL
}

func NewComponentNode() *ComponentNode {
	return &ComponentNode{}

}

func NewComponentNodeFromJSON(cnodebyte []byte) (*ComponentNode, error) {
	cnode := ComponentNode{}
	if err := json.Unmarshal(cnodebyte, &cnode); err != nil {
		return nil, err
	}
	return &cnode, nil
}

func ConstructComponentStateForUpdateByPayload(data interface{}) (*ComponentStateForUpdate, error) {
	var udata map[string]interface{}
	var ok bool
	var csfu ComponentStateForUpdate

	if udata, ok = data.(map[string]interface{}); !ok {
		return nil, errors.New("ConstructComponentStateForUpdateByPayload() failed, please check payload syntax.")
	}

	for k, v := range udata {
		switch k {
		case "before":
			csfu.Before = v.(interface{})
		case "after":
			csfu.After = v.(interface{})
		}
	}
	return &csfu, nil
}

func ConstructComponentNodeByMap(data interface{}) *ComponentNode {
	var cnode ComponentNode
	var udata map[string]interface{}
	var ok bool

	if udata, ok = data.(map[string]interface{}); !ok {
		return nil
	}
	for k, v := range udata {
		switch k {
		case "version":
			cnode.Version, _ = v.(float64)
		case "displayName":
			cnode.DisplayName, _ = v.(string)
		case "parentNode":
			cnode.ParentNode, _ = v.(string)
		case "showName":
			cnode.ShowName, _ = v.(string)
		case "childrenNode":
			childrenNode, _ := v.([]interface{})
			for _, node := range childrenNode {
				cnode.ChildrenNode = append(cnode.ChildrenNode, ConstructComponentNodeByMap(node))
			}
		case "type":
			cnode.Type, _ = v.(string)
		case "containerType":
			cnode.ContainerType, _ = v.(string)
		case "h":
			cnode.H, _ = v.(float64)
		case "w":
			cnode.W, _ = v.(float64)
		case "minH":
			cnode.MinH, _ = v.(float64)
		case "minW":
			cnode.MinW, _ = v.(float64)
		case "x":
			cnode.X, _ = v.(float64)
		case "y":
			cnode.Y, _ = v.(float64)
		case "z":
			cnode.Z, _ = v.(float64)
		case "props":
			cnode.Props, _ = v.(map[string]interface{})
		}
	}
	return &cnode
}

func (cnode *ComponentNode) UpdateParentNode(parentComponentNode *ComponentNode) {
	if parentComponentNode != nil {
		cnode.ParentNode = parentComponentNode.DisplayName
	}
}

func (cnode *ComponentNode) AppendChildrenNode(node *ComponentNode) {
	cnode.ChildrenNode = append(cnode.ChildrenNode, node)
}

func (cnode *ComponentNode) Serialization() ([]byte, error) {
	// build all relations for client output serialization
	return json.Marshal(cnode)
}

func (cnode *ComponentNode) SerializationForDatabase() ([]byte, error) {
	// the parentNode and childrenNode relation info are storaged in special column in database.
	// build these relations for client output serialization only.
	tmpParentNode := cnode.ParentNode
	tmpChildrenNode := cnode.ChildrenNode
	cnode.ParentNode = ""
	cnode.ChildrenNode = nil
	jsonbyte, err := json.Marshal(cnode)
	// recover
	cnode.ParentNode = tmpParentNode
	cnode.ChildrenNode = tmpChildrenNode
	if err != nil {
		return nil, err
	}
	return jsonbyte, nil

}

func BuildComponentTree(treeState *TreeState, treeStateMap map[int]*TreeState, parentComponentNode *ComponentNode) (*ComponentNode, error) {
	cnode := &ComponentNode{}
	var err error
	if cnode, err = treeState.ExportContentAsComponentState(); err != nil {
		return nil, err
	}
	var treestateIDs []int
	treestateIDs, err = treeState.ExportChildrenNodeRefIDs()
	if err != nil {
		return nil, err
	}
	for _, id := range treestateIDs {
		subcnode := &ComponentNode{}
		var err error
		// check if children nodes is exists
		subTreeState, ok := treeStateMap[id]
		if !ok {
			return nil, errors.New("TreeState relation has broken, can not find children node id: " + strconv.Itoa(id))
		}
		if subcnode, err = BuildComponentTree(subTreeState, treeStateMap, cnode); err != nil {
			return nil, err
		}
		cnode.AppendChildrenNode(subcnode)
	}
	cnode.UpdateParentNode(parentComponentNode)
	return cnode, nil

}
