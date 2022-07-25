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

type ComponentNode struct {
    DisplayName    string                 `json:"displayName"`
    ParentNode     string                 `json:"parentNode"`
    ShowName       string                 `json:"showName"`
    Cerror         bool                   `json:"error"`
    IsDragging     bool                   `json:"isDragging"`
    ChildrenNode   []*ComponentNode       `json:"childrenNode"`
    Ctype          string                 `json:"type"`
    ContainerType  map[string]interface{} `json:"containerType"`
    VerticalResize bool                   `json:"verticalResize"`
    H              int                    `json:"h"`
    W              int                    `json:"w"`
    MinH           int                    `json:"minH"`
    MinW           int                    `json:"minW"`
    X              int                    `json:"x"`
    Y              int                    `json:"y"`
    Z              int                    `json:"z"`
    Props          map[string]interface{} `json:"props"`
    PanelConfig    map[string]interface{} `json:"panelConfig"`
}

func NewComponentNodeFromJSON(cnodebyte []byte) (*ComponentNode, error) {
    cnode := ComponentNode{}
    if err := json.Unmarshal(cnodebyte, &cnode); err != nil {
        return nil, err
    }
    return &cnode, nil
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
    cnode.ParentNode = ""
    cnode.ChildrenNode = nil
    return json.Marshal(cnode)
}

func BuildComponentTree(treeState *TreeState, treeStateMap map[int]*TreeState, parentComponentNode *ComponentNode) (*ComponentNode, error) {
    cnode := &ComponentNode{}
    var err error
    if cnode, err = treeState.ExportContentAsComponentState(); err != nil {
        return nil, err
    }
    for _, id := range treeState.ChildrenNodeRefIDs {
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
