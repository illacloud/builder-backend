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

package action

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/illacloud/builder-backend/internal/repository"

	"go.uber.org/zap"
)

var type_array = [17]string{"transformer", "restapi", "graphql", "redis", "mysql", "mariadb", "postgresql", "mongodb",
	"tidb", "elasticsearch", "s3", "smtp", "supabasedb", "firebase", "clickhouse", "mssql", "huggingface"}
var type_map = map[string]int{
	"transformer":   0,
	"restapi":       1,
	"graphql":       2,
	"redis":         3,
	"mysql":         4,
	"mariadb":       5,
	"postgresql":    6,
	"mongodb":       7,
	"tidb":          8,
	"elasticsearch": 9,
	"s3":            10,
	"smtp":          11,
	"supabasedb":    12,
	"firebase":      13,
	"clickhouse":    14,
	"mssql":         15,
	"huggingface":   16,
}

type ActionService interface {
	CreateAction(action ActionDto) (ActionDto, error)
	DeleteAction(teamID int, id int) error
	UpdateAction(action ActionDto) (ActionDto, error)
	GetAction(teamID int, id int) (ActionDto, error)
	FindActionsByAppVersion(teamID int, app, version int) ([]ActionDto, error)
	RunAction(teamID int, action ActionDto) (interface{}, error)
	ValidateActionOptions(actionType string, options map[string]interface{}) error
}

type ActionDto struct {
	ID          int                    `json:"actionId"`
	UID         uuid.UUID              `json:"uid"`
	TeamID      int                    `json:"teamID"`
	App         int                    `json:"-"`
	Version     int                    `json:"-"`
	Resource    int                    `json:"resourceId,omitempty"`
	DisplayName string                 `json:"displayName" validate:"required"`
	Type        string                 `json:"actionType" validate:"oneof=transformer restapi graphql redis mysql mariadb postgresql mongodb tidb elasticsearch s3 smtp supabasedb firebase clickhouse mssql huggingface"`
	Template    map[string]interface{} `json:"content" validate:"required"`
	Transformer map[string]interface{} `json:"transformer" validate:"required"`
	TriggerMode string                 `json:"triggerMode" validate:"oneof=manually automate"`
	CreatedAt   time.Time              `json:"createdAt,omitempty"`
	CreatedBy   int                    `json:"createdBy,omitempty"`
	UpdatedAt   time.Time              `json:"updatedAt,omitempty"`
	UpdatedBy   int                    `json:"updatedBy,omitempty"`
}

func (a *ActionDto) InitUID() {
	a.UID = uuid.New()
}

func (a *ActionDto) SetTeamID(teamID int) {
	a.TeamID = teamID
}

type ActionServiceImpl struct {
	logger             *zap.SugaredLogger
	appRepository      repository.AppRepository
	actionRepository   repository.ActionRepository
	resourceRepository repository.ResourceRepository
}

func NewActionServiceImpl(logger *zap.SugaredLogger, appRepository repository.AppRepository,
	actionRepository repository.ActionRepository, resourceRepository repository.ResourceRepository) *ActionServiceImpl {
	return &ActionServiceImpl{
		logger:             logger,
		appRepository:      appRepository,
		actionRepository:   actionRepository,
		resourceRepository: resourceRepository,
	}
}

func (impl *ActionServiceImpl) CreateAction(action ActionDto) (ActionDto, error) {
	// validate app
	if appDto, err := impl.appRepository.RetrieveAppByID(action.TeamID, action.App); err != nil || appDto.ID != action.App {
		return ActionDto{}, errors.New("app not found")
	}
	// validate resource
	if rscDto, err := impl.resourceRepository.RetrieveByID(action.TeamID, action.Resource); (err != nil || rscDto.ID != action.Resource) && action.Type != type_array[0] {
		return ActionDto{}, errors.New("resource not found")
	}

	id, err := impl.actionRepository.Create(&repository.Action{
		ID:          action.ID,
		UID:         action.UID,
		TeamID:      action.TeamID,
		App:         action.App,
		Version:     action.Version,
		Resource:    action.Resource,
		Name:        action.DisplayName,
		Type:        type_map[action.Type],
		TriggerMode: action.TriggerMode,
		Transformer: action.Transformer,
		Template:    action.Template,
		CreatedAt:   action.CreatedAt,
		CreatedBy:   action.CreatedBy,
		UpdatedAt:   action.UpdatedAt,
		UpdatedBy:   action.UpdatedBy,
	})
	if err != nil {
		return ActionDto{}, err
	}
	action.ID = id

	// update app `updatedAt` field
	_ = impl.appRepository.UpdateUpdatedAt(&repository.App{
		ID:        action.App,
		UpdatedAt: time.Now().UTC(),
		UpdatedBy: action.UpdatedBy,
	})

	return action, nil
}

