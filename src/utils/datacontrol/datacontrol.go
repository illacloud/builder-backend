package datacontrol

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/illacloud/builder-backend/src/model"
	"github.com/illacloud/builder-backend/src/utils/supervisor"
)

func GetUserInfo(targetUserID int) (*model.User, error) {
	// init sdk
	instance := supervisor.GetInstance()

	// fetch raw data
	userRaw, errInGetTargetUser := instance.GetUser(targetUserID)
	if errInGetTargetUser != nil {
		return nil, errInGetTargetUser
	}

	// construct
	user, errInNewUser := model.NewUserByDataControlRawData(userRaw)
	if errInNewUser != nil {
		return nil, errInNewUser
	}
	return user, nil
}

func GetMultiUserInfo(targetUserIDs []int) (map[int]*model.User, error) {
	// empty input
	if len(targetUserIDs) == 0 {
		ret := make(map[int]*model.User, 0)
		return ret, nil
	}
	// init sdk
	instance := supervisor.GetInstance()

	// convert to query param
	targetUserIDsInString := make([]string, 0)
	for _, userIDInt := range targetUserIDs {
		userIDString := strconv.Itoa(userIDInt)
		if len(userIDString) != 0 {
			targetUserIDsInString = append(targetUserIDsInString, strconv.Itoa(userIDInt))
		}
	}
	requestParams := strings.Join(targetUserIDsInString, ",")
	fmt.Printf("[DUMP] datacontrol.GetMultiUserInfo.requestParams: %+v\n", requestParams)

	// fetch raw data
	usersRaw, errInGetTargetUser := instance.GetMultiUser(requestParams)
	fmt.Printf("[DUMP] datacontrol.GetMultiUserInfo.usersRaw: %+v\n", usersRaw)

	if errInGetTargetUser != nil {
		return nil, errInGetTargetUser
	}

	// construct
	users, errInNewUsers := model.NewUsersByDataControlRawData(usersRaw)
	if errInNewUsers != nil {
		return nil, errInNewUsers
	}
	return users, nil
}

func GetTeamInfoByIdentifier(targetTeamIdentifier string) (*model.Team, error) {
	// init sdk
	instance := supervisor.GetInstance()

	// fetch raw data
	teamRaw, errInGetTargetTeam := instance.GetTeamByIdentifier(targetTeamIdentifier)
	if errInGetTargetTeam != nil {
		return nil, errInGetTargetTeam
	}

	// construct
	team, errInNewTeam := model.NewTeamByDataControlRawData(teamRaw)
	if errInNewTeam != nil {
		return nil, errInNewTeam
	}
	return team, nil
}

func GetTeamInfoByID(targetTeamID int) (*model.Team, error) {
	// init sdk
	instance := supervisor.GetInstance()

	// fetch raw data
	teamRaw, errInGetTargetTeam := instance.GetTeamByID(targetTeamID)
	if errInGetTargetTeam != nil {
		return nil, errInGetTargetTeam
	}

	// construct
	team, errInNewTeam := model.NewTeamByDataControlRawData(teamRaw)
	if errInNewTeam != nil {
		return nil, errInNewTeam
	}
	return team, nil
}
