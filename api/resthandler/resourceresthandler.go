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

package resthandler

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/illa-family/builder-backend/pkg/connector"
	"github.com/illa-family/builder-backend/pkg/resource"
	"go.uber.org/zap"
	"net/http"
)

type TestConnectionRequest struct {
	Kind    string      `form:"phone" json:"kind" validate:"required"`
	Options MySQLOption `form:"options" json:"options" validate:"required"`
}

type MySQLOption struct {
	Host             string          `form:"host" json:"host" validate:"required"`
	Port             string          `form:"port" json:"port" validate:"required"`
	DatabaseName     string          `form:"databaseName" json:"databaseName"`
	DatabaseUsername string          `form:"databaseUsername" json:"databaseUsername" validate:"required"`
	DatabasePassword string          `form:"databasePassword" json:"databasePassword" validate:"required"`
	SSL              bool            `form:"ssl" json:"ssl"`
	SSH              bool            `form:"ssh" json:"ssh"`
	AdvancedOptions  AdvancedOptions `form:"advancedOptions" json:"advancedOptions" validate:"required"`
}

type AdvancedOptions struct {
	SSHHost       string `form:"sshHost" json:"sshHost"`
	SSHPort       string `form:"sshPort" json:"sshPort"`
	SSHUsername   string `form:"sshUsername" json:"sshUsername"`
	SSHPassword   string `form:"sshPassword" json:"sshPassword"`
	SSHPrivateKey string `form:"sshPrivateKey" json:"sshPrivateKey"`
	SSHPassphrase string `form:"sshPassphrase" json:"sshPassphrase"`
	ServerCert    string `form:"serverCert" json:"serverCert"`
	ClientKey     string `form:"clientKey" json:"clientKey"`
	ClientCert    string `form:"clientCert" json:"clientCert"`
}

type ResourceRestHandler interface {
	CreateResource(c *gin.Context)
	UpdateResource(c *gin.Context)
	DeleteResource(c *gin.Context)
	TestConnection(c *gin.Context)
}

type ResourceRestHandlerImpl struct {
	logger          *zap.SugaredLogger
	resourceService resource.ResourceService
}

func NewResourceRestHandlerImpl(logger *zap.SugaredLogger, resourceService resource.ResourceService) *ResourceRestHandlerImpl {
	return &ResourceRestHandlerImpl{
		logger:          logger,
		resourceService: resourceService,
	}
}

func (impl ResourceRestHandlerImpl) CreateResource(c *gin.Context) {
	c.JSON(http.StatusOK, "pass")
}

func (impl ResourceRestHandlerImpl) UpdateResource(c *gin.Context) {
	c.JSON(http.StatusOK, "pass")
}

func (impl ResourceRestHandlerImpl) DeleteResource(c *gin.Context) {
	c.JSON(http.StatusOK, "pass")
}

func (impl ResourceRestHandlerImpl) TestConnection(c *gin.Context) {
	// format data to Resource struct
	var requestInfo TestConnectionRequest
	err := json.NewDecoder(c.Request.Body).Decode(&requestInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": err.Error(),
			"data":    nil,
		})
		return
	}
	validate := validator.New()
	err = validate.Struct(requestInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": err.Error(),
			"data":    nil,
		})
		return
	}
	switch requestInfo.Kind {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", requestInfo.Options.DatabaseUsername,
			requestInfo.Options.DatabasePassword, requestInfo.Options.Host, requestInfo.Options.Port,
			requestInfo.Options.DatabaseName)
		err := connector.TestMySQLConnection(dsn)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "fail",
				"message": "test connection fail",
				"data":    nil,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "test connection success",
				"data":    nil,
			})
		}
		return
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "fail",
			"message": "not supported kind",
			"data":    nil,
		})
		return
	}
}
