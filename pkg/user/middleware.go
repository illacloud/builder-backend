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
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	supervisior "github.com/illacloud/builder-backend/internal/util/supervisior"
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

		sv, err := supervisior.NewSupervisior()
		fmt.Printf("err: %v\n", err)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			c.Next()
		}
		validated, errInValidate := sv.ValidateUserAccount(token)
		fmt.Printf("token: %v\n", token)
		fmt.Printf("errInValidate: %v\n", errInValidate)
		if errInValidate != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			c.Next()
		}
		if !validated {
			c.AbortWithStatus(http.StatusUnauthorized)
			c.Next()
		}
		c.Next()
	}
}
