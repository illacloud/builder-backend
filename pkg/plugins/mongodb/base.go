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

package mongodb

//import (
//	"fmt"
//
//	"github.com/mitchellh/mapstructure"
//	"go.mongodb.org/mongo-driver/mongo"
//)
//
//func (m *Connector) getConnectionWithOptions(resourceOptions map[string]interface{}) (*mongo.Client, error) {
//	if err := mapstructure.Decode(resourceOptions, &m.Resource); err != nil {
//		return nil, err
//	}
//	uri := ""
//	if m.Resource.ConnectionFormat == STANDARD_FORMAT {
//		uri = fmt.Sprintf("mongodb://%s:%s", m.Resource.Host, m.Resource.Port)
//	} else if m.Resource.ConnectionFormat == DNSSEEDLIST_FORMAT {
//		uri = fmt.Sprintf("mongodb+srv://%s", m.Resource.Host)
//	} else {
//		return nil, errors.New("unsupported connection format")
//	}
//	//clientOptions := options.Client().ApplyURI(uri)
//	return nil, nil
//}
