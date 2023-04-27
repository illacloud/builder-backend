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

import "database/sql"

func RetrieveToMap(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	// count of columns
	count := len(columns)
	mapData := make([]map[string]interface{}, 0)

	// value of every row
	values := make([]interface{}, count)
	// pointer of every row values
	valPointers := make([]interface{}, count)
	for rows.Next() {
		values = make([]interface{}, count)
		valPointers = make([]interface{}, count)
		// get pointer for every row
		for i := 0; i < count; i++ {
			valPointers[i] = &values[i]
		}

		// get query result
		rows.Scan(valPointers...)

		// value for every single row
		entry := make(map[string]interface{})

		for i, col := range columns {
			var v interface{}

			val := values[i]
			b, ok := val.([]byte)
			if ok {
				// []byte to string
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}
		mapData = append(mapData, entry)
	}

	return mapData, nil
}
