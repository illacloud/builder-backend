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

package clickhouse

type Resource struct {
	Host         string `validate:"required"`
	Port         int    `validate:"gt=0"`
	DatabaseName string `validate:"required"`
	Username     string
	Password     string
	SSL          SSLOptions `validate:"required"`
}

type SSLOptions struct {
	SSL        bool
	SelfSigned bool
	CACert     string `validate:"required_unless=SelfSigned false"`
	PrivateKey string
	ClientCert string
}

type Action struct {
	Query string
	Mode  string `validate:"required,oneof=gui sql"`
}
