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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildComponentTree(t *testing.T) {
	serilizationResult := `{"displayName":"cnode1","parentNode":"","showName":"","error":false,"isDragging":false,"childrenNode":[{"displayName":"cnode2","parentNode":"cnode1","showName":"","error":false,"isDragging":false,"childrenNode":[{"displayName":"cnode4","parentNode":"cnode2","showName":"","error":false,"isDragging":false,"childrenNode":[{"displayName":"cnode5","parentNode":"cnode4","showName":"","error":false,"isDragging":false,"childrenNode":null,"type":"","containerType":"","verticalResize":false,"h":0,"w":0,"minH":0,"minW":0,"x":0,"y":0,"z":0,"props":null,"panelConfig":null}],"type":"","containerType":"","verticalResize":false,"h":0,"w":0,"minH":0,"minW":0,"x":0,"y":0,"z":0,"props":null,"panelConfig":null}],"type":"","containerType":"","verticalResize":false,"h":0,"w":0,"minH":0,"minW":0,"x":0,"y":0,"z":0,"props":null,"panelConfig":null},{"displayName":"cnode3","parentNode":"cnode1","showName":"","error":false,"isDragging":false,"childrenNode":null,"type":"","containerType":"","verticalResize":false,"h":0,"w":0,"minH":0,"minW":0,"x":0,"y":0,"z":0,"props":null,"panelConfig":null}],"type":"","containerType":"","verticalResize":false,"h":0,"w":0,"minH":0,"minW":0,"x":0,"y":0,"z":0,"props":null,"panelConfig":null}`
	// init component node
	cnode1 := ComponentNode{
		DisplayName: "cnode1",
	}
	cnode1str, _ := json.Marshal(cnode1)
	cnode2 := ComponentNode{
		DisplayName: "cnode2",
	}
	cnode2str, _ := json.Marshal(cnode2)
	cnode3 := ComponentNode{
		DisplayName: "cnode3",
	}
	cnode3str, _ := json.Marshal(cnode3)
	cnode4 := ComponentNode{
		DisplayName: "cnode4",
	}
	cnode4str, _ := json.Marshal(cnode4)
	cnode5 := ComponentNode{
		DisplayName: "cnode5",
	}
	cnode5str, _ := json.Marshal(cnode5)

	// init children node ref ids
	cnrids1 := []int{2, 3}
	cnrids1str, _ := json.Marshal(cnrids1)

	cnrids2 := []int{4}
	cnrids2str, _ := json.Marshal(cnrids2)

	cnrids3 := []int{}
	cnrids3str, _ := json.Marshal(cnrids3)

	cnrids4 := []int{5}
	cnrids4str, _ := json.Marshal(cnrids4)

	cnrids5 := []int{}
	cnrids5str, _ := json.Marshal(cnrids5)

	// init tree state data
	treeState1 := TreeState{
		ID:                 1,
		StateType:          TREE_STATE_TYPE_COMPONENTS,
		ParentNodeRefID:    0,
		ChildrenNodeRefIDs: string(cnrids1str),
		Content:            string(cnode1str),
	}
	treeState2 := TreeState{
		ID:                 2,
		StateType:          TREE_STATE_TYPE_COMPONENTS,
		ParentNodeRefID:    0,
		ChildrenNodeRefIDs: string(cnrids2str),
		Content:            string(cnode2str),
	}
	treeState3 := TreeState{
		ID:                 3,
		StateType:          TREE_STATE_TYPE_COMPONENTS,
		ParentNodeRefID:    0,
		ChildrenNodeRefIDs: string(cnrids3str),
		Content:            string(cnode3str),
	}
	treeState4 := TreeState{
		ID:                 4,
		StateType:          TREE_STATE_TYPE_COMPONENTS,
		ParentNodeRefID:    0,
		ChildrenNodeRefIDs: string(cnrids4str),
		Content:            string(cnode4str),
	}
	treeState5 := TreeState{
		ID:                 5,
		StateType:          TREE_STATE_TYPE_COMPONENTS,
		ParentNodeRefID:    0,
		ChildrenNodeRefIDs: string(cnrids5str),
		Content:            string(cnode5str),
	}

	treeStateMap := map[int]*TreeState{1: &treeState1, 2: &treeState2, 3: &treeState3, 4: &treeState4, 5: &treeState5}
	cnodefin := &ComponentNode{}
	var err error
	cnodefin, err = BuildComponentTree(&treeState1, treeStateMap, nil)
	assert.Nil(t, err)
	// export
	var b []byte
	b, err = cnodefin.Serialization()
	assert.Nil(t, err)
	assert.Equal(t, serilizationResult, string(b), "the serlization result should be equal")
}

func TestNewComponentNodeFromJSON(t *testing.T) {
	serilizationData := `{"displayName":"cnode1","parentNode":"","showName":"","error":false,"isDragging":false,"childrenNode":[{"displayName":"cnode2","parentNode":"cnode1","showName":"","error":false,"isDragging":false,"childrenNode":[{"displayName":"cnode4","parentNode":"cnode2","showName":"","error":false,"isDragging":false,"childrenNode":[{"displayName":"cnode5","parentNode":"cnode4","showName":"","error":false,"isDragging":false,"childrenNode":null,"type":"","containerType":null,"verticalResize":false,"h":0,"w":0,"minH":0,"minW":0,"x":0,"y":0,"z":0,"props":null,"panelConfig":null}],"type":"","containerType":null,"verticalResize":false,"h":0,"w":0,"minH":0,"minW":0,"x":0,"y":0,"z":0,"props":null,"panelConfig":null}],"type":"","containerType":null,"verticalResize":false,"h":0,"w":0,"minH":0,"minW":0,"x":0,"y":0,"z":0,"props":null,"panelConfig":null},{"displayName":"cnode3","parentNode":"cnode1","showName":"","error":false,"isDragging":false,"childrenNode":null,"type":"","containerType":null,"verticalResize":false,"h":0,"w":0,"minH":0,"minW":0,"x":0,"y":0,"z":0,"props":null,"panelConfig":null}],"type":"","containerType":null,"verticalResize":false,"h":0,"w":0,"minH":0,"minW":0,"x":0,"y":0,"z":0,"props":null,"panelConfig":null}`
	_, err := NewComponentNodeFromJSON([]byte(serilizationData))
	assert.Nil(t, err)
}
