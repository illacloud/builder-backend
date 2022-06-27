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
	"github.com/illa-family/builder-backend/pkg/connector"
	"github.com/mitchellh/mapstructure"
)

var (
	RESTACTION = "RESTApi"
	SQLACTION  = "SQL"
)

type ActionFactory interface {
	Generate() *ActionAssemblyline
}

type ActionAssemblyline interface {
	Run() (interface{}, error)
}

type Factory struct {
	Type     string
	Template map[string]interface{}
	Resource *connector.Connector
}

func (f *Factory) Build() ActionAssemblyline {
	switch f.Type {
	case RESTACTION:
		restapiAction := &RestApiAction{
			Type:     f.Type,
			Resource: f.Resource,
		}
		mapstructure.Decode(f.Template, &restapiAction.RestApiTemplate)
		return restapiAction
	case SQLACTION:
		sqlAction := &SqlAction{
			Type:     f.Type,
			Resource: f.Resource,
		}
		mapstructure.Decode(f.Template, &sqlAction.SqlTemplate)
		return sqlAction
	default:
		return nil
	}
}
