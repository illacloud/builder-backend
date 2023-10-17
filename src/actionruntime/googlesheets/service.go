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

package googlesheets

import (
	"errors"
	"fmt"
	"sort"

	"github.com/go-playground/validator/v10"
	"github.com/illacloud/builder-backend/src/actionruntime/common"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/api/sheets/v4"
)

type Connector struct {
	resourceOptions Resource
	actionOptions   Action
}

type ActionRunner struct {
	opts    map[string]interface{}
	service *sheets.Service
}

func (g *Connector) ValidateResourceOptions(resourceOptions map[string]interface{}) (common.ValidateResult, error) {
	// format resource options
	if err := mapstructure.Decode(resourceOptions, &g.resourceOptions); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate Google Sheets resource options
	validate := validator.New()
	if err := validate.Struct(g.resourceOptions); err != nil {
		return common.ValidateResult{Valid: false}, err
	}
	return common.ValidateResult{Valid: true}, nil
}

func (g *Connector) ValidateActionTemplate(actionOptions map[string]interface{}) (common.ValidateResult, error) {
	// format action options
	if err := mapstructure.Decode(actionOptions, &g.actionOptions); err != nil {
		return common.ValidateResult{Valid: false}, err
	}

	// validate Google Sheets action options
	validate := validator.New()
	if err := validate.Struct(g.actionOptions); err != nil {
		return common.ValidateResult{Valid: false}, err
	}
	return common.ValidateResult{Valid: true}, nil
}

func (g *Connector) TestConnection(resourceOptions map[string]interface{}) (common.ConnectionResult, error) {
	return common.ConnectionResult{Success: true}, nil
}

func (g *Connector) GetMetaInfo(resourceOptions map[string]interface{}) (common.MetaInfoResult, error) {
	// get Google Drive service instance
	driveService, err := g.getDriveWithOpts(resourceOptions)
	if err != nil {
		return common.MetaInfoResult{Success: false}, err
	}

	// get all spreadsheet information
	query := "mimeType='application/vnd.google-apps.spreadsheet'"
	files, err := driveService.Files.List().Q(query).Do()
	if err != nil {
		return common.MetaInfoResult{Success: false}, err
	}

	// output spreadsheet information
	res := make([]map[string]interface{}, len(files.Files))
	for i, v := range files.Files {
		res[i] = map[string]interface{}{"id": v.Id, "name": v.Name}
	}

	return common.MetaInfoResult{
		Success: true,
		Schema:  map[string]interface{}{"spreadsheets": res},
	}, nil
}

func (g *Connector) Run(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	// get Google Sheets service instance
	svc, err := g.getSheetsWithOpts(resourceOptions)
	if err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	// format action options
	if err := mapstructure.Decode(actionOptions, &g.actionOptions); err != nil {
		return common.RuntimeResult{Success: false}, err
	}

	res := common.RuntimeResult{
		Success: false,
		Rows:    []map[string]interface{}{},
		Extra:   map[string]interface{}{},
	}

	// build ActionRunner
	actionRunner := &ActionRunner{
		service: svc,
		opts:    g.actionOptions.Opts,
	}

	// different methods call different functions
	switch g.actionOptions.Method {
	case READ_ACTION:
		res, err = actionRunner.Read()
		if err != nil {
			res.Success = false
			return res, err
		}
	case APPEND_ACTION:
		res, err = actionRunner.Append()
		if err != nil {
			res.Success = false
			return res, err
		}
	case UPDATE_ACTION:
		res, err = actionRunner.Update()
		if err != nil {
			res.Success = false
			return res, err
		}
	case BULKUPDATE_ACTION:
		res, err = actionRunner.BulkUpdate()
		if err != nil {
			res.Success = false
			return res, err
		}
	case DELETE_ACTION:
		res, err = actionRunner.DeleteSingleRow()
		if err != nil {
			res.Success = false
			return res, err
		}
	case CREATE_ACTION:
		res, err = actionRunner.CreateASpreadsheet()
		if err != nil {
			res.Success = false
			return res, err
		}
	case COPY_ACTION:
		res, err = actionRunner.CopyFromAToB()
		if err != nil {
			res.Success = false
			return res, err
		}
	case LIST_ACTION:
		driveService, err := g.getDriveWithOpts(resourceOptions)
		if err != nil {
			res.Success = false
			return res, err
		}
		// get all spreadsheet information
		query := "mimeType='application/vnd.google-apps.spreadsheet'"
		files, err := driveService.Files.List().Q(query).Do()
		if err != nil {
			res.Success = false
			return res, err
		}
		// output spreadsheet information
		filesArray := make([]map[string]interface{}, len(files.Files))
		for i, v := range files.Files {
			filesArray[i] = map[string]interface{}{"id": v.Id, "name": v.Name}
		}
		res.Rows = filesArray
	case GET_ACTION:
		res, err = actionRunner.GetSpreadsheetInfo()
		if err != nil {
			res.Success = false
			return res, err
		}
	default:
		return res, errors.New("invalid action method")
	}

	return res, nil
}

