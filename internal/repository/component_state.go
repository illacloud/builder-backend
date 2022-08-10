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
	"fmt"
	"reflect"
	"strconv"
)

type ComponentNode struct {
	DisplayName    string                 `json:"displayName"`
	ParentNode     string                 `json:"parentNode"`
	ShowName       string                 `json:"showName"`
	Error          bool                   `json:"error"`
	IsDragging     bool                   `json:"isDragging"`
	ChildrenNode   []*ComponentNode       `json:"childrenNode"`
	Type           string                 `json:"type"`
	ContainerType  string                 `json:"containerType"`
	VerticalResize bool                   `json:"verticalResize"`
	H              float64                `json:"h"`
	W              float64                `json:"w"`
	MinH           float64                `json:"minH"`
	MinW           float64                `json:"minW"`
	UnitW          float64                `json:"unitW"`
	UnitH          float64                `json:"unitH"`
	X              float64                `json:"x"`
	Y              float64                `json:"y"`
	Z              float64                `json:"z"`
	Props          map[string]interface{} `json:"props"`
	PanelConfig    map[string]interface{} `json:"panelConfig"`
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

func ConstructComponentNodeByMap(data interface{}) *ComponentNode {
	var cnode ComponentNode
	var udata map[string]interface{}
	var ok bool
	fmt.Printf("[DUMP] data reflect.TypeOf: %v\n", reflect.TypeOf(data))

	if udata, ok = data.(map[string]interface{}); !ok {
		return nil
	}
	fmt.Printf("[DUMP] udata: %v\n", udata)
	for k, v := range udata {
		switch k {
		case "displayName":
			cnode.DisplayName, _ = v.(string)
		case "parentNode":
			cnode.ParentNode, _ = v.(string)
		case "showName":
			cnode.ShowName, _ = v.(string)
		case "error":
			cnode.Error, _ = v.(bool)
		case "isDragging":
			cnode.IsDragging, _ = v.(bool)
		case "childrenNode":
			childrenNode, _ := v.([]interface{})
			for _, node := range childrenNode {
				cnode.ChildrenNode = append(cnode.ChildrenNode, ConstructComponentNodeByMap(node))
			}
		case "type":
			cnode.Type, _ = v.(string)
		case "containerType":
			cnode.ContainerType, _ = v.(string)
		case "verticalResize":
			cnode.VerticalResize, _ = v.(bool)
		case "h":
			fmt.Printf("[DUMP] data.v reflect.TypeOf: %v\n", reflect.TypeOf(v))
			fmt.Printf("[DUMP] data.v: %v\n", v)

			cnode.H, _ = v.(float64)
		case "w":
			cnode.W, _ = v.(float64)
		case "minH":
			cnode.MinH, _ = v.(float64)
		case "minW":
			cnode.MinW, _ = v.(float64)
		case "unitW":
			cnode.UnitW, _ = v.(float64)
		case "unitH":
			cnode.UnitH, _ = v.(float64)
		case "x":
			cnode.X, _ = v.(float64)
		case "y":
			cnode.Y, _ = v.(float64)
		case "z":
			cnode.Z, _ = v.(float64)
		case "props":
			cnode.Props, _ = v.(map[string]interface{})
		case "panelConfig":
			cnode.PanelConfig, _ = v.(map[string]interface{})
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
