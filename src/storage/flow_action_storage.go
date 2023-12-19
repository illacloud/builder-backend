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

package storage

import (
	"github.com/illacloud/builder-backend/src/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type FlowActionStorage struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewFlowActionStorage(logger *zap.SugaredLogger, db *gorm.DB) *FlowActionStorage {
	return &FlowActionStorage{
		logger: logger,
		db:     db,
	}
}

func (impl *FlowActionStorage) Create(action *model.FlowAction) (int, error) {
	if err := impl.db.Create(action).Error; err != nil {
		return 0, err
	}
	return action.ID, nil
}

func (impl *FlowActionStorage) Delete(teamID int, flowActionID int) error {
	if err := impl.db.Where("id = ? AND team_id = ?", flowActionID, teamID).Delete(&model.FlowAction{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *FlowActionStorage) UpdateWholeFlowAction(action *model.FlowAction) error {
	if err := impl.db.Model(action).Where("id = ?", action.ID).UpdateColumns(action).Error; err != nil {
		return err
	}
	return nil
}

func (impl *FlowActionStorage) RetrieveFlowActionByTeamIDAndID(teamID int, flowActionID int) (*model.FlowAction, error) {
	var action *model.FlowAction
	if err := impl.db.Where("team_id = ? AND id = ?", teamID, flowActionID).Find(&action).Error; err != nil {
		return nil, err
	}
	return action, nil
}

func (impl *FlowActionStorage) RetrieveAll(teamID int, workflowID int, version int) ([]*model.FlowAction, error) {
	var actions []*model.FlowAction
	if err := impl.db.Where("team_id = ? AND workflow_id = ? AND version = ?", teamID, workflowID, version).Find(&actions).Error; err != nil {
		return nil, err
	}
	return actions, nil
}

func (impl *FlowActionStorage) RetrieveByType(teamID int, workflowID int, version int, actionType int) ([]*model.FlowAction, error) {
	var actions []*model.FlowAction
	if err := impl.db.Where("team_id = ? AND workflow_id = ? AND version = ? AND type = ? ", teamID, workflowID, version, actionType).Find(&actions).Error; err != nil {
		return nil, err
	}
	return actions, nil
}

func (impl *FlowActionStorage) RetrieveByID(teamID int, flowActionID int) (*model.FlowAction, error) {
	var action *model.FlowAction
	if err := impl.db.Where("id = ? AND team_id = ?", flowActionID, teamID).First(&action).Error; err != nil {
		return &model.FlowAction{}, err
	}
	return action, nil
}

func (impl *FlowActionStorage) RetrieveFlowActionsByTeamIDWorkflowIDAndVersion(teamID int, workflowID int, version int) ([]*model.FlowAction, error) {
	var actions []*model.FlowAction
	if err := impl.db.Where("team_id = ? AND workflow_id = ? AND version = ?", teamID, workflowID, version).Find(&actions).Error; err != nil {
		return nil, err
	}
	return actions, nil
}

func (impl *FlowActionStorage) RetrieveFlowActionsByTeamIDWorkflowIDVersionAndType(teamID int, workflowID int, version int, actionType int) ([]*model.FlowAction, error) {
	var actions []*model.FlowAction
	if err := impl.db.Where("team_id = ? AND workflow_id = ? AND version = ? AND type = ?", teamID, workflowID, version, actionType).Find(&actions).Error; err != nil {
		return nil, err
	}
	return actions, nil
}

func (impl *FlowActionStorage) RetrieveFlowActionByTeamIDFlowActionID(teamID int, flowActionID int) (*model.FlowAction, error) {
	var action *model.FlowAction
	if err := impl.db.Where("team_id = ? AND id = ?", teamID, flowActionID).First(&action).Error; err != nil {
		return nil, err
	}
	return action, nil
}

func (impl *FlowActionStorage) DeleteFlowActionsByWorkflow(teamID int, workflowID int) error {
	if err := impl.db.Where("team_id = ? AND workflow_id = ?", teamID, workflowID).Delete(&model.FlowAction{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *FlowActionStorage) DeleteFlowActionByTeamIDAndFlowActionID(teamID int, flowActionID int) error {
	if err := impl.db.Where("team_id = ? AND id = ?", teamID, flowActionID).Delete(&model.FlowAction{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *FlowActionStorage) CountFlowActionByTeamID(teamID int) (int, error) {
	var count int64
	if err := impl.db.Model(&model.FlowAction{}).Where("team_id = ?", teamID).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func (impl *FlowActionStorage) DeleteAllFlowActionsByTeamIDWorkflowIDAndVersion(teamID int, workflowID int, targetVersion int) error {
	if err := impl.db.Where("team_id = ? AND workflow_id = ? AND version = ?", teamID, workflowID, targetVersion).Delete(&model.FlowAction{}).Error; err != nil {
		return err
	}
	return nil
}
