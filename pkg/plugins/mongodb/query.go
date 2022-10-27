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

import (
	"github.com/illa-family/builder-backend/pkg/plugins/common"

	"go.mongodb.org/mongo-driver/mongo"
)

type QueryRunner struct {
	client *mongo.Client
	query  Query
}

func (q *QueryRunner) aggregate() (common.RuntimeResult, error) {
	return common.RuntimeResult{Success: false}, nil
}
