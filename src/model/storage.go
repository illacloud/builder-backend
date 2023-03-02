package model

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Storage struct {
	AppStorage *AppStorage
}

func NewStorage(postgresDriver *gorm.DB, logger *zap.SugaredLogger) *Storage {
	appStorage := NewAppStorageImpl(postgresDriver, logger)
	return &Storage{
		AppStorage: appStorage,
	}
}
