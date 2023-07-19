package datacontrol

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/illacloud/builder-backend/internal/repository"

	supervisior "github.com/illacloud/builder-backend/internal/util/supervisior"
)

func GetUserInfo(targetUserID int) (*repository.User, error) {
	// init sdk
	instance, err := supervisior.NewSupervisior()
	if err != nil {
		return nil, err
	}

	// fetch raw data
	userRaw, errInGetTargetUser := instance.GetUser(targetUserID)
	if errInGetTargetUser != nil {
		return nil, errInGetTargetUser
	}

	// construct
	user, errInNewUser := repository.NewUserByDataControlRawData(userRaw)
	if errInNewUser != nil {
		return nil, errInNewUser
	}
	return user, nil
}

func GetMultiUserInfo(targetUserIDs []int) (map[int]*repository.User, error) {
	// empty input
	if len(targetUserIDs) == 0 {
		ret := make(map[int]*repository.User, 0)
		return ret, nil
	}
	// init sdk
	instance, err := supervisior.NewSupervisior()
	if err != nil {
		return nil, err
	}

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
	users, errInNewUsers := repository.NewUsersByDataControlRawData(usersRaw)
	if errInNewUsers != nil {
		return nil, errInNewUsers
	}
	return users, nil
}

func GetTeamInfoByIdentifier(targetTeamIdentifier string) (*repository.Team, error) {
	// init sdk
	instance, err := supervisior.NewSupervisior()
	if err != nil {
		return nil, err
	}

	// fetch raw data
	teamRaw, errInGetTargetTeam := instance.GetTeamByIdentifier(targetTeamIdentifier)
	if errInGetTargetTeam != nil {
		return nil, errInGetTargetTeam
	}

	// construct
	team, errInNewTeam := repository.NewTeamByDataControlRawData(teamRaw)
	if errInNewTeam != nil {
		return nil, errInNewTeam
	}
	return team, nil
}
