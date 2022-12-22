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

package firebase

import (
	"context"
	"encoding/json"

	firebase "firebase.google.com/go/v4"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/api/option"
)

func (f *Connector) getConnectionWithOptions(resourceOptions map[string]interface{}) (*firebase.App, error) {
	if err := mapstructure.Decode(resourceOptions, &f.ResourceOpts); err != nil {
		return nil, err
	}

	// build firebase service account
	privateKey, err := json.Marshal(f.ResourceOpts.PrivateKey)
	if err != nil {
		return nil, err
	}
	sa := option.WithCredentialsJSON(privateKey)

	// build firebase config for realtime database
	config := &firebase.Config{DatabaseURL: f.ResourceOpts.DatabaseURL}

	// new firebase app
	app, err := firebase.NewApp(context.Background(), config, sa)
	if err != nil {
		return nil, err
	}

	return app, nil
}
