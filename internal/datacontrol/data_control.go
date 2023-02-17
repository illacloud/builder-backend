package datacontrol

import (
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
