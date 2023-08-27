package illadrive

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/illacloud/builder-backend/src/utils/config"
	"github.com/illacloud/builder-backend/src/utils/tokenvalidator"
)

const (
	BASEURL = "http://127.0.0.1:9002/api/v1"
	// api route part
	CREATE_TEAM_DRIVE = "/drive/teams/%d"
	DELETE_TEAM_DRIVE = "/drive/teams/%d"
)

type IllaDriveRestAPI struct {
	Config    *config.Config
	Validator *tokenvalidator.RequestTokenValidator
}

func NewIllaDriveRestAPI() (*IllaDriveRestAPI, error) {
	requestTokenValidator := tokenvalidator.NewRequestTokenValidator()
	return &IllaDriveRestAPI{
		Config:    config.GetInstance(),
		Validator: requestTokenValidator,
	}, nil
}

func (r *IllaDriveRestAPI) CreateTeamDrive(teamID int) (int, error) {
	teamIDInString := strconv.Itoa(teamID)
	client := resty.New()
	resp, err := client.R().
		SetHeader("Request-Token", r.Validator.GenerateValidateToken(teamIDInString)).
		Post(r.Config.IllaDriveInternalRestAPI + fmt.Sprintf(CREATE_TEAM_DRIVE, teamID))
	fmt.Printf("[IllaDriveRestAPI.CreateTeamDrive()]  response: %+v, err: %+v", resp, err)
	if resp.StatusCode() != http.StatusOK {
		if err != nil {
			return 0, errors.New("request illa drive failed")
		}
		return 0, errors.New("create team drive failed")
	}
	// decode target team drive record id
	var driveID map[string]int
	errInUnmarshal := json.Unmarshal([]byte(resp.String()), &driveID)
	if errInUnmarshal != nil {
		return 0, errInUnmarshal
	}
	driveIDInt, errInAssert := driveID["id"]
	if !errInAssert {
		return 0, errors.New("can not assert drive id in int")
	}
	return driveIDInt, nil
}

func (r *IllaDriveRestAPI) DeleteTeamDrive(teamID int) error {
	teamIDInString := strconv.Itoa(teamID)
	client := resty.New()
	resp, err := client.R().
		SetHeader("Request-Token", r.Validator.GenerateValidateToken(teamIDInString)).
		Delete(r.Config.IllaDriveInternalRestAPI + fmt.Sprintf(DELETE_TEAM_DRIVE, teamID))
	fmt.Printf("[IllaDriveRestAPI.DeleteTeamDrive()]  response: %+v, err: %+v", resp, err)
	if resp.StatusCode() != http.StatusOK {
		if err != nil {
			return errors.New("request illa drive failed")
		}
		return errors.New("delete team drive failed")
	}
	return nil
}
