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
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/illa-family/builder-backend/internal/repository"
	"github.com/illa-family/builder-backend/pkg/app"
	"go.uber.org/zap"
)

type KVStateService interface {
	CreateKVState(kvstate KVStateDto) (KVStateDto, error)
	DeleteKVState(kvstateID int) error
	UpdateKVState(kvstate KVStateDto) (KVStateDto, error)
	GetKVStateByID(kvstateID int) (KVStateDto, error)
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

type KVStateServiceImpl struct {
	logger            *zap.SugaredLogger
	kvstateRepository repository.kvstateRepository
}

func NewKVStateServiceImpl(logger *zap.SugaredLogger,
	kvstateRepository repository.kvstateRepository) *KVStateServiceImpl {
	return &KVStateServiceImpl{
		logger:            logger,
		kvstateRepository: kvstateRepository,
	}
}

func (impl *KVStateServiceImpl) CreateKVState(kvstate KVStateDto) (KVStateDto, error) {
	// TODO: validate the version
	validate := validator.New()
	if err := validate.Struct(kvstate); err != nil {
		return KVStateDto{}, err
	}
	kvstate.CreatedAt = time.Now().UTC()
	kvstate.UpdatedAt = time.Now().UTC()
	if err := impl.kvstateRepository.Create(&repository.KVState{
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

func (impl *KVStateServiceImpl) DeleteKVState(kvstateID int) error {
	if err := impl.kvstateRepository.Delete(kvstateID); err != nil {
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
	if err := impl.kvstateRepository.Update(&repository.KVState{
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
	res, err := impl.kvstateRepository.RetrieveById(kvstateID)
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
	kvstates, err := impl.kvstateRepository.RetrieveAllTypeKVStatesByApp(app.ID, version)
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
	kvstates, err := impl.kvstateRepository.RetrieveKVStatesByApp(app.ID, statetype, version)
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

// @todo: should this method be in a transaction?
func (impl *KVStateServiceImpl) ReleaseKVStateByApp(app *app.AppDto) error {
	// get edit version K-V state from database
	kvstates, err := impl.kvstateRepository.RetrieveAllTypeKVStatesByApp(app.ID, repository.APP_EDIT_VERSION)
	if err != nil {
		return err
	}
	// set version as minaline version
	for serial, _ := range kvstates {
		kvstates[serial].Version = app.MainlineVersion
	}
	// and put them to the database as duplicate
	for _, kvstate := range kvstates {
		if err := impl.kvstateRepository.Create(kvstate); err != nil {
			return err
		}
	}
	return nil
}
