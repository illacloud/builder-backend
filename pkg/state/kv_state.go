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

package state

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/illacloud/builder-backend/internal/repository"
	"github.com/illacloud/builder-backend/pkg/app"
	"go.uber.org/zap"
)

type KVStateService interface {
	CreateKVState(kvstate KVStateDto) (KVStateDto, error)
	DeleteKVState(kvstateID int) error
	UpdateKVState(kvstate KVStateDto) (KVStateDto, error)
	GetKVStateByID(kvstateID int) (KVStateDto, error)
	GetKVStateByKey(kvStateDto *KVStateDto) (*KVStateDto, error)
	GetAllTypeKVStateByApp(app *app.AppDto, version int) ([]*KVStateDto, error)
	GetKVStateByApp(app *app.AppDto, statetype int, version int) ([]*KVStateDto, error)
	ReleaseKVStateByApp(app *app.AppDto) error
}

type KVStateDto struct {
	ID        int       `json:"id"`
	StateType int       `json:"state_type"`
	AppRefID  int       `json:"app_ref_id"`
	Version   int       `json:"version"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy int       `json:"created_by"`
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy int       `json:"updated_by"`
}

func NewKVStateDto() *KVStateDto {
	return &KVStateDto{}
}

type KVStateServiceImpl struct {
	logger            *zap.SugaredLogger
	kvStateRepository repository.KVStateRepository
}

func (kvsd *KVStateDto) ConstructByMap(data interface{}) error {
	udata, ok := data.(map[string]interface{})
	if !ok {
		err := errors.New("KVStateDto ConstructByMap failed, please check your input.")
		return err
	}
	displayName, mapok := udata["displayName"].(string)
	if !mapok {
		err := errors.New("KVStateDto ConstructByMap failed, can not find displayName field.")
		return err
	}
	// fild
	kvsd.Key = displayName
	jsonbyte, _ := json.Marshal(udata)
	kvsd.Value = string(jsonbyte)
	return nil
}

func (kvsd *KVStateDto) ConstructForDependenciesState(data interface{}) error {
	jsonbyte, _ := json.Marshal(data)
	kvsd.Value = string(jsonbyte)
	return nil
}

func (kvsd *KVStateDto) ConstructByKvState(kvState *repository.KVState) {
	kvsd.ID = kvState.ID
	kvsd.StateType = kvState.StateType
	kvsd.AppRefID = kvState.AppRefID
	kvsd.Version = kvState.Version
	kvsd.Key = kvState.Key
	kvsd.Value = kvState.Value
	kvsd.CreatedAt = kvState.CreatedAt
	kvsd.CreatedBy = kvState.CreatedBy
	kvsd.UpdatedAt = kvState.UpdatedAt
	kvsd.UpdatedBy = kvState.UpdatedBy
}

func (kvsd *KVStateDto) ConstructWithDisplayNameForDelete(displayNameInterface interface{}) error {
	dnis, ok := displayNameInterface.(string)
	if !ok {
		err := errors.New("ConstructWithDisplayNameForDelete() can not resolve displayName.")
		return err
	}
	kvsd.Key = dnis
	return nil
}

func (kvsd *KVStateDto) ConstructWithID(id int) {
	kvsd.ID = id
}

func (kvsd *KVStateDto) ConstructWithType(stateType int) {
	kvsd.StateType = stateType
}

func (kvsd *KVStateDto) ConstructByApp(app *app.AppDto) {
	kvsd.AppRefID = app.ID
}

func (kvsd *KVStateDto) ConstructWithEditVersion() {
	kvsd.Version = repository.APP_EDIT_VERSION
}

func (kvsd *KVStateDto) ConstructWithKey(key string) {
	kvsd.Key = key
}

func (kvsd *KVStateDto) ConstructWithValue(value string) {
	kvsd.Value = value
}

func NewKVStateServiceImpl(logger *zap.SugaredLogger,
	kvStateRepository repository.KVStateRepository) *KVStateServiceImpl {
	return &KVStateServiceImpl{
		logger:            logger,
		kvStateRepository: kvStateRepository,
	}
}

func (impl *KVStateServiceImpl) CreateKVState(kvstate *KVStateDto) (*KVStateDto, error) {
	// TODO: validate the version
	validate := validator.New()
	if err := validate.Struct(kvstate); err != nil {
		return nil, err
	}
	kvstate.CreatedAt = time.Now().UTC()
	kvstate.UpdatedAt = time.Now().UTC()
	if err := impl.kvStateRepository.Create(&repository.KVState{
		ID:        kvstate.ID,
		StateType: kvstate.StateType,
		AppRefID:  kvstate.AppRefID,
		Version:   kvstate.Version,
		Key:       kvstate.Key,
		Value:     kvstate.Value,
		CreatedAt: kvstate.CreatedAt,
		CreatedBy: kvstate.CreatedBy,
		UpdatedAt: kvstate.UpdatedAt,
		UpdatedBy: kvstate.UpdatedBy,
	}); err != nil {
		return nil, err
	}
	return kvstate, nil
}

func (impl *KVStateServiceImpl) DeleteKVState(kvstateID int) error {
	if err := impl.kvStateRepository.Delete(kvstateID); err != nil {
		return err
	}
	return nil
}

func (impl *KVStateServiceImpl) UpdateKVState(kvstate KVStateDto) (KVStateDto, error) {
	validate := validator.New()
	if err := validate.Struct(kvstate); err != nil {
		return KVStateDto{}, err
	}
	kvstate.UpdatedAt = time.Now().UTC()
	if err := impl.kvStateRepository.Update(&repository.KVState{
		ID:        kvstate.ID,
		StateType: kvstate.StateType,
		AppRefID:  kvstate.AppRefID,
		Version:   kvstate.Version,
		Key:       kvstate.Key,
		Value:     kvstate.Value,
		CreatedAt: kvstate.CreatedAt,
		CreatedBy: kvstate.CreatedBy,
		UpdatedAt: kvstate.UpdatedAt,
		UpdatedBy: kvstate.UpdatedBy,
	}); err != nil {
		return KVStateDto{}, err
	}
	return kvstate, nil
}

func (impl *KVStateServiceImpl) GetKVStateByID(kvstateID int) (KVStateDto, error) {
	res, err := impl.kvStateRepository.RetrieveByID(kvstateID)
	if err != nil {
		return KVStateDto{}, err
	}
	resDto := KVStateDto{
		ID:        res.ID,
		StateType: res.StateType,
		AppRefID:  res.AppRefID,
		Version:   res.Version,
		Key:       res.Key,
		Value:     res.Value,
		CreatedAt: res.CreatedAt,
		CreatedBy: res.CreatedBy,
		UpdatedAt: res.UpdatedAt,
		UpdatedBy: res.UpdatedBy,
	}
	return resDto, nil
}

func (impl *KVStateServiceImpl) GetAllTypeKVStateByApp(app *app.AppDto, version int) ([]*KVStateDto, error) {
	kvstates, err := impl.kvStateRepository.RetrieveAllTypeKVStatesByApp(app.ID, version)
	if err != nil {
		return nil, err
	}
	kvstatesdto := make([]*KVStateDto, len(kvstates))
	for _, kvstate := range kvstates {
		kvstatesdto = append(kvstatesdto, &KVStateDto{
			ID:        kvstate.ID,
			StateType: kvstate.StateType,
			AppRefID:  kvstate.AppRefID,
			Version:   kvstate.Version,
			Key:       kvstate.Key,
			Value:     kvstate.Value,
			CreatedAt: kvstate.CreatedAt,
			CreatedBy: kvstate.CreatedBy,
			UpdatedAt: kvstate.UpdatedAt,
			UpdatedBy: kvstate.UpdatedBy,
		})
	}
	return kvstatesdto, nil
}

func (impl *KVStateServiceImpl) GetKVStateByApp(app *app.AppDto, statetype int, version int) ([]*KVStateDto, error) {
	kvstates, err := impl.kvStateRepository.RetrieveKVStatesByApp(app.ID, statetype, version)
	if err != nil {
		return nil, err
	}
	kvstatesdto := make([]*KVStateDto, len(kvstates))
	for _, kvstate := range kvstates {
		kvstatesdto = append(kvstatesdto, &KVStateDto{
			ID:        kvstate.ID,
			StateType: kvstate.StateType,
			AppRefID:  kvstate.AppRefID,
			Version:   kvstate.Version,
			Key:       kvstate.Key,
			Value:     kvstate.Value,
			CreatedAt: kvstate.CreatedAt,
			CreatedBy: kvstate.CreatedBy,
			UpdatedAt: kvstate.UpdatedAt,
			UpdatedBy: kvstate.UpdatedBy,
		})
	}
	return kvstatesdto, nil
}

func (impl *KVStateServiceImpl) GetKVStateByKey(kvStateDto *KVStateDto) (*KVStateDto, error) {
	// get id by key
	var err error
	var inDBKVState *repository.KVState
	if inDBKVState, err = impl.kvStateRepository.RetrieveEditVersionByAppAndKey(kvStateDto.AppRefID, kvStateDto.StateType, kvStateDto.Key); err != nil {
		// not exists
		return nil, err
	}
	inDBKVStateDto := NewKVStateDto()
	inDBKVStateDto.ConstructByKvState(inDBKVState)
	return inDBKVStateDto, nil
}

// @todo: should this method be in a transaction?
func (impl *KVStateServiceImpl) ReleaseKVStateByApp(appDto *app.AppDto) error {
	// get edit version K-V state from database
	kvstates, err := impl.kvStateRepository.RetrieveAllTypeKVStatesByApp(appDto.ID, repository.APP_EDIT_VERSION)
	if err != nil {
		return err
	}
	// set version as mainline version
	for serial, _ := range kvstates {
		kvstates[serial].Version = appDto.MainlineVersion
	}
	// and put them to the database as duplicate
	for _, kvstate := range kvstates {
		if err := impl.kvStateRepository.Create(kvstate); err != nil {
			return err
		}
	}
	return nil
}

func (impl *KVStateServiceImpl) DeleteKVStateByKey(kvStateDto *KVStateDto) error {
	// get kvstate from database by kvstate.Key
	kvstate, err := impl.kvStateRepository.RetrieveEditVersionByAppAndKey(kvStateDto.AppRefID, kvStateDto.StateType, kvStateDto.Key)
	if err != nil {
		return err
	}
	kvstateid := kvstate.ID

	// delete by id
	if err := impl.kvStateRepository.Delete(kvstateid); err != nil {
		return err
	}
	return nil
}

func (impl *KVStateServiceImpl) DeleteAllEditKVStateByStateType(kvStateDto *KVStateDto) error {
	// get kvstate from database by kvstate.Key
	err := impl.kvStateRepository.DeleteAllKVStatesByAppVersionAndType(kvStateDto.AppRefID, repository.APP_EDIT_VERSION, kvStateDto.StateType)
	if err != nil {
		return err
	}
	return nil
}

func (impl *KVStateServiceImpl) UpdateKVStateByID(kvStateDto *KVStateDto) error {
	// update by id
	if _, err := impl.UpdateKVState(*kvStateDto); err != nil {
		return err
	}
	return nil
}

func (impl *KVStateServiceImpl) UpdateKVStateByKey(kvStateDto *KVStateDto) error {
	// get kvstate from database by kvstate.Key
	kvstate, err := impl.kvStateRepository.RetrieveEditVersionByAppAndKey(kvStateDto.AppRefID, kvStateDto.StateType, kvStateDto.Key)
	if err != nil {
		return err
	}
	kvStateDto.ID = kvstate.ID

	// update by id
	if _, err := impl.UpdateKVState(*kvStateDto); err != nil {
		return err
	}
	return nil
}
