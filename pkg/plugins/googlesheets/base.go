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
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/mitchellh/mapstructure"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const (
	SERVICE_ACCOUNT_AUTH = "serviceAccount"
	OAUTH2_AUTH          = "oauth2"

	READ_ACTION       = "read"
	APPEND_ACTION     = "append"
	UPDATE_ACTION     = "update"
	BULKUPDATE_ACTION = "bulkUpdate"
	DELETE_ACTION     = "delete"
	CREATE_ACTION     = "create"
	COPY_ACTION       = "copy"
	LIST_ACTION       = "list"
	GET_ACTION        = "get"
)

func (g *Connector) getSheetsWithOpts(resourceOpts map[string]interface{}) (*sheets.Service, error) {
	if err := mapstructure.Decode(resourceOpts, &g.ResourceOpts); err != nil {
		return nil, err
	}
	switch g.ResourceOpts.Authentication {
	case SERVICE_ACCOUNT_AUTH:
		var saOpts SAOpts
		if err := mapstructure.Decode(g.ResourceOpts.Opts, &saOpts); err != nil {
			return nil, err
		}
		return getSheetsWithKey(saOpts.PrivateKey)
	case OAUTH2_AUTH:
		return getSheetsWithOAuth2()
	default:
		return nil, errors.New("unsupported authentication method")
	}
}

func getSheetsWithKey(privateKey string) (*sheets.Service, error) {
	config, err := google.JWTConfigFromJSON([]byte(privateKey), sheets.SpreadsheetsScope)
	if err != nil {
		return nil, err
	}

	// create an OAuth2 client using JWT configuration.
	ctx := context.Background()
	client := config.Client(ctx)

	// create a Google Sheets service instance using an OAuth2 client
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return srv, nil
}

func getSheetsWithOAuth2() (*sheets.Service, error) {
	// TODO
	return nil, nil
}

func (g *Connector) getDriveWithOpts(resourceOpts map[string]interface{}) (*drive.Service, error) {
	if err := mapstructure.Decode(resourceOpts, &g.ResourceOpts); err != nil {
		return nil, err
	}
	switch g.ResourceOpts.Authentication {
	case SERVICE_ACCOUNT_AUTH:
		var saOpts SAOpts
		if err := mapstructure.Decode(g.ResourceOpts.Opts, &saOpts); err != nil {
			return nil, err
		}
		return getDriveWithKey(saOpts.PrivateKey)
	case OAUTH2_AUTH:
		return getDriveWithOAuth2()
	default:
		return nil, errors.New("unsupported authentication method")
	}
}

func getDriveWithKey(privateKey string) (*drive.Service, error) {
	config, err := google.JWTConfigFromJSON([]byte(privateKey), drive.DriveScope)
	if err != nil {
		return nil, err
	}

	// create an OAuth2 client using JWT configuration.
	ctx := context.Background()
	client := config.Client(ctx)

	// create a Google Drive service instance using an OAuth2 client
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return srv, nil
}

func getDriveWithOAuth2() (*drive.Service, error) {
	// TODO
	return nil, nil
}

func interfaceToString(i interface{}) string {
	switch v := i.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
