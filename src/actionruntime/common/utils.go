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
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/DmitriyVTitov/size"
	parser_template "github.com/illacloud/builder-backend/src/utils/parser/template"
)

const DEFAULT_QUERY_AND_EXEC_TIMEOUT = 30 * time.Second
const SQL_RESULT_MEMORY_LIMIT = 20971520   // 20 * 1024 * 1024 bytes
const SQL_RESULT_MEMORY_CHECK_SAMPLE = 100 // check 100 item bytes and calculate max item capacity

func RetrieveToMap(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	// rewrite columns for duplicate name
	renamedColumns := make([]string, 0)
	columnNameHitMap := make(map[string]int, 0)
	columnNamePosMap := make(map[string]int, 0)
	for pos, column := range columns {
		hitColumnTimes, hitColumn := columnNameHitMap[column]
		cloName := column
		if hitColumn {
			cloName += fmt.Sprintf("_%d", hitColumnTimes)
			// rewrite first column to "_0"
			if columnNameHitMap[column] == 1 {
				firstHitPos := columnNamePosMap[column]
				renamedColumns[firstHitPos] += "_0"
			}
			columnNameHitMap[column]++
		}
		columnNameHitMap[cloName] = 1
		columnNamePosMap[cloName] = pos
		renamedColumns = append(renamedColumns, cloName)
	}
	// count of columns
	count := len(renamedColumns)
	mapData := make([]map[string]interface{}, 0)

	// value of every row
	values := make([]interface{}, count)
	// pointer of every row values
	valPointers := make([]interface{}, count)
	iteratorNums := 0
	tableDataCapacity := 10000
	for rows.Next() {
		iteratorNums++
		// get pointer for every row
		for i := 0; i < count; i++ {
			valPointers[i] = &values[i]
		}

		// get query result
		rows.Scan(valPointers...)
		valPointersInJSONByte, _ := json.Marshal(valPointers)
		fmt.Printf("[DUMP] valPointersInJSONByte: %s\n", valPointersInJSONByte)
		// value for every single row
		entry := make(map[string]interface{})

		for i, col := range renamedColumns {
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
		// check tableData size by sample
		if iteratorNums == SQL_RESULT_MEMORY_CHECK_SAMPLE {
			tableDataSizeBySample := size.Of(mapData)
			tableDataCapacity = (SQL_RESULT_MEMORY_LIMIT / tableDataSizeBySample) * SQL_RESULT_MEMORY_CHECK_SAMPLE
		}
		if iteratorNums > tableDataCapacity {
			log.Printf("[ERROR] RetrieveToMap result exceeds 20MiB by iteratorNums: %d, size: %d", iteratorNums, size.Of(mapData))
			return nil, errors.New("returned result exceeds 20MiB, please adjust the query limit to reduce the number of results")
		}
	}

	return mapData, nil
}

func RetrieveToMapByDriverRows(rows driver.Rows) ([]map[string]interface{}, error) {
	columns := rows.Columns()
	mapData := make([]map[string]interface{}, 0)
	// rewrite columns for duplicate name
	renamedColumns := make([]string, 0)
	columnNameHitMap := make(map[string]int, 0)
	columnNamePosMap := make(map[string]int, 0)
	for pos, column := range columns {
		hitColumnTimes, hitColumn := columnNameHitMap[column]
		cloName := column
		if hitColumn {
			cloName += fmt.Sprintf("_%d", hitColumnTimes)
			// rewrite first column to "_0"
			if columnNameHitMap[column] == 1 {
				firstHitPos := columnNamePosMap[column]
				renamedColumns[firstHitPos] += "_0"
			}
			columnNameHitMap[column]++
		}
		columnNameHitMap[cloName] = 1
		columnNamePosMap[cloName] = pos
		renamedColumns = append(renamedColumns, cloName)
	}

	// value of every row
	values := make([]driver.Value, len(renamedColumns))
	iteratorNums := 0
	tableDataCapacity := 10000
	// get all values
	for {
		iteratorNums++
		errInFetchNextRows := rows.Next(values)
		if errInFetchNextRows != nil {
			break
		}

		// value for every single row
		entry := make(map[string]interface{})

		for i, col := range renamedColumns {
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
		// check tableData size by sample
		if iteratorNums == SQL_RESULT_MEMORY_CHECK_SAMPLE {
			tableDataSizeBySample := size.Of(mapData)
			tableDataCapacity = (SQL_RESULT_MEMORY_LIMIT / tableDataSizeBySample) * SQL_RESULT_MEMORY_CHECK_SAMPLE
		}
		if iteratorNums > tableDataCapacity {
			log.Printf("[ERROR] RetrieveToMap result exceeds 20MiB by iteratorNums: %d, size: %d", iteratorNums, size.Of(mapData))
			return nil, errors.New("returned result exceeds 20MiB, please adjust the query limit to reduce the number of results")
		}
	}

	return mapData, nil
}

func ProcessTemplateByContext(template interface{}, context map[string]interface{}) (interface{}, error) {
	fmt.Printf("[IDUMP] template: %+v\n", template)
	processorMethod := func(template string, context map[string]interface{}) (interface{}, error) {
		// check if value is json string
		var valueInJson interface{}
		errInUnmarshal := json.Unmarshal([]byte(template), &valueInJson)
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
			processedTemplate, errInAssembleTemplate := parser_template.AssembleTemplateWithVariable(template, context)
			if errInAssembleTemplate != nil {
				return nil, errInAssembleTemplate
			}
			return processedTemplate, nil
		}
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
			processedTemplate, errInProcess := ProcessTemplateByContext(value, context)
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
			processedTemplate, errInProcess := ProcessTemplateByContext(value, context)
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
