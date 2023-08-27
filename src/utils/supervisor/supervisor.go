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

package supervisor

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	resty "github.com/go-resty/resty/v2"
	"github.com/illacloud/builder-backend/src/utils/config"
	"github.com/illacloud/builder-backend/src/utils/idconvertor"
	"github.com/illacloud/builder-backend/src/utils/tokenvalidator"
)

const (
	BASEURL = "http://127.0.0.1:9001/api/v1"
	// access control part
	VALIDATE_USER_ACCOUNT = "/accessControl/account/validateResult"
	GET_TEAM_PERMISSIONS  = "/accessControl/teams/%s/permissions"
	CAN_ACCESS            = "/accessControl/teams/%s/unitType/%s/unitID/%s/attribute/canAccess/%s"
	CAN_MANAGE            = "/accessControl/teams/%s/unitType/%s/unitID/%s/attribute/canManage/%s"
	CAN_MANAGE_SPECIAL    = "/accessControl/teams/%s/unitType/%s/unitID/%s/attribute/canManageSpecial/%s"
	CAN_MODIFY            = "/accessControl/teams/%s/unitType/%s/unitID/%s/attribute/canModify/%s/from/%s/to/%s"
	CAN_DELETE            = "/accessControl/teams/%s/unitType/%s/unitID/%s/attribute/canDelete/%s"
	// data control part
	GET_USER               = "/dataControl/users/%s"
	GET_MULTI_USER         = "/dataControl/users/multi/%s"
	GET_TEAM_BY_IDENTIFIER = "/dataControl/teams/byIdentifier/%s"
	GET_TEAM_BY_ID         = "/dataControl/teams/%d"
)

var instance *Supervisor
var once sync.Once

func GetInstance() *Supervisor {
	once.Do(func() {
		if instance == nil {
			instance = NewSupervisor() // not thread safe
		}
	})
	return instance
}

type Supervisor struct {
	API       string
	Validator *tokenvalidator.RequestTokenValidator
}

func NewSupervisor() *Supervisor {
	conf := config.GetInstance()
	return &Supervisor{
		API:       conf.GetIllaSupervisorInternalRestAPI(),
		Validator: tokenvalidator.NewRequestTokenValidator(),
	}
}

func (supervisor *Supervisor) ValidateUserAccount(token string) (bool, error) {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization-Token", token).
		SetHeader("Request-Token", supervisor.Validator.GenerateValidateToken(token)).
		Get(supervisor.API + VALIDATE_USER_ACCOUNT)
	fmt.Printf("response: %+v, err: %+v", resp, err)
	if resp.StatusCode() != http.StatusOK {
		if err != nil {
			return false, errors.New("request illa supervisor failed.")
		}
		return false, nil
	}
	return true, nil
}

func (supervisor *Supervisor) GetTeamPermissions(teamID int) (string, error) {
	teamIDString := idconvertor.ConvertIntToString(teamID)
	client := resty.New()
	resp, err := client.R().
		SetHeader("Request-Token", supervisor.Validator.GenerateValidateToken(teamIDString)).
		Get(supervisor.API + fmt.Sprintf(GET_TEAM_PERMISSIONS, teamIDString))
	if resp.StatusCode() != http.StatusOK {
		if err != nil {
			return "", errors.New("request illa supervisor failed.")
		}
		return "", errors.New("validate failed.")
	}
	fmt.Printf("response: %+v, err: %+v", resp, err)
	return resp.String(), nil
}

func (supervisor *Supervisor) CanAccess(token string, teamID int, unitType int, unitID int, attributeID int) (bool, error) {
	teamIDString := idconvertor.ConvertIntToString(teamID)
	unitTypeString := idconvertor.ConvertIntToString(unitType)
	unitIDString := idconvertor.ConvertIntToString(unitID)
	attributeIDString := idconvertor.ConvertIntToString(attributeID)

	fmt.Printf(fmt.Sprintf(CAN_ACCESS, teamIDString, unitTypeString, unitIDString, attributeIDString))
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization-Token", token).
		SetHeader("Request-Token", supervisor.Validator.GenerateValidateToken(token, teamIDString, unitTypeString, unitIDString, attributeIDString)).
		Get(supervisor.API + fmt.Sprintf(CAN_ACCESS, teamIDString, unitTypeString, unitIDString, attributeIDString))
	if resp.StatusCode() != http.StatusOK {
		if err != nil {
			return false, errors.New("request illa supervisor failed: " + err.Error())
		}
		return false, nil
	}
	return true, nil
}

func (supervisor *Supervisor) CanManage(token string, teamID int, unitType int, unitID int, attributeID int) (bool, error) {
	teamIDString := idconvertor.ConvertIntToString(teamID)
	unitTypeString := idconvertor.ConvertIntToString(unitType)
	unitIDString := idconvertor.ConvertIntToString(unitID)
	attributeIDString := idconvertor.ConvertIntToString(attributeID)

	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization-Token", token).
		SetHeader("Request-Token", supervisor.Validator.GenerateValidateToken(token, teamIDString, unitTypeString, unitIDString, attributeIDString)).
		Get(supervisor.API + fmt.Sprintf(CAN_MANAGE, teamIDString, unitTypeString, unitIDString, attributeIDString))
	if resp.StatusCode() != http.StatusOK {
		if err != nil {
			return false, errors.New("request illa supervisor failed: " + err.Error())
		}
		return false, nil
	}
	return true, nil
}

