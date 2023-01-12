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

package illacloudbackendsdk

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

const (
	BASEURL          	  = "http://127.0.0.1:9001/api/v1"
	VALIDATE_USER_ACCOUNT = "/accessControl/account/validateResult"
	GET_TEAM_PERMISSIONS  = "/accessControl/team/%s/permissions"
	CAN_ACCESS 			  = "/accessControl/team/%s/unitType/%s/unitID/%s/attribute/canAccess/%s"
	CAN_MANAGE 			  = "/accessControl/team/%s/unitType/%s/unitID/%s/attribute/canManage/%s"
	CAN_MANAGE_SPECIAL    = "/accessControl/team/%s/unitType/%s/unitID/%s/attribute/canManageSpecial/%s"
	CAN_MODIFY 			  = "/accessControl/team/%s/unitType/%s/unitID/%s/attribute/canModify/%s/from/%s/to/%s"
	CAN_DELETE 			  = "/accessControl/team/%s/unitType/%s/unitID/%s/attribute/canDelete/%s"
)

type IllaCloudSDK struct {
	Validator: *RequestTokenValidator, 
}


func NewIllaCloudSDK() (*IllaCloudSDK, error) {
	v, err := NewRequestTokenValidator()
	if err != nil {
		return nil, err
	}
	return &IllaCloudSDK{
		Validator: v
	}, nil
}


func (sdk *IllaCloudSDK) ValidateUserAccount(token string) (bool, error) {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization-Token", token).
		SetHeader("Request-Token", sdk.Validator.GenerateValidateToken(token)).
		Get(BASEURL + VALIDATE_USER_ACCOUNT)
	if resp.StatusCode() != http.StatusOK {
		if err != nil {
			return false, errors.New("request illa cloud failed.")
		}
		return false, errors.New("validate failed.")
	}
	fmt.Printf("response: %+v, err: %+v", resp, err)
	return true, nil
}


func (sdk *IllaCloudSDK) GetTeamPermissions(teamID int) (string, error) {
	teamIDString := strconv.Itoa(teamID)
	client := resty.New()
	resp, err := client.R().
		SetHeader("Request-Token", sdk.Validator.GenerateValidateToken(teamIDString)).
		Get(BASEURL + fmt.Sprintf(GET_TEAM_PERMISSIONS, teamIDString))
	if resp.StatusCode() != http.StatusOK {
		if err != nil {
			return "", errors.New("request illa cloud failed.")
		}
		return "", errors.New("validate failed.")
	}
	fmt.Printf("response: %+v, err: %+v", resp, err)
	return resp, nil
}


func (sdk *IllaCloudSDK) CanAccess(token string, teamID int, unitType int, unitID int, attributeID int) (bool, error) {
	teamIDString := strconv.Itoa(teamID)
	unitTypeString := strconv.Itoa(unitType)
	unitIDString := strconv.Itoa(unitID)
	attributeIDString := strconv.Itoa(attributeID)

	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization-Token", token).
		SetHeader("Request-Token", sdk.Validator.GenerateValidateToken(token, teamIDString, unitTypeString, unitIDString, attributeIDString)).
		Get(BASEURL + fmt.Sprintf(CAN_ACCESS, teamIDString, unitTypeString, unitIDString, attributeIDString))
	if resp.StatusCode() != http.StatusOK {
		if err != nil {
			return false, errors.New("request illa cloud failed: " + err.Error())
		}
		return false, nil
	}
	return true, nil
}

func (sdk *IllaCloudSDK) CanManage(token string, teamID int, unitType int, unitID int, attributeID int) (bool, error) {
	teamIDString := strconv.Itoa(teamID)
	unitTypeString := strconv.Itoa(unitType)
	unitIDString := strconv.Itoa(unitID)
	attributeIDString := strconv.Itoa(attributeID)

	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization-Token", token).
		SetHeader("Request-Token", sdk.Validator.GenerateValidateToken(token, teamIDString, unitTypeString, unitIDString, attributeIDString)).
		Get(BASEURL + fmt.Sprintf(CAN_MANAGE, teamIDString, unitTypeString, unitIDString, attributeIDString))
	if resp.StatusCode() != http.StatusOK {
		if err != nil {
			return false, errors.New("request illa cloud failed: " + err.Error())
		}
		return false, nil
	}
	return true, nil
}

func (sdk *IllaCloudSDK) CanManageSpecial(token string, teamID int, unitType int, unitID int, attributeID int) (bool, error) {
	teamIDString := strconv.Itoa(teamID)
	unitTypeString := strconv.Itoa(unitType)
	unitIDString := strconv.Itoa(unitID)
	attributeIDString := strconv.Itoa(attributeID)

	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization-Token", token).
		SetHeader("Request-Token", sdk.Validator.GenerateValidateToken(token, teamIDString, unitTypeString, unitIDString, attributeIDString)).
		Get(BASEURL + fmt.Sprintf(CAN_MANAGE_SPECIAL, teamIDString, unitTypeString, unitIDString, attributeIDString))
	if resp.StatusCode() != http.StatusOK {
		if err != nil {
			return false, errors.New("request illa cloud failed: " + err.Error())
		}
		return false, nil
	}
	return true, nil
}

func (sdk *IllaCloudSDK) CanModify(token string, teamID int, unitType int, unitID int, attributeID int, fromID int, toID int) (bool, error) {
	teamIDString := strconv.Itoa(teamID)
	unitTypeString := strconv.Itoa(unitType)
	unitIDString := strconv.Itoa(unitID)
	attributeIDString := strconv.Itoa(attributeID)

	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization-Token", token).
		SetHeader("Request-Token", sdk.Validator.GenerateValidateToken(token, teamIDString, unitTypeString, unitIDString, attributeIDString, fromID, toID)).
		Get(BASEURL + fmt.Sprintf(CAN_MODIFY, teamIDString, unitTypeString, unitIDString, attributeIDString, fromID, toID))
	if resp.StatusCode() != http.StatusOK {
		if err != nil {
			return false, errors.New("request illa cloud failed: " + err.Error())
		}
		return false, nil
	}
	return true, nil
}

func (sdk *IllaCloudSDK) CanDelete(token string, teamID int, unitType int, unitID int, attributeID int) (bool, error) {
	teamIDString := strconv.Itoa(teamID)
	unitTypeString := strconv.Itoa(unitType)
	unitIDString := strconv.Itoa(unitID)
	attributeIDString := strconv.Itoa(attributeID)

	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization-Token", token).
		SetHeader("Request-Token", sdk.Validator.GenerateValidateToken(token, teamIDString, unitTypeString, unitIDString, attributeIDString)).
		Get(BASEURL + fmt.Sprintf(CAN_DELETE, teamIDString, unitTypeString, unitIDString, attributeIDString))
	if resp.StatusCode() != http.StatusOK {
		if err != nil {
			return false, errors.New("request illa cloud failed: " + err.Error())
		}
		return false, nil
	}
	return true, nil
}
