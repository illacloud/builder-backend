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

package connector

import "database/sql"

var (
	MYSQL_RESOURCE = "mysql"
)

type BaseConnector interface {
	Generate() BaseConnection
}

type BaseConnection interface {
	Format(connector *Connector) error
	Connection() (*sql.DB, error)
}

type Connector struct {
	Type    string
	Options map[string]interface{}
}

func (c *Connector) Generate() BaseConnection {
	switch c.Type {
	case MYSQL_RESOURCE:
		return &MySQLConnection{Kind: c.Type}
	default:
		return nil
	}
}
