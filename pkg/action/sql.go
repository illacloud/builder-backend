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
	"github.com/pkg/errors"
)

type SqlAction struct {
	Type        string
	SqlTemplate SqlTemplate
	Resource    *connector.Connector
}

type SqlTemplate struct {
	Sql string
}

func (s *SqlAction) Run() (interface{}, error) {
	dbResource := s.Resource.Generate()
	if dbResource == nil {
		err := errors.New("invalid ResourceType: unsupported type")
		return nil, err
	}
	if err := dbResource.Format(s.Resource); err != nil {
		return nil, err
	}
	dbConn, err := dbResource.Connection()
	defer dbConn.Close()

	if err != nil {
		return nil, err
	}
	res, err := dbConn.Exec(s.SqlTemplate.Sql)
	if err != nil {
		return nil, err
	}
	return res, nil
}
