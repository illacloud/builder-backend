// Copyright 2023 Illa Soft, Inc.
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

package couchdb

type resource struct {
	Host     string `validate:"required"`
	Port     string `validate:"required"`
	Username string
	Password string
	SSL      bool
}

type action struct {
	Method   string `validate:"required,oneof=listRecords retrieveRecord createRecord updateRecord deleteRecord find getView"`
	Database string
	Opts     map[string]interface{}
}