func (supervisor *Supervisor) CanManageSpecial(token string, teamID int, unitType int, unitID int, attributeID int) (bool, error) {
	teamIDString := idconvertor.ConvertIntToString(teamID)
	unitTypeString := idconvertor.ConvertIntToString(unitType)
	unitIDString := idconvertor.ConvertIntToString(unitID)
	attributeIDString := idconvertor.ConvertIntToString(attributeID)

	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization-Token", token).
		SetHeader("Request-Token", supervisor.Validator.GenerateValidateToken(token, teamIDString, unitTypeString, unitIDString, attributeIDString)).
		Get(supervisor.API + fmt.Sprintf(CAN_MANAGE_SPECIAL, teamIDString, unitTypeString, unitIDString, attributeIDString))
	if resp.StatusCode() != http.StatusOK {
		if err != nil {
			return false, errors.New("request illa supervisor failed: " + err.Error())
		}
		return false, nil
	}
	return true, nil
}

func (supervisor *Supervisor) CanModify(token string, teamID int, unitType int, unitID int, attributeID int, fromID int, toID int) (bool, error) {
	teamIDString := idconvertor.ConvertIntToString(teamID)
	unitTypeString := idconvertor.ConvertIntToString(unitType)
	unitIDString := idconvertor.ConvertIntToString(unitID)
	attributeIDString := idconvertor.ConvertIntToString(attributeID)
	fromIDString := idconvertor.ConvertIntToString(fromID)
	toIDString := idconvertor.ConvertIntToString(toID)

	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization-Token", token).
		SetHeader("Request-Token", supervisor.Validator.GenerateValidateToken(token, teamIDString, unitTypeString, unitIDString, attributeIDString, fromIDString, toIDString)).
		Get(supervisor.API + fmt.Sprintf(CAN_MODIFY, teamIDString, unitTypeString, unitIDString, attributeIDString, fromID, toID))
	if resp.StatusCode() != http.StatusOK {
		if err != nil {
			return false, errors.New("request illa supervisor failed: " + err.Error())
		}
		return false, nil
	}
	return true, nil
}

func (supervisor *Supervisor) CanDelete(token string, teamID int, unitType int, unitID int, attributeID int) (bool, error) {
	teamIDString := idconvertor.ConvertIntToString(teamID)
	unitTypeString := idconvertor.ConvertIntToString(unitType)
	unitIDString := idconvertor.ConvertIntToString(unitID)
	attributeIDString := idconvertor.ConvertIntToString(attributeID)

	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization-Token", token).
		SetHeader("Request-Token", supervisor.Validator.GenerateValidateToken(token, teamIDString, unitTypeString, unitIDString, attributeIDString)).
		Get(supervisor.API + fmt.Sprintf(CAN_DELETE, teamIDString, unitTypeString, unitIDString, attributeIDString))
	if resp.StatusCode() != http.StatusOK {
		if err != nil {
			return false, errors.New("request illa supervisor failed: " + err.Error())
		}
		return false, nil
	}
	return true, nil
}

func (supervisor *Supervisor) GetUser(targetUserID int) (string, error) {
	targetUserIDString := idconvertor.ConvertIntToString(targetUserID)
	client := resty.New()
	resp, err := client.R().
		SetHeader("Request-Token", supervisor.Validator.GenerateValidateToken(targetUserIDString)).
		Get(supervisor.API + fmt.Sprintf(GET_USER, targetUserIDString))
	if resp.StatusCode() != http.StatusOK {
		if err != nil {
			return "", errors.New("request illa supervisor failed.")
		}
		return "", errors.New("validate failed.")
	}
	return resp.String(), nil
}

func (supervisor *Supervisor) GetMultiUser(targetUserIDsInString string) (string, error) {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Request-Token", supervisor.Validator.GenerateValidateToken(targetUserIDsInString)).
		Get(supervisor.API + fmt.Sprintf(GET_MULTI_USER, targetUserIDsInString))
	if resp.StatusCode() != http.StatusOK {
		if err != nil {
			return "", errors.New("request illa supervisor failed.")
		}
		return "", errors.New("validate failed.")
	}
	return resp.String(), nil
}

func (supervisor *Supervisor) GetTeamByIdentifier(targetTeamIdentifier string) (string, error) {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Request-Token", supervisor.Validator.GenerateValidateToken(targetTeamIdentifier)).
		Get(supervisor.API + fmt.Sprintf(GET_TEAM_BY_IDENTIFIER, targetTeamIdentifier))
	if resp.StatusCode() != http.StatusOK {
		if err != nil {
			return "", errors.New("request illa supervisor failed.")
		}
		return "", errors.New("validate failed.")
	}
	return resp.String(), nil
}

func (supervisor *Supervisor) GetTeamByID(targetTeamID int) (string, error) {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Request-Token", supervisor.Validator.GenerateValidateToken(strconv.Itoa(targetTeamID))).
		Get(supervisor.API + fmt.Sprintf(GET_TEAM_BY_ID, targetTeamID))
	if resp.StatusCode() != http.StatusOK {
		if err != nil {
			return "", errors.New("request illa supervisor failed.")
		}
		return "", errors.New("validate failed.")
	}
	return resp.String(), nil
}
