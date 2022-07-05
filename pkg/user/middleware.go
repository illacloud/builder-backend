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
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, _ := c.Cookie("access_token")
		refreshToken, _ := c.Cookie("refresh_token")

		validAccessToken, _ := ValidateAccessToken(accessToken)
		if validAccessToken {
			userId, err := ExtractUserIdFromToken(accessToken)
			if err != nil {
				c.AbortWithStatus(http.StatusUnauthorized)
			}
			c.Set("userId", userId)
		} else {
			// if access token expires, validate refresh token
			validRefreshToken, err := ValidateRefreshToken(refreshToken)
			if err != nil || !validRefreshToken {
				c.AbortWithStatus(http.StatusUnauthorized)
			} else {
				userId, err := ExtractUserIdFromToken(accessToken)
				if err != nil {
					c.AbortWithStatus(http.StatusUnauthorized)
				}
				c.Set("userId", userId)
				newAccessToken, _ := CreateAccessToken(userId)
				newRefreshToken, _ := CreateRefreshToken(newAccessToken)
				c.SetCookie("access_token", newAccessToken, 7200, "/", "localhost", false, true)
				c.SetCookie("refresh_token", newRefreshToken, 259200, "/", "localhost", false, true)
			}
		}

		c.Next()
	}
}
