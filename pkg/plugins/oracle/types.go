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

package oracle

type Resource struct {
	Host     string `mapstructure:"host" validate:"required"`
	Port     string `mapstructure:"port" validate:"required"`
	Type     string `mapstructure:"connectionType" validate:"oneof=SID Service"`
	Name     string `mapstructure:"name"`
	SSL      bool   `mapstructure:"ssl"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type Action struct {
	Mode string                 `mapstructure:"mode" validate:"oneof=sql gui"`
	Opts map[string]interface{} `mapstructure:"opts"`
}

type SQL struct {
	Raw string `mapstructure:"raw"`
}

type GUIBulkOpts struct {
	Table   string                   `mapstructure:"table"`
	Type    string                   `mapstructure:"actionType"`
	Records []map[string]interface{} `mapstructure:"records"`
	Key     string                   `mapstructure:"primaryKey"`
}
