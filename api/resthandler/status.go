// Copyright 2022 The ILLA Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by Statuslicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resthandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type StatusRestHandler interface {
	GetStatus(c *gin.Context)
}

type StatusRestHandlerImpl struct {
	logger *zap.SugaredLogger
}

func NewStatusRestHandlerImpl(logger *zap.SugaredLogger) *StatusRestHandlerImpl {
	return &StatusRestHandlerImpl{
		logger: logger,
	}
}

func (impl StatusRestHandlerImpl) GetStatus(c *gin.Context) {
	c.JSON(http.StatusOK, nil)
}
