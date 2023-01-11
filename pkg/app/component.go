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

package app

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/illacloud/builder-backend/internal/repository"
)

var initialComponet = "{\"displayName\":\"root\",\"showName\":\"root\",\"error\":false,\"isDragging\":false,\"isResizing\":false,\"type\":\"DOT_PANEL\",\"parentNode\":\"\",\"childrenNode\":[],\"containerType\":\"EDITOR_DOT_PANEL\",\"verticalResize\":true,\"h\":0,\"w\":0,\"minH\":0,\"minW\":0,\"unitW\":0,\"unitH\":0,\"x\":-1,\"y\":-1,\"z\":0,\"props\":{\"currentPageIndex\":0,\"pageSortedKey\":[\"page1\"]}}"

type ComponentNode struct {
	DisplayName    string                 `json:"displayName"`
	ParentNode     string                 `json:"parentNode"`
	ShowName       string                 `json:"showName"`
	Cerror         bool                   `json:"error"`
	IsDragging     bool                   `json:"isDragging"`
	IsResizing     bool                   `json:"isResizing"`
	ChildrenNode   []*ComponentNode       `json:"childrenNode"`
	Ctype          string                 `json:"type"`
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

func newComponentNodeFromJSON(cnodebyte []byte) (*ComponentNode, error) {
	cnode := ComponentNode{}
	if err := json.Unmarshal(cnodebyte, &cnode); err != nil {
		return nil, err
	}
	return &cnode, nil
}

func (cnode *ComponentNode) appendChildrenNode(node *ComponentNode) {
	cnode.ChildrenNode = append(cnode.ChildrenNode, node)
}

func (cnode *ComponentNode) updateParentNode(parentComponentNode *ComponentNode) {
	if parentComponentNode != nil {
		cnode.ParentNode = parentComponentNode.DisplayName
	}
}

func buildComponentTree(treeState *repository.TreeState, treeStateMap map[int]*repository.TreeState, parentComponentNode *ComponentNode) (*ComponentNode, error) {
	cnode := &ComponentNode{}
	var err error
	cnode, err = newComponentNodeFromJSON([]byte(treeState.Content))
	if err != nil {
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
		if subcnode, err = buildComponentTree(subTreeState, treeStateMap, cnode); err != nil {
			return nil, err
		}
		cnode.appendChildrenNode(subcnode)
	}
	cnode.updateParentNode(parentComponentNode)
	return cnode, nil
}
