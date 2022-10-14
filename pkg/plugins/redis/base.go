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

package redis

import (
	"crypto/tls"

	"github.com/go-redis/redis/v8"
	"github.com/mitchellh/mapstructure"
)

func (r *Connector) getConnectionWithOptions(resourceOptions map[string]interface{}) (*redis.Client, error) {
	if err := mapstructure.Decode(resourceOptions, &r.Resource); err != nil {
		return nil, err
	}

	options := redis.Options{
		Addr:     r.Resource.Host + ":" + r.Resource.Port,
		Username: r.Resource.DatabaseUsername,
		Password: r.Resource.DatabasePassword,
		DB:       r.Resource.DatabaseIndex,
	}
	if r.Resource.SSL {
		tlsConfig := tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: r.Resource.Host,
		}
		options.TLSConfig = &tlsConfig
	}
	rdb := redis.NewClient(&options)

	return rdb, nil
}
