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
	"time"

	"github.com/google/uuid"
)

const TREE_STATE_SUMMIT_ID = 0
const TREE_STATE_SUMMIT_NAME = "root"

type TreeState struct {
	ID                 int       `json:"id" 							 gorm:"column:id;type:bigserial"`
	UID                uuid.UUID `json:"uid" 							 gorm:"column:uid;type:uuid;not null"`
	TeamID             int       `json:"teamID" 						 gorm:"column:team_id;type:bigserial"`
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
