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

package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	cloudsdk "github.com/illacloud/builder-backend/internal/util/illacloudbackendsdk"
)

func JWTAuth(authenticator Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// fetch content
		accessToken := c.Request.Header["Authorization"]
		var token string
		if len(accessToken) != 1 {
			c.AbortWithStatus(http.StatusUnauthorized)
		} else {
			token = accessToken[0]
		}

		// init deploy mode
		deployMode := os.Getenv("ILLA_DEPLOY_MODE")
		if deployMode == const.DEPLOY_MODE_CLOUD {
			sdk, err := cloudsdk.NewIllaCloudSDK()
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
			}
			validated, errInValidate := sdk.ValidateUserAccount(token)
			if errInValidate != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
			}
			if !validated {
				c.AbortWithStatus(http.StatusUnauthorized)
			}
			c.Next()
		}

		// local auth method
		userID, userUID, extractErr := authenticator.ExtractUserIDFromToken(token)
		validAccessToken, validaAccessErr := authenticator.ValidateAccessToken(token)
		validUser, validUserErr := authenticator.ValidateUser(userID, userUID)

		if validAccessToken && validUser && validaAccessErr == nil && extractErr == nil && validUserErr == nil {
			c.Set("userID", userID)
		} else {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		c.Next()
	}
}
