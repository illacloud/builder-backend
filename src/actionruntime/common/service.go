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

package common

type DataConnector interface {
	ValidateResourceOptions(resourceOptions map[string]interface{}) (ValidateResult, error)
	ValidateActionTemplate(actionOptions map[string]interface{}) (ValidateResult, error)
	TestConnection(resourceOptions map[string]interface{}) (ConnectionResult, error)
	GetMetaInfo(resourceOptions map[string]interface{}) (MetaInfoResult, error)
	Run(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (RuntimeResult, error)
}
