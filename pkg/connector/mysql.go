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

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mitchellh/mapstructure"
)

type MySQLConnection struct {
	Kind    string
	Options MySQLOption
}

type MySQLOption struct {
	Host             string
	Port             string
	DatabaseName     string
	DatabaseUsername string
	DatabasePassword string
	SSL              bool
	SSH              bool
	AdvancedOptions  AdvancedOptions
}

type AdvancedOptions struct {
	SSHHost       string
	SSHPort       string
	SSHUsername   string
	SSHPassword   string
	SSHPrivateKey map[string]interface{}
	SSHPassphrase string
	ServerCert    map[string]interface{}
	ClientKey     map[string]interface{}
	ClientCert    map[string]interface{}
}

func (m *MySQLConnection) Format(connector *Connector) error {
	if err := mapstructure.Decode(connector.Options, &m.Options); err != nil {
		return err
	}
	return nil
}

func (m *MySQLConnection) Connection() (*sql.DB, error) {
	var db *sql.DB
	var err error
	if m.Options.SSH {
		db, err = m.ConnectionViaSSH()
	} else if m.Options.SSL {
		db, err = m.ConnectionViaSSL()
	} else {
		db, err = m.ConnectionPure()
	}
	return db, err
}

func (m *MySQLConnection) ConnectionViaSSH() (*sql.DB, error) {
	// TODO: to complete code
	return nil, nil
}

func (m *MySQLConnection) ConnectionViaSSL() (*sql.DB, error) {
	// TODO: to complete code
	return nil, nil
}

func (m *MySQLConnection) ConnectionPure() (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", m.Options.DatabaseUsername, m.Options.DatabasePassword,
		m.Options.Host, m.Options.Port, m.Options.DatabaseName)
	db, err := sql.Open("mysql", dsn+"?timeout=5s")
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}
