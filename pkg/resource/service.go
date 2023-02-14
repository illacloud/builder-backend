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
	"errors"
	"time"

	"github.com/illacloud/builder-backend/internal/repository"

	"go.uber.org/zap"
)

var type_array = [21]string{"restapi", "graphql", "redis", "mysql", "mariadb", "postgresql", "mongodb", "tidb",
	"elasticsearch", "s3", "smtp", "supabasedb", "firebase", "clickhouse", "mssql", "huggingface", "dynamodb", "snowflake",
	"couchdb", "hfendpoint", "oracle"}
var type_map = map[string]int{
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
	"dynamodb":      17,
	"snowflake":     18,
	"couchdb":       19,
	"hfendpoint":    20,
	"oracle":        21,
}

type ResourceService interface {
	CreateResource(resource ResourceDto) (ResourceDto, error)
	DeleteResource(id int) error
	UpdateResource(resource ResourceDto) (ResourceDto, error)
	GetResource(id int) (ResourceDto, error)
	FindAllResources() ([]ResourceDto, error)
	TestConnection(resource ResourceDto) (bool, error)
	ValidateResourceOptions(resourceType string, options map[string]interface{}) error
	GetMetaInfo(id int) (map[string]interface{}, error)
}

type ResourceDto struct {
	ID        int                    `json:"resourceId"`
	Name      string                 `json:"resourceName" validate:"required"`
	Type      string                 `json:"resourceType" validate:"oneof=restapi graphql redis mysql mariadb postgresql mongodb tidb elasticsearch s3 smtp supabasedb firebase clickhouse mssql huggingface dynamodb snowflake couchdb hfendpoint oracle"`
	Options   map[string]interface{} `json:"content" validate:"required"`
	CreatedAt time.Time              `json:"createdAt,omitempty"`
	CreatedBy int                    `json:"createdBy,omitempty"`
	UpdatedAt time.Time              `json:"updatedAt,omitempty"`
	UpdatedBy int                    `json:"updatedBy,omitempty"`
}

func (resourced *ResourceDto) ConstructByMap(data interface{}) {

	udata, ok := data.(map[string]interface{})
	if !ok {
		return
	}
	for k, v := range udata {
		switch k {
		case "id":
			idf, _ := v.(float64)
			resourced.ID = int(idf)
		case "name":
			resourced.Name, _ = v.(string)
		case "type":
			resourced.Type, _ = v.(string)
		case "options":
			resourced.Options, _ = v.(map[string]interface{})
		}
	}
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
	res, err := impl.resourceRepository.RetrieveAllByUpdatedTime()
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

func (impl *ResourceServiceImpl) TestConnection(resource ResourceDto) (bool, error) {
	rscFactory := Factory{Type: resource.Type}
	dbResource := rscFactory.Generate()
	if dbResource == nil {
		return false, errors.New("invalid ResourceType: unsupported type")
	}
	if _, err := dbResource.ValidateResourceOptions(resource.Options); err != nil {
		return false, err
	}
	connRes, err := dbResource.TestConnection(resource.Options)
	if err != nil || !connRes.Success {
		return false, errors.New("connection failed")
	}
	return true, nil
}

func (impl *ResourceServiceImpl) ValidateResourceOptions(resourceType string, options map[string]interface{}) error {
	rscFactory := Factory{Type: resourceType}
	dbResource := rscFactory.Generate()
	if dbResource == nil {
		return errors.New("invalid ResourceType: unsupported type")
	}
	if _, err := dbResource.ValidateResourceOptions(options); err != nil {
		return err
	}
	return nil
}

func (impl *ResourceServiceImpl) GetMetaInfo(id int) (map[string]interface{}, error) {
	rsc, err := impl.resourceRepository.RetrieveByID(id)
	if err != nil {
		return map[string]interface{}{}, err
	}
	rscFactory := Factory{Type: type_array[rsc.Type-1]}
	dbResource := rscFactory.Generate()
	if dbResource == nil {
		return map[string]interface{}{}, errors.New("invalid ResourceType: unsupported type")
	}
	if _, err := dbResource.ValidateResourceOptions(rsc.Options); err != nil {
		return map[string]interface{}{}, err
	}
	res, err := dbResource.GetMetaInfo(rsc.Options)
	if err != nil {
		return map[string]interface{}{}, err
	}

	return map[string]interface{}{"schema": res.Schema, "resourceName": rsc.Name}, nil
}
