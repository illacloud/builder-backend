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
	"time"

	"github.com/illacloud/builder-backend/internal/util"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const TREE_STATE_SUMMIT_ID = 0
const TREE_STATE_SUMMIT_NAME = "root"

type TreeState struct {
	ID                 int       `json:"id" 							 gorm:"column:id;type:bigserial"`
	StateType          int       `json:"state_type" 					 gorm:"column:state_type;type:bigint"`
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

type TreeStateRepository interface {
	Create(treestate *TreeState) (int, error)
	Delete(treestateID int) error
	Update(treestate *TreeState) error
	RetrieveByID(treestateID int) (*TreeState, error)
	RetrieveTreeStatesByVersion(versionID int) ([]*TreeState, error)
	RetrieveTreeStatesByName(name string) ([]*TreeState, error)
	RetrieveTreeStatesByApp(apprefid int, statetype int, version int) ([]*TreeState, error)
	RetrieveEditVersionByAppAndName(apprefid int, statetype int, name string) (*TreeState, error)
	RetrieveAllTypeTreeStatesByApp(apprefid int, version int) ([]*TreeState, error)
	DeleteAllTypeTreeStatesByApp(apprefid int) error
}

type TreeStateRepositoryImpl struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewTreeState() *TreeState {
	return &TreeState{}
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

func NewTreeStateRepositoryImpl(logger *zap.SugaredLogger, db *gorm.DB) *TreeStateRepositoryImpl {
	return &TreeStateRepositoryImpl{
		logger: logger,
		db:     db,
	}
}

func (impl *TreeStateRepositoryImpl) Create(treestate *TreeState) (int, error) {
	if err := impl.db.Create(treestate).Error; err != nil {
		return 0, err
	}
	return treestate.ID, nil
}

func (impl *TreeStateRepositoryImpl) Delete(treestateID int) error {
	if err := impl.db.Delete(&TreeState{}, treestateID).Error; err != nil {
		return err
	}
	return nil
}

func (impl *TreeStateRepositoryImpl) Update(treestate *TreeState) error {
	if err := impl.db.Model(treestate).UpdateColumns(TreeState{
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
	}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *TreeStateRepositoryImpl) RetrieveByID(treestateID int) (*TreeState, error) {
	treestate := &TreeState{}
	if err := impl.db.First(treestate, treestateID).Error; err != nil {
		return &TreeState{}, err
	}
	return treestate, nil
}

func (impl *TreeStateRepositoryImpl) RetrieveTreeStatesByVersion(version int) ([]*TreeState, error) {
	var treestates []*TreeState
	if err := impl.db.Where("version = ?", version).Find(&treestates).Error; err != nil {
		return nil, err
	}
	return treestates, nil
}

func (impl *TreeStateRepositoryImpl) RetrieveTreeStatesByName(name string) ([]*TreeState, error) {
	var treestates []*TreeState
	if err := impl.db.Where("name = ?", name).Find(&treestates).Error; err != nil {
		return nil, err
	}
	return treestates, nil
}

func (impl *TreeStateRepositoryImpl) RetrieveTreeStatesByApp(apprefid int, statetype int, version int) ([]*TreeState, error) {
	var treestates []*TreeState
	if err := impl.db.Where("app_ref_id = ? AND state_type = ? AND version = ?", apprefid, statetype, version).Find(&treestates).Error; err != nil {
		return nil, err
	}
	return treestates, nil
}

func (impl *TreeStateRepositoryImpl) RetrieveEditVersionByAppAndName(apprefid int, statetype int, name string) (*TreeState, error) {
	var treestate *TreeState
	if err := impl.db.Where("app_ref_id = ? AND state_type = ? AND version = ? AND name = ?", apprefid, statetype, APP_EDIT_VERSION, name).First(&treestate).Error; err != nil {
		return nil, err
	}
	return treestate, nil
}

func (impl *TreeStateRepositoryImpl) RetrieveAllTypeTreeStatesByApp(apprefid int, version int) ([]*TreeState, error) {
	var treestates []*TreeState
	if err := impl.db.Where("app_ref_id = ? AND version = ?", apprefid, version).Find(&treestates).Error; err != nil {
		return nil, err
	}
	return treestates, nil
}

func (impl *TreeStateRepositoryImpl) DeleteAllTypeTreeStatesByApp(apprefid int) error {
	if err := impl.db.Where("app_ref_id = ?", apprefid).Delete(&TreeState{}).Error; err != nil {
		return err
	}
	return nil
}
