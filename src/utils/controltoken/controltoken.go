package controltoken

import (
	"github.com/illacloud/builder-backend/src/utils/config"
)

func IsControlTokenAvaliable(token string) bool {
	conf := config.GetInstance()
	controlToken := conf.GetControlToken()
	if token == controlToken {
		return true
	}
	return false
}
