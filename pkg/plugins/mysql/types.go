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

package mysql

type MySQLOptions struct {
	Host             string     `validate:"required"`
	Port             string     `validate:"required"`
	DatabaseName     string     `validate:"required"`
	DatabaseUsername string     `validate:"required"`
	DatabasePassword string     `validate:"required"`
	SSL              SSLOptions `validate:"required,omitempty"`
	SSH              SSHOptions `validate:"required,omitempty"`
}

type SSLOptions struct {
	SSL        bool
	ServerCert string `validate:"required_unless=SSL false"`
	ClientKey  string `validate:"required_unless=SSL false"`
	ClientCert string `validate:"required_unless=SSL false"`
}

type SSHOptions struct {
	SSH           bool
	SSHHost       string `validate:"required_unless=SSH false"`
	SSHPort       string `validate:"required_unless=SSH false"`
	SSHUsername   string `validate:"required_unless=SSH false"`
	SSHPassword   string `validate:"required_unless=SSH false"`
	SSHPrivateKey string `validate:"required_unless=SSH false"`
	SSHPassphrase string `validate:"required_unless=SSH false"`
}

type MySQLQuery struct {
	Mode  string `validate:"required,oneof=gui sql"`
	Query string
}
