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

package elasticsearch

import (
	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/mitchellh/mapstructure"
)

func (e *Connector) getConnectionWithOptions(resourceOptions map[string]interface{}) (*es.Client, error) {
	if err := mapstructure.Decode(resourceOptions, &e.ResourceOpts); err != nil {
		return nil, err
	}

	esCfg := es.Config{
		Addresses: []string{
			e.ResourceOpts.Host + ":" + e.ResourceOpts.Port,
		},
		Username: e.ResourceOpts.Username,
		Password: e.ResourceOpts.Password,
	}
	esClient, err := es.NewClient(esCfg)
	if err != nil {
		return nil, err
	}
	return esClient, err
}
