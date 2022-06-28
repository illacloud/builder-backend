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

package resource

import (
	"database/sql"
	"errors"
	"time"

	"github.com/illa-family/builder-backend/internal/repository"
	"github.com/illa-family/builder-backend/pkg/connector"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ResourceService interface {
	CreateResource(resource ResourceDto) (ResourceDto, error)
	DeleteResource(resourceId uuid.UUID) error
	UpdateResource(resource ResourceDto) (ResourceDto, error)
	GetResource(resourceId uuid.UUID) (ResourceDto, error)
	FindAllResources() ([]ResourceDto, error)
	OpenConnection(resource ResourceDto) (*sql.DB, error)
}

type ResourceDto struct {
	ResourceId   uuid.UUID              `json:"resourceId"`
	ResourceName string                 `json:"resourceName,omitempty" validate:"required"`
	ResourceType string                 `json:"resourceType,omitempty" validate:"required"`
	Options      map[string]interface{} `json:"options,omitempty" validate:"required"`
	CreatedBy    uuid.UUID              `json:"createdBy,omitempty"`
	CreatedAt    time.Time              `json:"createdAt,omitempty"`
	UpdatedBy    uuid.UUID              `json:"updatedBy,omitempty"`
	UpdatedAt    time.Time              `json:"updatedAt,omitempty"`
}

type ResourceServiceImpl struct {
	logger             *zap.SugaredLogger
	resourceRepository repository.ResourceRepository
}

func NewResourceServiceImpl(logger *zap.SugaredLogger, resourceRepository repository.ResourceRepository) *ResourceServiceImpl {
	return &ResourceServiceImpl{
		logger:             logger,
		resourceRepository: resourceRepository,
	}
}

func (impl *ResourceServiceImpl) CreateResource(resource ResourceDto) (ResourceDto, error) {
	validate := validator.New()
	if err := validate.Struct(resource); err != nil {
		return ResourceDto{}, err
	}
	resource.CreatedAt = time.Now().UTC()
	resource.UpdatedAt = time.Now().UTC()
	if err := impl.resourceRepository.Create(&repository.Resource{
		ID:        resource.ResourceId,
		Name:      resource.ResourceName,
		Kind:      resource.ResourceType,
		Options:   resource.Options,
		CreatedBy: resource.CreatedBy,
		CreatedAt: resource.CreatedAt,
		UpdatedBy: resource.UpdatedBy,
		UpdatedAt: resource.UpdatedAt,
	}); err != nil {
		return ResourceDto{}, err
	}
	return resource, nil
}

func (impl *ResourceServiceImpl) DeleteResource(resourceId uuid.UUID) error {
	if err := impl.resourceRepository.Delete(resourceId); err != nil {
		return err
	}
	return nil
}

func (impl *ResourceServiceImpl) UpdateResource(resource ResourceDto) (ResourceDto, error) {
	validate := validator.New()
	if err := validate.Struct(resource); err != nil {
		return ResourceDto{}, err
	}
	resource.UpdatedAt = time.Now().UTC()
	if err := impl.resourceRepository.Update(&repository.Resource{
		ID:        resource.ResourceId,
		Name:      resource.ResourceName,
		Kind:      resource.ResourceType,
		Options:   resource.Options,
		CreatedBy: resource.CreatedBy,
		CreatedAt: resource.CreatedAt,
		UpdatedBy: resource.UpdatedBy,
		UpdatedAt: resource.UpdatedAt,
	}); err != nil {
		return ResourceDto{}, err
	}
	return resource, nil
}

func (impl *ResourceServiceImpl) GetResource(resourceId uuid.UUID) (ResourceDto, error) {
	res, err := impl.resourceRepository.RetrieveById(resourceId)
	if err != nil {
		return ResourceDto{}, err
	}
	resDto := ResourceDto{
		ResourceId:   res.ID,
		ResourceName: res.Name,
		ResourceType: res.Kind,
		Options:      res.Options,
		CreatedBy:    res.CreatedBy,
		CreatedAt:    res.CreatedAt,
		UpdatedBy:    res.UpdatedBy,
		UpdatedAt:    res.UpdatedAt,
	}
	return resDto, nil
}

func (impl *ResourceServiceImpl) FindAllResources() ([]ResourceDto, error) {
	res, err := impl.resourceRepository.RetrieveAll()
	if err != nil {
		return nil, err
	}
	resDtoSlice := make([]ResourceDto, 0, len(res))
	for _, value := range res {
		resDtoSlice = append(resDtoSlice, ResourceDto{
			ResourceId:   value.ID,
			ResourceName: value.Name,
			ResourceType: value.Kind,
			Options:      value.Options,
			CreatedBy:    value.CreatedBy,
			CreatedAt:    value.CreatedAt,
			UpdatedBy:    value.UpdatedBy,
			UpdatedAt:    value.UpdatedAt,
		})
	}
	return resDtoSlice, nil
}

func (impl *ResourceServiceImpl) OpenConnection(resource ResourceDto) (*sql.DB, error) {
	resourceConn := &connector.Connector{
		Type:    resource.ResourceType,
		Options: resource.Options,
	}
	dbResource := resourceConn.Generate()
	if dbResource == nil {
		err := errors.New("invalid ResourceType: unsupported type")
		return nil, err
	}
	if err := dbResource.Format(resourceConn); err != nil {
		return nil, err
	}
	dbConn, err := dbResource.Connection()
	if err != nil {
		return nil, err
	}
	return dbConn, nil
}