func (r *ActionRunner) Read() (common.RuntimeResult, error) {
	// format read action options
	var readOpts ReadOpts
	if err := mapstructure.Decode(r.opts, &readOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate read action options
	validate := validator.New()
	if err := validate.Struct(readOpts); err != nil {
		return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": err.Error()}}}, nil
	}

	readRange := readOpts.A1Notation
	if readOpts.RangeType == "limit" {
		sheetName := ""
		if readOpts.SheetName == "" {
			sheetName = "Sheet1"
		}

		// get the total number of rows of a spreadsheet
		resp, err := r.service.Spreadsheets.Get(readOpts.Spreadsheet).Do()
		if err != nil {
			return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": err.Error()}}}, nil
		}

		totalRows := 0
		for _, sheet := range resp.Sheets {
			if sheet.Properties.Title == sheetName {
				totalRows = int(sheet.Properties.GridProperties.RowCount)
				break
			}
		}

		// calculate the range in A1 notation
		startRow := readOpts.Offset + 1
		endRow := 0
		if readOpts.Limit == 0 {
			endRow = totalRows
		} else {
			endRow = startRow + readOpts.Limit
		}
		if endRow > totalRows {
			endRow = totalRows
		}

		if endRow == 0 {
			readRange = fmt.Sprintf("%s!A%d:Z", sheetName, startRow)
		} else {
			readRange = fmt.Sprintf("%s!A%d:Z%d", sheetName, startRow, endRow)
		}

	}

	valuesResp, err := r.service.Spreadsheets.Values.Get(readOpts.Spreadsheet, readRange).Do()
	if err != nil {
		return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": err.Error()}}}, nil
	}

	if len(valuesResp.Values) == 0 {
		return common.RuntimeResult{Success: true}, nil
	}

	// Create a slice of maps to store the data
	data := make([]map[string]interface{}, len(valuesResp.Values)-1)

	// Extract column headers
	headers := valuesResp.Values[0]

	// Iterate through the rows, skipping the header row
	for i, row := range valuesResp.Values[1:] {
		data[i] = make(map[string]interface{}, len(headers))
		for j, cell := range row {
			if j >= len(headers) {
				break
			}
			header := interfaceToString(headers[j])
			data[i][header] = cell
		}
	}

	return common.RuntimeResult{Success: true, Rows: data}, nil
}

func (r *ActionRunner) Append() (common.RuntimeResult, error) {
	// format append action options
	var appendOpts AppendOpts
	if err := mapstructure.Decode(r.opts, &appendOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate append action options
	validate := validator.New()
	if err := validate.Struct(appendOpts); err != nil {
		return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": err.Error()}}}, nil
	}

	sheet := "Sheet1"
	if appendOpts.SheetName != "" {
		sheet = appendOpts.SheetName
	}
	// get the last non-empty row in the sheet
	resp, err := r.service.Spreadsheets.Values.Get(appendOpts.Spreadsheet, sheet).Do()
	if err != nil {
		return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": err.Error()}}}, nil
	}

	// calculate the range to append based on the existing data
	rangeToAppend := fmt.Sprintf("%s!A%d", sheet, len(resp.Values)+1)
	valuesToAppend := make([][]interface{}, len(appendOpts.Values)+1)
	if len(resp.Values) == 0 {
		keys := make([]string, 0, 0)
		if len(appendOpts.Values) != 0 {
			for k := range appendOpts.Values[0] {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			rowValues := make([]interface{}, 0)
			for _, k := range keys {
				rowValues = append(rowValues, k)
			}
			valuesToAppend[0] = rowValues
		}
		// convert the input data format to the required format for appending
		for i, row := range appendOpts.Values {
			rowValues := make([]interface{}, 0)
			for _, k := range keys {
				rowValues = append(rowValues, row[k])
			}
			valuesToAppend[i+1] = rowValues
		}
	} else {
		keys := make([]string, 0, len(resp.Values[0]))
		for _, k := range resp.Values[0] {
			keys = append(keys, interfaceToString(k))
		}
		valuesToAppend = make([][]interface{}, len(appendOpts.Values))
		for i, row := range appendOpts.Values {
			rowValues := make([]interface{}, 0)
			for _, k := range keys {
				rowValues = append(rowValues, row[k])
			}
			valuesToAppend[i] = rowValues
		}
	}

	rb := &sheets.ValueRange{
		MajorDimension: "ROWS",
		Values:         valuesToAppend,
	}

	appendResp, err := r.service.Spreadsheets.Values.Append(appendOpts.Spreadsheet, rangeToAppend, rb).ValueInputOption("RAW").Do()
	if err != nil {
		return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": err.Error()}}}, nil
	}
	res := make([]map[string]interface{}, 1, 1)
	res[0] = map[string]interface{}{
		"spreadsheetId": appendResp.SpreadsheetId,
		"tableRange":    appendResp.TableRange,
		"updates": map[string]interface{}{
			"spreadsheetId":  appendResp.SpreadsheetId,
			"updatedRange":   appendResp.Updates.UpdatedRange,
			"updatedRows":    appendResp.Updates.UpdatedRows,
			"updatedColumns": appendResp.Updates.UpdatedColumns,
			"updatedCells":   appendResp.Updates.UpdatedCells,
		},
	}

	return common.RuntimeResult{Success: true, Rows: res}, nil
}

func (r *ActionRunner) Update() (common.RuntimeResult, error) {
	// format update action options
	var updateOpts UpdateOpts
	if err := mapstructure.Decode(r.opts, &updateOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate update action options
	validate := validator.New()
	if err := validate.Struct(updateOpts); err != nil {
		return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": err.Error()}}}, nil
	}

	if updateOpts.SheetName == "" {
		updateOpts.SheetName = "Sheet1"
	}

	// get the header row in the sheet
	readRange := fmt.Sprintf("%s!A1:Z1", updateOpts.SheetName)
	resp, err := r.service.Spreadsheets.Values.Get(updateOpts.Spreadsheet, readRange).Do()
	if err != nil {
		return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": err.Error()}}}, nil
	}
	keys := make([]string, 0, 0)
	if len(resp.Values) != 0 {
		keys = make([]string, 0, len(resp.Values[0]))
		for _, k := range resp.Values[0] {
			keys = append(keys, interfaceToString(k))
		}
	}

	// convert the input data format to the required format for updating.
	valuesToUpdate := make([][]interface{}, len(updateOpts.Values))
	for i, row := range updateOpts.Values {
		rowValues := make([]interface{}, 0)
		for _, k := range keys {
			rowValues = append(rowValues, row[k])
		}
		valuesToUpdate[i] = rowValues
	}

	res := make([]map[string]interface{}, 1, 1)

	if updateOpts.FilterType == "a1" {
		rb := &sheets.ValueRange{
			MajorDimension: "ROWS",
			Values:         valuesToUpdate,
		}

		resp, err := r.service.Spreadsheets.Values.Update(updateOpts.Spreadsheet, updateOpts.A1Notation, rb).ValueInputOption("RAW").Do()
		res[0] = map[string]interface{}{
			"spreadsheetId": resp.SpreadsheetId,
			"updates": map[string]interface{}{
				"spreadsheetId":  resp.SpreadsheetId,
				"updatedRange":   resp.UpdatedRange,
				"updatedRows":    resp.UpdatedRows,
				"updatedColumns": resp.UpdatedColumns,
				"updatedCells":   resp.UpdatedCells,
			},
		}
		if err != nil {
			return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": err.Error()}}}, nil
		}
	} else if updateOpts.FilterType == "filter" {
		return updateSpreadsheetByFilters(r.service, updateOpts.Spreadsheet, updateOpts.SheetName, updateOpts.Filters, updateOpts.Values)
	}

	return common.RuntimeResult{Success: true, Rows: res}, nil
}

func (r *ActionRunner) BulkUpdate() (common.RuntimeResult, error) {
	// format bulkUpdate action options
	var bulkUpdateOpts BulkUpdateOpts
	if err := mapstructure.Decode(r.opts, &bulkUpdateOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate bulkUpdate action options
	validate := validator.New()
	if err := validate.Struct(bulkUpdateOpts); err != nil {
		return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": err.Error()}}}, nil
	}

	if bulkUpdateOpts.SheetName == "" {
		bulkUpdateOpts.SheetName = "Sheet1"
	}

	// read the data from the sheet
	readRange := fmt.Sprintf("%s!A1:Z", bulkUpdateOpts.SheetName)
	resp, err := r.service.Spreadsheets.Values.Get(bulkUpdateOpts.Spreadsheet, readRange).Do()
	if err != nil {
		return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": err.Error()}}}, nil
	}
	keys := make(map[string]int)
	if len(resp.Values) != 0 {
		for i, k := range resp.Values[0] {
			keys[interfaceToString(k)] = i
		}
	}

	if len(resp.Values) == 0 {
		return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": "no data found"}}}, nil
	}

	// create a map to store row numbers for each primary key
	rowNumbers := make(map[string]int)

	// iterate through rows to find the primary key column index and row numbers
	primaryKeyIndex := -1
	for rowIndex, row := range resp.Values {
		for colIndex, cell := range row {
			if rowIndex == 0 && cell == bulkUpdateOpts.PrimaryKey {
				primaryKeyIndex = colIndex
			} else if rowIndex > 0 && colIndex == primaryKeyIndex {
				rowNumbers[interfaceToString(cell)] = rowIndex + 1
			}
		}
	}

	if primaryKeyIndex == -1 {
		return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": "primary key column not found"}}}, nil
	}

	// create the bulk update request
	updateRequests := []*sheets.Request{}

	for _, value := range bulkUpdateOpts.RowsArray {
		primaryKeyValue, ok := value[bulkUpdateOpts.PrimaryKey]
		if !ok {
			return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": "primary key value missing in provided values"}}}, nil
		}

		rowNumber, ok := rowNumbers[interfaceToString(primaryKeyValue)]
		if !ok {
			return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": "primary key value not found in the sheet"}}}, nil
		}

		for colName, cellValue := range value {
			if colName != bulkUpdateOpts.PrimaryKey {
				cellValueString := interfaceToString(cellValue)
				updateRequest := &sheets.Request{
					UpdateCells: &sheets.UpdateCellsRequest{
						Range: &sheets.GridRange{
							SheetId:          0,
							StartColumnIndex: int64(keys[colName]),
							EndColumnIndex:   int64(keys[colName] + 1),
							StartRowIndex:    int64(rowNumber - 1),
							EndRowIndex:      int64(rowNumber),
						},
						Rows: []*sheets.RowData{
							{
								Values: []*sheets.CellData{
									{
										UserEnteredValue: &sheets.ExtendedValue{
											StringValue: &cellValueString,
										},
									},
								},
							},
						},
						Fields: "*",
					},
				}
				updateRequests = append(updateRequests, updateRequest)
			}
		}
	}

	// send the bulk update request
	batchUpdate := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: updateRequests,
	}

	batchUpdateResp, err := r.service.Spreadsheets.BatchUpdate(bulkUpdateOpts.Spreadsheet, batchUpdate).Do()
	if err != nil {
		return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": err.Error()}}}, nil
	}
	res := make([]map[string]interface{}, 1, 1)
	res[0] = map[string]interface{}{
		"spreadsheetId":      batchUpdateResp.SpreadsheetId,
		"updatedSpreadsheet": batchUpdateResp.UpdatedSpreadsheet,
		"replies":            batchUpdateResp.Replies,
	}

	return common.RuntimeResult{Success: true, Rows: res}, nil
}

func (r *ActionRunner) DeleteSingleRow() (common.RuntimeResult, error) {
	// format delete action options
	var deleteOpts DeleteOpts
	if err := mapstructure.Decode(r.opts, &deleteOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate delete action options
	validate := validator.New()
	if err := validate.Struct(deleteOpts); err != nil {
		return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": err.Error()}}}, nil
	}

	if deleteOpts.SheetName == "" {
		deleteOpts.SheetName = "Sheet1"
	}

	// get sheet id
	var sheetID int64
	spreadsheet, err := r.service.Spreadsheets.Get(deleteOpts.Spreadsheet).Do()
	if err != nil {
		return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": err.Error()}}}, nil
	}

	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == deleteOpts.SheetName {
			sheetID = sheet.Properties.SheetId
		}
	}

	deleteDimensionRequest := &sheets.DeleteDimensionRequest{
		Range: &sheets.DimensionRange{
			SheetId:    sheetID,
			Dimension:  "ROWS",
			StartIndex: int64(deleteOpts.RowIndex),
			EndIndex:   int64(deleteOpts.RowIndex) + 1,
		},
	}

	requests := []*sheets.Request{
		{
			DeleteDimension: deleteDimensionRequest,
		},
	}

	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}

	batchUpdateResp, err := r.service.Spreadsheets.BatchUpdate(deleteOpts.Spreadsheet, batchUpdateRequest).Do()
	if err != nil {
		return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": err.Error()}}}, nil
	}
	res := make([]map[string]interface{}, 1, 1)
	res[0] = map[string]interface{}{
		"spreadsheetId":      batchUpdateResp.SpreadsheetId,
		"updatedSpreadsheet": batchUpdateResp.UpdatedSpreadsheet,
		"replies":            batchUpdateResp.Replies,
	}

	return common.RuntimeResult{Success: true, Rows: res}, nil
}

func (r *ActionRunner) CreateASpreadsheet() (common.RuntimeResult, error) {
	// format create action options
	var createOpts CreateOpts
	if err := mapstructure.Decode(r.opts, &createOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate create action options
	validate := validator.New()
	if err := validate.Struct(createOpts); err != nil {
		return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": err.Error()}}}, nil
	}

	newSpreadsheet := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: createOpts.Title,
		},
	}

	createdSpreadsheet, err := r.service.Spreadsheets.Create(newSpreadsheet).Do()
	if err != nil {
		return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": err.Error()}}}, nil
	}

	res := make([]map[string]interface{}, 1, 1)
	res[0] = map[string]interface{}{
		"spreadsheetId":  createdSpreadsheet.SpreadsheetId,
		"spreadsheetUrl": createdSpreadsheet.SpreadsheetUrl,
		"sheets":         createdSpreadsheet.Sheets,
		"properties":     createdSpreadsheet.Properties,
	}

	return common.RuntimeResult{Success: true, Rows: res}, nil
}

func (r *ActionRunner) CopyFromAToB() (common.RuntimeResult, error) {
	// format copy action options
	var copyOpts CopyOpts
	if err := mapstructure.Decode(r.opts, &copyOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate copy action options
	validate := validator.New()
	if err := validate.Struct(copyOpts); err != nil {
		return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": err.Error()}}}, nil
	}

	if copyOpts.SheetName == "" {
		copyOpts.SheetName = "Sheet1"
	}
	if copyOpts.ToSheet == "" {
		copyOpts.ToSheet = "Sheet1"
	}

	// get sheet id
	var sheetID int64
	spreadsheet, err := r.service.Spreadsheets.Get(copyOpts.Spreadsheet).Do()
	if err != nil {
		return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": err.Error()}}}, nil
	}

	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == copyOpts.SheetName {
			sheetID = sheet.Properties.SheetId
		}
	}

	copySheetRequest := &sheets.CopySheetToAnotherSpreadsheetRequest{
		DestinationSpreadsheetId: copyOpts.ToSpreadsheet,
	}

	copyResp, err := r.service.Spreadsheets.Sheets.CopyTo(copyOpts.Spreadsheet, sheetID, copySheetRequest).Do()
	if err != nil {
		return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": err.Error()}}}, nil
	}

	updateSheetPropertiesRequest := &sheets.UpdateSheetPropertiesRequest{
		Properties: &sheets.SheetProperties{
			SheetId: copyResp.SheetId,
			Title:   copyOpts.ToSheet,
		},
		Fields: "title",
	}

	requests := []*sheets.Request{
		{
			UpdateSheetProperties: updateSheetPropertiesRequest,
		},
	}

	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}

	batchUpdateResp, err := r.service.Spreadsheets.BatchUpdate(copyOpts.ToSpreadsheet, batchUpdateRequest).Do()
	if err != nil {
		return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": err.Error()}}}, nil
	}
	res := make([]map[string]interface{}, 1, 1)
	res[0] = map[string]interface{}{
		"spreadsheetId":      batchUpdateResp.SpreadsheetId,
		"updatedSpreadsheet": batchUpdateResp.UpdatedSpreadsheet,
		"replies":            batchUpdateResp.Replies,
	}

	return common.RuntimeResult{Success: true, Rows: res}, nil
}

func (r *ActionRunner) GetSpreadsheetInfo() (common.RuntimeResult, error) {
	// format get action options
	var getOpts GetOpts
	if err := mapstructure.Decode(r.opts, &getOpts); err != nil {
		return common.RuntimeResult{Success: false}, err
	}
	// validate get action options
	validate := validator.New()
	if err := validate.Struct(getOpts); err != nil {
		return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": err.Error()}}}, nil
	}

	spreadsheet, err := r.service.Spreadsheets.Get(getOpts.Spreadsheet).IncludeGridData(false).Do()
	if err != nil {
		return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": err.Error()}}}, nil
	}

	res := make([]map[string]interface{}, 1, 1)
	res[0] = map[string]interface{}{
		"spreadsheetId":  spreadsheet.SpreadsheetId,
		"spreadsheetUrl": spreadsheet.SpreadsheetUrl,
		"sheets":         spreadsheet.Sheets,
		"properties":     spreadsheet.Properties,
	}

	return common.RuntimeResult{Success: true, Rows: res}, nil
}

func updateSpreadsheetByFilters(srv *sheets.Service, spreadsheetID, sheetName string, filters []Filter, values []map[string]interface{}) (common.RuntimeResult, error) {
	// get the sheet data
	readRange := fmt.Sprintf("%s!A1:Z", sheetName)
	response, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": err.Error()}}}, nil
	}

	// Find the matching rows
	matchingRows := findMatchingRows(response.Values, filters)

	// Update the matching rows with the new values
	var updateRows []*sheets.ValueRange
	for i, rowIndex := range matchingRows {
		row := response.Values[rowIndex]
		updateValues(row, values[i], response.Values[0])
		updateRange := fmt.Sprintf("%s!A%d:Z%d", sheetName, rowIndex+1, rowIndex+1)
		updateRow := &sheets.ValueRange{
			Range:  updateRange,
			Values: [][]interface{}{row},
		}
		updateRows = append(updateRows, updateRow)
	}

	// Execute the batch update request
	batchUpdate := &sheets.BatchUpdateValuesRequest{
		ValueInputOption: "RAW",
		Data:             updateRows,
	}

	resp, err := srv.Spreadsheets.Values.BatchUpdate(spreadsheetID, batchUpdate).Do()
	res := make([]map[string]interface{}, 1, 1)
	res[0] = map[string]interface{}{
		"spreadsheetId": resp.SpreadsheetId,
		"updates": map[string]interface{}{
			"spreadsheetId":  resp.SpreadsheetId,
			"updatedSheets":  resp.TotalUpdatedSheets,
			"updatedRows":    resp.TotalUpdatedRows,
			"updatedColumns": resp.TotalUpdatedColumns,
			"updatedCells":   resp.TotalUpdatedCells,
		},
	}
	if err != nil {
		return common.RuntimeResult{Success: false, Rows: []map[string]interface{}{0: {"message": err.Error()}}}, nil
	}

	return common.RuntimeResult{Success: true, Rows: res}, nil
}

// findMatchingRows is a helper function that returns the indices of rows matching the filters.
func findMatchingRows(sheetData [][]interface{}, filters []Filter) []int {
	var matchingRows []int

	for rowIndex, row := range sheetData {
		matches := true
		for _, filter := range filters {
			columnIndex := getColumnIndex(sheetData[0], filter.Key)
			if columnIndex == -1 || interfaceToString(row[columnIndex]) != filter.Value {
				matches = false
				break
			}
		}
		if matches {
			matchingRows = append(matchingRows, rowIndex)
		}
	}

	return matchingRows
}

// getColumnIndex is a helper function that returns the index of a column by its name.
func getColumnIndex(header []interface{}, columnName string) int {
	for i, col := range header {
		if interfaceToString(col) == columnName {
			return i
		}
	}
	return -1
}

// updateValues is a helper function that updates the values of a row based on the given data.
func updateValues(row []interface{}, values map[string]interface{}, header []interface{}) {
	for key, value := range values {
		columnIndex := getColumnIndex(header, key)
		if columnIndex != -1 {
			row[columnIndex] = value
		}
	}
}
