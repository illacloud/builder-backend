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

import (
	"fmt"
	"net/url"

	_ "github.com/go-kivik/couchdb/v4"
	"github.com/go-kivik/kivik/v4"
	"github.com/mitchellh/mapstructure"
)

const (
	LIST_METHOD     = "listRecords"
	RETRIEVE_METHOD = "retrieveRecord"
	CREATE_METHOD   = "createRecord"
	UPDATE_METHOD   = "updateRecord"
	DELETE_METHOD   = "deleteRecord"
	FIND_METHOD     = "find"
	GET_METHOD      = "getView"
)

func (c *Connector) getClient(resourceOptions map[string]interface{}) (*kivik.Client, error) {
	// format resource options
	if err := mapstructure.Decode(resourceOptions, &c.resourceOptions); err != nil {
		return nil, err
	}

	protocolStr := "http"
	if c.resourceOptions.SSL {
		protocolStr = "https"
	}
	escapedPassword := url.QueryEscape(c.resourceOptions.Password)
	dsn := fmt.Sprintf("%s://%s:%s@%s:%s/", protocolStr, c.resourceOptions.Username, escapedPassword,
		c.resourceOptions.Host, c.resourceOptions.Port)
	client, err := kivik.New("couch", dsn)
	if err != nil {
		return nil, err
	}

	return client, nil
}
