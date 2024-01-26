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

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"

	parser_template "github.com/illacloud/builder-backend/src/utils/parser/template"
)

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

func RetrieveToMapByDriverRows(rows driver.Rows) ([]map[string]interface{}, error) {
	columns := rows.Columns()
	mapData := make([]map[string]interface{}, 0)

	// value of every row
	values := make([]driver.Value, len(columns))
	// get all values
	for {
		errInFetchNextRows := rows.Next(values)
		if errInFetchNextRows != nil {
			break
		}

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

func ProcessTemplateByContext(template interface{}, context map[string]interface{}) (interface{}, error) {
	fmt.Printf("[IDUMP] template: %+v\n", template)
	processorMethod := func(template interface{}, context map[string]interface{}) (interface{}, error) {
		valueInMap, valueIsMap := template.(map[string]interface{})
		if valueIsMap {
			processedValue, errInPreprocessTemplate := ProcessTemplateByContext(valueInMap, context)
			if errInPreprocessTemplate != nil {
				return nil, errInPreprocessTemplate
			}
			return processedValue, nil
		}

		// check if value is string, then process it
		valueInString, valueIsString := template.(string)
		if valueIsString {
			// check if value is json string
			var valueInJson interface{}
			errInUnmarshal := json.Unmarshal([]byte(valueInString), &valueInJson)
			itIsJSONString := errInUnmarshal == nil
			if itIsJSONString {
				// json string, process it as array or map
				processedValue, errInPreprocessTemplate := ProcessTemplateByContext(valueInJson, context)
				if errInPreprocessTemplate != nil {
					return nil, errInPreprocessTemplate
				}
				processedValueInJSON, _ := json.Marshal(processedValue)
				return string(processedValueInJSON), nil
			} else {
				// jsut a normal string
				processedTemplate, errInAssembleTemplate := parser_template.AssembleTemplateWithVariable(valueInString, context)
				if errInAssembleTemplate != nil {
					return nil, errInAssembleTemplate
				}
				return processedTemplate, nil
			}
		}
		return template, nil
	}

	// assert input
	if template == nil {
		return template, nil
	}

	inputInSLice, inputIsSlice := template.([]interface{})
	inputInMap, inputIsMap := template.(map[string]interface{})
	inputInString, inputIsString := template.(string)

	// process it
	if inputIsSlice {
		fmt.Printf("[IDUMP] inputIsSlice: %+v\n", inputIsSlice)

		newSlice := make([]interface{}, 0)
		for _, value := range inputInSLice {
			processedTemplate, errInProcess := processorMethod(value, context)
			if errInProcess != nil {
				return nil, errInProcess
			}
			newSlice = append(newSlice, processedTemplate)
		}
		return newSlice, nil
	}
	if inputIsMap {
		fmt.Printf("[IDUMP] inputIsMap: %+v\n", inputIsMap)

		newMap := make(map[string]interface{}, 0)
		for key, value := range inputInMap {
			processedTemplate, errInProcess := processorMethod(value, context)
			if errInProcess != nil {
				return nil, errInProcess
			}
			newMap[key] = processedTemplate
		}
		return newMap, nil
	}
	if inputIsString {
		fmt.Printf("[IDUMP] inputIsString: %+v\n", inputIsString)
		if inputInString == "" {
			fmt.Printf("[IDUMP] inputIsEmptyString: %+v\n", inputInString == "")
			return nil, nil
		}
		return processorMethod(inputInString, context)
	}
	return template, nil
}