func (impl *ActionServiceImpl) DeleteAction(teamID int, actionID int) error {
	action, _ := impl.actionRepository.RetrieveByID(teamID, actionID)

	if err := impl.actionRepository.Delete(teamID, actionID); err != nil {
		return err
	}

	// update app `updatedAt` field
	_ = impl.appRepository.UpdateUpdatedAt(&repository.App{
		ID:        action.App,
		UpdatedAt: time.Now().UTC(),
		UpdatedBy: action.UpdatedBy,
	})

	return nil
}

func (impl *ActionServiceImpl) UpdateAction(action ActionDto) (ActionDto, error) {
	// validate app
	if appDto, err := impl.appRepository.RetrieveAppByID(action.TeamID, action.App); err != nil || appDto.ID != action.App {
		return ActionDto{}, errors.New("app not found")
	}
	// validate resource
	if rscDto, err := impl.resourceRepository.RetrieveByID(action.TeamID, action.Resource); (err != nil || rscDto.ID != action.Resource) && action.Type != type_array[0] {
		return ActionDto{}, errors.New("resource not found")
	}

	if err := impl.actionRepository.Update(&repository.Action{
		ID:          action.ID,
		UID:         action.UID,
		TeamID:      action.TeamID,
		Resource:    action.Resource,
		Name:        action.DisplayName,
		Type:        type_map[action.Type],
		TriggerMode: action.TriggerMode,
		Transformer: action.Transformer,
		Template:    action.Template,
		UpdatedAt:   action.UpdatedAt,
		UpdatedBy:   action.UpdatedBy,
	}); err != nil {
		return ActionDto{}, err
	}

	// update app `updatedAt` field
	_ = impl.appRepository.UpdateUpdatedAt(&repository.App{
		ID:        action.App,
		UpdatedAt: time.Now().UTC(),
		UpdatedBy: action.UpdatedBy,
	})

	return action, nil
}

func (impl *ActionServiceImpl) GetAction(teamID int, actionID int) (ActionDto, error) {
	res, err := impl.actionRepository.RetrieveByID(teamID, actionID)
	if err != nil {
		return ActionDto{}, err
	}
	resDto := ActionDto{
		ID:          res.ID,
		UID:         res.UID,
		TeamID:      res.TeamID,
		Resource:    res.Resource,
		DisplayName: res.Name,
		Type:        type_array[res.Type],
		TriggerMode: res.TriggerMode,
		Transformer: res.Transformer,
		Template:    res.Template,
		CreatedBy:   res.CreatedBy,
		CreatedAt:   res.CreatedAt,
		UpdatedBy:   res.UpdatedBy,
		UpdatedAt:   res.UpdatedAt,
	}
	return resDto, nil
}

func (impl *ActionServiceImpl) FindActionsByAppVersion(teamID int, appID int, version int) ([]ActionDto, error) {
	res, err := impl.actionRepository.RetrieveActionsByAppVersion(teamID, appID, version)
	if err != nil {
		return nil, err
	}

	resDtoSlice := make([]ActionDto, 0, len(res))
	for _, value := range res {
		resDtoSlice = append(resDtoSlice, ActionDto{
			ID:          value.ID,
			UID:         value.UID,
			TeamID:      value.TeamID,
			Resource:    value.Resource,
			DisplayName: value.Name,
			Type:        type_array[value.Type],
			TriggerMode: value.TriggerMode,
			Transformer: value.Transformer,
			Template:    value.Template,
			CreatedBy:   value.CreatedBy,
			CreatedAt:   value.CreatedAt,
			UpdatedBy:   value.UpdatedBy,
			UpdatedAt:   value.UpdatedAt,
		})
	}
	return resDtoSlice, nil
}

func (impl *ActionServiceImpl) RunAction(teamID int, action ActionDto) (interface{}, error) {
	if action.Resource == 0 {
		return nil, errors.New("no resource")
	}
	rsc, err := impl.resourceRepository.RetrieveByID(teamID, action.Resource)
	if rsc.ID == 0 {
		return nil, errors.New("resource not found")
	}
	if err != nil {
		return nil, err
	}
	actionFactory := Factory{Type: action.Type}
	actionAssemblyLine := actionFactory.Build()
	if actionAssemblyLine == nil {
		return nil, errors.New("invalid ActionType:: unsupported type")
	}
	if _, err := actionAssemblyLine.ValidateResourceOptions(rsc.Options); err != nil {
		return nil, errors.New("invalid resource content")
	}
	if _, err := actionAssemblyLine.ValidateActionOptions(action.Template); err != nil {
		return nil, errors.New("invalid action content")
	}
	res, err := actionAssemblyLine.Run(rsc.Options, action.Template)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (impl *ActionServiceImpl) ValidateActionOptions(actionType string, options map[string]interface{}) error {
	if actionType == TRANSFORMER_ACTION {
		return nil
	}
	actionFactory := Factory{Type: actionType}
	actionAssemblyLine := actionFactory.Build()
	if actionAssemblyLine == nil {
		return errors.New("invalid ActionType:: unsupported type")
	}
	if _, err := actionAssemblyLine.ValidateActionOptions(options); err != nil {
		return errors.New("invalid action content")
	}
	return nil
}
