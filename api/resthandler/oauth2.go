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

package resthandler

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/illacloud/builder-backend/internal/idconvertor"
	"github.com/illacloud/builder-backend/pkg/resource"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
)

type OAuth2RestHandler interface {
	GoogleOAuth2(c *gin.Context)
}

type OAuth2RestHandlerImpl struct {
	logger          *zap.SugaredLogger
	resourceService resource.ResourceService
}

func NewOAuth2RestHandlerImpl(logger *zap.SugaredLogger, resourceService resource.ResourceService) *OAuth2RestHandlerImpl {
	return &OAuth2RestHandlerImpl{
		logger:          logger,
		resourceService: resourceService,
	}
}

func (impl OAuth2RestHandlerImpl) GoogleOAuth2(c *gin.Context) {
	stateToken := c.Query("state")
	code := c.Query("code")
	errorOAuth2Callback := c.Query("error")
	if stateToken == "" {
		c.JSON(400, nil)
		return
	}
	teamID, userID, resourceID, url, err := extractGSOAuth2Token(stateToken)
	if err != nil {
		c.Redirect(302, fmt.Sprintf("%s?status=%d&resourceID=%d", url, 2, resourceID))
		return
	}
	if errorOAuth2Callback != "" || code == "" {
		c.Redirect(302, fmt.Sprintf("%s?status=%d&resourceID=%d", url, 2, resourceID))
		return
	}
	// get resource
	res, err := impl.resourceService.GetResource(teamID, resourceID)
	if err != nil {
		c.Redirect(302, fmt.Sprintf("%s?status=%d&resourceID=%d", url, 2, resourceID))
		return
	}
	if res.Type != "googlesheets" {
		c.Redirect(302, fmt.Sprintf("%s?status=%d&resourceID=%d", url, 2, resourceID))
		return
	}
	var googleSheetsResource GoogleSheetsResource
	if err := mapstructure.Decode(res.Options, &googleSheetsResource); err != nil {
		c.Redirect(302, fmt.Sprintf("%s?status=%d&resourceID=%d", url, 2, resourceID))
		return
	}
	if googleSheetsResource.Authentication != "oauth2" {
		c.Redirect(302, fmt.Sprintf("%s?status=%d&resourceID=%d", url, 2, resourceID))
		return
	}

	// exchange oauth2 refresh token and access token
	client := resty.New()
	resp, err := client.R().
		SetFormData(map[string]string{
			"client_id":     os.Getenv("ILLA_GS_CLIENT_ID"),
			"client_secret": os.Getenv("ILLA_GS_CLIENT_SECRET"),
			"code":          code,
			"grant_type":    "authorization_code",
			"redirect_uri":  os.Getenv("ILLA_GS_REDIRECT_URI"),
		}).
		Post("https://oauth2.googleapis.com/token")
	if resp.IsSuccess() {
		type ExchangeTokenSuccessResponse struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			Expiry       int    `json:"expires_in"`
			Scope        string `json:"scope"`
			TokenType    string `json:"token_type"`
		}
		var exchangeTokenSuccessResponse ExchangeTokenSuccessResponse
		if err := json.Unmarshal(resp.Body(), &exchangeTokenSuccessResponse); err != nil {
			c.Redirect(302, fmt.Sprintf("%s?status=%d&resourceID=%d", url, 2, resourceID))
			return
		}
		googleSheetsResource.Opts.RefreshToken = exchangeTokenSuccessResponse.RefreshToken
		googleSheetsResource.Opts.TokenType = exchangeTokenSuccessResponse.TokenType
		googleSheetsResource.Opts.AccessToken = exchangeTokenSuccessResponse.AccessToken
		googleSheetsResource.Opts.Status = 2
	} else if resp.IsError() {
		c.Redirect(302, fmt.Sprintf("%s?status=%d&resourceID=%d", url, 2, resourceID))
		return
	}

	// update resource and return response
	if _, err := impl.resourceService.UpdateResource(resource.ResourceDto{
		ID:   idconvertor.ConvertStringToInt(res.ID),
		Name: res.Name,
		Type: res.Type,
		Options: map[string]interface{}{
			"authentication": googleSheetsResource.Authentication,
			"opts": map[string]interface{}{
				"accessType":   googleSheetsResource.Opts.AccessType,
				"accessToken":  googleSheetsResource.Opts.AccessToken,
				"tokenType":    googleSheetsResource.Opts.TokenType,
				"refreshToken": googleSheetsResource.Opts.RefreshToken,
				"status":       googleSheetsResource.Opts.Status,
			},
		},
		UpdatedAt: time.Now().UTC(),
		UpdatedBy: userID,
	}); err != nil {
		c.Redirect(302, fmt.Sprintf("%s?status=%d&resourceID=%d", url, 2, resourceID))
		return
	}

	// redirect
	c.Redirect(302, fmt.Sprintf("%s?status=%d&resourceID=%d", url, 1, resourceID))
	return
}
