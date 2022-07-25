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

    "go.uber.org/zap"
)

var type_array = [7]string{"restapi", "graphql", "redis", "mysql", "mariadb", "postgresql", "mongodb"}
var type_map = map[string]int{
    "restapi":    1,
    "graphql":    2,
    "redis":      3,
    "mysql":      4,
    "mariadb":    5,
    "postgresql": 6,
    "mongodb":    7,
}

type ResourceService interface {
    CreateResource(resource ResourceDto) (ResourceDto, error)
    DeleteResource(id int) error
    UpdateResource(resource ResourceDto) (ResourceDto, error)
    GetResource(id int) (ResourceDto, error)
    FindAllResources() ([]ResourceDto, error)
    OpenConnection(resource ResourceDto) (*sql.DB, error)
}

type ResourceDto struct {
    ID        int                    `json:"resourceId"`
    Name      string                 `json:"resourceName,omitempty" validate:"required"`
    Type      string                 `json:"resourceType,omitempty" validate:"oneof=restapi graphql redis mysql mariadb postgresql mongodb"`
    Options   map[string]interface{} `json:"options,omitempty" validate:"required"`
    CreatedAt time.Time              `json:"createdAt,omitempty"`
    CreatedBy int                    `json:"createdBy,omitempty"`
    UpdatedAt time.Time              `json:"updatedAt,omitempty"`
    UpdatedBy int                    `json:"updatedBy,omitempty"`
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
    ID, err := impl.resourceRepository.Create(&repository.Resource{
        Name:      resource.Name,
        Type:      type_map[resource.Type],
        Options:   resource.Options,
        CreatedAt: resource.CreatedAt,
        CreatedBy: resource.CreatedBy,
        UpdatedAt: resource.UpdatedAt,
        UpdatedBy: resource.UpdatedBy,
    })
    if err != nil {
        return ResourceDto{}, err
    }
    resource.ID = ID
    return resource, nil
}

func (impl *ResourceServiceImpl) DeleteResource(id int) error {
    if err := impl.resourceRepository.Delete(id); err != nil {
        return err
    }
    return nil
}

func (impl *ResourceServiceImpl) UpdateResource(resource ResourceDto) (ResourceDto, error) {
    if err := impl.resourceRepository.Update(&repository.Resource{
        ID:        resource.ID,
        Name:      resource.Name,
        Type:      type_map[resource.Type],
        Options:   resource.Options,
        CreatedAt: resource.CreatedAt,
        CreatedBy: resource.CreatedBy,
        UpdatedAt: resource.UpdatedAt,
        UpdatedBy: resource.UpdatedBy,
    }); err != nil {
        return ResourceDto{}, err
    }
    return resource, nil
}

func (impl *ResourceServiceImpl) GetResource(id int) (ResourceDto, error) {
    res, err := impl.resourceRepository.RetrieveByID(id)
    if err != nil {
        return ResourceDto{}, err
    }
    resDto := ResourceDto{
        ID:        res.ID,
        Name:      res.Name,
        Type:      type_array[res.Type-1],
        Options:   res.Options,
        CreatedAt: res.CreatedAt,
        CreatedBy: res.CreatedBy,
        UpdatedAt: res.UpdatedAt,
        UpdatedBy: res.UpdatedBy,
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
            ID:        value.ID,
            Name:      value.Name,
            Type:      type_array[value.Type-1],
            Options:   value.Options,
            CreatedAt: value.CreatedAt,
            CreatedBy: value.CreatedBy,
            UpdatedAt: value.UpdatedAt,
            UpdatedBy: value.UpdatedBy,
        })
    }
    return resDtoSlice, nil
}

func (impl *ResourceServiceImpl) OpenConnection(resource ResourceDto) (*sql.DB, error) {
    resourceConn := &connector.Connector{
        Type:    resource.Type,
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
