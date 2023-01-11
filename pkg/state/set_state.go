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
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either expsetStates or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package state

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/illacloud/builder-backend/internal/repository"
	"github.com/illacloud/builder-backend/pkg/app"
	"go.uber.org/zap"
)

type SetStateService interface {
	CreateSetState(setState *SetStateDto) (*SetStateDto, error)
	DeleteSetState(setStateID int) error
	DeleteSetStateByValue(setStateDto *SetStateDto) error
	UpdateSetState(setState *SetStateDto) (*SetStateDto, error)
	UpdateSetStateByValue(beforeSetStateDto *SetStateDto, afterSetStateDto *SetStateDto) error
	GetSetStateByID(setStateID int) (SetStateDto, error)
	GetByValue(setStateDto *SetStateDto) (*SetStateDto, error)
}

type SetStateDto struct {
	ID        int       `json:"id"`
	StateType int       `json:"state_type"`
	AppRefID  int       `json:"app_ref_id"`
	Version   int       `json:"version"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy int       `json:"created_by"`
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy int       `json:"updated_by"`
}

func NewSetStateDto() *SetStateDto {
	return &SetStateDto{}
}

type SetStateServiceImpl struct {
	logger             *zap.SugaredLogger
	setStateRepository repository.SetStateRepository
}

func (setsd *SetStateDto) ConstructBySetState(setState *repository.SetState) {
	setsd.ID = setState.ID
	setsd.StateType = setState.StateType
	setsd.AppRefID = setState.AppRefID
	setsd.Version = setState.Version
	setsd.Value = setState.Value
	setsd.CreatedAt = setState.CreatedAt
	setsd.CreatedBy = setState.CreatedBy
	setsd.UpdatedAt = setState.UpdatedAt
	setsd.UpdatedBy = setState.UpdatedBy
}

func (setsd *SetStateDto) ConstructWithDisplayNameForDelete(displayNameInterface interface{}) error {
	dnis, ok := displayNameInterface.(string)
	if !ok {
		err := errors.New("ConstructWithDisplayNameForDelete() can not resolve displayName.")
		return err
	}
	setsd.Value = dnis
	return nil
}

func (setsd *SetStateDto) ConstructWithValueBeforeUpdate(dnsfu *repository.DisplayNameStateForUpdate) {
	setsd.Value = dnsfu.Before
}

func (setsd *SetStateDto) ConstructWithValueAfterUpdate(dnsfu *repository.DisplayNameStateForUpdate) {
	setsd.Value = dnsfu.After
}

func (setsd *SetStateDto) ConstructWithType(stateType int) {
	setsd.StateType = stateType
}

func (setsd *SetStateDto) ConstructByApp(app *app.AppDto) {
	setsd.AppRefID = app.ID
}

func (setsd *SetStateDto) ConstructWithEditVersion() {
	setsd.Version = repository.APP_EDIT_VERSION
}

func (setsd *SetStateDto) ConstructWithValue(value string) {
	setsd.Value = value
}

func NewSetStateServiceImpl(logger *zap.SugaredLogger,
	setStateRepository repository.SetStateRepository) *SetStateServiceImpl {
	return &SetStateServiceImpl{
		logger:             logger,
		setStateRepository: setStateRepository,
	}
}

func (impl *SetStateServiceImpl) CreateSetState(setState *SetStateDto) (*SetStateDto, error) {
	// TODO: validate the version
	validate := validator.New()
	if err := validate.Struct(setState); err != nil {
		return nil, err
	}
	setState.CreatedAt = time.Now().UTC()
	setState.UpdatedAt = time.Now().UTC()
	if err := impl.setStateRepository.Create(&repository.SetState{
		ID:        setState.ID,
		StateType: setState.StateType,
		AppRefID:  setState.AppRefID,
		Version:   setState.Version,
		Value:     setState.Value,
		CreatedAt: setState.CreatedAt,
		CreatedBy: setState.CreatedBy,
		UpdatedAt: setState.UpdatedAt,
		UpdatedBy: setState.UpdatedBy,
	}); err != nil {
		return nil, err
	}
	return setState, nil
}

func (impl *SetStateServiceImpl) DeleteSetState(setStateID int) error {
	if err := impl.setStateRepository.Delete(setStateID); err != nil {
		return err
	}
	return nil
}

func (impl *SetStateServiceImpl) DeleteSetStateByValue(setStateDto *SetStateDto) error {
	setState := &repository.SetState{
		StateType: setStateDto.StateType,
		AppRefID:  setStateDto.AppRefID,
		Version:   setStateDto.Version,
		Value:     setStateDto.Value,
	}
	if err := impl.setStateRepository.DeleteByValue(setState); err != nil {
		return err
	}
	return nil
}

func (impl *SetStateServiceImpl) UpdateSetState(setState *SetStateDto) (*SetStateDto, error) {
	validate := validator.New()
	if err := validate.Struct(setState); err != nil {
		return nil, err
	}
	setState.UpdatedAt = time.Now().UTC()
	if err := impl.setStateRepository.Update(&repository.SetState{
		ID:        setState.ID,
		StateType: setState.StateType,
		AppRefID:  setState.AppRefID,
		Version:   setState.Version,
		Value:     setState.Value,
		CreatedAt: setState.CreatedAt,
		CreatedBy: setState.CreatedBy,
		UpdatedAt: setState.UpdatedAt,
		UpdatedBy: setState.UpdatedBy,
	}); err != nil {
		return nil, err
	}
	return setState, nil
}

func (impl *SetStateServiceImpl) UpdateSetStateByValue(beforeSetStateDto *SetStateDto, afterSetStateDto *SetStateDto) error {
	validate := validator.New()
	if err := validate.Struct(beforeSetStateDto); err != nil {
		return err
	}
	if err := validate.Struct(afterSetStateDto); err != nil {
		return err
	}

	// init model
	afterSetStateDto.UpdatedAt = time.Now().UTC()
	beforeSetState := &repository.SetState{
		StateType: beforeSetStateDto.StateType,
		AppRefID:  beforeSetStateDto.AppRefID,
		Version:   beforeSetStateDto.Version,
		Value:     beforeSetStateDto.Value,
	}
	afterSetState := &repository.SetState{
		StateType: beforeSetStateDto.StateType,
		AppRefID:  beforeSetStateDto.AppRefID,
		Value:     afterSetStateDto.Value,
		UpdatedAt: afterSetStateDto.UpdatedAt,
	}

	if err := impl.setStateRepository.UpdateByValue(beforeSetState, afterSetState); err != nil {
		return err
	}
	return nil
}

func (impl *SetStateServiceImpl) GetSetStateByID(setStateID int) (SetStateDto, error) {
	setState, err := impl.setStateRepository.RetrieveByID(setStateID)
	if err != nil {
		return SetStateDto{}, err
	}
	ret := SetStateDto{
		ID:        setState.ID,
		StateType: setState.StateType,
		AppRefID:  setState.AppRefID,
		Version:   setState.Version,
		Value:     setState.Value,
		CreatedAt: setState.CreatedAt,
		CreatedBy: setState.CreatedBy,
		UpdatedAt: setState.UpdatedAt,
		UpdatedBy: setState.UpdatedBy,
	}
	return ret, nil
}

func (impl *SetStateServiceImpl) GetByValue(setStateDto *SetStateDto) (*SetStateDto, error) {
	setState := &repository.SetState{
		StateType: setStateDto.StateType,
		AppRefID:  setStateDto.AppRefID,
		Version:   setStateDto.Version,
		Value:     setStateDto.Value,
	}

	inDBSetState, err := impl.setStateRepository.RetrieveByValue(setState)
	if err != nil {
		return nil, err
	}
	inDBSetStateDto := NewSetStateDto()
	inDBSetStateDto.ConstructBySetState(inDBSetState)
	return inDBSetStateDto, nil
}
