package controltoken

import (
	"github.com/illacloud/illa-resource-manager-backend/src/utils/config"
)

func IsControlTokenAvaliable(token string) bool {
	conf := config.GetInstance()
	controlToken := conf.GetControlToken()
	if token == controlToken {
		return true
	}
	return false
}
