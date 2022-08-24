// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/illa-family/builder-backend/api/resthandler"
	"github.com/illa-family/builder-backend/api/router"
	"github.com/illa-family/builder-backend/internal/repository"
	"github.com/illa-family/builder-backend/internal/util"
	"github.com/illa-family/builder-backend/pkg/action"
	"github.com/illa-family/builder-backend/pkg/app"
	"github.com/illa-family/builder-backend/pkg/db"
	"github.com/illa-family/builder-backend/pkg/resource"
	"github.com/illa-family/builder-backend/pkg/room"
	"github.com/illa-family/builder-backend/pkg/smtp"
	"github.com/illa-family/builder-backend/pkg/user"
)

// Injectors from wire.go:

func Initialize() (*Server, error) {
	config, err := GetAppConfig()
	if err != nil {
		return nil, err
	}
	engine := gin.New()
	sugaredLogger := util.NewSugardLogger()
	dbConfig, err := db.GetConfig()
	if err != nil {
		return nil, err
	}
	gormDB, err := db.NewDbConnection(dbConfig, sugaredLogger)
	if err != nil {
		return nil, err
	}
	userRepositoryImpl := repository.NewUserRepositoryImpl(gormDB, sugaredLogger)
	smtpConfig, err := smtp.GetConfig()
	if err != nil {
		return nil, err
	}
	smtpServer := smtp.NewSMTPServer(smtpConfig)
	userServiceImpl := user.NewUserServiceImpl(userRepositoryImpl, sugaredLogger, smtpServer)
	userRestHandlerImpl := resthandler.NewUserRestHandlerImpl(sugaredLogger, userServiceImpl)
	userRouterImpl := router.NewUserRouterImpl(userRestHandlerImpl)
	appRepositoryImpl := repository.NewAppRepositoryImpl(sugaredLogger, gormDB)
	kvStateRepositoryImpl := repository.NewKVStateRepositoryImpl(sugaredLogger, gormDB)
	treeStateRepositoryImpl := repository.NewTreeStateRepositoryImpl(sugaredLogger, gormDB)
	setStateRepositoryImpl := repository.NewSetStateRepositoryImpl(sugaredLogger, gormDB)
	actionRepositoryImpl := repository.NewActionRepositoryImpl(sugaredLogger, gormDB)
	appServiceImpl := app.NewAppServiceImpl(sugaredLogger, appRepositoryImpl, userRepositoryImpl, kvStateRepositoryImpl, treeStateRepositoryImpl, setStateRepositoryImpl, actionRepositoryImpl)
	appRestHandlerImpl := resthandler.NewAppRestHandlerImpl(sugaredLogger, appServiceImpl)
	appRouterImpl := router.NewAppRouterImpl(appRestHandlerImpl)
	roomServiceImpl := room.NewRoomServiceImpl(sugaredLogger)
	roomRestHandlerImpl := resthandler.NewRoomRestHandlerImpl(sugaredLogger, roomServiceImpl)
	roomRouterImpl := router.NewRoomRouterImpl(roomRestHandlerImpl)
	resourceRepositoryImpl := repository.NewResourceRepositoryImpl(sugaredLogger, gormDB)
	actionServiceImpl := action.NewActionServiceImpl(sugaredLogger, appRepositoryImpl, actionRepositoryImpl, resourceRepositoryImpl)
	actionRestHandlerImpl := resthandler.NewActionRestHandlerImpl(sugaredLogger, actionServiceImpl)
	actionRouterImpl := router.NewActionRouterImpl(actionRestHandlerImpl)
	resourceServiceImpl := resource.NewResourceServiceImpl(sugaredLogger, resourceRepositoryImpl)
	resourceRestHandlerImpl := resthandler.NewResourceRestHandlerImpl(sugaredLogger, resourceServiceImpl)
	resourceRouterImpl := router.NewResourceRouterImpl(resourceRestHandlerImpl)
	restRouter := router.NewRESTRouter(sugaredLogger, userRouterImpl, appRouterImpl, roomRouterImpl, actionRouterImpl, resourceRouterImpl)
	server := NewServer(config, engine, restRouter, sugaredLogger)
	return server, nil
}
